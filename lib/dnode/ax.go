package dnode

import (
	"ax-distiller/lib/chrome"
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
	"unique"

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

// FromAXTree converts an AX tree into a dnode tree.
func FromAXTree(root *chrome.AXNode, km Keymap) *Node {
	return convertAXTree(km, root, 0, 0)
}

// ToAXTree converts a dnode tree into an AX tree.
// - This assumes node is the AX_NODE container
func ToAXTree(km Keymap, node *Node) *chrome.AXNode {
	out := &chrome.AXNode{}

	var lastChild *chrome.AXNode

	cur := node.FirstChild
	for cur != nil {
		text, ok := km.StringOf(cur.FullKey)
		if !ok {
			panic("unknown key")
		}

		if strings.HasPrefix(text, "role:") {
			out.Role = unique.Make(text[5:])
		} else if strings.HasPrefix(text, "attr:") {
			value := ""
			if cur.FirstChild != nil {
				value, _ = km.StringOf(cur.FirstChild.FullKey)
			}
			out.Properties = append(out.Properties, chrome.Prop{
				Name:  unique.Make(text[5:]),
				Value: value,
			})
		} else if strings.HasPrefix(text, "AX_NODE") {
			child := ToAXTree(km, cur)
			if lastChild != nil {
				lastChild.NextSibling = child
			} else {
				out.FirstChild = child
			}
			lastChild = child
		}

		cur = cur.NextSibling
	}

	if out.Role.Value() == "" {
		out.Role = unique.Make("UNKNOWN")
	}

	return out
}

type axKey struct {
	childIdx uint64
}

var axKeyHash = bytes.NewBuffer(make([]byte, 1+8))

func (k axKey) Key() uint64 {
	axKeyHash.Truncate(0)
	binary.Write(axKeyHash, binary.LittleEndian, byte(0))
	binary.Write(axKeyHash, binary.LittleEndian, k.childIdx)
	return xxh3.Hash(axKeyHash.Bytes())
}

func (k axKey) String() string {
	return fmt.Sprintf("AX_NODE:%d", k.childIdx)
}

type roleKey struct {
	role unique.Handle[string]
}

func (k roleKey) Key() uint64 {
	buff := append([]byte{1}, []byte(k.role.Value())...)
	return xxh3.Hash(buff)
}

func (k roleKey) String() string {
	return fmt.Sprintf("role:%s", k.role.Value())
}

type attrKey struct {
	attr unique.Handle[string]
}

func (k attrKey) Key() uint64 {
	buff := append([]byte{2}, []byte(k.attr.Value())...)
	return xxh3.Hash(buff)
}

func (k attrKey) String() string {
	return fmt.Sprintf("attr:%s", k.attr.Value())
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
