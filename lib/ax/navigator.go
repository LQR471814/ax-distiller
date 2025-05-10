package ax

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"time"

	"github.com/chromedp/cdproto/accessibility"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	easyjson "github.com/mailru/easyjson"
)

type Navigator struct {
	Timeout time.Duration

	ctx    context.Context
	cancel func()
}

func NewNavigator(ctx context.Context) (Navigator, error) {
	dataTemp := "./data/chrome-data"
	err := os.RemoveAll(dataTemp)
	if err != nil {
		return Navigator{}, err
	}
	err = os.Mkdir(dataTemp, 0777)
	if err != nil {
		return Navigator{}, err
	}

	allocatorCtx, _ := chromedp.NewExecAllocator(
		ctx,
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Env("APPIMAGELAUNCHER_DISABLE=1"),
			chromedp.ExecPath("./data/thorium-browser"),
			chromedp.UserDataDir(dataTemp),
			chromedp.Flag("load-extension", "./data/ublock"),
			chromedp.Flag("headless", false),
			chromedp.Flag("blink-settings", "imagesEnabled=false"),
			chromedp.Flag("disable-extensions", false),
			chromedp.Flag("disable-remote-fonts", true),
			chromedp.Flag("no-sandbox", true),
		)...,
	)
	ctx, cancel := chromedp.NewContext(allocatorCtx)

	err = chromedp.Run(ctx)
	if err != nil {
		return Navigator{}, err
	}

	return Navigator{
		Timeout: 20 * time.Second,
		ctx:     ctx,
		cancel:  cancel,
	}, nil
}

func (n Navigator) Navigate(u *url.URL) (Page, error) {
	ctx, cancel := chromedp.NewContext(n.ctx)

	axTree := Node{}

	page := Page{
		url:    u,
		cancel: cancel,
		Tree:   axTree,
	}

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
	if err != nil {
		cancel()
		return Page{}, err
	}

	chromedp.ListenTarget(ctx, func(ev any) {
		// fmt.Printf("%T\n", ev)
		switch typed := ev.(type) {
		case *accessibility.EventLoadComplete:
			slog.Info("[event] load complete", "id", typed.Root.NodeID)
		case *accessibility.EventNodesUpdated:
			roles := make([]string, len(typed.Nodes))
			for i, n := range typed.Nodes {
				roles[i] = n.Role.Value.String()
			}
			slog.Info("[event] nodes updated", "roles", roles)
		// case *dom.EventAttributeModified:
		// 	slog.Info("[event] attribute modified", "id", typed.NodeID, "attr", typed.Name)
		// case *dom.EventAttributeRemoved:
		// 	slog.Info("[event] attribute removed", "id", typed.NodeID, "attr", typed.Name)
		// case *dom.EventCharacterDataModified:
		// 	slog.Info("[event] character data modified", "id", typed.NodeID, "data", typed.CharacterData)
		case *dom.EventChildNodeCountUpdated:
			slog.Info("[event] DOM node count updated", "id", typed.NodeID, "count", typed.ChildNodeCount)
			go func() {
				params := easyjson.RawMessage(fmt.Sprintf(`{"nodeId": %d}`, typed.NodeID.Int64()))
				result := &getNodesResult{}

				err := cdp.Execute(page.ctx, accessibility.CommandGetAXNodeAndAncestors, &params, result)
				if err != nil {
					slog.Error("[nav] get partial ax tree", "err", err)
					return
				}

				type node struct {
					Name string
					Role string
				}
				nodes := make([]node, len(result.Nodes))
				for i, r := range result.Nodes {
					if r.Name != nil {
						nodes[i].Name = r.Name.Value.(string)
					}
					nodes[i].Role = r.Role.Value.(string)
				}
				slog.Info("[nav] got partial ax tree", "nodes", nodes)
			}()
		case *dom.EventChildNodeInserted:
			slog.Info("[event] DOM node inserted", "parent_id", typed.ParentNodeID, "id", typed.Node.NodeID)
		case *dom.EventChildNodeRemoved:
			slog.Info("[event] DOM node removed", "parent_id", typed.ParentNodeID, "id", typed.NodeID)
		}
	})

	return page, nil
}
