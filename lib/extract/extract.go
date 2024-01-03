package extract

import (
	mdrender "ax-distiller/lib/md-render"
	"context"
	"log/slog"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/cdproto/accessibility"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
)

type Extractor struct {
	Timeout    time.Duration
	Preprocess func(AXNode) (AXNode, bool)
	ctx        context.Context
	cancel     func()
}

func NewExtractor() (Extractor, error) {
	os.Mkdir("debug", 0777)

	dataTemp := "./data/chrome-data"
	err := os.RemoveAll(dataTemp)
	if err != nil {
		return Extractor{}, err
	}
	err = os.Mkdir(dataTemp, 0777)
	if err != nil {
		return Extractor{}, err
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
	if err != nil {
		return Extractor{}, err
	}

	err = chromedp.Run(ctx)
	if err != nil {
		return Extractor{}, err
	}

	return Extractor{
		Timeout: 20 * time.Second,
		ctx:     ctx,
		cancel:  cancel,
	}, nil
}

func (e Extractor) Extract(url *url.URL) ([]mdrender.Node, AXNode, error) {
	currentCtx, cancelCurrent := chromedp.NewContext(e.ctx)
	defer cancelCurrent()

	withTimeout, cancelTimeout := context.WithTimeout(currentCtx, e.Timeout)
	defer cancelTimeout()

	axTree := AXNode{}
	var mdTree []mdrender.Node
	err := chromedp.Run(
		withTimeout,
		chromedp.Navigate(url.String()),
		chromedp.ActionFunc(func(pageCtx context.Context) error {
			err := accessibility.Enable().Do(pageCtx)
			if err != nil {
				return err
			}

			axTree, err = getAccessibilityTree(pageCtx)
			if err != nil {
				return err
			}

			if e.Preprocess != nil {
				newTree, ok := e.Preprocess(axTree)
				if ok {
					axTree = newTree
				}
			}

			mdTree = MarkdownFromAXTree(pageCtx, axTree, url)
			return nil
		}),
	)

	return mdTree, axTree, err
}

func (e Extractor) Context() context.Context {
	return e.ctx
}

func (e Extractor) Destroy() {
	e.cancel()
}

func filterParagraphElements(nodes []mdrender.Node) []mdrender.ParagraphElement {
	result := []mdrender.ParagraphElement{}
	for _, n := range nodes {
		cast, ok := n.(mdrender.ParagraphElement)
		if ok {
			result = append(result, cast)
		}
	}
	return result
}

type traversalState struct {
	baseUrl        *url.URL
	underMain      bool
	underParagraph bool
	underTable     bool
	listItemDepth  int
}

func MarkdownFromAXTree(pageCtx context.Context, root AXNode, baseUrl *url.URL) []mdrender.Node {
	return convertMdFromAx(pageCtx, root, traversalState{
		baseUrl: baseUrl,
	})
}

func convertMdFromAx(
	pageCtx context.Context,
	root AXNode,
	state traversalState,
) []mdrender.Node {
	childState := state
	nodes := []mdrender.Node{}

	switch string(root.Role) {
	case "RootWebArea":
		header := strings.Trim(root.Name, " \t\n")
		nodes = append(nodes, mdrender.Header{
			Order: 1,
			Content: mdrender.PlainText{
				Content: header,
			},
		})
	case "main":
		childState.underMain = true
	case "StaticText":
		text := strings.Trim(root.Name, " \t\n")
		return []mdrender.Node{
			mdrender.PlainText{
				Content: text,
			},
		}
	case "heading":
		children := []mdrender.Node{}
		for _, child := range root.Children {
			children = append(
				children,
				convertMdFromAx(pageCtx, child, state)...,
			)
		}

		order := 1
		node, err := dom.DescribeNode().
			WithBackendNodeID(cdp.BackendNodeID(root.DomNodeId)).
			Do(pageCtx)
		if err != nil {
			slog.Warn("could not get DOM node", "id", root.DomNodeId, "err", err.Error())
		} else {
			parsed, err := strconv.ParseInt(strings.ReplaceAll(node.NodeName, "H", ""), 10, 32)
			if err != nil {
				slog.Warn("could not parse heading order", "tagName", node.NodeName)
			} else {
				order = int(parsed)
			}
		}

		return []mdrender.Node{
			mdrender.Header{
				Order: order,
				Content: mdrender.Paragraph{
					Elements: filterParagraphElements(children),
				},
			},
		}
	case "list":
		listType := mdrender.LIST_UNORDERED
		node, err := dom.DescribeNode().
			WithBackendNodeID(cdp.BackendNodeID(root.DomNodeId)).
			Do(pageCtx)
		if err != nil {
			slog.Warn("could not get DOM node", "id", root.DomNodeId, "err", err.Error())
		} else if node.NodeName == "OL" {
			listType = mdrender.LIST_ORDERED
		}

		childState.listItemDepth++

		children := []mdrender.ListItem{}
		for _, child := range root.Children {
			childNodes := convertMdFromAx(pageCtx, child, childState)
			for _, childNode := range childNodes {
				cast, ok := childNode.(mdrender.ListItem)
				if ok {
					children = append(children, cast)
				}
			}
		}

		return []mdrender.Node{
			mdrender.List{
				Type:  listType,
				Items: children,
			},
		}
	case "paragraph", "caption", "note", "listitem":
		if state.underParagraph {
			children := []mdrender.Node{}
			for _, child := range root.Children {
				children = append(children, convertMdFromAx(pageCtx, child, state)...)
			}
			return children
		}

		childState.underParagraph = true

		children := []mdrender.ParagraphElement{}
		for _, child := range root.Children {
			children = append(
				children,
				filterParagraphElements(convertMdFromAx(
					pageCtx, child, childState,
				))...,
			)
		}

		return []mdrender.Node{
			mdrender.Paragraph{
				Elements: children,
			},
		}
	case "code":
		child := root.Children[0]
		if child.Role != "StaticText" {
			return []mdrender.Node{}
		}

		return []mdrender.Node{
			mdrender.BlockCode{
				Content: child.Name,
			},
		}
	case "link":
		children := []mdrender.Node{}
		for _, child := range root.Children {
			children = append(
				children,
				convertMdFromAx(pageCtx, child, state)...,
			)
		}

		href := "ERROR"
		node, err := dom.DescribeNode().
			WithBackendNodeID(cdp.BackendNodeID(root.DomNodeId)).
			Do(pageCtx)
		if err != nil {
			slog.Warn("could not get DOM node", "id", root.DomNodeId, "err", err.Error())
		} else {
			href = node.AttributeValue("href")
		}

		linkUrl := href
		parsed, err := url.Parse(href)
		if err != nil {
			slog.Warn("could not parse href", "href", href)
		} else if state.baseUrl != nil {
			linkUrl = state.baseUrl.ResolveReference(parsed).String()
		}

		return []mdrender.Node{
			mdrender.Link{
				URL: linkUrl,
				Title: mdrender.Paragraph{
					Elements: filterParagraphElements(children),
				},
			},
		}
	case "table":
		childState.underTable = true

		table := [][]mdrender.CanBeDecorated{}
		for _, row := range root.Children {
			if row.Role != "row" {
				continue
			}

			rowMd := []mdrender.CanBeDecorated{}
			for _, cell := range row.Children {
				cellNodes := convertMdFromAx(pageCtx, cell, childState)
				elements := filterParagraphElements(cellNodes)
				if len(elements) == 0 {
					continue
				}
				rowMd = append(rowMd, mdrender.Paragraph{Elements: elements})
			}

			table = append(table, rowMd)
		}

		return []mdrender.Node{
			mdrender.Table{Rows: table},
		}
	}

	for _, child := range root.Children {
		nodes = append(
			nodes,
			convertMdFromAx(pageCtx, child, childState)...,
		)
	}

	return nodes
}
