package main

import (
	"ax-distiller/lib/dnode"
	"encoding/gob"
	"os"
	"path/filepath"
	"strconv"
)

type TreeStore struct {
	dir    string
	keymap dnode.Keymap
}

func (s TreeStore) Save() (err error) {
	f, err := os.Create(filepath.Join(
		s.dir,
		"keymap.bin",
	))
	if err != nil {
		return
	}
	defer f.Close()

	encoder := gob.NewEncoder(f)
	err = encoder.Encode(s.keymap)
	return
}

func (s TreeStore) Keymap() dnode.Keymap {
	return s.keymap
}

func (s TreeStore) Add(domain string, ht dnode.HashTree) (err error) {
	if domain == "" {
		domain = "unknown.host"
	}

	err = os.MkdirAll(filepath.Join(s.dir, domain), 0777)
	if err != nil {
		return
	}
	f, err := os.Create(filepath.Join(
		s.dir,
		domain,
		strconv.FormatUint(ht.Root, 10),
	))
	if err != nil {
		return
	}
	defer f.Close()

	encoder := gob.NewEncoder(f)
	err = encoder.Encode(ht)
	return
}

func NewTreeStore(dir string) (ts TreeStore, err error) {
	err = os.MkdirAll(dir, 0777)
	if err != nil {
		return
	}
	ts = TreeStore{
		dir:    dir,
		keymap: dnode.NewKeymap(),
	}
	return
}
