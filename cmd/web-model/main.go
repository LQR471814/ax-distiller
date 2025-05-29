package main

import (
	"ax-distiller/lib/chrome"
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/url"
	"os"
	"os/signal"

	"github.com/chromedp/cdproto/accessibility"
	"github.com/chromedp/chromedp"
	"github.com/lmittmann/tint"
)

func main() {
	flag.Parse()
	u := flag.Arg(0)

	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level: slog.LevelDebug,
		}),
	))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	ctx, cancel, err := chrome.NewBrowser(ctx)
	defer cancel()
	if err != nil {
		fatalerr("create new browser", err)
	}

	if u != "" {
		parsed, err := url.Parse(u)
		if err != nil {
			log.Fatal(err)
		}
		start(ctx, parsed)
	}

	<-ctx.Done()
}

func fatalerr(message string, err error) {
	slog.Error(fmt.Sprintf("[main] %s", message), "err", err)
	os.Exit(1)
}

func start(ctx context.Context, parsed *url.URL) {
	// sets := make(map[accessibility.NodeID]ax.NodeSet)

	err := chromedp.Run(
		ctx,
		chromedp.Navigate(parsed.String()),
		accessibility.Enable(),
		chromedp.ActionFunc(func(pageCtx context.Context) error {
			ax := chrome.AX{
				PageCtx: pageCtx,
			}

			ax.Listen()
			return nil
		}),
	)
	if err != nil {
		fatalerr("run chromedp", err)
	}
}
