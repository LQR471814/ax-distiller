package main

import (
	"ax-distiller/lib/chrome/ax"
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
)

type Collector struct {
	tabctx context.Context
	store  TreeStore
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
	err = c.store.Add(parsed, tree)
	if err != nil {
		return
	}

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

			go c.handleDomChange()
		}
	}
}

func NewCollector(tabctx context.Context, store TreeStore) Collector {
	return Collector{
		tabctx: tabctx,
		store:  store,
	}
}
