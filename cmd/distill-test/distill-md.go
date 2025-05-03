package main

import (
	"ax-distiller/lib/ax"
	"ax-distiller/lib/markdown"
	"bytes"
	"io"
	"log/slog"
)

func getTextWithWhitespace(page ax.Page, node ax.Node) string {
	info, err := page.GetDomInfo(node.DomNodeId)
	if err == nil {
		return info.NodeValue
	}
	if node.DomNodeId != 0 {
		slog.Warn(
			"Failed to get dom node info",
			"role", node.Role,
			"id", node.DomNodeId,
			"err", err.Error(),
		)
	}
	return node.Name
}

func collectText(page ax.Page, node ax.Node, out io.Writer) {
	if node.Role == "StaticText" {
		out.Write([]byte(getTextWithWhitespace(page, node)))
		return
	}
	for _, c := range node.Children {
		collectText(page, c, out)
	}
}

func collectTextNodes(page ax.Page, node ax.Node, out *[]markdown.Inline) {
	if node.Role == "StaticText" {
		*out = append(*out, markdown.Text{
			Text: getTextWithWhitespace(page, node),
		})
		return
	}
	for _, c := range node.Children {
		collectTextNodes(page, c, out)
	}
}

func resolveHref(page ax.Page, href string) string {
	u, err := page.URL().Parse(href)
	if err != nil {
		slog.Warn("Failed to resolve href", "href", href, "base", page.URL().String())
		return ""
	}
	return u.String()
}

func parseInlineNode(page ax.Page, node ax.Node) markdown.Inline {
	switch node.Role {
	case "StaticText":
		return markdown.Text{Text: getTextWithWhitespace(page, node)}
	case "code":
		return markdown.Code{Text: getTextWithWhitespace(page, node)}
	case "deletion":
		children := []markdown.Inline{}
		collectTextNodes(page, node, &children)
		return markdown.HtmlTag{Tag: "del", Children: children}
	case "insertion":
		children := []markdown.Inline{}
		collectTextNodes(page, node, &children)
		return markdown.HtmlTag{Tag: "ins", Children: children}
	case "emphasis":
		children := []markdown.Inline{}
		collectTextNodes(page, node, &children)
		return markdown.Italic{Children: children}
	case "mark":
		children := []markdown.Inline{}
		collectTextNodes(page, node, &children)
		return markdown.HtmlTag{Tag: "mark", Children: children}
	case "subscript":
		children := []markdown.Inline{}
		collectTextNodes(page, node, &children)
		return markdown.HtmlTag{Tag: "sub", Children: children}
	case "superscript":
		children := []markdown.Inline{}
		collectTextNodes(page, node, &children)
		return markdown.HtmlTag{Tag: "sup", Children: children}
	case "time":
		children := []markdown.Inline{}
		collectTextNodes(page, node, &children)
		return markdown.HtmlTag{Tag: "time", Children: children}
	case "link":
		info, err := page.GetDomInfo(node.DomNodeId)
		if err != nil {
			slog.Warn("Failed to get link dom node", "id", node.DomNodeId, "err", err.Error())
			return nil
		}
		href, _ := info.Attribute("href")
		children := []markdown.Inline{}
		collectTextNodes(page, node, &children)
		if href == "" {
			return markdown.Fragment{Children: children}
		}
		return markdown.Link{
			Children: children,
			Href:     resolveHref(page, href),
		}
	case "image":
		info, err := page.GetDomInfo(node.DomNodeId)
		if err != nil {
			slog.Warn("Failed to get image dom node", "id", node.DomNodeId, "err", err.Error())
			return nil
		}
		src, _ := info.Attribute("src")
		return markdown.Link{
			Href: resolveHref(page, src),
			Children: []markdown.Inline{
				markdown.Text{Text: node.Name},
			},
			Image: true,
		}
	case "MathMLMath", "math":
		html, err := page.GetHTML(node.DomNodeId)
		if err != nil {
			slog.Warn(
				"Failed to retrieve MathML representation of math node.",
				"id", node.DomNodeId, "err", err.Error(),
			)
			return nil
		}
		return markdown.Text{Text: html}
	}
	return nil
}

