package dnode

import "fmt"

type KeymapValue interface {
	Key() uint64
	String() string
}

type keymapEntry struct {
	text   string
	parent uint64
}

type Keymap struct {
	textmap map[uint64]keymapEntry
}

func (km Keymap) FullKey(parentFullKey uint64, v KeymapValue) (fullkey uint64) {
	fullkey = CompositeHash(parentFullKey, v.Key())
	if km.textmap == nil {
		return
	}
	km.textmap[fullkey] = keymapEntry{
		text:   v.String(),
		parent: parentFullKey,
	}
	return
}

func (km Keymap) StringOf(fullkey uint64) (string, bool) {
	res, ok := km.textmap[fullkey]
	return res.text, ok
}

func (km Keymap) PathOf(fullkey uint64) string {
	// fullkey == 0 is a special value representing the root
	if fullkey == 0 {
		return ""
	}
	res, ok := km.textmap[fullkey]
	if !ok {
		return ""
	}
	path := km.PathOf(res.parent)
	return fmt.Sprintf("%s/%s", path, res.text)
}

// NewKeymap creates a new keymap, if size < 0, then the keymap is considered disabled.
func NewKeymap(size int) Keymap {
	if size < 0 {
		return Keymap{}
	}
	return Keymap{
		textmap: make(map[uint64]keymapEntry, size),
	}
}
