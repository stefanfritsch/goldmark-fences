package fences

import (
	"fmt"

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
	fenceID           string   // The ID of the fence. This enables nested fences with indentation
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
	line, _ := reader.PeekLine()
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
	node.SetAttributeString("data-fence", []byte(fenceID))

	attrs, ok := parser.ParseAttributes(reader)
	if ok {
		for _, attr := range attrs {
			node.SetAttribute(attr.Name, attr.Value)
		}
	}

	fdata := &fenceData{
		fenceID:           fenceID,
		char:              fenceChar,
		indent:            findent,
		length:            oFenceLength,
		node:              node,
		contentIndent:     0,
		contentHasStarted: false,
	}

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

	if close, _ := hasClosingTag(line, w, pos, fdata); w < fdata.indent || close {
		return node, parser.NoChildren
	}

	return node, parser.HasChildren
}

func (b *fencedContainerParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	// ========================================================================== //
	// Get fenceID from node

	rawFenceID, ok := node.AttributeString("data-fence")
	if !ok {
		// huhu: don't panic in production
		panic("fenceID is missing")
	}
	fenceID := string(rawFenceID.([]byte))

	// ========================================================================== //
	// 	Get fenceData for current fenceID

	rawdata := pc.Get(fencedContainerInfoKey)
	fdataMap := rawdata.([]*fenceData)

	var fdata *fenceData
	var flevel int
	for flevel = 0; flevel < len(fdataMap); flevel++ {
		fdata = fdataMap[flevel]
		if fdata.fenceID == fenceID {
			break
		}
	}

	// ========================================================================== //
	// 	Set indentation level if it hasn't been set yet

	line, segment := reader.PeekLine()
	w, pos := util.IndentWidth(line, reader.LineOffset())

	if !fdata.contentHasStarted && !util.IsBlank(line[pos:]) {
		fdata.contentHasStarted = true
		fdata.contentIndent = w

		fdataMap[flevel] = fdata
		pc.Set(fencedContainerInfoKey, fdataMap)
	}

	// ========================================================================== //
	// Are we closing the node?
	// * Either the indentation is below the indentation of the opening tags
	// * or it is at the level of the opening tags but the content was indented
	// * or there is a closing tag and we're in the deepest fenced block
	// indentClose :=
	// 	!util.IsBlank(line) &&
	// 		(w < fdata.indent || (w == fdata.indent && fdata.contentIndent > 0))
	close, newline := hasClosingTag(line, w, pos, fdata)

	if close && flevel == len(fdataMap)-1 {
		reader.Advance(segment.Stop - segment.Start - newline + segment.Padding)
		fdataMap = fdataMap[:flevel]
		node.SetAttributeString("data-fence", []byte(fmt.Sprint(flevel)))

		if len(fdataMap) == 0 {
			return parser.Close
		} else {
			pc.Set(fencedContainerInfoKey, fdataMap)
			return parser.Close
		}
	}

	if fdata.contentIndent > 0 {
		dontJumpLineEnd := segment.Stop - segment.Start - 1
		if fdata.contentIndent < dontJumpLineEnd {
			dontJumpLineEnd = fdata.contentIndent
		}

		reader.Advance(dontJumpLineEnd)
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

func hasClosingTag(line []byte, w int, pos int, fdata *fenceData) (bool, int) {
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
