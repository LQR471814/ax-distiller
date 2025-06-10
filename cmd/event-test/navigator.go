package main

import (
	"ax-distiller/lib/chrome"
	"context"
	"log/slog"
	"reflect"

	"github.com/bytedance/sonic"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

type Navigator struct {
	page          *rod.Page
	tree          AXTree
	root          proto.DOMBackendNodeID
	backendLookup map[proto.DOMNodeID]proto.DOMBackendNodeID
	jobCancelers  []func()
	initiating    bool
}

func NewNavigator(page *rod.Page) *Navigator {
	return &Navigator{
		page:          page,
		tree:          NewAXTree(),
		backendLookup: make(map[proto.DOMNodeID]proto.DOMBackendNodeID),
	}
}

const (
	// event_page_lifecycleEvent = "Page.lifecycleEvent"
	// event_page_frameNavigated = "Page.frameNavigated"
	// event_dom_documentUpdated   = "DOM.documentUpdated"
	event_dom_childNodeRemoved  = "DOM.childNodeRemoved"
	event_dom_childNodeInserted = "DOM.childNodeInserted"
	event_dom_setChildNodes     = "DOM.setChildNodes"
	event_ax_loadComplete       = "Accessibility.loadComplete"
)

func (n *Navigator) HandleEvents() {
	err := sonic.Pretouch(reflect.TypeOf(proto.AccessibilityAXNode{}))
	if err != nil {
		panic(err)
	}
	err = sonic.Pretouch(reflect.TypeOf(proto.DOMNode{}))
	if err != nil {
		panic(err)
	}

	for msg := range n.page.Event() {
		// data is not public... so must use reflect
		buff := reflect.ValueOf(msg).Elem().FieldByName("data").Bytes()

		var err error

		switch msg.Method {
		// case event_page_lifecycleEvent,
		// 	event_page_frameNavigated:
		// case event_dom_documentUpdated:
		// 	n.handleNavigation()

		case event_dom_childNodeRemoved:
			var out proto.DOMChildNodeRemoved
			err = sonic.Unmarshal(buff, &out)
			if err != nil {
				panic(err)
			}
			n.handleDOMRemoval(&out)

		case event_dom_childNodeInserted:
			var out proto.DOMChildNodeInserted
			err = sonic.Unmarshal(buff, &out)
			if err != nil {
				panic(err)
			}
			err = n.handleDOMInsert(&out)

		case event_dom_setChildNodes:
			var out proto.DOMSetChildNodes
			err = sonic.Unmarshal(buff, &out)
			if err != nil {
				panic(err)
			}
			n.handleDOMSetChildNodes(&out)

		case event_ax_loadComplete:
			var out proto.AccessibilityLoadComplete
			err = sonic.Unmarshal(buff, &out)
			if err != nil {
				panic(err)
			}
			err = n.handleAXLoadComplete(&out)
		}

		if err != nil {
			slog.Error("[event] handle event", "err", err)
		}
	}
}

func (n *Navigator) handleNavigation() (err error) {
	if n.initiating {
		return
	}
	n.initiating = true
	defer func() {
		n.initiating = false
	}()

	for _, cancel := range n.jobCancelers {
		cancel()
	}

	ctx := n.newContext()
	res, err := chrome.CDPCommand(ctx, n.page, proto.DOMGetDocument{})
	if err != nil {
		return
	}
	err = n.loadFullPage(ctx, res.Root.BackendNodeID)
	if err != nil {
		return
	}

	slog.Info("[nav] handle navigation")
	return
}

func (n *Navigator) handleDOMInsert(e *proto.DOMChildNodeInserted) (err error) {
	ctx := n.newContext()
	err = n.tree.AddSubtree(ctx, n.page, e.Node.BackendNodeID)
	if err != nil {
		return
	}

	depth := -1
	// this will trigger a series of DOM.setChildNodes events, which will provide the
	// NodeID -> BackendNodeID mappings for the rest of the subtree
	err = chrome.CDPProcedure(ctx, n.page, proto.DOMRequestChildNodes{
		NodeID: e.Node.NodeID,
		Depth:  &depth,
	})
	if err != nil {
		return
	}
	n.backendLookup[e.Node.NodeID] = e.Node.BackendNodeID

	slog.Info("[nav] DOM child inserted", "id", e.Node.BackendNodeID)
	return
}

func (n *Navigator) handleDOMRemoval(e *proto.DOMChildNodeRemoved) {
	parentID := n.backendLookup[e.ParentNodeID]
	nodeID := n.backendLookup[e.NodeID]
	slog.Info(
		"[nav] DOM child removed",
		"id", nodeID,
		"parent", parentID,
	)
	n.tree.RemoveSubtree(parentID, nodeID)
	return
}

func (n *Navigator) handleDOMSetChildNodes(e *proto.DOMSetChildNodes) {
	for _, node := range e.Nodes {
		n.backendLookup[node.NodeID] = node.BackendNodeID
	}
}

func (n *Navigator) loadFullPage(ctx context.Context, root proto.DOMBackendNodeID) (err error) {
	n.tree.RemoveSubtree(0, n.root)
	n.root = root

	err = n.tree.AddSubtree(ctx, n.page, root)
	if err != nil {
		return
	}

	depth := -1
	docRes, err := chrome.CDPCommand(ctx, n.page, proto.DOMGetDocument{Depth: &depth})
	if err != nil {
		return
	}
	queue := []*proto.DOMNode{docRes.Root}
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		n.backendLookup[node.NodeID] = node.BackendNodeID
		for _, child := range node.Children {
			queue = append(queue, child)
		}
	}
	return
}

func (n *Navigator) handleAXLoadComplete(e *proto.AccessibilityLoadComplete) (err error) {
	slog.Info("[nav] AX load complete, loading...")
	err = n.loadFullPage(n.newContext(), e.Root.BackendDOMNodeID)
	slog.Info("[nav] AX load complete, load finished")
	return
}

func (n *Navigator) newContext() context.Context {
	ctx, cancel := context.WithCancel(n.page.GetContext())
	n.jobCancelers = append(n.jobCancelers, cancel)
	return ctx
}
