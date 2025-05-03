package main

import (
	"ax-distiller/lib/ax"
	"bytes"
	"fmt"
	"io"
)

const INDENT_SIZE = 2

func debugAx(node ax.Node, out io.Writer, depth int) {
	offset := make([]byte, INDENT_SIZE * depth)
	for i := 0; i < INDENT_SIZE * depth; i++ {
		offset[i] = ' '
	}
	out.Write(offset)
	out.Write([]byte(fmt.Sprintf("%s \"%s\" (%d)\n", node.Role, node.Name, node.DomNodeId)))
	for _, c := range node.Children {
		debugAx(c, out, depth+1)
	}
}

func dumpPageAx(page ax.Page) string {
	buff := bytes.NewBuffer(nil)
	debugAx(page.Tree, buff, 0)
	return buff.String()
}
