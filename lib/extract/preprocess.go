package extract

// Extract the first "main" node from the tree.
func ExtractMain(root AXNode) (AXNode, bool) {
	queue := []AXNode{root}

	for len(queue) > 0 {
		lowerQueue := []AXNode{}

		for _, node := range queue {
			if string(node.Role) == "main" {
				return node, true
			}
			lowerQueue = append(lowerQueue, node.Children...)
		}

		queue = lowerQueue
	}

	return AXNode{}, false
}
