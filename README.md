# Goldmark-Fences

## Overview

[Goldmark](https://github.com/yuin/goldmark) is a fast markdown renderer for Go. Fences are a markdown extension that allows you to wrap parts of a markdown in a `<div>` or other html-tag with custom classes, ids or other html attributes.

Fences are also called "fenced divs" or "fenced containers." I will use fences because it's the shortest of the available options.

### An Example

```markdown
:::{.blue}
## Life Inside Fences

We are now inside a div with the css-class "blue". This can be used to style this block

:::{#insideme .red data="important"}
fences can be nested and given ids as well as classes
:::
:::
```

Now add the following css to your stylesheet:

```css
.blue { background-color: steelblue; padding: 5px; }
#insideme { color: yellow; }
.red { background-color: red; }
```

And the fenced part will look like this (this is an image as github doesn't allow custom css in READMEs):

![](assets/Screenshot%202022-10-14%20001453.png)

## Full Code Example

A full code example to use the extension with goldmark could look like this:

```go
func main() {
	src := []byte(`
## Hello
We now try out fences:

:::{#big-div .add-border}
This paragraph is inside the fenced block.

This as well.
:::
        `)

	markdown := goldmark.New(
		goldmark.WithExtensions(
			&fences.Extender{},
		),
	)

	doc := markdown.Parser().Parse(text.NewReader(src))
	markdown.Renderer().Render(os.Stdout, src, doc)
}
```

## Possible Use Cases

You can use fences to e.g.

* style your output
* wrap semantic units into blocks - e.g. put your toc into one div - which allows you to...
* move parts outside the normal layout - e.g. create a floating toc on the left
* fulfill the requirements of third party libs regarding attributes like `data=` etc.
* use it for other extensions -- e.g. to wrap all sections in divs for a semantic output structure
* create random html creations
