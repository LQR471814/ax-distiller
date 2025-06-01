package main

import (
	"ax-distiller/lib/chrome"
	"ax-distiller/lib/chrome/ax"
	"context"
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/url"
	"os"
	"os/signal"

	"github.com/chromedp/cdproto/accessibility"
	"github.com/chromedp/chromedp"
)

func main() {
	flag.Parse()
	u := flag.Arg(0)

	if u == "" {
		fatalerr(
			"parse cli args",
			fmt.Errorf("you must provide a url to dump as the first command line argument"),
		)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	ctx, cancel, err := chrome.NewBrowser(ctx)
	defer cancel()
	if err != nil {
		fatalerr("create new browser", err)
	}

	parsed, err := url.Parse(u)
	if err != nil {
		fatalerr("parse url", err)
	}

	var tree *ax.Node

	err = chromedp.Run(
		ctx,
		accessibility.Enable(),
		chromedp.Navigate(parsed.String()),
		chromedp.ActionFunc(func(pageCtx context.Context) error {
			ax := ax.API{
				PageCtx: pageCtx,
			}
			tree, err = ax.FetchFullTree()
			return err
		}),
	)
	if err != nil {
		fatalerr("run chromedp", err)
	}

	tree = ax.FilterWhitespace(tree)
	if tree == nil {
		fatalerr("empty tree", fmt.Errorf("tree is nil"))
	}

	serialized, err := xml.MarshalIndent(tree, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(serialized))
}

func fatalerr(message string, err error) {
	slog.Error(fmt.Sprintf("[main] %s", message), "err", err)
	os.Exit(1)
}
