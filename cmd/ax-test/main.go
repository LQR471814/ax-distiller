package main

import (
	"ax-distiller/lib/chrome"
	"context"
	"encoding/xml"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/chromedp/cdproto/accessibility"
	"github.com/chromedp/cdproto/css"
	"github.com/chromedp/cdproto/log"
	"github.com/chromedp/cdproto/network"
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

	trees := []*chrome.AXNode{}

	err = chromedp.Run(
		ctx,
		accessibility.Enable(),
		network.Disable(),
		css.Disable(),
		log.Disable(),
		// dom.Disable(),
		chromedp.Navigate("https://flights.google.com"),
		chromedp.ActionFunc(func(pageCtx context.Context) error {
			ax := chrome.AX{PageCtx: pageCtx}
			t, err := ax.FetchFullAXTree()
			if err != nil {
				fatalerr("fetch full ax tree", err)
			}
			trees = append(trees, t)
			return nil
		}),
		chromedp.Click("input[aria-label='Where from?']"),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.ActionFunc(func(pageCtx context.Context) error {
			ax := chrome.AX{PageCtx: pageCtx}
			t, err := ax.FetchFullAXTree()
			if err != nil {
				fatalerr("fetch full ax tree", err)
			}
			trees = append(trees, t)
			return nil
		}),
		chromedp.Click("div[role=presentation]"),
		chromedp.Click("input[aria-label='Where to? ']"),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.ActionFunc(func(pageCtx context.Context) error {
			ax := chrome.AX{PageCtx: pageCtx}
			t, err := ax.FetchFullAXTree()
			if err != nil {
				fatalerr("fetch full ax tree", err)
			}
			trees = append(trees, t)
			return nil
		}),
		chromedp.Click("div[role=presentation]"),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.ActionFunc(func(pageCtx context.Context) error {
			ax := chrome.AX{PageCtx: pageCtx}
			t, err := ax.FetchFullAXTree()
			if err != nil {
				fatalerr("fetch full ax tree", err)
			}
			trees = append(trees, t)
			return nil
		}),
	)
	if err != nil {
		fatalerr("run chromedp", err)
	}

	for i, t := range trees {
		buff, err := xml.MarshalIndent(t, "", "  ")
		if err != nil {
			fatalerr("marshal", err)
		}
		err = os.WriteFile(fmt.Sprintf("%d.xml", i), buff, 0600)
		if err != nil {
			fatalerr("write", err)
		}
	}
}

func fatalerr(message string, err error) {
	slog.Error(fmt.Sprintf("[main] %s", message), "err", err)
	os.Exit(1)
}
