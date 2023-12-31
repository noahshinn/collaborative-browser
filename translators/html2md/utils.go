package html2md

import (
	"collaborativebrowser/utils/slicesx"
	"collaborativebrowser/utils/stringsx"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

func parseInnerText(childTexts []string) string {
	s := strings.Join(childTexts, "")
	re := regexp.MustCompile(`[^a-zA-Z0-9\\s]+`)
	return re.ReplaceAllString(s, "")
}

func buildAttrMapFromNode(n *html.Node) map[string]string {
	attrMap := make(map[string]string)
	for _, attr := range n.Attr {
		attrMap[attr.Key] = attr.Val
	}
	return attrMap
}

func shouldVisitNode(n *html.Node) bool {
	if n == nil {
		return false
	}
	if n.Type != html.ElementNode {
		return true
	}
	if n.Data == "input" || n.Data == "textarea" {
		for _, attr := range n.Attr {
			if attr.Key == "type" && attr.Val == "hidden" {
				return false
			}
		}
	}
	for _, attr := range n.Attr {
		if attr.Key == "aria-hidden" && attr.Val == "true" {
			return false
		}
		if attr.Key == "style" {
			hiddenStyles := []string{"opacity: 0", "font-size: 0", "width: 0", "height: 0", "display: none", "visibility: hidden"}
			for _, style := range hiddenStyles {
				if strings.Contains(attr.Val, style) {
					return false
				}
			}
		}
	}
	return true
}

func renderSelectable(typ SelectableType, virtualID string, primaryContent string, secondaryContent string) string {
	var suffix string
	if secondaryContent != "" {
		suffix = ", " + secondaryContent
	}
	return fmt.Sprintf("[%s%s, type=%s](%s)", primaryContent, suffix, typ, virtualID)
}

var whitelistedImportantAttributeValues = map[string][]string{
	"name": {"login", "search"},
}

func isClickable(n *html.Node, attrMap map[string]string) bool {
	if _, ok := attrMap["data-vid"]; !ok {
		return false
	} else if n.Data != "a" && n.Data != "button" {
		return false
	} else if ariaHidden, ok := attrMap["aria-hidden"]; ok && ariaHidden == "true" {
		return false
	}
	if ariaLabel, ok := attrMap["aria-label"]; ok && ariaLabel != "" {
		return true
	} else if name, ok := attrMap["name"]; ok && name != "" {
		if importantNameValues, ok := whitelistedImportantAttributeValues["name"]; ok {
			if slicesx.Contains(importantNameValues, name) {
				return true
			}
		}
	}

	if typ, ok := attrMap["type"]; ok && typ == "submit" {
		return true
	}
	if n.Data == "a" {
		if href, ok := attrMap["href"]; ok && href != "" {
			return true
		} else if role, ok := attrMap["role"]; ok && role == "button" {
			return true
		} else if name, ok := attrMap["name"]; ok && name == "login" {
			return true
		}
	} else if n.Data == "button" {
		if ariaExpanded, ok := attrMap["aria-expanded"]; ok {
			return ariaExpanded == "true" || ariaExpanded == "false"
		} else if inForm := isInForm(n); inForm {
			return true
		}
	}
	return false
}

func isInputable(n *html.Node, attrMap map[string]string) bool {
	if _, ok := attrMap["data-vid"]; !ok {
		return false
	} else if n.Data != "input" && n.Data != "textarea" {
		return false
	} else if typ, ok := attrMap["type"]; !ok || typ == "hidden" {
		return false
	} else if ariaHidden, ok := attrMap["aria-hidden"]; ok && ariaHidden == "true" {
		return false
	} else if placeholder, ok := attrMap["placeholder"]; ok && placeholder != "" {
		return true
	} else if ariaLabel, ok := attrMap["aria-label"]; ok && ariaLabel != "" {
		return true
	} else if value, ok := attrMap["value"]; ok && value != "" {
		return true
	} else if autocapitalize, ok := attrMap["autocapitalize"]; ok && autocapitalize == "on" || autocapitalize == "sentences" || autocapitalize == "words" || autocapitalize == "characters" {
		return true
	} else if autocomplete, ok := attrMap["autocomplete"]; ok && autocomplete != "off" {
		return true
	} else if spellcheck, ok := attrMap["spellcheck"]; ok && spellcheck == "true" {
		return true
	}

	if n.Data == "input" {
		if role, ok := attrMap["role"]; ok && role == "combobox" {
			return true
		}
	} else if n.Data == "textarea" {
		if rows, ok := attrMap["rows"]; ok && rows != "" {
			if rowsInt, err := strconv.Atoi(rows); err == nil && rowsInt > 0 {
				return true
			}
		}
	}
	return false
}

// TODO: this is currently more conservative than isClickable
func getLabelForClickable(n *html.Node, attrMap map[string]string, childContent []string) (label string, isClickable bool) {
	if n.Data != "a" && n.Data != "button" {
		return "", false
	} else if ariaLabel, ok := attrMap["aria-label"]; ok && ariaLabel != "" {
		return ariaLabel, true
	}
	innerText := parseInnerText(childContent)
	var importantAttributePairs []string
	var prefix string
	for attr, values := range whitelistedImportantAttributeValues {
		if value, ok := attrMap[attr]; ok && slicesx.Contains(values, value) {
			importantAttributePairs = append(importantAttributePairs, fmt.Sprintf("%s=%s", attr, value))
		}
	}
	if len(importantAttributePairs) > 0 {
		prefix = fmt.Sprintf("%s, ", strings.Join(importantAttributePairs, ", "))
	}
	if n.Data == "a" {
		href, ok := attrMap["href"]
		if !ok {
			return "", false
		}
		strippedQueryParams := stripQueryParamsFromPossibleFullURL(href)
		if innerText == "" {
			return fmt.Sprintf("%shref=%s", prefix, strippedQueryParams), true
		} else {
			return fmt.Sprintf("%sinner-text=%s, href=%s", prefix, innerText, strippedQueryParams), true
		}
	}
	if typ, ok := attrMap["type"]; (!ok || typ != "submit") && innerText == "" {
		return "", false
	} else {
		if typ == "submit" {
			prefix = prefix + "type=submit"
		}
		if innerText != "" {
			prefix = prefix + "inner-text=" + innerText
		}
		return strings.TrimRight(prefix, ", "), true
	}
}

// TODO: this is currently more conservative than isInputable
func getLabelForInputable(n *html.Node, attrMap map[string]string) (label string, isInputable bool) {
	if n.Data != "input" && n.Data != "textarea" {
		return "", false
	} else if placeholder, ok := attrMap["placeholder"]; ok && placeholder != "" {
		return placeholder, true
	} else if ariaLabel, ok := attrMap["aria-label"]; ok && ariaLabel != "" {
		return ariaLabel, true
	} else if autocompleteType, ok := attrMap["autocomplete"]; ok && autocompleteType != "" && autocompleteType != "off" {
		return autocompleteType, true
	}
	return "", false
}

func isInForm(n *html.Node) bool {
	if n.Data == "form" {
		return true
	}
	if n.Parent == nil {
		return false
	}
	return isInForm(n.Parent)
}

func cleanup(mdText string) string {
	// remove extra newlines
	s := stringsx.ReduceNewlines(mdText, 2)

	// remove extra spaces
	s = strings.ReplaceAll(s, "  ", " ")

	// remove trailing spaces for each line
	lines := strings.Split(s, "\n")
	s = strings.Join(slicesx.Map(lines, func(str string, _ int) string {
		return strings.TrimSpace(str)
	}), "\n")

	// remove leading or trailing spaces
	return strings.TrimSpace(s)
}

func stripQueryParamsFromPossibleFullURL(link string) string {
	parsedURL, err := url.Parse(link)
	if err != nil {
		return link
	}
	parsedURL.RawQuery = ""
	return parsedURL.String()
}
