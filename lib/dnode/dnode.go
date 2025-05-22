package dnode

import (
	"encoding/binary"

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

type Tree struct {
	FromHash    map[uint64]*Node
	FromFullKey map[uint64][]*Node
}

func NewTree(size int) Tree {
	return Tree{
		FromHash:    make(map[uint64]*Node, size),
		FromFullKey: make(map[uint64][]*Node, size),
	}
}

func (s Tree) Register(node *Node) (*Node, uint64) {
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
