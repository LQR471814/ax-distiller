package ax

import (
	"encoding/binary"

	"github.com/zeebo/xxh3"
)

type genericNode interface {
	Key() uint64
	NextSibling() genericNode
	FirstChild() genericNode
}

type treeSetNode struct {
	Key             uint64
	NextSiblingHash uint64
	FirstChildHash  uint64
}

type treeSet struct {
	FromHash map[uint64]treeSetNode
	FromKey  map[uint64][]treeSetNode
}

func newTreeSet(size int) treeSet {
	return treeSet{
		FromHash: make(map[uint64]treeSetNode, size),
		FromKey:  make(map[uint64][]treeSetNode, size),
	}
}

func (s treeSet) Register(node genericNode) (hash uint64) {
	nshash := uint64(0)
	ns := node.NextSibling()
	if ns != nil {
		nshash = s.Register(ns)
	}
	fchash := uint64(0)
	fc := node.FirstChild()
	if fc != nil {
		fchash = s.Register(fc)
	}

	key := node.Key()
	buff := make([]byte, 8*3)
	binary.BigEndian.AppendUint64(buff, key)
	binary.BigEndian.AppendUint64(buff, nshash)
	binary.BigEndian.AppendUint64(buff, fchash)
	hash = xxh3.Hash(buff)

	tsnode := treeSetNode{
		Key:             key,
		NextSiblingHash: nshash,
		FirstChildHash:  fchash,
	}
	_, exists := s.FromHash[hash]
	if !exists {
		s.FromHash[hash] = tsnode
		s.FromKey[key] = append(s.FromKey[key], tsnode)
	}

	return
}
