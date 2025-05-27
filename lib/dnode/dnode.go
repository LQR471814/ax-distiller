package dnode

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/zeebo/xxh3"
)

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

func (s DiffTree) register(node *Node) (resolved *Node, hash, nshash uint64) {
	if node == nil {
		return
	}

	if node.NextSibling != nil {
		ns, nsownhash, nsnshash := s.register(node.NextSibling)
		node.NextSibling = ns
		nshash = CompositeHash(nsownhash, nsnshash)
	}

	if node.FirstChild != nil {
		fc, fcownhash, fcnshash := s.register(node.FirstChild)
		node.FirstChild = fc
		// the node's hash should reflect the hashes of all its children
		hash = CompositeHash(node.FullKey, CompositeHash(fcownhash, fcnshash))
	} else {
		hash = node.FullKey
	}

	existing, ok := s.FromHash[hash]
	if ok {
		fmt.Println("resolved", node.FullKey)
		resolved = existing
	} else {
		resolved = node
		s.FromHash[hash] = node
		s.FromFullKey[node.FullKey] = append(s.FromFullKey[node.FullKey], node)
	}

	return
}

func (s DiffTree) Register(node *Node) (resolved *Node, hash uint64) {
	resolved, hash, _ = s.register(node)
	return
}

func printNode(out *strings.Builder, debugMap map[uint64]DebugEntry, node *Node, depth int) {
	if node == nil {
		return
	}

	entry := debugMap[node.FullKey]
	for range depth {
		out.WriteString("  ")
	}
	out.WriteString(fmt.Sprintf("<%s>", entry.Name))
	if node.FirstChild != nil {
		out.WriteString("\n")
	}

	printNode(out, debugMap, node.FirstChild, depth+1)

	if node.FirstChild != nil {
		for range depth {
			out.WriteString("  ")
		}
		out.WriteString(fmt.Sprintf("</%s>", entry.Name))
	}
	out.WriteString("\n")

	printNode(out, debugMap, node.NextSibling, depth)
}

func Print(debugMap map[uint64]DebugEntry, node *Node) string {
	var builder strings.Builder
	printNode(&builder, debugMap, node, 0)
	return builder.String()
}
