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

	allocatorCtx, cancel1 := chromedp.NewExecAllocator(
		ctx,
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Env("APPIMAGELAUNCHER_DISABLE=1"),
			chromedp.ExecPath("./data/chrome"),
			chromedp.UserDataDir(dataTemp),
			chromedp.Flag("load-extension", "./data/ublock"),
			chromedp.Flag("headless", false),
			chromedp.Flag("disable-extensions", false),
			chromedp.Flag("disable-blink-features", "AutomationControlled"),
			// improve performance
			chromedp.DisableGPU,
			chromedp.NoSandbox,
			chromedp.NoDefaultBrowserCheck,
			chromedp.Flag("disable-remote-fonts", true),
			chromedp.Flag("disable-background-networking", true),
			chromedp.Flag("disable-dev-shm-usage", true),
			chromedp.Flag("disable-sync", true),
			chromedp.Flag("disable-translate", true),
			chromedp.Flag("disable-default-apps", true),
			chromedp.Flag("mute-audio", true),
			chromedp.Flag("hide-scrollbars", true),
		)...,
	)
	cdpCtx, cancel2 := chromedp.NewContext(
		allocatorCtx,
		// chromedp.WithBrowserOption(
		// 	chromedp.WithBrowserDebugf(func(s string, a ...any) {
		// 		var parsed struct {
		// 			Method string `json:"method"`
		// 		}
		// 		err = json.Unmarshal([]byte(a[0].([]uint8)), &parsed)
		// 		if err != nil {
		// 			slog.Error("parse msg", "msg", s, "err", err)
		// 			return
		// 		}
		// 		if strings.Contains(strings.ToLower(parsed.Method), "accessibility") {
		// 			log.Printf(s, a...)
		// 		}
		// 	}),
		// ),
	)

	cancel = func() {
		cancel2()
		cancel1()
	}

	err = chromedp.Run(cdpCtx)
	return
}
