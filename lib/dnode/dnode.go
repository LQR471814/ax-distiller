package dnode

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/zeebo/xxh3"
)

// Node represents a single data value whose key (not full key, but individual key)
// is unique among its siblings, but not necessarily unique among the whole tree.
type Node struct {
	// FullKey must be the key computed with the FullKey(key, parentKey) function
	FullKey     uint64
	NextSibling *Node
	FirstChild  *Node
}

func CompositeHash(a, b uint64) uint64 {
	combo := [2]uint64{a, b}
	// "unsafe" cast of [2]uint64 -> []byte
	bslice := unsafe.Slice((*byte)(unsafe.Pointer(&combo)), unsafe.Sizeof(combo))
	hash := xxh3.Hash(bslice)
	return hash
}

type HashedNode struct {
	FullKey         uint64
	NextSiblingHash uint64
	FirstChildHash  uint64
}

type HashTree struct {
	FromHash map[uint64]HashedNode
	Root     uint64
}

func NewHashTree(node *Node, size int) HashTree {
	tree := HashTree{
		FromHash: make(map[uint64]HashedNode, size),
	}
	tree.Root, _ = tree.register(node)
	return tree
}

func (s HashTree) register(node *Node) (hash, nshash uint64) {
	if node == nil {
		return
	}

	hashedNode := HashedNode{
		FullKey: node.FullKey,
	}

	if node.NextSibling != nil {
		nsownhash, nsnshash := s.register(node.NextSibling)
		hashedNode.NextSiblingHash = nsownhash
		nshash = CompositeHash(nsownhash, nsnshash)
	}
	if node.FirstChild != nil {
		fcownhash, fcnshash := s.register(node.FirstChild)
		hashedNode.FirstChildHash = fcownhash
		// the node's hash should reflect the hashes of all its children
		hash = CompositeHash(node.FullKey, CompositeHash(fcownhash, fcnshash))
	} else {
		hash = node.FullKey
	}

	s.FromHash[hash] = hashedNode
	return
}

func printNode(out *strings.Builder, km Keymap, node *Node, depth int) {
	if node == nil {
		return
	}

	name, _ := km.StringOf(node.FullKey)
	for range depth {
		out.WriteString("  ")
	}
	out.WriteString(fmt.Sprintf("<%s>", name))
	if node.FirstChild != nil {
		out.WriteString("\n")
	}

	printNode(out, km, node.FirstChild, depth+1)

	if node.FirstChild != nil {
		for range depth {
			out.WriteString("  ")
		}
		out.WriteString(fmt.Sprintf("</%s>", name))
	}
	out.WriteString("\n")

	printNode(out, km, node.NextSibling, depth)
}

func Print(km Keymap, node *Node) string {
	var builder strings.Builder
	printNode(&builder, km, node, 0)
	return builder.String()
}
