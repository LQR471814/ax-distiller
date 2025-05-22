package dnode

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/zeebo/xxh3"
)

type Node struct {
	// FullKey must be the key computed with the FullKey(key, parentKey) function
	FullKey     uint64
	NextSibling *Node
	FirstChild  *Node
}

func FullKey(parentKey, key uint64) uint64 {
	buff := make([]byte, 8*2)
	buff = binary.BigEndian.AppendUint64(buff, parentKey)
	buff = binary.BigEndian.AppendUint64(buff, key)
	fullKey := xxh3.Hash(buff)
	return fullKey
}

type DiffTree struct {
	FromHash    map[uint64]*Node
	FromFullKey map[uint64][]*Node
}

func NewDiffTree(size int) DiffTree {
	return DiffTree{
		FromHash:    make(map[uint64]*Node, size),
		FromFullKey: make(map[uint64][]*Node, size),
	}
}

func (s DiffTree) Register(node *Node) (*Node, uint64) {
	key := node.FullKey

	// resolve next sibling to cached item if it already exists
	nshash := uint64(0)
	ns := node.NextSibling
	if ns != nil {
		ns, nshash = s.Register(ns)
	}

	// resolve first child to cached item if it already exists
	fchash := uint64(0)
	fc := node.FirstChild
	if fc != nil {
		fc, fchash = s.Register(fc)
	}

	node.NextSibling = ns
	node.FirstChild = fc

	buff := make([]byte, 8*3)
	buff = binary.BigEndian.AppendUint64(buff, key)
	buff = binary.BigEndian.AppendUint64(buff, nshash)
	buff = binary.BigEndian.AppendUint64(buff, fchash)
	hash := xxh3.Hash(buff)

	existing, exists := s.FromHash[hash]
	if !exists {
		s.FromHash[hash] = node
		s.FromFullKey[key] = append(s.FromFullKey[key], node)
		return node, hash
	}

	return existing, hash
}

func printNode(out *strings.Builder, debugMap map[uint64]string, node *Node, depth int) {
	if node == nil {
		return
	}

	text := debugMap[node.FullKey]
	for range depth {
		out.WriteString("  ")
	}
	out.WriteString(fmt.Sprintf("<%s>", text))
	if node.FirstChild != nil {
		out.WriteString("\n")
	}

	printNode(out, debugMap, node.FirstChild, depth+1)

	if node.FirstChild != nil {
		for range depth {
			out.WriteString("  ")
		}
		out.WriteString(fmt.Sprintf("</%s>", text))
	}
	out.WriteString("\n")

	printNode(out, debugMap, node.NextSibling, depth)
}

func Print(debugMap map[uint64]string, node *Node) string {
	var builder strings.Builder
	printNode(&builder, debugMap, node, 0)
	return builder.String()
}
