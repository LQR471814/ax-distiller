package axextract

import "strings"

// FilterWhitespace removes all "whitespace" nodes which are defined as:
//   - name empty or with only whitespace
//   - description emtpy or with only whitespace
//   - all children are also whitespace nodes
func FilterWhitespace(node *AXNode) (isWhitespace bool) {
	nameEmpty := strings.Trim(node.Name, " \t\n") == ""
	descEmpty := strings.Trim(node.Description, " \t\n") == ""

	var newChildren []AXNode
	for _, child := range node.Children {
		isWhitespace := FilterWhitespace(&child)
		if isWhitespace {
			continue
		}
		newChildren = append(newChildren, child)
	}
	childrenEmpty := len(newChildren) == 0

	node.Children = newChildren
	return nameEmpty && descEmpty && childrenEmpty
}
