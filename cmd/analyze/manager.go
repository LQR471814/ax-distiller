package main

import (
	"ax-distiller/lib/dnode"
	"context"
	"encoding/gob"
	"log/slog"
	"os"
	"path/filepath"
)

type CommonNode struct {
	Tree        dnode.HashTree
	NextSibling *CommonNode
	FirstChild  *CommonNode
}

type CommonTree struct {
	FromHash map[uint64]*CommonNode
	Root     uint64
}

func (t CommonTree) Add(parent, child dnode.HashTree) {
	if child.Root == t.Root {
		t.Root = parent.Root
	}

	_, childExists := t.FromHash[child.Root]
	if childExists {
		return
	}

	existingParent, parentExists := t.FromHash[parent.Root]
	if !parentExists {
		existingParent = &CommonNode{
			Tree: parent,
		}
	}

	cur := existingParent.FirstChild
	for cur != nil {
		if cur.NextSibling == nil {
			break
		}
	}
	if cur != nil {
		cur.NextSibling = &CommonNode{
			Tree: child,
		}
	}
}

func NewCommonTree() CommonTree {
	return CommonTree{
		FromHash: make(map[uint64]*CommonNode),
	}
}

type Manager struct {
	analyzer *Analyzer
	loader   Loader
}

func (m Manager) analyzeDomain(ctx context.Context, d DomainGroup) CommonTree {
	tree := NewCommonTree()

	next := NewAnalysisOutput()
	m.analyzer.SetOutput(next)

	var queued []dnode.HashTree
	for t := range d.Trees {
		for _, other := range queued {
			// slog.Info("cmp", "self", len(t.FromHash), "other", len(other.FromHash))
			m.analyzer.Input() <- Comparison{
				Self:  t,
				Other: other,
			}
		}
		queued = append(queued, t)
	}

	for _, key := range next.Keys() {
		res := next.Get(key)
		tree.Add(res.Commonality, res.Comparison.Self)
		tree.Add(res.Commonality, res.Comparison.Other)
	}

	slog.Info("keys", "k", next.Keys())

	depth := 0
	for len(next.Keys()) >= 2 {
		depth++
		slog.Info("[manager] common tree layer", "layer", depth)

		current := next
		next = NewAnalysisOutput()
		m.analyzer.SetOutput(next)

		count := 0
		for i := 1; i < len(current.Keys()); i++ {
			slog.Info("comparisons", "count", count)

			self := current.Get(current.Keys()[i])
			for j := i - 1; j >= 0; j-- {
				// if count >= 10 {
				// 	slog.Info("output", "keys", len(next.Keys()))
				// 	<-ctx.Done()
				// 	return tree
				// }

				other := current.Get(current.Keys()[j])
				m.analyzer.Input() <- Comparison{
					Self:  self.Commonality,
					Other: other.Commonality,
				}
				count++
			}
		}

		for _, key := range next.Keys() {
			res := next.Get(key)
			tree.Add(res.Commonality, res.Comparison.Self)
			tree.Add(res.Commonality, res.Comparison.Other)
		}
	}

	return tree
}

func (m Manager) writeAnalysis(domain string, ct CommonTree) (err error) {
	f, err := os.Create(filepath.Join(m.loader.Dir(), domain, "analysis.bin"))
	if err != nil {
		return
	}
	defer f.Close()
	encoder := gob.NewEncoder(f)
	err = encoder.Encode(ct)
	return
}

func (m Manager) Run(ctx context.Context) (err error) {
	domains, err := m.loader.LoadDomains()
	if err != nil {
		return
	}

	m.analyzer.StartWorkers(ctx)

	for d := range domains {
		slog.Info("[manager] running comparisons", "domain", d.Domain)

		commonTree := m.analyzeDomain(ctx, d)
		err = m.writeAnalysis(d.Domain, commonTree)
		if err != nil {
			slog.Error("[manager] write analysis", "err", err)
		}
	}
	return
}

func NewManager(loader Loader) (man Manager, err error) {
	km, err := loader.LoadKeymap()
	if err != nil {
		return
	}
	man = Manager{
		loader:   loader,
		analyzer: NewAnalyzer(km),
	}
	return
}
