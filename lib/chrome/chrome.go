package chrome

import (
	"context"
	"os"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

func NewBrowser(ctx context.Context) (browser *rod.Browser, err error) {
	dataTemp := "./data/chrome-data"
	err = os.RemoveAll(dataTemp)
	if err != nil {
		return
	}
	err = os.Mkdir(dataTemp, 0777)
	if err != nil {
		return
	}

	launch := launcher.New().Bin("./data/chrome").
		Env("APPIMAGELAUNCHER_DISABLE=1").
		UserDataDir(dataTemp).
		Headless(false).
		Set("display", os.Getenv("DISPLAY")).
		Set("load-extension", "./data/ublock").
		Set("disable-extensions", "false").
		Set("disable-blink-features", "AutomationControlled").
		Set("disable-gpu", "true").
		Set("no-sandbox", "true").
		Set("no-default-browser-check", "true").
		Set("disable-remote-fonts", "true").
		Set("disable-background-networking", "true").
		Set("disable-dev-shm-usage", "true").
		Set("disable-sync", "true").
		Set("disable-translate", "true").
		Set("disable-default-apps", "true").
		Set("mute-audio", "true").
		Set("hide-scrollbars", "true")

	controlURL := launch.MustLaunch()
	browser = rod.New().ControlURL(controlURL).MustConnect()

	return
}

// BlockGraphics applies a routing configuration to the given page which blocks
// the loading of images and videos.
func BlockGraphics(page *rod.Page) {
	router := page.HijackRequests()

	for _, ext := range []string{
		"*.png",
		"*.jpg",
		"*.jpeg",
		"*.bmp",
		"*.gif",
		"*.webp",
		"*.heic",
		"*.heif",
		"*.tiff",
		"*.tif",
		"*.mp4",
		"*.avi",
		"*.mov",
		"*.mkv",
		"*.webm",
		"*.ts",
		"*.ogv",
	} {
		router.MustAdd(ext, blockImageOrMedia)
	}

	// since we are only hijacking a specific page, even using the "*" won't affect much of the performance
	go router.Run()
}

func blockImageOrMedia(ctx *rod.Hijack) {
	switch ctx.Request.Type() {
	case proto.NetworkResourceTypeImage, proto.NetworkResourceTypeMedia:
		ctx.Response.Fail(proto.NetworkErrorReasonBlockedByClient)
		return
	}
	ctx.ContinueRequest(&proto.FetchContinueRequest{})
}
