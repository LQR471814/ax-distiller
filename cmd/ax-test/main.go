package main

import (
	"ax-distiller/lib/chrome"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"reflect"
	"time"

	"github.com/chromedp/cdproto/accessibility"
	"github.com/chromedp/cdproto/css"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/log"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	ctx, cancel, err := chrome.NewBrowser(ctx)
	defer cancel()
	if err != nil {
		fatalerr("create new browser", err)
	}

	err = chromedp.Run(
		ctx,
		accessibility.Enable(),
		network.Disable(),
		page.Disable(),
		css.Disable(),
		log.Disable(),
		dom.Disable(),
		runtime.Disable(),
		chromedp.Navigate("https://flights.google.com"),
		chromedp.ActionFunc(func(pageCtx context.Context) error {
			// ax := chrome.AX{
			// 	PageCtx: pageCtx,
			// }

			// _, err = ax.FetchFullAXTree()
			// if err != nil {
			// 	fatalerr("fetch full ax tree", err)
			// }

			_, err = accessibility.QueryAXTree().WithRole("combobox").Do(pageCtx)
			if err != nil {
				panic(err)
			}

			chromedp.ListenTarget(pageCtx, func(ev any) {
				slog.Info(fmt.Sprintf("[event] %s", reflect.TypeOf(ev).String()))
				switch typed := ev.(type) {
				case *accessibility.EventLoadComplete:
					slog.Info("[event] load complete", "id", typed.Root.NodeID)
				case *accessibility.EventNodesUpdated:
					roles := make([]string, len(typed.Nodes))
					for i, n := range typed.Nodes {
						roles[i] = n.Role.Value.String()
					}
					slog.Info("[event] nodes updated", "roles", roles)
				}
			})
			return nil
		}),
		chromedp.Sleep(20*time.Second),
	)
	if err != nil {
		fatalerr("run chromedp", err)
	}
}

func fatalerr(message string, err error) {
	slog.Error(fmt.Sprintf("[main] %s", message), "err", err)
	os.Exit(1)
}
