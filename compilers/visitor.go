package compilers

import (
	"golang.org/x/net/html"
)

type NodeVisitor interface {
	Visit(*html.Node) string
}

type Translater interface {
	NodeVisitor

	Translate(string) (string, error)
}