func collectInlineNodes(page ax.Page, node ax.Node, out *[]markdown.Inline) {
	inline := parseInlineNode(page, node)
	if inline != nil {
		*out = append(*out, inline)
		return
	}
	for _, c := range node.Children {
		collectInlineNodes(page, c, out)
	}
}

func collectTableItems(page ax.Page, node ax.Node, out *[]markdown.TableRow) {
	for _, c := range node.Children {
		switch c.Role {
		case "rowgroup":
			collectTableItems(page, c, out)
		case "row":
			cells := [][]markdown.Inline{}
			for _, cell := range c.Children {
				switch cell.Role {
				case "rowheader", "columnheader", "cell", "gridcell":
					inlineNodes := []markdown.Inline{}
					collectInlineNodes(page, cell, &inlineNodes)
					cells = append(cells, inlineNodes)
				}
			}
			*out = append(*out, markdown.TableRow{
				Cells: cells,
			})
		}
	}
}

func collectListItems(page ax.Page, node ax.Node, out *[]markdown.ListItem) {
	for _, c := range node.Children {
		switch c.Role {
		case "listitem":
			inlineNodes := []markdown.Inline{}
			collectInlineNodes(page, c, &inlineNodes)
			*out = append(*out, markdown.ListItem{
				InlineRegion: inlineNodes,
			})
		case "list":
			items := []markdown.ListItem{}
			collectListItems(page, c, &items)
			*out = append(*out, markdown.ListItem{
				List: markdown.List{
					Items: items,
				},
			})
		}
	}
}

func convertToMd(page ax.Page, node ax.Node, blocks *[]markdown.Block) {
	accumulator := []markdown.Inline{}

	for _, child := range node.Children {
		inlineNode := parseInlineNode(page, child)
		if inlineNode != nil {
			accumulator = append(accumulator, inlineNode)
			continue
		}

		if len(accumulator) > 0 {
			*blocks = append(*blocks, markdown.Section{
				Children: accumulator,
			})
			accumulator = []markdown.Inline{}
		}

		// block nodes
		switch child.Role {
		case "table", "grid":
			rows := []markdown.TableRow{}
			collectTableItems(page, child, &rows)
			*blocks = append(*blocks, markdown.Table{
				Rows: rows,
			})
		case "list":
			items := []markdown.ListItem{}
			collectListItems(page, child, &items)
			*blocks = append(*blocks, markdown.List{
				Items: items,
			})
		case "heading":
			order := 2

			// this doesn't work for some reason??

			// info, err := page.GetDomInfo(node.DomNodeId)
			// if err != nil {
			// 	slog.Warn(
			// 		"Could not get dom info of header",
			// 		"id", node.DomNodeId, "err", err.Error(),
			// 	)
			// }
			// if info != nil && len(info.NodeName) == 2 {
			// 	charCode := int(info.NodeName[1])
			// 	if charCode >= '1' && charCode <= '6' {
			// 		order = int(info.NodeName[1] - '0')
			// 	}
			// }

			contents := []markdown.Inline{}
			collectInlineNodes(page, child, &contents)
			*blocks = append(*blocks, markdown.Header{
				Children: contents,
				Order:    order,
			})
		case "Pre":
			buff := bytes.NewBuffer(nil)
			collectText(page, child, buff)
			*blocks = append(*blocks, markdown.CodeBlock{
				Text: buff.String(),
			})
		default:
			convertToMd(page, child, blocks)
		}
	}

	if len(accumulator) > 0 {
		*blocks = append(*blocks, markdown.Section{
			Children: accumulator,
		})
		accumulator = []markdown.Inline{}
	}
}

func pageToMd(page ax.Page) []markdown.Block {
	tree := page.Tree
	blocks := []markdown.Block{}
	convertToMd(page, tree, &blocks)
	return blocks
}
