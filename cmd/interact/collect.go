package main

import (
	"ax-distiller/lib/chrome"
	"context"
	"log/slog"
	"net/url"
	"reflect"
	"time"

	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
)

type Collector struct {
	tabctx context.Context
	store  TreeStore
}

func (c Collector) worker() {
	domupdate := make(chan struct{}, 16)

	go func() {
		chromedp.ListenTarget(c.tabctx, func(ev any) {
			switch ev.(type) {
			case *dom.EventDocumentUpdated,
				*dom.EventChildNodeCountUpdated,
				*dom.EventCharacterDataModified:
				slog.Info("[collect] dom updated", "type", reflect.TypeOf(ev).String())
				go func() {
					domupdate <- struct{}{}
				}()
			}
		})
	}()

	timer := time.NewTimer(time.Second)
	for {
		select {
		case <-c.tabctx.Done():
			return
		case <-domupdate:
			timer.Reset(time.Second)
		case <-timer.C:
			go func() {
				var currentURL string

				slog.Info("[collect] begin fetch ax tree")

				err := chromedp.Run(
					c.tabctx,
					chromedp.Evaluate(`window.location.href`, &currentURL),
					chromedp.ActionFunc(func(ctx context.Context) (err error) {
						ax := chrome.AX{PageCtx: ctx}
						t, err := ax.FetchFullAXTree()
						if err != nil {
							return
						}

						parsed, err := url.Parse(currentURL)
						if err != nil {
							return
						}

						err = c.store.Add(parsed, t)
						if err != nil {
							return
						}
						return
					}),
				)
				if err != nil {
					slog.Warn("[collect] fetch ax tree", "err", err)
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
