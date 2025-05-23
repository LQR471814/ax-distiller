package dnode

import (
	"encoding/binary"
)

func (s DiffTree) ToBytes(dt DiffTree) []byte {

	buffer := make([]byte, len(dt.FromHash))
	l := len(dt.FromHash)
	binary.BigEndian.AppendUint64(buffer, uint64(l))

	for key, value := range dt.FromHash {
		binary.BigEndian.AppendUint64(buffer, key)
		if value.FirstChild == nil {
			binary.BigEndian.AppendUint64(buffer, 0)

		} else {
			binary.BigEndian.AppendUint64(buffer, value.FirstChild.FullKey)
		}

		if value.NextSibling == nil {
			binary.BigEndian.AppendUint64(buffer, 0)
		} else {
			binary.BigEndian.AppendUint64(buffer, value.NextSibling.FullKey)
		}
	}

	return buffer
}
