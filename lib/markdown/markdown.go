package markdown

import (
	"fmt"
	"regexp"
	"strings"
)

// block elements

type Header struct {
	Children []Inline
	Order    int
}

type Section struct {
	Children []Inline
}

type Table struct {
	Rows []TableRow
}

type TableRow struct {
	Cells [][]Inline
}

type List struct {
	Items []ListItem
}

// this is a union, either InlineRegion or List will be defined
type ListItem struct {
	InlineRegion []Inline
	List         List
}

type CodeBlock struct {
	Language string
	Text     string
}

type Block interface {
	block()
}

func (Header) block()    {}
func (Section) block()   {}
func (Table) block()     {}
func (CodeBlock) block() {}
func (List) block()      {}

// inline elements

type HtmlTag struct {
	Children []Inline
	Tag      string
}

type Text struct {
	Text string
}

type Link struct {
	Children []Inline
	Href     string
	Image    bool
}

type Fragment struct {
	Children []Inline
}

type Italic struct {
	Children []Inline
}

type Bold struct {
	Children []Inline
}

type Code struct {
	Text string
}

type Inline interface {
	inline()
}

func (HtmlTag) inline()  {}
func (Text) inline()     {}
func (Link) inline()     {}
func (Fragment) inline() {}
func (Code) inline()     {}
func (Bold) inline()     {}
func (Italic) inline()   {}

// rendering logic

var consecutiveWhitespace = regexp.MustCompile(`\s\s+`)

// inline whitespace handling follows the following specification
// https://developer.mozilla.org/en-US/docs/Web/API/Document_Object_Model/Whitespace
// this follows a simplified algorithm, essentially:
// 1. convert tabs/newlines to spaces
// 2. collapse consecutive spaces into a single space
// 3. remove spaces from the start and end of a block layout (only done in renderSection)
func renderInlineRegion(nodes []Inline) string {
	out := ""

	for _, i := range nodes {
		switch el := i.(type) {
		case Text:
			out += el.Text
		case Fragment:
			out += renderInlineRegion(el.Children)
		case HtmlTag:
			text := renderInlineRegion(el.Children)
			if text == "" {
				continue
			}
			out += fmt.Sprintf("<%s>%s</%s>", el.Tag, text, el.Tag)
		case Code:
			if el.Text == "" {
				continue
			}
			if strings.Contains(el.Text, "`") {
				return fmt.Sprintf("<code>%s</code>", el.Text)
			}
			return fmt.Sprintf("`%s`", el.Text)
		case Bold:
			text := renderInlineRegion(el.Children)
			if text == "" {
				continue
			}
			if strings.Contains(text, "*") {
				return fmt.Sprintf("<b>%s</b>", text)
			}
			return fmt.Sprintf("**%s**", text)
		case Italic:
			text := renderInlineRegion(el.Children)
			if text == "" {
				continue
			}
			if strings.Contains(text, "*") {
				return fmt.Sprintf("<i>%s</i>", text)
			}
			return fmt.Sprintf("*%s*", text)
		case Link:
			text := renderInlineRegion(el.Children)
			if text == "" {
				continue
			}

			start := text[0]
			end := text[len(text)-1]

			if start == ' ' {
				out += " "
			}
			if el.Image {
				out += "!"
			}
			out += fmt.Sprintf("[%s](%s)", strings.Trim(text, " \t\n"), el.Href)
			if end == ' ' {
				out += " "
			}
		}
	}

	result := consecutiveWhitespace.ReplaceAllString(out, " ")
	if result == "" {
		return ""
	}

	return result
}

func renderList(l List, depth int) string {
	out := ""
	for _, i := range l.Items {
		if i.InlineRegion != nil {
			text := renderInlineRegion(i.InlineRegion)
			if text == "" {
				continue
			}
			for i := 0; i < depth; i++ {
				out += "   "
			}
			out += "- " + text + "\n"
			continue
		}
		out += renderList(i.List, depth+1)
	}
	return out
}

func renderBlock(b Block) string {
	switch el := b.(type) {
	case Header:
		rendered := renderInlineRegion(el.Children)
		if rendered == "" {
			return ""
		}

		prefix := ""
		for i := 0; i < el.Order; i++ {
			prefix += "#"
		}

		return fmt.Sprintf("%s %s\n\n", prefix, rendered)
	case Section:
		region := renderInlineRegion(el.Children)
		trimmed := strings.Trim(region, " ")
		if trimmed == "" {
			return ""
		}
		return trimmed + "\n\n"
	case CodeBlock:
		if el.Text == "" {
			return ""
		}
		return fmt.Sprintf("```%s\n%s```\n\n", el.Language, el.Text)
	case Table:
		maxColumnCount := 0
		for _, row := range el.Rows {
			if len(row.Cells) > maxColumnCount {
				maxColumnCount = len(row.Cells)
			}
		}
		if maxColumnCount == 0 {
			return ""
		}

		table := [][]string{}
		columnLengths := make([]int, maxColumnCount)
		for _, r := range el.Rows {
			emptyRow := true
			row := make([]string, len(r.Cells))
			for col, cell := range r.Cells {
				cellText := renderInlineRegion(cell)
				if len(cellText) > 0 {
					emptyRow = false
				}
				row[col] = cellText

				if len(cellText) > columnLengths[col] {
					columnLengths[col] = len(cellText)
				}
			}
			if emptyRow {
				continue
			}
			table = append(table, row)
		}

		text := ""
		for rowIdx, row := range table {
			text += "|"
			for col := 0; col < maxColumnCount; col++ {
				if columnLengths[col] == 0 {
					continue
				}
				cell := ""
				if col < len(row) {
					cell = row[col]
				}
				text += fmt.Sprintf(" %-*s |", columnLengths[col], cell)
			}
			text += "\n"
			if rowIdx == 0 {
				text += "|"
				for _, length := range columnLengths {
					if length == 0 {
						continue
					}
					text += " " + strings.Repeat("-", length) + " |"
				}
				text += "\n"
			}
		}
		if text == "" {
			return ""
		}

		return text + "\n"
	case List:
		text := renderList(el, 0)
		if text == "" {
			return ""
		}
		return text + "\n"
	}
	return ""
}

func Render(blocks []Block) string {
	out := ""
	for _, b := range blocks {
		out += renderBlock(b)
	}
	return out
}
