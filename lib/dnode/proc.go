package dnode

import (
	"encoding/binary"
)

func (s DiffTree) ToBytes(dt DiffTree) []byte {
	buffer := make([]byte, 0, len(dt.FromHash)*8*3)

	buffer = binary.BigEndian.AppendUint64(buffer, uint64(len(dt.FromHash)))

	for key, value := range dt.FromHash {
		buffer = binary.BigEndian.AppendUint64(buffer, key)

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
