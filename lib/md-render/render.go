package mdrender

import (
	"ax-distiller/lib/utils"
	"fmt"
	"strings"
)

type renderState struct {
	listDepth int
}

func Render(nodes []Node) string {
	text := ""
	for _, n := range nodes {
		rendered := render(n, renderState{})
		if rendered == "" {
			continue
		}
		text += rendered + "\n\n"
	}
	return text
}

func render(node Node, state renderState) string {
	text := ""

	switch typedNode := node.(type) {
	case Header:
		if typedNode.Order > 6 {
			panic("header order must be <= 6")
		}

		content := render(typedNode.Content, state)
		if content == "" {
			return ""
		}

		for i := 0; i < typedNode.Order; i++ {
			text += "#"
		}
		text += " "
		text += content
	case Paragraph:
		for i, e := range typedNode.Elements {
			rendered := render(e, state)
			if rendered == "" {
				continue
			}
			if i > 0 {
				text += " "
			}
			text += rendered
			if i >= len(typedNode.Elements) {
				text += " "
			}
		}
	case Link:
		inner := ""
		switch typedNode.Title.(type) {
		case Link:
			inner = render(typedNode.Title.(Link).Title, state)
		default:
			inner = render(typedNode.Title, state)
		}
		if inner == "" {
			break
		}

		if typedNode.Image {
			text += "!"
		}
		text += fmt.Sprintf("[%s](%s)", inner, typedNode.URL)
	case LineBreak:
		text = "<br>"
	case HorizontalRule:
		text = "<hr>"
	case InlineCode:
		content := strings.ReplaceAll(typedNode.Content, "`", "\\`")
		if content == "" {
			break
		}
		text = fmt.Sprintf("`%s`", content)
	case BlockCode:
		content := strings.ReplaceAll(
			utils.RemoveInvisible(strings.TrimRight(typedNode.Content, " \n\t")),
			"`", "\\`",
		)
		if content == "" {
			break
		}
		text = fmt.Sprintf(
			"```%s\n%s\n```", typedNode.Language,
			content,
		)
	case PlainText:
		text = utils.RemoveInvisible(strings.Trim(typedNode.Content, " \n\t"))
	case DecoratedText:
		inside := render(typedNode.Content, state)
		if inside == "" {
			return ""
		}
		switch typedNode.Type {
		case DECOR_BOLD:
			text = fmt.Sprintf("**%s**", strings.ReplaceAll(inside, "*", "_"))
		case DECOR_ITALIC:
			text = fmt.Sprintf("*%s*", strings.ReplaceAll(inside, "*", "_"))
		case DECOR_UNDERLINE:
			text = fmt.Sprintf("<ins>%s</ins>", inside)
		}
	case List:
		indent := ""
		for i := 0; i < state.listDepth; i++ {
			indent += "   "
		}

		prefix := "-"
		if typedNode.Type == LIST_ORDERED {
			prefix = "1."
		}

		for i, item := range typedNode.Items {
			content := render(item, renderState{
				listDepth: state.listDepth + 1,
			})
			if content == "" {
				continue
			}

			switch item.(type) {
			case List:
				text += content
			default:
				if indent != "" {
					text += indent
				}
				text += fmt.Sprintf("%s %s", prefix, content)
			}
			if i < len(typedNode.Items)-1 {
				text += "\n"
			}
		}
	case Table:
		maxColumnCount := 0
		for _, row := range typedNode.Rows {
			if len(row) > maxColumnCount {
				maxColumnCount = len(row)
			}
		}
		if maxColumnCount == 0 {
			break
		}

		table := [][]string{}
		columnLengths := make([]int, maxColumnCount)
		for _, cellList := range typedNode.Rows {
			emptyRow := true
			row := make([]string, len(cellList))
			for col, cell := range cellList {
				cellText := render(cell, state)
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
	}

	return text
}
