package mdrender

import (
	"regexp"
	"strings"
)

type layoutVariant = uint8

const (
	layout_inline layoutVariant = iota
	layout_block
)

type layoutNode interface {
	layoutNode()
}

type container struct {
	variant  layoutVariant
	children []layoutNode
}

type display struct {
	content string
}

func (container) layoutNode() {}
func (display) layoutNode()   {}

var consecutiveSpaceRegex = regexp.MustCompile("\\s\\s+")

// inline whitespace handling follows the following specification
// https://developer.mozilla.org/en-US/docs/Web/API/Document_Object_Model/Whitespace
// this follows a simplified algorithm, essentially:
// 1. convert tabs/newlines to spaces
// 2. collapse consecutive spaces into a single space
// 3. remove spaces from the start and end of a block layout (only done under block layout)
func renderLayout(nodes []layoutNode) string {
	inlineContext := true
	for _, child := range nodes {
		switch typedNode := child.(type) {
		case container:
			if typedNode.variant != layout_inline {
				inlineContext = false
				break
			}
		}
	}

	text := ""
	if inlineContext {
		for _, child := range nodes {
			switch typedNode := child.(type) {
			case container:
				text += renderLayout(typedNode.children)
			case display:
				text += typedNode.content
			}
		}
		text = consecutiveSpaceRegex.ReplaceAllString(text, " ")
	} else {
		for _, child := range nodes {
			switch typedNode := child.(type) {
			case container:
				text += renderLayout(typedNode.children)
			case display:
				text += typedNode.content
			}
			text += "\n\n"
		}
		text = strings.TrimSpace(text)
	}

	return text
}
