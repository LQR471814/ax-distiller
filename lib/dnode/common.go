package dnode

// FindCommon finds the common nodes in each tree.
// - hash must be a node's hash from self.
func FindCommon(km Keymap, self, other HashTree, hash uint64) *Node {
	if hash == 0 {
		return nil
	}

	node, exists := self.FromHash[hash]
	if !exists {
		panic("self node's hash does not exist in self")
	}

	var commonNS *Node
	if node.NextSiblingHash != 0 {
		commonNS = FindCommon(km, self, other, node.NextSiblingHash)
	}

	_, nodeAndSubtreeEqual := other.FromHash[hash]
	if nodeAndSubtreeEqual {
		return &Node{
			FullKey:     node.Original.FullKey,
			NextSibling: commonNS,
			FirstChild:  node.Original.FirstChild,
		}
	}

	_, nodeEqual := other.FromFullKey[node.Original.FullKey]
	if !nodeEqual {
		return commonNS
	}

	out := &Node{
		FullKey:     node.Original.FullKey,
		NextSibling: commonNS,
	}
	if node.FirstChildHash != 0 {
		out.FirstChild = FindCommon(km, self, other, node.FirstChildHash)
	}

	return out
}
