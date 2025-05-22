package main

import (
	"ax-distiller/lib/chrome"
	"ax-distiller/lib/markdown"
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/url"
	"os"
	"os/signal"
	"sync"

	"github.com/chromedp/cdproto/accessibility"
	"github.com/chromedp/chromedp"
)

func main() {
	urls := []string{
		"https://music.youtube.com/channel/UCWuBpAte4YHm_oELpzoM2qg",
		"https://en.wikipedia.org/wiki/Quantum_mechanics",
		"https://www.npr.org/2024/03/29/1198909601/lost-animals-moles-rats-being-rediscovered",
		"https://pkg.go.dev/github.com/chromedp/chromedp#section-readme",
		"https://www.w3schools.com/tags/ref_colornames.asp",
		"https://code.whatever.social/questions/53692326/convert-relative-to-absolute-urls-in-go",
		"https://safereddit.com/r/golang/comments/181ebuq/anybody_who_has_used_chromedp_similar_libraries",
		"https://github.com/LQR471814",
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	ctx, cancel, err := chrome.NewBrowser(ctx)
	defer cancel()
	if err != nil {
		fatalerr("create new browser", err)
	}

	wg := sync.WaitGroup{}
	for _, u := range urls {
		wg.Add(1)
		go func(u string) {
			parsed, err := url.Parse(u)
			if err != nil {
				log.Fatal(err)
			}

			var tree chrome.AXNode

			err = chromedp.Run(
				ctx,
				accessibility.Enable(),
				chromedp.Navigate(parsed.String()),
				chromedp.ActionFunc(func(pageCtx context.Context) error {
					ax := chrome.AX{
						PageCtx: pageCtx,
					}
					tree, err = ax.FetchFullAXTree()
					return err
				}),
			)

			onlyTextTree, _ := onlyTextContent(tree)
			noTextTree := noTextContent(tree)

			tree = onlyTextTree
			md := pageToMd(page)
			mdout := markdown.Render(md)
			err = os.WriteFile(fmt.Sprintf("out_textonly_%s.md", parsed.Host), []byte(mdout), 0777)
			if err != nil {
				log.Fatal(err)
			}

			page.Tree = noTextTree
			md = pageToMd(page)
			mdout = markdown.Render(md)
			err = os.WriteFile(fmt.Sprintf("out_notext_%s.md", parsed.Host), []byte(mdout), 0777)
			if err != nil {
				log.Fatal(err)
			}

			// html := pageToHtml(page)
			// err = os.WriteFile(fmt.Sprintf("out_%s.html", parsed.Host), []byte(html), 0777)
			// if err != nil {
			// 	log.Fatal(err)
			// }
			//

			// pageDump := dumpPageAx(page)
			// err = os.WriteFile(fmt.Sprintf("dump_%s.txt", parsed.Host), []byte(pageDump), 0777)
			// if err != nil {
			// 	log.Fatal(err)
			// }

			wg.Done()
		}(u)
	}

	wg.Wait()
}

func fatalerr(message string, err error) {
	slog.Error(fmt.Sprintf("[main] %s", message), "err", err)
	os.Exit(1)
}
