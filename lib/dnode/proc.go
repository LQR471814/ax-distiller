package dnode

import (
	"encoding/binary"
)

func (s DiffTree) ToBytes(rootHash uint64) []byte {
	buffer := make([]byte, 0, len(s.FromHash)*8*3+16)

	buffer = binary.LittleEndian.AppendUint64(buffer, uint64(len(s.FromHash)))
	buffer = binary.LittleEndian.AppendUint64(buffer, rootHash)

	for hash, node := range s.FromHash {
		buffer = binary.LittleEndian.AppendUint64(buffer, node.FullKey)
		buffer = binary.LittleEndian.AppendUint64(buffer, hash)

		if node.FirstChild == nil {
			binary.LittleEndian.AppendUint64(buffer, 0)
		} else {
			binary.LittleEndian.AppendUint64(buffer, node.FirstChild.FullKey)
		}

		if node.NextSibling == nil {
			binary.LittleEndian.AppendUint64(buffer, 0)
		} else {
			binary.LittleEndian.AppendUint64(buffer, node.NextSibling.FullKey)
		}
	}
	return buffer
}
