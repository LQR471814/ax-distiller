package ax

import (
	"ax-distiller/lib/chrome"
	"context"

	"github.com/chromedp/cdproto/accessibility"
	"github.com/zeebo/xxh3"
)

type Action interface {
	Do(pageCtx context.Context) error
}

// ActionWaitForChanges waits until the AX tree stops changing for 3 seconds
type ActionWaitForChanges struct {
}

// ActionClick clicks on an element given its selector
type ActionClick struct {
	Selector string
}

// ActionScroll scrolls on an element given its selector
type ActionScroll struct {
	Selector string
}

// implement heuristic selector construction, so called "element fingerprinting"
// that element fingerprinting only needs to uniquely identify an element in contrast
// to the other elements in the page
// so: "am I the only element with this specific style or class name or attribute"

// NodeSetMapping stores the mapping of accessibility node IDs to NodeSet.
type NodeSetMapping map[accessibility.NodeID]*NodeSet

// NodeSet is a set of possible children for a given node.
type NodeSet struct {
	// Key is a unique identifier corresponding to the AX role of the given node.
	Key uint64
	// Children contains the possible children for the given node.
	Children [][]NodeSet

	Mapping NodeSetMapping
}

// Add adds a possible children set to the current NodeSet.
func (cn *NodeSet) Add(children []chrome.AXNode) {
	newChildren := make([]NodeSet, len(children))
	for i, c := range children {
		newChildren[i] = NodeSet{
			Key:     xxh3.Hash([]byte(c.Role)),
			Mapping: cn.Mapping,
		}
		cn.Mapping[c.ID] = &newChildren[i]
	}
	for _, existing := range cn.Children {
		if childrenStructurallyEqual(newChildren, existing) {
			for i, c := range children {
				(&existing[i]).Add(c.Children)
			}
			return
		}
	}
	for i, c := range children {
		(&newChildren[i]).Add(c.Children)
	}
	cn.Children = append(cn.Children, newChildren)
}

func childrenStructurallyEqual(a, b []NodeSet) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Key != b[i].Key {
			return false
		}
	}
	return true
}
