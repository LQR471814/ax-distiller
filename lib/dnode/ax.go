package dnode

import (
	"ax-distiller/lib/chrome"
	"fmt"
	"unsafe"

	"github.com/zeebo/xxh3"
)

func convertAXTree(km Keymap, n *chrome.AXNode, childIdx, parentKey uint64) (dn *Node) {
	if n == nil {
		return nil
	}

	axFullKey := km.FullKey(parentKey, axKey{childIdx})
	dn = &Node{
		FullKey: axFullKey,
	}
	if n.NextSibling != nil {
		dn.NextSibling = convertAXTree(km, n.NextSibling, childIdx+1, axFullKey)
	}

	role := &Node{
		FullKey: km.FullKey(axFullKey, roleKey{n.Role}),
	}
	dn.FirstChild = role

	var endProp *Node

	var nextAttr *Node
	for i := len(n.Properties) - 1; i >= 0; i-- {
		p := n.Properties[i]
		dprop := &Node{
			FullKey:     km.FullKey(axFullKey, attrKey{p.Name}),
			NextSibling: nextAttr,
		}
		if endProp == nil {
			endProp = dprop
		}

		if p.Value != "" {
			dprop.FirstChild = &Node{
				FullKey: km.FullKey(dprop.FullKey, valueKey{p.Value}),
			}
		}
		nextAttr = dprop
	}
	role.NextSibling = nextAttr

	if endProp == nil {
		endProp = role
	}
	if n.FirstChild != nil {
		endProp.NextSibling = convertAXTree(km, n.FirstChild, 0, axFullKey)
	}
	return
}

func ConvertAXTree(root *chrome.AXNode, km Keymap) *Node {
	return convertAXTree(km, root, 0, 0)
}

type axKey struct {
	childIdx uint64
}

func (k axKey) Key() uint64 {
	combo := struct {
		Type byte
		Idx  uint64
	}{Type: 0, Idx: k.childIdx}
	buff := unsafe.Slice((*byte)(unsafe.Pointer(&combo)), unsafe.Sizeof(combo))
	return xxh3.Hash(buff)
}

func (k axKey) String() string {
	return fmt.Sprintf("AX_NODE:%d", k.childIdx)
}

type roleKey struct {
	role string
}

func (k roleKey) Key() uint64 {
	buff := append([]byte{1}, []byte(k.role)...)
	return xxh3.Hash(buff)
}

func (k roleKey) String() string {
	return fmt.Sprintf("role:%s", k.role)
}

type attrKey struct {
	attr string
}

func (k attrKey) Key() uint64 {
	buff := append([]byte{2}, []byte(k.attr)...)
	return xxh3.Hash(buff)
}

func (k attrKey) String() string {
	return fmt.Sprintf("attr:%s", k.attr)
}

type valueKey struct {
	value string
}

func (k valueKey) Key() uint64 {
	return xxh3.Hash([]byte(k.value))
}

func (k valueKey) String() string {
	return k.value
}
