package ax

import (
	"context"
	"net/url"

	"github.com/chromedp/cdproto/accessibility"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

func (n Navigator) Navigate(u *url.URL) (Page, error) {
	err := chromedp.Run(
		ctx,
		network.Disable(),
		dom.Enable(),
		accessibility.Enable(),
		chromedp.Navigate(u.String()),
		chromedp.ActionFunc(func(pageCtx context.Context) error {
			page.ctx = pageCtx
			page.FetchAXTree()

			// root, err := accessibility.GetRootAXNode().Do(pageCtx)
			// if err != nil {
			// 	return fmt.Errorf("get root: %w", err)
			// }
			// stack := []*accessibility.Node{root}
			// for len(stack) > 0 {
			// 	current := stack[len(stack)-1]
			// 	stack = stack[:len(stack)-1]
			// 	children, err := accessibility.GetChildAXNodes(current.NodeID).Do(pageCtx)
			// 	if err != nil {
			// 		slog.Warn("get children", "err", err)
			// 		continue
			// 	}
			// 	stack = append(stack, children...)
			// }

			return nil
		}),
	)
}
