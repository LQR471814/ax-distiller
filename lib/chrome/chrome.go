package chrome

import (
	"context"
	"os"

	"github.com/chromedp/chromedp"
)

func NewBrowser(ctx context.Context) (cdpCtx context.Context, cancel func(), err error) {
	dataTemp := "./data/chrome-data"
	err = os.RemoveAll(dataTemp)
	if err != nil {
		return
	}
	err = os.Mkdir(dataTemp, 0777)
	if err != nil {
		return
	}

	allocatorCtx, _ := chromedp.NewExecAllocator(
		ctx,
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Env("APPIMAGELAUNCHER_DISABLE=1"),
			chromedp.ExecPath("./data/thorium-browser"),
			chromedp.UserDataDir(dataTemp),
			chromedp.Flag("load-extension", "./data/ublock"),
			chromedp.Flag("headless", false),
			chromedp.Flag("disable-extensions", false),
			chromedp.Flag("disable-remote-fonts", true),
			chromedp.Flag("disable-blink-features", "AutomationControlled"),
			chromedp.Flag("no-sandbox", true),
		)...,
	)
	cdpCtx, cancel = chromedp.NewContext(allocatorCtx)

	err = chromedp.Run(cdpCtx)
	return
}
