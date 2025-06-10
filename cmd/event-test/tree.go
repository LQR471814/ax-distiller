package main

import (
	"ax-distiller/lib/chrome"
	"context"
	"fmt"
	"log/slog"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

type AXNode struct {
	Underlying  *proto.AccessibilityAXNode
	FirstChild  *AXNode
	NextSibling *AXNode
}

type AXTree struct {
	fromBackendID map[proto.DOMBackendNodeID]*AXNode
}

func NewAXTree() AXTree {
	return AXTree{
		fromBackendID: make(map[proto.DOMBackendNodeID]*AXNode),
	}
}

/*
- when subtree is requested, every node's BackendNodeId must be recorded. (regardless of ignorance or not)
- when DOM node is inserted, corresponding AX tree under the given BackendNodeID is fetched
- when DOM node is removed, parent is updated by removing the child
- post-processing logic that converts tree or subtree into just the non-ignored nodes
*/

func (t AXTree) AddSubtree(ctx context.Context, page *rod.Page, nodeID proto.DOMBackendNodeID) (err error) {
	fetchRelatives := false
	rootRes, err := chrome.CDPCommand(ctx, page, proto.AccessibilityGetPartialAXTree{
		BackendNodeID:  nodeID,
		FetchRelatives: &fetchRelatives,
	})
	if err != nil {
		return
	}
	if len(rootRes.Nodes) != 1 {
		panic(fmt.Errorf("assert root.Nodes == 1: %d", len(rootRes.Nodes)))
	}
	root := rootRes.Nodes[0]

	currentNode := &AXNode{Underlying: root}
	t.fromBackendID[root.BackendDOMNodeID] = currentNode

	type queryJob struct {
		parent *AXNode
		node   *proto.AccessibilityAXNode
	}

	queue := []queryJob{{
		parent: nil,
		node:   root,
	}}
	seenmap := make(map[proto.AccessibilityAXNodeID]struct{})

	for len(queue) > 0 {
		select {
		case <-ctx.Done():
			return
		default:
		}

		job := queue[0]
		queue = queue[1:]
		seenmap[job.node.NodeID] = struct{}{}

		if job.parent != nil {
			currentNode = &AXNode{
				Underlying:  job.node,
				NextSibling: job.parent.FirstChild,
			}
			job.parent.FirstChild = currentNode
		}

		// if NodeID is a negative number, we know it is a leaf
		if job.node.NodeID[0] == '-' {
			continue
		}

		// this always causes "invalid ID for some reason"
		if job.node.NodeID == "0" {
			continue
		}

		var res proto.AccessibilityGetChildAXNodesResult
		res, err = chrome.CDPCommand(
			ctx,
			page,
			proto.AccessibilityGetChildAXNodes{ID: job.node.NodeID},
		)
		if err != nil {
			slog.Warn("[axtree] get child ax nodes", "id", job.node.NodeID, "role", job.node.Role, "err", err)
			err = nil
			continue
		}

		for _, child := range res.Nodes {
			_, seen := seenmap[child.NodeID]
			if seen {
				continue
			}
			queue = append(queue, queryJob{
				parent: currentNode,
				node:   child,
			})
		}
	}

	return
}

func (t AXTree) RemoveSubtree(parentID, nodeID proto.DOMBackendNodeID) {
	par, exists := t.fromBackendID[parentID]
	if exists {
		// remove from parent references if parent exists
		child := par.FirstChild
		var prev *AXNode
		for child != nil {
			if child.Underlying.BackendDOMNodeID == nodeID {
				next := child.NextSibling
				if prev != nil {
					prev.NextSibling = next
				} else {
					par.FirstChild = next
				}
				break
			}
			prev = child
			child = child.NextSibling
		}
	}

	root, exists := t.fromBackendID[nodeID]
	if !exists {
		return
	}

	queue := []*AXNode{root}
	for len(queue) > 0 {
		first := queue[0]
		queue = queue[1:]

		delete(t.fromBackendID, first.Underlying.BackendDOMNodeID)

		child := first.FirstChild
		for child != nil {
			queue = append(queue, child)
			child = child.NextSibling
		}
	}
}
