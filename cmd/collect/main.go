package main

import (
	"ax-distiller/lib/chrome"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"github.com/chromedp/cdproto/accessibility"
	"github.com/chromedp/cdproto/css"
	"github.com/chromedp/cdproto/log"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	tabctx, cancel, err := chrome.NewBrowser(ctx)
	defer cancel()
	if err != nil {
		fatalerr("create new browser", err)
	}

	err = chromedp.Run(
		tabctx,
		network.Disable(),
		css.Disable(),
		log.Disable(),
		accessibility.Enable(),
	)
	if err != nil {
		fatalerr("enable capabilities", err)
	}

	ts, err := NewTreeStore("data/treestore")
	if err != nil {
		fatalerr("create treestore", err)
	}
	defer func() {
		err := ts.Save()
		if err != nil {
			fatalerr("save treestore", err)
		}
	}()

	collector := NewCollector(tabctx, ts)
	go collector.worker()

	<-tabctx.Done()
}

func fatalerr(message string, err error) {
	slog.Error(fmt.Sprintf("[main] %s", message), "err", err)
	os.Exit(1)
}
