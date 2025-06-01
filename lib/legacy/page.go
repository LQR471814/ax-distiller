package irrelevant

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/runtime"
)

type Page struct {
	url    *url.URL
	ctx    context.Context
	cancel func()

	mapping map[string]cdpNode
	root    Node

	Tree Node
}

func (p Page) ShowDebugInfo() error {
	err := dom.Enable().
		WithIncludeWhitespace(dom.EnableIncludeWhitespaceNone).
		Do(p.ctx)
	if err != nil {
		return err
	}
	return p.recurseDebug(p.Tree)
}

func (p Page) recurseDebug(node Node) error {
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

func (p Page) debugDomNode(node Node) error {
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

const script = `(e) => {
	if (!window.objs) {
		window.objs = []
	}
	window.objs.push(e)
	if (e.style) {
		e.style.border = "solid %s 1px"
	}
}`
