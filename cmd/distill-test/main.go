package main

import (
	"ax-distiller/lib/axextract"
	"ax-distiller/lib/markdown"
	"fmt"
	"log"
	"net/url"
	"os"
	"sync"
)

func main() {
	urls := []string{
		// "https://music.youtube.com/channel/UCWuBpAte4YHm_oELpzoM2qg",
		// "https://en.wikipedia.org/wiki/Quantum_mechanics",
		// "https://www.npr.org/2024/03/29/1198909601/lost-animals-moles-rats-being-rediscovered",
		// "https://pkg.go.dev/github.com/chromedp/chromedp#section-readme",
		// "https://www.w3schools.com/tags/ref_colornames.asp",
		// "https://code.whatever.social/questions/53692326/convert-relative-to-absolute-urls-in-go",
		// "https://safereddit.com/r/golang/comments/181ebuq/anybody_who_has_used_chromedp_similar_libraries",
		"https://github.com/LQR471814",
	}

	navigator, err := axextract.NewNavigator()
	if err != nil {
		log.Fatal(err)
	}

	wg := sync.WaitGroup{}
	for _, u := range urls {
		wg.Add(1)
		go func(u string) {
			parsed, err := url.Parse(u)
			if err != nil {
				log.Fatal(err)
			}

			page, err := navigator.Navigate(parsed)
			if err != nil {
				log.Fatal(err)
			}

			html := pageToHtml(page)
			err = os.WriteFile(fmt.Sprintf("out_%s.html", parsed.Host), []byte(html), 0777)
			if err != nil {
				log.Fatal(err)
			}

			md := pageToMd(page)
			mdout := markdown.Render(md)
			err = os.WriteFile(fmt.Sprintf("out_%s.md", parsed.Host), []byte(mdout), 0777)
			if err != nil {
				log.Fatal(err)
			}

			pageDump := dumpPageAx(page)
			err = os.WriteFile(fmt.Sprintf("dump_%s.txt", parsed.Host), []byte(pageDump), 0777)
			if err != nil {
				log.Fatal(err)
			}

			wg.Done()
		}(u)
	}

	wg.Wait()
}
