package fences

import (
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type fencedContainerParser struct {
}

var defaultFencedContainerParser = &fencedContainerParser{}

// NewFencedContainerParser returns a new BlockParser that
// parses fenced code blocks.
func NewFencedContainerParser() parser.BlockParser {
	return defaultFencedContainerParser
}

type fenceData struct {
	char              byte     // Currently, this is always ":"
	indent            int      // The indentation of the opening (and closing) tags (:::{})
	length            int      // The length of the fence, e.g. is it ::: or ::::?
	node              ast.Node // The node of the fence
	contentIndent     int      // The indentation of the content relative to the previous fenced block. The first line of the content is taken as its indentation. If you want a fence with just a code block you need to use backticks
	contentHasStarted bool     // Only used as an indicator if contentIndent has been set already
}

var fencedContainerInfoKey = parser.NewContextKey()

func (b *fencedContainerParser) Trigger() []byte {
	return []byte{':'}
}

func (b *fencedContainerParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, segment := reader.PeekLine()
	pos := pc.BlockOffset()
	if pos < 0 || line[pos] != ':' {
		return nil, parser.NoChildren
	}
	findent := pos
	fenceChar := line[pos]
	i := pos
	for ; i < len(line) && line[i] == fenceChar; i++ {
	}
	oFenceLength := i - pos
	if oFenceLength < 3 {
		return nil, parser.NoChildren
	}

	// ========================================================================== //
	// 	Without attributes we return

	if i >= len(line)-1 {
		// If there are no attributes we can't create a div because we won't know
		// if a ":::" ends the last fenced container or opens a new one
		return nil, parser.NoChildren
	}

	rest := line[i:]
	left := i + util.TrimLeftSpaceLength(rest)
	right := len(line) - 1 - util.TrimRightSpaceLength(rest)

	if left >= right {
		// As above:
		// If there are no attributes we can't create a div because we won't know
		// if a ":::" ends the last fenced container or opens a new one
		return nil, parser.NoChildren
	}

	// ========================================================================== //
	// 	With attributes we construct the node

	reader.Advance(left)
	node := NewFencedContainer()

	fenceID := genRandomString(24)
	node.SetAttributeString("data-fenceid", []byte(fenceID))

	attrs, ok := parser.ParseAttributes(reader)
	if ok {
		for _, attr := range attrs {
			node.SetAttribute(attr.Name, attr.Value)
		}
	}

	fdata := &fenceData{
		char:              fenceChar,
		indent:            findent,
		length:            oFenceLength,
		node:              node,
		contentIndent:     0,
		contentHasStarted: false,
	}

	var fdataMap map[string]*fenceData

	if oldData := pc.Get(fencedContainerInfoKey); oldData != nil {
		fdataMap = oldData.(map[string]*fenceData)
		fdataMap[fenceID] = fdata
	} else {
		fdataMap = map[string]*fenceData{fenceID: fdata}
	}
	pc.Set(fencedContainerInfoKey, fdataMap)

	// check if it's an empty block
	line, _ = reader.PeekLine()
	w, pos := util.IndentWidth(line, reader.LineOffset())

	if close, _ := b.closes(line, segment, w, pos, node, fdata); close {
		return node, parser.NoChildren
	}

	return node, parser.HasChildren
}

func (b *fencedContainerParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	rawdata := pc.Get(fencedContainerInfoKey)
	fdataMap := rawdata.(map[string]*fenceData)

	rawFenceID, ok := node.AttributeString("data-fenceid")
	if !ok {
		// huhu: don't panic in production
		panic("fenceID is missing")
	}
	fenceID := string(rawFenceID.([]byte))
	fdata := fdataMap[fenceID]

	line, segment := reader.PeekLine()
	w, pos := util.IndentWidth(line, reader.LineOffset())

	if !fdata.contentHasStarted && !util.IsBlank(line[pos:]) {
		fdata.contentHasStarted = true
		fdata.contentIndent = w

		fdataMap[fenceID] = fdata
		pc.Set(fencedContainerInfoKey, fdataMap)
	}

	if close, newline := b.closes(line, segment, w, pos, node, fdata); close {
		reader.Advance(segment.Stop - segment.Start - newline + segment.Padding)
		delete(fdataMap, fenceID)

		if len(fdataMap) == 0 {
			return parser.Close
		} else {
			pc.Set(fencedContainerInfoKey, fdataMap)
			return parser.Close
		}
	}

	if fdata.contentIndent > 0 {
		reader.Advance(fdata.contentIndent)
	}

	return parser.Continue | parser.HasChildren
}

func (b *fencedContainerParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {
}

func (b *fencedContainerParser) CanInterruptParagraph() bool {
	return true
}

func (b *fencedContainerParser) CanAcceptIndentedLine() bool {
	return false
}

func (b *fencedContainerParser) closes(line []byte, segment text.Segment, w int, pos int, node ast.Node, fdata *fenceData) (bool, int) {

	// don't close anything but the last node
	if node != fdata.node {
		return false, 1
	}

	// If the indentation is lower, we assume the user forgot to close the block
	if w < fdata.indent {
		return true, 1
	}

	// else, check for the correct number of closing chars and provide the info
	// necessary to advance the reader
	if w == fdata.indent {
		i := pos
		for ; i < len(line) && line[i] == fdata.char; i++ {
		}
		length := i - pos

		if length >= fdata.length && util.IsBlank(line[i:]) {
			newline := 1
			if line[len(line)-1] != '\n' {
				newline = 0
			}

			return true, newline
		}
	}

	return false, 0
}
