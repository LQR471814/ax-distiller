package main

import (
	"ax-distiller/lib/dnode"
	"context"
	"log/slog"
	"runtime"
	"sync"
)

type Comparison struct {
	Self  dnode.HashTree
	Other dnode.HashTree
}

type AnalysisResult struct {
	Commonality dnode.HashTree
	Comparison  Comparison
}

type AnalysisOutput struct {
	lock  sync.Mutex
	order []uint64
	data  map[uint64]AnalysisResult
}

func NewAnalysisOutput() *AnalysisOutput {
	return &AnalysisOutput{
		data: make(map[uint64]AnalysisResult),
	}
}

func (a *AnalysisOutput) Write(res AnalysisResult) {
	a.lock.Lock()
	defer a.lock.Unlock()

	_, exists := a.data[res.Commonality.Root]
	if exists {
		return
	}
	a.data[res.Commonality.Root] = res
	a.order = append(a.order, res.Commonality.Root)
}

func (a *AnalysisOutput) Keys() []uint64 {
	a.lock.Lock()
	defer a.lock.Unlock()

	return a.order
}

func (a *AnalysisOutput) Get(hash uint64) AnalysisResult {
	a.lock.Lock()
	defer a.lock.Unlock()

	return a.data[hash]
}

type Analyzer struct {
	keymap dnode.Keymap
	jobs   chan Comparison
	output *AnalysisOutput
}

func (a *Analyzer) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case cmp := <-a.jobs:
			common := dnode.FindCommon(a.keymap, cmp.Self, cmp.Other, cmp.Self.Root)
			ht := dnode.NewHashTree(common)
			a.output.Write(AnalysisResult{
				Commonality: ht,
				Comparison:  cmp,
			})
		}
	}
}

func (a *Analyzer) StartWorkers(ctx context.Context) {
	slog.Info("[analyzer] starting workers...", "count", runtime.NumCPU())
	for range runtime.NumCPU() {
		go a.worker(ctx)
	}
}

func (a *Analyzer) Input() chan Comparison {
	return a.jobs
}

func (a *Analyzer) SetOutput(output *AnalysisOutput) {
	a.output = output
}

func NewAnalyzer(
	keymap dnode.Keymap,
) *Analyzer {
	return &Analyzer{
		keymap: keymap,
		jobs:   make(chan Comparison, runtime.NumCPU()*2),
	}
}
