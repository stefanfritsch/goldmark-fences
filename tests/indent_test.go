package fences_test

import (
	"os"
	"strings"

	fences "github.com/stefanfritsch/goldmark-fences"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/text"
)

func Example_indent_2() {
	srcString := `
  ## Hello

  This is outside

  :::{}
      This is inside
      
      still inside
  :::
  Outside again.`

	src := []byte(srcString)

	markdown := goldmark.New(
		goldmark.WithExtensions(
			&fences.Extender{},
		),
	)

	doc := markdown.Parser().Parse(text.NewReader(src))
	markdown.Renderer().Render(os.Stdout, src, doc)

	// Output:
	// <h2>Hello</h2>
	// <p>This is outside</p>
	// <div data-fence="0">
	// <p>This is inside</p>
	// <p>still inside</p>
	// </div>
	// <p>Outside again.</p>
}

func Example_indent() {
	srcString := strings.ReplaceAll(`
## Hello

This is outside

§§§
this is unindented code
§§§

    this is indented code

:::{#big-div .add-border}

    This is indented
    
    :::{.background-green .font-big}
        ## This is indented within indented
        
        §§§
        this is unindented code in an indented block
        §§§
        
            This is indented code in an indented block
    :::
    
    ::: {.background-yellow}
    This is not indented
    :::
:::`, "§", "`")
	src := []byte(srcString)

	markdown := goldmark.New(
		goldmark.WithExtensions(
			&fences.Extender{},
		),
	)

	doc := markdown.Parser().Parse(text.NewReader(src))
	markdown.Renderer().Render(os.Stdout, src, doc)

	// Output:
	// <h2>Hello</h2>
	// <p>This is outside</p>
	// <pre><code>this is unindented code
	// </code></pre>
	// <pre><code>this is indented code
	// </code></pre>
	// <div data-fence="0" id="big-div" class="add-border">
	// <p>This is indented</p>
	// <div data-fence="1" class="background-green font-big">
	// <h2>This is indented within indented</h2>
	// <pre><code>this is unindented code in an indented block
	// </code></pre>
	// <pre><code>This is indented code in an indented block
	// </code></pre>
	// </div>
	// <div data-fence="1" class="background-yellow">
	// <p>This is not indented</p>
	// </div>
	// </div>
}
