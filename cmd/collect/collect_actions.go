package main

import (
	"ax-distiller/lib/chrome/ax"
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"strings"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

type Action interface {
	Node() *ax.Node
	Do(ctx context.Context, nodeID cdp.NodeID) error
	Color() string
}

var action_node_key = fmt.Sprintf("data-action-%d", rand.Uint64())

type ClickAction struct {
	Target *ax.Node
}

func (a ClickAction) Node() *ax.Node {
	return a.Target
}

func (a ClickAction) Do(ctx context.Context, nodeID cdp.NodeID) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("click action: %w", err)
		}
	}()

	err = dom.SetAttributeValue(nodeID, action_node_key, fmt.Sprint(a.Target.DomNodeId)).
		Do(ctx)
	if err != nil {
		return
	}

	err = chromedp.Click(fmt.Sprintf(
		"[%s=%d]",
		action_node_key,
		a.Target.DomNodeId,
	)).Do(ctx)
	return
}

func (a ClickAction) Color() string {
	return "red"
}

/*
action notes:
- all things under category "command" can be clicked
- the stuff under category "input" should be handled by setting the values they correspond to
  - combobox
  - radio
  - checkbox
    - switch
  - spinbutton
  - option
    - treeitem
  - slider
  - textbox
    - searchbox
- stuff under category "dialog/alertdialog" should be handled appropriately
- scrollbar can probably be scrolled. or you can just look for all overflow-scroll divs or something
*/

func (c Collector) getAction(node *ax.Node) Action {
	switch node.Role.Value() {
	case "button", "link":
		return ClickAction{
			Target: node,
		}
	}
	return nil
}

func (c Collector) findActions(node *ax.Node, out *[]Action) {
	found := c.getAction(node)
	if found != nil {
		*out = append(*out, found)
		return
	}
	child := node.FirstChild
	for child != nil {
		c.findActions(child, out)
		child = child.NextSibling
	}
}

const debugNodeScript = `(e) => {
	e.style.border = "1px solid %s"
}`

func (c Collector) debugNodeAction(node *ax.Node, color string) (action chromedp.Action) {
	action = chromedp.ActionFunc(func(ctx context.Context) (err error) {
		obj, err := dom.ResolveNode().
			WithBackendNodeID(cdp.BackendNodeID(node.DomNodeId)).
			Do(ctx)
		if err != nil {
			if strings.Contains(err.Error(), "node with given id found") {
				slog.Warn("node not found", "id", node.DomNodeId)
				err = nil
				return
			}
			if strings.Contains(err.Error(), "nodeID or backendNodeId must be specified") {
				slog.Warn("unspecified backendNodeId", "id", node.DomNodeId)
				err = nil
				return
			}
			return
		}

		_, jserr, err := runtime.CallFunctionOn(fmt.Sprintf(
			debugNodeScript,
			color,
		)).
			WithObjectID(obj.ObjectID).
			WithArguments([]*runtime.CallArgument{
				{ObjectID: obj.ObjectID},
			}).
			Do(ctx)
		if err != nil {
			return
		}
		if jserr != nil {
			err = jserr
			return
		}
		return
	})
	return
}
