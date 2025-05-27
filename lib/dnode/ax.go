package dnode

import (
	"ax-distiller/lib/chrome"
	"fmt"
	"unsafe"

	"github.com/zeebo/xxh3"
)

type DebugEntry struct {
	Name   string `json:"name"`
	Parent uint64 `json:"parent"`
}

type AXTree struct {
	DebugText map[uint64]DebugEntry
	Root      *Node
}

func (t AXTree) axFullKey(childIdx uint64) uint64 {
	combo := struct {
		Type byte
		Idx  uint64
	}{Type: 0, Idx: childIdx}
	buff := unsafe.Slice((*byte)(unsafe.Pointer(&combo)), unsafe.Sizeof(combo))
	return xxh3.Hash(buff)
}

func (t AXTree) roleFullKey(role string) uint64 {
	buff := append([]byte{1}, []byte(role)...)
	return xxh3.Hash(buff)
}

func (t AXTree) attrFullKey(name string) uint64 {
	buff := append([]byte{2}, []byte(name)...)
	return xxh3.Hash(buff)
}

func (t AXTree) convertNode(n *chrome.AXNode, childIdx, parentKey uint64) (dn *Node) {
	if n == nil {
		return nil
	}

	fullKey := CompositeHash(parentKey, t.axFullKey(childIdx))
	dn = &Node{
		FullKey: fullKey,
	}
	if t.DebugText != nil {
		t.DebugText[fullKey] = DebugEntry{
			Name:   fmt.Sprintf("AX_NODE:%d", childIdx),
			Parent: parentKey,
		}
	}
	if n.NextSibling != nil {
		dn.NextSibling = t.convertNode(n.NextSibling, childIdx+1, fullKey)
	}

	role := &Node{
		FullKey: CompositeHash(fullKey, t.roleFullKey(n.Role)),
	}
	if t.DebugText != nil {
		t.DebugText[role.FullKey] = DebugEntry{
			Name:   fmt.Sprintf("role:%s", n.Role),
			Parent: parentKey,
		}
	}
	dn.FirstChild = role

	var endProp *Node

	var nextAttr *Node
	for i := len(n.Properties) - 1; i >= 0; i-- {
		p := n.Properties[i]
		dprop := &Node{
			FullKey:     CompositeHash(fullKey, t.attrFullKey(p.Name)),
			NextSibling: nextAttr,
		}
		if t.DebugText != nil {
			t.DebugText[dprop.FullKey] = DebugEntry{
				Name:   fmt.Sprintf("attr:%s", p.Name),
				Parent: parentKey,
			}
		}
		if endProp == nil {
			endProp = dprop
		}

		if p.Value != "" {
			dprop.FirstChild = &Node{
				FullKey: CompositeHash(dprop.FullKey, xxh3.Hash([]byte(p.Value))),
			}
			if t.DebugText != nil {
				t.DebugText[dprop.FirstChild.FullKey] = DebugEntry{
					Name:   p.Value,
					Parent: parentKey,
				}
			}
		}
		nextAttr = dprop
	}
	role.NextSibling = nextAttr

	if endProp == nil {
		endProp = role
	}
	if n.FirstChild != nil {
		endProp.NextSibling = t.convertNode(n.FirstChild, 0, fullKey)
	}
	return
}

func NewAXTree(root *chrome.AXNode, debugText map[uint64]DebugEntry) AXTree {
	ax := AXTree{
		DebugText: debugText,
	}
	ax.Root = ax.convertNode(root, 0, 0)
	return ax
}
