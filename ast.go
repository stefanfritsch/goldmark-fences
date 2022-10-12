package fences

import (
	"github.com/yuin/goldmark/ast"
)

// A FencedContainer struct represents a fenced code block of Markdown text.
type FencedContainer struct {
	ast.BaseBlock
	element string
}

// Dump implements Node.Dump .
func (n *FencedContainer) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

// KindFencedContainer is a NodeKind of the FencedContainer node.
var KindFencedContainer = ast.NewNodeKind("FencedContainer")

// Kind implements Node.Kind.
func (n *FencedContainer) Kind() ast.NodeKind {
	return KindFencedContainer
}

// NewFencedContainer return a new FencedContainer node.
func NewFencedContainer() *FencedContainer {
	return &FencedContainer{
		BaseBlock: ast.BaseBlock{},
	}
}
