package main

import (
	"ax-distiller/lib/chrome"
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/lmittmann/tint"
)

func main() {
	logger := slog.New(tint.NewHandler(os.Stderr, &tint.Options{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	err := start()
	if err != nil {
		slog.Error("[main] create new browser", "err", err)
		os.Exit(1)
	}
}

func start() (err error) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	browser, err := chrome.NewBrowser(ctx)
	if err != nil {
		return
	}

	pool := rod.NewPagePool(3)
	defer pool.Cleanup(func(p *rod.Page) { p.MustClose() })

	getPage := func() (page *rod.Page, err error) {
		page, err = pool.Get(func() (*rod.Page, error) {
			return browser.MustIncognito().MustPage(), nil
		})
		if err != nil {
			return
		}

		err = chrome.CDPProcedure(nil, page, proto.NetworkDisable{})
		if err != nil {
			return
		}
		err = chrome.CDPProcedure(nil, page, proto.LogDisable{})
		if err != nil {
			return
		}
		err = chrome.CDPProcedure(nil, page, proto.DOMEnable{})
		if err != nil {
			return
		}
		err = chrome.CDPProcedure(nil, page, proto.CSSEnable{})
		if err != nil {
			return
		}
		err = chrome.CDPProcedure(nil, page, proto.AccessibilityEnable{})
		if err != nil {
			return
		}
		chrome.BlockGraphics(page)

		return
	}

	page, err := getPage()
	if err != nil {
		return
	}

	nav := NewNavigator(page)
	go nav.HandleEvents()

	// go page.EachEvent(
	// 	func(e *proto.AccessibilityNodesUpdated) {
	// 		var filtered []*proto.AccessibilityAXNode
	// 		for _, n := range e.Nodes {
	// 			if n.Role.Value.String() == "RootWebArea" {
	// 				continue
	// 			}
	// 			filtered = append(filtered, n)
	// 		}
	// 		if len(filtered) == 0 {
	// 			return
	// 		}
	//
	// 		type node struct {
	// 			role       string
	// 			backend_id int
	// 		}
	// 		nodes := make([]node, len(filtered))
	// 		for i, n := range filtered {
	// 			nodes[i] = node{
	// 				role:       n.Role.Value.String(),
	// 				backend_id: int(n.BackendDOMNodeID),
	// 			}
	// 		}
	// 		slog.Info("[main] nodes updated", "nodes", nodes)
	//
	// 		for _, n := range filtered {
	// 			subcSubtree(n.NodeID, page)
	// 		}
	// 	},
	// )()

	page.MustNavigate("http://localhost:8000")

	<-ctx.Done()

	return
}
