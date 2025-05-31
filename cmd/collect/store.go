package main

import (
	"ax-distiller/lib/chrome"
	"ax-distiller/lib/dnode"
	"encoding/gob"
	"net/url"
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

func (s TreeStore) Add(currentUrl *url.URL, tree *chrome.AXNode) (err error) {
	dn := dnode.FromAXTree(tree, s.keymap)
	ht := dnode.NewHashTree(dn)

	hostname := currentUrl.Hostname()
	if hostname == "" {
		hostname = "unknown.host"
	}

	err = os.MkdirAll(filepath.Join(s.dir, hostname), 0777)
	if err != nil {
		return
	}
	f, err := os.Create(filepath.Join(
		s.dir,
		hostname,
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
