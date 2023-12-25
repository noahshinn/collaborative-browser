package html2md

import (
	"collaborativebrowser/translators"
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/net/html"
)

const DefaultMaxListDisplaySize = 5

type HTML2MDTranslator struct {
	maxListDisplaySize int
}

type Options struct {
	maxListDisplaySize int
}

type SelectableType string

const (
	SelectableTypeButton SelectableType = "button"
	SelectableTypeLink   SelectableType = "link"
	SelectableTypeInput  SelectableType = "input"
	SelectableTypeTextA  SelectableType = "textarea"
)

func NewHTML2MDTranslator(options *Options) translators.Translator {
	maxListDisplaySize := DefaultMaxListDisplaySize
	if options != nil {
		if options.maxListDisplaySize > 0 {
			maxListDisplaySize = options.maxListDisplaySize
		}
	}
	return &HTML2MDTranslator{
		maxListDisplaySize: maxListDisplaySize,
	}
}

func (t *HTML2MDTranslator) Translate(text string) (string, error) {
	doc, err := html.Parse(strings.NewReader(text))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing HTML: %v\n", err)
		return "", err
	}
	mdText := t.Visit(doc)
	return cleanup(mdText), nil
}

func (t *HTML2MDTranslator) Visit(n *html.Node) string {
	if !shouldVisitNode(n) {
		return ""
	}
	switch n.Type {
	case html.TextNode:
		return n.Data
	case html.ElementNode:
		content := t.visitChildren(n)
		attrMap := buildAttrMapFromNode(n)
		virtualID := attrMap["data-vid"]
		switch n.Data {
		case "button":
			if !isClickable(n, attrMap) {
				return strings.Join(content, "\n")
			} else if label, isClickable := getLabelForClickable(n, attrMap, content); !isClickable {
				return strings.Join(content, "\n")
			} else {
				return renderSelectable(SelectableTypeButton, virtualID, label, "")
			}
		case "input", "textarea":
			if !isInputable(n, attrMap) {
				return strings.Join(content, "\n")
			} else if label, isInputable := getLabelForInputable(n, attrMap); !isInputable {
				return strings.Join(content, "\n")
			} else {
				return renderSelectable(SelectableType(n.Data), virtualID, label, "")
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
			return "\n\n###### " + strings.Join(content, "")
		case "title":
			return "# " + strings.Join(content, "")
		case "img":
			if alt, ok := attrMap["alt"]; !ok || strings.TrimSpace(alt) == "" {
				return ""
			} else {
				return fmt.Sprintf("![%s](<img>)", alt)
			}
		case "video":
			return "<video>"
		case "a":
			if !isClickable(n, attrMap) {
				return strings.Join(content, "\n")
			} else if label, isClickable := getLabelForClickable(n, attrMap, content); !isClickable {
				return strings.Join(content, "\n")
			} else {
				return renderSelectable(SelectableTypeLink, virtualID, label, "")
			}
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
		case "nav":
			items := []string{}
			for _, c := range content {
				s := strings.TrimSpace(c)
				if s != "" {
					items = append(items, "- "+s)
				}
			}
			if len(items) == 0 {
				return ""
			}
			return fmt.Sprintf("## Nav Bar\n\n%s", strings.Join(items, "\n"))
		case "sup":
			return "^{" + strings.Join(content, "") + "}"
		case "div", "section", "body", "header", "form", "dialog", "ul", "ol", "small", "bdi", "template", "summary", "details", "dl", "dt", "dd", "main", "tbody", "table", "tr", "td":
			return strings.Join(content, "\n")
		case "p", "span", "g", "figure", "desc", "footer", "html", "legend", "fieldset", "center", "picture":
			return strings.Join(content, " ")
		case "head", "script", "style", "iframe", "svg", "meso-native", "meso-display-ad", "grammarly-desktop-integration", "path", "noscript", "link", "meta", "label", "circle", "rect", "image", "polygon", "source", "use", "canvas":
			return ""
		default:
			log.Printf("Found unknown element: %v\n", n.Data)
			return strings.Join(content, "\n")
		}
	case html.CommentNode, html.DoctypeNode:
		return ""
	case html.DocumentNode:
		content := t.visitChildren(n)
		return strings.Join(content, "\n")
	default:
		log.Printf("Unknown html node type: %v\n", n.Data)
		return ""
	}
}

func (t *HTML2MDTranslator) visitChildren(n *html.Node) []string {
	content := []string{}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		content = append(content, t.Visit(c))
	}
	return content
}
