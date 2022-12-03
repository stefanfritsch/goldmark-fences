package fences

import (
	"regexp"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

// A Config struct has configurations for the HTML based renderers.
type Config struct {
	Writer    html.Writer
	HardWraps bool
	XHTML     bool
	Unsafe    bool
}

// HeadingAttributeFilter defines attribute names which heading elements can have
var FencedContainerAttributeFilter = html.GlobalAttributeFilter

// A Renderer struct is an implementation of renderer.NodeRenderer that renders
// nodes as (X)HTML.
type Renderer struct {
	Config
}

// RegisterFuncs implements NodeRenderer.RegisterFuncs .
func (r *Renderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindFencedContainer, r.renderFencedContainer)
}

func (r *Renderer) renderFencedContainer(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*FencedContainer)
	if entering {
		n.element = "div"
		if isNav(node) {
			n.element = "nav"
		}

		if n.Attributes() != nil {
			_, _ = w.WriteString("<" + n.element)
			html.RenderAttributes(w, n, FencedContainerAttributeFilter)
			_, _ = w.WriteString(">\n")
		} else {
			_, _ = w.WriteString("<" + n.element + ">\n")
		}
	} else {
		_, _ = w.WriteString("</" + n.element + ">\n")
	}
	return ast.WalkContinue, nil
}

func isNav(node ast.Node) bool {
	class, ok := node.AttributeString("class")
	if !ok {
		return false
	}
	if navChk.Match(class.([]byte)) {
		return true
	}

	return false
}

// check for the .nav class
var navChk = regexp.MustCompile(`(^| |\.)elem-nav($| )`)
