package fences_test

import (
	"os"

	fences "github.com/stefanfritsch/goldmark-fences"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/text"
)

func Example_noAttribute() {
	src := []byte(`
## Hello

The following contains no classes

:::{}
This is fenced
:::

But you need something to denote opening ":"

:::
This is not fenced because we can't differentiate between nested and closing ":"
:::

`)

	markdown := goldmark.New(
		goldmark.WithExtensions(
			&fences.Extender{},
		),
	)

	doc := markdown.Parser().Parse(text.NewReader(src))
	markdown.Renderer().Render(os.Stdout, src, doc)

	// Output:
	// <h2>Hello</h2>
	// <p>The following contains no classes</p>
	// <div data-fence="0">
	// <p>This is fenced</p>
	// </div>
	// <p>But you need something to denote opening &quot;:&quot;</p>
	// <p>:::
	// This is not fenced because we can't differentiate between nested and closing &quot;:&quot;
	// :::</p>
}
