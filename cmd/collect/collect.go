package main

import (
	"ax-distiller/lib/chrome/ax"
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
)

type Collector struct {
	tabctx       context.Context
	actionQueues map[uint64][]uint32
	store        TreeStore
}

func (c Collector) fetchActionNodeIDs(actions []Action) (nodeIds []cdp.NodeID, err error) {
	backendToFrontendID := make(map[uint64]cdp.NodeID)
	actionBackendIDs := make([]cdp.BackendNodeID, len(actions))
	for i, elem := range actions {
		actionBackendIDs[i] = cdp.BackendNodeID(elem.Node().DomNodeId)
	}
	nodeIds, err = dom.PushNodesByBackendIDsToFrontend(actionBackendIDs).Do(c.tabctx)
	if err != nil {
		return
	}
	for i, elem := range nodeIds {
		backendToFrontendID[uint64(actionBackendIDs[i])] = elem
	}
	return
}

func (c Collector) findAndTakeAction(currentURL *url.URL, tree *ax.Node, websiteHash uint64) (err error) {
	var actions []Action
	c.findActions(tree, &actions)

	nodeIDs, err := c.fetchActionNodeIDs(actions)
	if err != nil {
		return
	}

	queue := c.actionQueues[websiteHash]
	if len(queue) == 0 {
		actionIDs := make([]uint32, len(actions))
		for i := range actions {
			actionIDs[i] = uint32(i)
		}
		c.actionQueues[websiteHash] = actionIDs
		queue = actionIDs
	}

	actionIdx := queue[0]
	rotated := make([]uint32, len(queue))
	copy(rotated, queue[1:])
	rotated[len(rotated)-1] = actionIdx
	c.actionQueues[websiteHash] = rotated

	err = chromedp.Run(c.tabctx, chromedp.ActionFunc(func(ctx context.Context) error {
		return actions[actionIdx].Do(ctx, nodeIDs[actionIdx])
	}))
	return
}

func (c Collector) handleDomChange() (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("handle dom change: %w", err)
		}
	}()

	var currentURL string
	var tree *ax.Node

	err = chromedp.Run(
		c.tabctx,
		chromedp.Evaluate(`window.location.href`, &currentURL),
		chromedp.ActionFunc(func(ctx context.Context) (err error) {
			ax := ax.API{PageCtx: ctx}
			tree, err = ax.FetchFullTree()
			return
		}),
	)
	if err != nil {
		return
	}

	parsed, err := url.Parse(currentURL)
	if err != nil {
		return
	}

	websiteHash, err := c.store.Add(parsed, tree)
	if err != nil {
		return
	}
	c.findAndTakeAction(parsed, tree, websiteHash)

	slog.Info("[collect] fetched and stored ax tree")
	return
}

func (c Collector) worker() {
	domUpdate := make(chan struct{}, 16)
	// requestsUpdate := make(chan struct{}, 16)
	// activeRequests := atomic.Int32{}

	go func() {
		chromedp.ListenTarget(c.tabctx, func(ev any) {
			switch ev.(type) {
			// case *network.EventRequestWillBeSent:
			// 	go func() {
			// 		activeRequests.Add(1)
			// 		requestsUpdate <- struct{}{}
			// 	}()
			// case *network.EventResponseReceived:
			// 	go func() {
			// 		activeRequests.Add(-1)
			// 		requestsUpdate <- struct{}{}
			// 	}()
			case *dom.EventDocumentUpdated,
				*dom.EventChildNodeCountUpdated,
				*dom.EventCharacterDataModified:
				go func() {
					domUpdate <- struct{}{}
				}()
			}
		})
	}()

	timer := time.NewTimer(time.Second)
	for {
		select {
		case <-c.tabctx.Done():
			return
		case <-domUpdate:
			timer.Reset(time.Second)
		case <-timer.C:
			// slog.Info("[collect] waiting for the network to settle...")
			// if activeRequests.Load() > 1 {
			// 	for {
			// 		select {
			// 		case <-c.tabctx.Done():
			// 			return
			// 		case <-requestsUpdate:
			// 			if activeRequests.Load() <= 1 {
			// 				break
			// 			}
			// 		}
			// 	}
			// }

			go func() {
				err := c.handleDomChange()
				if err != nil {
					slog.Error("[collect] handle dom change", "err", err)
				}
			}()
		}
	}
}

func NewCollector(tabctx context.Context, store TreeStore) Collector {
	return Collector{
		tabctx: tabctx,
		store:  store,
	}
}
