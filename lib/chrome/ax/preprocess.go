package ax

import (
	"strings"
)

// FilterWhitespace removes all "whitespace" nodes which are defined as:
//   - name empty or with only whitespace
//   - description empty or with only whitespace
//   - all children are also whitespace nodes
//
// FilterWhitespace returns the next node with content
func FilterWhitespace(node *Node) *Node {
	if node.FirstChild != nil {
		node.FirstChild = FilterWhitespace(node.FirstChild)
	}
	if node.NextSibling != nil {
		node.NextSibling = FilterWhitespace(node.NextSibling)
	}

	nameEmpty := strings.Trim(node.Name, " \t\n") == ""
	descEmpty := strings.Trim(node.Description, " \t\n") == ""
	urlEmpty := true
	for _, p := range node.Properties {
		if p.Name.Value() == "url" && p.Value != "" {
			urlEmpty = false
			break
		}
	}
	hasContent := !nameEmpty || !descEmpty || !urlEmpty

	// if self is not whitespace node
	if hasContent {
		return node
	}

	if node.FirstChild != nil && node.NextSibling != nil {
		return node
	}
	if node.FirstChild == nil && node.NextSibling == nil {
		return nil
	}
	if node.NextSibling != nil {
		return node.NextSibling
	}

	return node
}
