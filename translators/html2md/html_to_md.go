package html2md

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"webbot/browser/virtualid"
	"webbot/translators"
	"webbot/utils/stringsx"

	"golang.org/x/net/html"
)

const DefaultMaxListDisplaySize = 5

type HTML2MDTranslator struct {
	virtualIDGenerator virtualid.VirtualIDGenerator
	maxListDisplaySize int
}

type Options struct {
	maxListDisplaySize int
	virtualIDGenerator virtualid.VirtualIDGenerator
}

func NewHTML2MDTranslator(options *Options) translators.Translator {
	maxListDisplaySize := DefaultMaxListDisplaySize
	virtualIDGenerator := virtualid.NewIncrIntVirtualIDGenerator()
	if options != nil {
		if options.maxListDisplaySize > 0 {
			maxListDisplaySize = options.maxListDisplaySize
		}
		if options.virtualIDGenerator != nil {
			virtualIDGenerator = options.virtualIDGenerator
		}
	}
	return &HTML2MDTranslator{
		virtualIDGenerator: virtualIDGenerator,
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

func (t *HTML2MDTranslator) parseInnerText(childTexts []string) string {
	s := strings.Join(childTexts, "")
	re := regexp.MustCompile(`[^a-zA-Z0-9\\s]+`)
	return re.ReplaceAllString(s, "")
}

func (t *HTML2MDTranslator) Visit(n *html.Node) string {
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
			innerText := t.parseInnerText(content)
			var id string
			for _, attr := range n.Attr {
				if attr.Key == "data-vid" {
					id = attr.Val
				}
			}
			if innerText == "" || id == "" {
				return ""
			}
			return fmt.Sprintf("[%s](%s)", innerText, id)
		case "input":
			var typ, name, ariaLabel, id string
			for _, attr := range n.Attr {
				switch attr.Key {
				case "type":
					typ = attr.Val
				case "name":
					name = attr.Val
				case "aria-label":
					ariaLabel = attr.Val
				case "data-vid":
					id = attr.Val
				default:
				}
			}
			if typ == "hidden" || id == "" {
				return ""
			} else if typ == "submit" {
				return fmt.Sprintf("[%s](%s)", name, id)
			} else {
				text := name
				if text == "" {
					text = ariaLabel
				}
				if strings.TrimSpace(text) == "" {
					return fmt.Sprintf("[type=%s](%s)", typ, id)
				}
				return fmt.Sprintf("[%s, type=%s](%s)", text, typ, id)
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
			var href, id string
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					href = attr.Val
				} else if attr.Key == "data-vid" {
					id = attr.Val
				}
			}
			innerText := t.parseInnerText(content)
			if strings.TrimSpace(href) == "" || innerText == "" || id == "" {
				return ""
			}
			return fmt.Sprintf("[%s, href=%s](%s)", strings.TrimSpace(strings.Join(content, "")), href, id)
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
		case "div", "section", "body", "header", "form", "dialog", "ul", "ol":
			return strings.Join(content, "\n")
		case "p", "span", "g", "figure", "desc", "footer", "html", "main", "legend", "fieldset", "center":
			return strings.Join(content, "")
		case "head", "script", "style", "iframe", "svg", "meso-native", "meso-display-ad", "grammarly-desktop-integration", "path", "noscript", "link", "meta", "label", "circle", "rect", "image":
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

func (t *HTML2MDTranslator) visitChildren(n *html.Node) []string {
	content := []string{}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		content = append(content, t.Visit(c))
	}
	return content
}

func shouldVisit(n *html.Node) bool {
	if n == nil {
		return false
	}
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
	s := stringsx.ReduceNewlines(mdText, 2)
	s = strings.ReplaceAll(s, "  ", " ")
	return strings.TrimSpace(s)
}
