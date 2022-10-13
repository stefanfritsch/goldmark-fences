package fences

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// Extender allows you to use fenced divs / fenced containers / fences in markdown
//
// Fences are a way to wrap other elements in divs and giving those divs
// attributes such as ids or css classes using the same syntax as with headings.
//
// :::{#big-div .add-border}
// this is some text
//
// ## with a header
//
// :::{.background-green .font-big}
// ```R
// X <- as.data.table(iris)
// X[Species != "virginica", mean(Sepal.Length), Species]
// ```
// :::
// :::
type Extender struct {
	priority int // optional int != 0. the priority value for parser and renderer. Defaults to 100.
}

// This implements the Extend method for goldmark-fences.Extender
func (e *Extender) Extend(md goldmark.Markdown) {
	priority := 100

	if e.priority != 0 {
		priority = e.priority
	}
	md.Parser().AddOptions(
		parser.WithBlockParsers(
			util.Prioritized(&fencedContainerParser{}, priority),
		),
	)
	md.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(&Renderer{}, priority),
		),
	)
}
