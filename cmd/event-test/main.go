package main

import (
	"ax-distiller/lib/chrome"
	"ax-distiller/lib/chrome/ax"
	"context"
	"log/slog"
	"os"
	"os/signal"
	"reflect"

	"github.com/chromedp/cdproto/accessibility"
	"github.com/chromedp/cdproto/css"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/log"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

func main() {
	err := start()
	if err != nil {
		slog.Error("[main] create new browser", "err", err)
		os.Exit(1)
	}
}

func start() (err error) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	tabctx, cancel, err := chrome.NewBrowser(ctx)
	defer cancel()

	err = chromedp.Run(
		tabctx,
		network.Disable(),
		log.Disable(),
		css.Disable(),
		accessibility.Enable(),
	)
	if err != nil {
		return
	}

	// api := ax.API{
	// 	PageCtx: tabctx,
	// }
	// api.Listen()

	chromedp.ListenTarget(tabctx, func(ev any) {
		switch event := ev.(type) {
		case *accessibility.EventLoadComplete:
			go func() {
				err := chromedp.Run(tabctx, chromedp.ActionFunc(func(ctx context.Context) (err error) {
					api := ax.API{PageCtx: ctx}
					return api.SubscribeFullTree(event.Root.NodeID.String())
				}))
				if err != nil {
					slog.Error("[main] get full ax tree", "err", err)
					return
				}
				slog.Info("[main] fetch full tree")
			}()
			slog.Info("[main] accessibility load complete")
		case *accessibility.EventNodesUpdated:
			slog.Info("[main] nodes updated", "nodes", event.Nodes)
		case *dom.EventChildNodeCountUpdated:
			slog.Info("[main] child node count updated", "node", event.NodeID)
		case *dom.EventChildNodeInserted:
			slog.Info("[main] child node inserted", "node", event.Node.NodeID, "parent", event.ParentNodeID)
		default:
			slog.Info("[main] event received", "event", reflect.TypeOf(ev).String())
		}
	})

	<-tabctx.Done()

	return
}
