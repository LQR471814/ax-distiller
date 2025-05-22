package chrome

import (
	"context"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
)

type DOM struct {
	PageCtx context.Context
}

func (d DOM) GetDomInfo(nodeId int64) (*cdp.Node, error) {
	return dom.DescribeNode().
		WithBackendNodeID(cdp.BackendNodeID(nodeId)).
		Do(d.PageCtx)
}

func (d DOM) GetHTML(nodeId int64) (string, error) {
	return dom.GetOuterHTML().
		WithBackendNodeID(cdp.BackendNodeID(nodeId)).
		Do(d.PageCtx)
}
