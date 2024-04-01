package axextract

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/chromedp/cdproto/accessibility"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

type Navigator struct {
	Timeout time.Duration

	ctx    context.Context
	cancel func()
}

func NewNavigator() (Navigator, error) {
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
		context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Env("APPIMAGELAUNCHER_DISABLE=1"),
			chromedp.ExecPath("./data/thorium-browser"),
			chromedp.UserDataDir(dataTemp),
			chromedp.Flag("load-extension", "./data/ublock"),
			chromedp.Flag("headless", false),
			chromedp.Flag("blink-settings", "imagesEnabled=false"),
			chromedp.Flag("disable-extensions", false),
			chromedp.Flag("disable-remote-fonts", true),
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

type Page struct {
	url    *url.URL
	ctx    context.Context
	cancel func()
	tree   AXNode
}

func (p Page) URL() *url.URL {
	return p.url
}

func (p Page) GetDomInfo(nodeId int64) (*cdp.Node, error) {
	return dom.DescribeNode().
		WithBackendNodeID(cdp.BackendNodeID(nodeId)).
		Do(p.ctx)
}

func (p Page) GetHTML(nodeId int64) (string, error) {
	return dom.GetOuterHTML().
		WithBackendNodeID(cdp.BackendNodeID(nodeId)).
		Do(p.ctx)
}

func (p Page) Tree() AXNode {
	return p.tree
}

func (p Page) Close() {
	p.cancel()
}

const script = `(e) => {
	if (!window.objs) {
		window.objs = []
	}
	window.objs.push(e)
	if (e.style) {
		e.style.border = "solid %s 1px"
	}
}`

func (p Page) debugDomNode(node AXNode) error {
	if node.DomNodeId < 0 {
		return nil
	}

	obj, err := dom.ResolveNode().
		WithBackendNodeID(cdp.BackendNodeID(node.DomNodeId)).
		Do(p.ctx)
	if err != nil {
		if strings.Contains(err.Error(), "node with given id found") {
			slog.Warn("Node not found", "id", node.DomNodeId)
			return nil
		}
		if strings.Contains(err.Error(), "nodeId or backendNodeId must be specified") {
			slog.Warn("Unspecified backendNodeId", "id", node.DomNodeId)
			return nil
		}
		return err
	}

	color := "pink"
	category := RoleCategoryMap[node.Role]
	if len(category) > 0 {
		switch category[0] {
		case CATEGORY_STRUCTURE:
			color = "black"
		case CATEGORY_GENERIC:
			color = ""
		case CATEGORY_SECTIONHEAD:
			color = "aqua"
		case CATEGORY_SECTION:
			color = "blue"
		case CATEGORY_LANDMARK:
			color = "green"
		case CATEGORY_DIALOG:
			color = "purple"
		case CATEGORY_WIDGET:
			color = "red"
		case CATEGORY_DOCUMENT:
			color = "yellow"
		}
	}

	if color == "" {
		return nil
	}

	_, exception, err := runtime.CallFunctionOn(fmt.Sprintf(script, color)).
		WithObjectID(obj.ObjectID).
		WithArguments([]*runtime.CallArgument{
			{ObjectID: obj.ObjectID},
		}).
		Do(p.ctx)
	if err != nil {
		return err
	}
	if exception != nil {
		return exception
	}

	return nil
}

func (p Page) recurseDebug(node AXNode) error {
	err := p.debugDomNode(node)
	if err != nil {
		return err
	}

	for _, c := range node.Children {
		err := p.recurseDebug(c)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p Page) ShowDebugInfo() error {
	err := dom.Enable().
		WithIncludeWhitespace(dom.EnableIncludeWhitespaceNone).
		Do(p.ctx)
	if err != nil {
		return err
	}
	return p.recurseDebug(p.tree)
}

func (n Navigator) Navigate(u *url.URL) (Page, error) {
	currentCtx, cancelCurrent := chromedp.NewContext(n.ctx)
	withTimeout, cancelTimeout := context.WithTimeout(currentCtx, n.Timeout)

	axTree := AXNode{}
	var ctx context.Context

	cancel := func() {
		cancelTimeout()
		cancelCurrent()
	}

	err := chromedp.Run(
		withTimeout,
		chromedp.Navigate(u.String()),
		chromedp.ActionFunc(func(pageCtx context.Context) error {
			err := accessibility.Enable().Do(pageCtx)
			if err != nil {
				return err
			}

			axTree, err = getAccessibilityTree(pageCtx)
			if err != nil {
				return err
			}

			ctx = pageCtx
			return nil
		}),
	)
	if err != nil {
		cancel()
		return Page{}, err
	}

	return Page{
		url:    u,
		ctx:    ctx,
		cancel: cancel,
		tree:   axTree,
	}, nil
}
