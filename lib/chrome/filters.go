package chrome

import "strings"

// FilterWhitespace removes all "whitespace" nodes which are defined as:
//   - name empty or with only whitespace
//   - description empty or with only whitespace
//   - all children are also whitespace nodes
func FilterWhitespace(node *AXNode) *AXNode {
	nameEmpty := strings.Trim(node.Name, " \t\n") == ""
	descEmpty := strings.Trim(node.Description, " \t\n") == ""

	var firstChild *AXNode
	if node.FirstChild != nil {
		firstChild = FilterWhitespace(node.FirstChild)
	}

	var nextSibling *AXNode
	if node.NextSibling != nil {
		nextSibling = FilterWhitespace(node.NextSibling)
	}

	if nameEmpty && descEmpty {
		cur := firstChild
		for cur != nil {
			if cur.NextSibling == nil {
				cur.NextSibling = nextSibling
				break
			}
			cur = cur.NextSibling
		}
		return firstChild
	}

	node.FirstChild = firstChild
	node.NextSibling = nextSibling
	return node
}
