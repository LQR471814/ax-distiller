package main

import (
	"ax-distiller/lib/dnode"
	"encoding/gob"
	"iter"
	"log/slog"
	"os"
	"path/filepath"
)

type Loader struct {
	dir string
}

func (l Loader) Dir() string {
	return l.dir
}

func (l Loader) LoadKeymap() (km dnode.Keymap, err error) {
	kmfile, err := os.OpenFile(filepath.Join(l.dir, "keymap.bin"), os.O_RDONLY, 0400)
	if err != nil {
		return
	}
	defer kmfile.Close()

	decoder := gob.NewDecoder(kmfile)
	err = decoder.Decode(&km)
	return
}

type DomainGroup struct {
	Domain string
	Trees  iter.Seq[dnode.HashTree]
}

func (l Loader) readTree(path string) (ht dnode.HashTree, err error) {
	var f *os.File
	f, err = os.OpenFile(path, os.O_RDONLY, 0400)
	if err != nil {
		return
	}
	defer f.Close()

	decoder := gob.NewDecoder(f)
	decoder.Decode(&ht)

	return
}

func (l Loader) loadTrees(domain string) (trees iter.Seq[dnode.HashTree], err error) {
	entries, err := os.ReadDir(filepath.Join(l.dir, domain))
	if err != nil {
		return
	}
	trees = func(yield func(dnode.HashTree) bool) {
		for _, e := range entries {
			ht, err := l.readTree(filepath.Join(
				l.dir,
				domain,
				e.Name(),
			))
			if err != nil {
				slog.Warn("[loader] load tree", "err", err)
				continue
			}
			if !yield(ht) {
				break
			}
		}
	}
	return
}

func (l Loader) LoadDomains() (domains iter.Seq[DomainGroup], err error) {
	entries, err := os.ReadDir(l.dir)
	if err != nil {
		return
	}
	domains = func(yield func(DomainGroup) bool) {
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}

			trees, err := l.loadTrees(e.Name())
			if err != nil {
				slog.Warn("[loader] load domain", "err", err)
				continue
			}

			if !yield(DomainGroup{
				Domain: e.Name(),
				Trees:  trees,
			}) {
				break
			}
		}
	}
	return
}

func NewLoader(dir string) Loader {
	return Loader{
		dir: dir,
	}
}
