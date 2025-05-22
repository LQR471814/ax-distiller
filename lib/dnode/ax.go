package dnode

import (
	"ax-distiller/lib/chrome"
	"fmt"
	"strings"

	"github.com/zeebo/xxh3"
)

type AXTree struct {
	DebugText map[uint64]string
	Root      *Node
}

var ax_node_id = xxh3.Hash([]byte("AX_NODE"))

func (t AXTree) convertNode(n *chrome.AXNode, parentKey uint64) (dn *Node) {
	if n == nil {
		return nil
	}

	fullKey := FullKey(parentKey, ax_node_id)
	dn = &Node{
		FullKey: fullKey,
	}
	if t.DebugText != nil {
		t.DebugText[fullKey] = "ax_node"
	}
	if n.NextSibling != nil {
		dn.NextSibling = t.convertNode(n.NextSibling, fullKey)
	}

	role := &Node{
		FullKey: FullKey(fullKey, xxh3.Hash([]byte(n.Role))),
	}
	if t.DebugText != nil {
		t.DebugText[role.FullKey] = fmt.Sprintf("role:%s", n.Role)
	}
	dn.FirstChild = role

	var endProp *Node

	var nextAttr *Node
	for i := len(n.Properties) - 1; i >= 0; i-- {
		p := n.Properties[i]
		dprop := &Node{
			FullKey:     FullKey(fullKey, xxh3.Hash([]byte(p.Name))),
			NextSibling: nextAttr,
		}
		if t.DebugText != nil {
			t.DebugText[dprop.FullKey] = fmt.Sprintf("attr:%s", p.Name)
		}
		if endProp == nil {
			endProp = dprop
		}

		if p.Value != "" {
			dprop.FirstChild = &Node{
				FullKey: FullKey(dprop.FullKey, xxh3.Hash([]byte(p.Value))),
			}
			if t.DebugText != nil {
				t.DebugText[dprop.FirstChild.FullKey] = p.Value
			}
		}
		nextAttr = dprop
	}
	role.NextSibling = nextAttr

	if endProp == nil {
		endProp = role
	}
	if n.FirstChild != nil {
		endProp.NextSibling = t.convertNode(n.FirstChild, fullKey)
	}
	return
}

func (t AXTree) indent(out *strings.Builder, depth int) {
	for range depth {
		out.WriteString("  ")
	}
}

func (t AXTree) marshal(out *strings.Builder, node *Node, depth int) {
	if node == nil {
		return
	}

	text := t.DebugText[node.FullKey]
	t.indent(out, depth)
	out.WriteString(fmt.Sprintf("<%s>", text))
	if node.FirstChild != nil {
		out.WriteString("\n")
	}

	t.marshal(out, node.FirstChild, depth+1)

	if node.FirstChild != nil {
		t.indent(out, depth)
		out.WriteString(fmt.Sprintf("</%s>", text))
	}
	out.WriteString("\n")

	t.marshal(out, node.NextSibling, depth)
}

func (t AXTree) String() string {
	var builder strings.Builder
	t.marshal(&builder, t.Root, 0)
	return builder.String()
}

func NewAXTree(root *chrome.AXNode, debug bool) AXTree {
	ax := AXTree{}
	if debug {
		ax.DebugText = make(map[uint64]string)
	}
	ax.Root = ax.convertNode(root, 0)
	return ax
}
