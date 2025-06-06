package dnode

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"

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

type HashedNode struct {
	Hash     uint64
	Original *Node
	// NextSiblingHash that equals 0 means there is no next sibling.
	NextSiblingHash uint64
	// FirstChildHash that equals 0 means there is no first child.
	FirstChildHash uint64
}

type HashTree struct {
	FromHash    map[uint64]HashedNode
	FromFullKey map[uint64]HashedNode
	Root        uint64
}

func NewHashTree(node *Node) HashTree {
	tree := HashTree{
		FromHash:    make(map[uint64]HashedNode),
		FromFullKey: make(map[uint64]HashedNode),
	}
	tree.Root, _ = tree.register(node)
	return tree
}

var hsRegisterHash = bytes.NewBuffer(make([]byte, 8*2))

func (s HashTree) register(node *Node) (hash, nshash uint64) {
	if node == nil {
		return
	}

	hashedNode := HashedNode{
		Original: node,
	}

	if node.NextSibling != nil {
		nsownhash, nsnshash := s.register(node.NextSibling)
		hashedNode.NextSiblingHash = nsownhash

		hsRegisterHash.Truncate(0)
		binary.Write(hsRegisterHash, binary.LittleEndian, nsownhash)
		binary.Write(hsRegisterHash, binary.LittleEndian, nsnshash)
		nshash = xxh3.Hash(hsRegisterHash.Bytes())
	}
	if node.FirstChild != nil {
		fcownhash, fcnshash := s.register(node.FirstChild)
		hashedNode.FirstChildHash = fcownhash

		hsRegisterHash.Truncate(0)
		binary.Write(hsRegisterHash, binary.LittleEndian, fcownhash)
		binary.Write(hsRegisterHash, binary.LittleEndian, fcnshash)
		childHash := xxh3.Hash(hsRegisterHash.Bytes())

		hsRegisterHash.Truncate(0)
		binary.Write(hsRegisterHash, binary.LittleEndian, node.FullKey)
		binary.Write(hsRegisterHash, binary.LittleEndian, childHash)
		hash = xxh3.Hash(hsRegisterHash.Bytes())
	} else {
		hash = node.FullKey
	}

	hashedNode.Hash = hash
	s.FromHash[hash] = hashedNode
	s.FromFullKey[node.FullKey] = hashedNode
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
