package main

import (
	"ax-distiller/lib/ax"
	"bytes"
	"fmt"
	"io"
	"os"
)

const template = `
<!DOCTYPE html>
<html>
	<body>
		%s
	</body>
</html>
`

func convertToHtml(page ax.Page, node ax.Node, out io.Writer) {
	category := ax.RoleCategoryMap[node.Role]
	if len(category) > 0 {
		switch category[0] {
		case ax.CATEGORY_STRUCTURE:
			out.Write([]byte("<div>"))
			defer out.Write([]byte("</div>"))
		case ax.CATEGORY_SECTIONHEAD:
			out.Write([]byte("<h2 style=\"border: 1px solid aqua; margin: 0.5rem 0px\">"))
			defer out.Write([]byte("</h2>"))
		case ax.CATEGORY_SECTION:
			out.Write([]byte("<div style=\"border: 1px solid blue; margin: 4px\">"))
			defer out.Write([]byte("</div>"))
		case ax.CATEGORY_LANDMARK:
			out.Write([]byte("<div style=\"border: 1px solid green; margin: 4px\">"))
			defer out.Write([]byte("</div>"))
		case ax.CATEGORY_WIDGET:
			out.Write([]byte(fmt.Sprintf("<a href=\"/interact?id=%d\">", node.DomNodeId)))
			defer out.Write([]byte("</a>"))
		case ax.CATEGORY_DOCUMENT:
			out.Write([]byte("<div>"))
			defer out.Write([]byte("</div>"))
		}
	}

	switch node.Role {
	case "StaticText":
		out.Write([]byte(node.Name + " "))
	case "image":
		info, err := page.GetDomInfo(node.DomNodeId)
		if err != nil {
			fmt.Fprintln(os.Stderr, "WARN", err.Error())
			return
		}

		src, ok := info.Attribute("src")
		if !ok {
			return
		}

		u, err := page.URL().Parse(src)
		if err != nil {
			fmt.Fprintln(os.Stderr, "WARN", err.Error())
			return
		}

		out.Write([]byte(fmt.Sprintf("<a style=\"font-style: italic\" href=\"%s\">[IMAGE]</a>", u.String())))
		return
	}

	for _, c := range node.Children {
		convertToHtml(page, c, out)
	}
}

func pageToHtml(page ax.Page) string {
	tree := page.Tree
	out := bytes.NewBuffer(nil)
	convertToHtml(page, tree, out)
	return fmt.Sprintf(template, out.String())
}

