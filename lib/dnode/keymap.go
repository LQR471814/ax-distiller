package dnode

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/zeebo/xxh3"
)

type KeymapValue interface {
	Key() uint64
	String() string
}

type keymapEntry struct {
	Text   string
	Parent uint64
}

type Keymap struct {
	Textmap map[uint64]keymapEntry
}

var keymapHash = bytes.NewBuffer(make([]byte, 8*2))

func (km Keymap) FullKey(parentFullKey uint64, v KeymapValue) (fullkey uint64) {
	keymapHash.Truncate(0)
	binary.Write(keymapHash, binary.LittleEndian, parentFullKey)
	binary.Write(keymapHash, binary.LittleEndian, v.Key())
	fullkey = xxh3.Hash(keymapHash.Bytes())

	if km.Textmap == nil {
		return
	}
	km.Textmap[fullkey] = keymapEntry{
		Text:   v.String(),
		Parent: parentFullKey,
	}
	return
}

func (km Keymap) StringOf(fullkey uint64) (string, bool) {
	res, ok := km.Textmap[fullkey]
	return res.Text, ok
}

func (km Keymap) PathOf(fullkey uint64) string {
	// fullkey == 0 is a special value representing the root
	if fullkey == 0 {
		return ""
	}
	res, ok := km.Textmap[fullkey]
	if !ok {
		return ""
	}
	path := km.PathOf(res.Parent)
	return fmt.Sprintf("%s/%s", path, res.Text)
}

func NewKeymap() Keymap {
	return Keymap{
		Textmap: make(map[uint64]keymapEntry),
	}
}
