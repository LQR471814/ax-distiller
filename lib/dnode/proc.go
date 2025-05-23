package dnode

import (
	"encoding/binary"
)

func (s DiffTree) ToBytes(rootHash uint64) []byte {
	buffer := make([]byte, 0, len(s.FromHash)*8*3 + 16)

	buffer = binary.SmallEndian.AppendUint64(buffer, uint64(len(s.FromHash) ))
	buffer = binary.SmallEndian.AppendUint64(buffer, rootHash)

	for key, value := range s.FromHash {
		buffer = binary.SmallEndian.AppendUint64(buffer, key)

		if value.FirstChild == nil {
			binary.SmallEndian.AppendUint64(buffer, 0)
		} else {
			binary.SmallEndian.AppendUint64(buffer, value.FirstChild.FullKey)
		}

		if value.NextSibling == nil {
			binary.SmallEndian.AppendUint64(buffer, 0)
		} else {
			binary.SmallEndian.AppendUint64(buffer, value.NextSibling.FullKey)
		}
	}
	return buffer
}
