package html2md

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"golang.org/x/net/html"
)

type HTML2MDTranslater struct{}

// TODO: implement ID generator
var idCounter int = 0

const maxListDisplaySize = 5

func (t *HTML2MDTranslater) Translate(text string) (string, error) {
	doc, err := html.Parse(strings.NewReader(text))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing HTML: %v\n", err)
		return "", err
	}
	mdText := t.Visit(doc)
	return cleanup(mdText), nil
}

func (t *HTML2MDTranslater) Visit(n *html.Node) string {
	if !shouldVisit(n) {
		return ""
	}
	switch n.Type {
	case html.TextNode:
		return n.Data
	case html.ElementNode:
		content := t.visitChildren(n)
		switch n.Data {
		case "button":
			text := strings.Join(content, "")
			id := idCounter
			idCounter++
			return fmt.Sprintf("[%s](ID_%d)", text, id)
		case "input":
			typ := ""
			name := ""
			ariaLabel := ""
			for _, attr := range n.Attr {
				switch attr.Key {
				case "type":
					typ = attr.Val
				case "name":
					name = attr.Val
				case "aria-label":
					ariaLabel = attr.Val
				default:
				}
			}
			id := idCounter
			idCounter++
			if typ == "hidden" {
				return ""
			} else if typ == "submit" {
				return fmt.Sprintf("[%s](ID_%d)", name, id)
			} else {
				text := name
				if text == "" {
					text = ariaLabel
				}
				return fmt.Sprintf("[%s, type=%s](ID_%d)", text, typ, id)
			}
		case "b", "strong":
			return "**" + strings.Join(content, "") + "**"
		case "i", "em":
			return "_" + strings.Join(content, "") + "_"
		case "h1":
			return "\n\n## " + strings.Join(content, "")
		case "h2":
			return "\n\n### " + strings.Join(content, "")
		case "h3":
			return "\n\n#### " + strings.Join(content, "")
		case "h4":
			return "\n\n##### " + strings.Join(content, "")
		case "h5":
			return "\n\n###### " + strings.Join(content, "")
		case "h6":
			return "\n\n####### " + strings.Join(content, "")
		case "title":
			return "# " + strings.Join(content, "")
		case "img":
			alt := ""
			for _, attr := range n.Attr {
				if attr.Key == "alt" {
					alt = attr.Val
				}
			}
			if strings.TrimSpace(alt) == "" {
				return ""
			}
			return fmt.Sprintf("![%s](<img>)", alt)
		case "a":
			return fmt.Sprintf("[%s](%s)", strings.Join(content, ""), trimURL(n.Attr[0].Val))
		case "li":
			text := strings.Join(content, "")
			if strings.TrimSpace(text) == "" {
				return ""
			}
			return "- " + text
		case "code":
			return "code"
		case "pre":
			return "```" + strings.Join(content, "") + "```"
		case "br":
			return "\n"
		case "hr":
			return "---"
		case "del":
			return "~~" + strings.Join(content, "") + "~~"
		case "ul", "ol":
			if len(content) > maxListDisplaySize {
				// TODO: improve the button display
				id := idCounter
				idCounter++
				return strings.Join(content[:maxListDisplaySize], "\n") + fmt.Sprintf("\n\n[See more](ID_%d)", id)
			}
			return strings.Join(content, "\n")
		case "div", "section", "body", "header", "form":
			return strings.Join(content, "\n")
		case "p", "span", "g", "figure", "desc", "footer", "html", "main", "legend", "fieldset":
			return strings.Join(content, "")
		case "head", "script", "style", "iframe", "svg", "meso-native", "meso-display-ad", "grammarly-desktop-integration", "path", "noscript", "link", "meta", "label":
			return ""
		default:
			fmt.Printf("Unknown element: %v\n", n.Data)
			return strings.Join(content, "\n")
		}
	case html.CommentNode, html.DoctypeNode:
		return ""
	case html.DocumentNode:
		content := t.visitChildren(n)
		return strings.Join(content, "\n")
	default:
		fmt.Printf("Unknown node type: %v\n", n.Data)
		return ""
	}
}

func (t *HTML2MDTranslater) visitChildren(n *html.Node) []string {
	content := []string{}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		content = append(content, t.Visit(c))
	}
	return content
}

func shouldVisit(n *html.Node) bool {
	if n.Type != html.ElementNode {
		return true
	}
	if n.Data == "input" {
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

func cleanup(mdText string) string {
	for strings.Contains(mdText, "\n\n\n") || strings.Contains(mdText, "  ") {
		mdText = strings.ReplaceAll(mdText, "\n\n\n", "\n\n")
		mdText = strings.ReplaceAll(mdText, "  ", " ")
	}
	return strings.TrimSpace(mdText)
}

func trimURL(inputURL string) string {
	u, err := url.Parse(inputURL)
	if err != nil {
		return inputURL
	}
	u.Scheme = ""
	u.RawQuery = ""
	u.User = nil
	if u.Opaque != "" {
		return u.Opaque
	}
	return u.Host + u.Path
}
