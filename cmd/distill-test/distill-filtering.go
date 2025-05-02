package main

import "ax-distiller/lib/ax"

func onlyTextContent(node ax.AXNode) (ax.AXNode, bool) {
	if node.Role == "paragraph" || node.Role == "heading" {
		return node, true
	}

	var children []ax.AXNode
	for _, c := range node.Children {
		transformed, keep := onlyTextContent(c)
		if keep {
			children = append(children, transformed)
		}
	}

	node.Children = children
	if len(children) > 0 {
		return node, true
	}

	return node, false
}

func noTextContent(node ax.AXNode) ax.AXNode {
	var children []ax.AXNode
	for _, c := range node.Children {
		if c.Role == "paragraph" || c.Role == "heading" {
			continue
		}
		children = append(children, noTextContent(c))
	}
	node.Children = children
	return node
}
