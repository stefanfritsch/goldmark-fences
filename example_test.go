package fences_test

import (
	"os"

	fences "github.com/stefanfritsch/goldmark-fences"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/text"
)

func Example() {
	src := []byte(`
## Hello

The following contains an id and a class

:::{#big-div .add-border}

And the next fence contains two classes.

:::{.background-green .font-big}
## This is nested within nested fences

here we close the inner fence:
:::

and finally the outer one:
:::`)

	markdown := goldmark.New(
		goldmark.WithExtensions(
			&fences.Extender{},
		),
	)

	doc := markdown.Parser().Parse(text.NewReader(src))
	markdown.Renderer().Render(os.Stdout, src, doc)

	// Output:
	// <h2>Hello</h2>
	// <p>The following contains an id and a class</p>
	// <div id="big-div" class="add-border">
	// <p>And the next fence contains two classes.</p>
	// <div class="background-green font-big">
	// <h2>This is nested within nested fences</h2>
	// <p>here we close the inner fence:</p>
	// </div>
	// <p>and finally the outer one:</p>
	// </div>
}
