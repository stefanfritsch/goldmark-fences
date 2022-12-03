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
	char   byte
	indent int
	length int
	node   ast.Node
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

	node := NewFencedContainer()
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

	reader.Advance(left)
	attrs, ok := parser.ParseAttributes(reader)

	if ok {
		for _, attr := range attrs {
			node.SetAttribute(attr.Name, attr.Value)
		}
	}

	fdata := &fenceData{fenceChar, findent, oFenceLength, node}
	var fdataMap []*fenceData

	if oldData := pc.Get(fencedContainerInfoKey); oldData != nil {
		fdataMap = oldData.([]*fenceData)
		fdataMap = append(fdataMap, fdata)
	} else {
		fdataMap = []*fenceData{fdata}
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
	fdataMap := rawdata.([]*fenceData)
	fdata := fdataMap[len(fdataMap)-1]

	line, segment := reader.PeekLine()
	w, pos := util.IndentWidth(line, reader.LineOffset())

	if close, newline := b.closes(line, segment, w, pos, node, fdata); close {
		reader.Advance(segment.Stop - segment.Start - newline + segment.Padding)
		fdataMap = fdataMap[:len(fdataMap)-1]

		if len(fdataMap) == 0 {
			return parser.Close
		} else {
			pc.Set(fencedContainerInfoKey, fdataMap)
			return parser.Close
		}
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
