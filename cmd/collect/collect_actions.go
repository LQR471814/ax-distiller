package main

import (
	"ax-distiller/lib/chrome/ax"
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/css"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

type Action interface {
	Node() *ax.Node
	Do() error
	Color() string
}

var action_node_key = fmt.Sprintf("data-action-%d", rand.Uint64())

type ClickAction struct {
	ctx    context.Context
	Target *ax.Node
}

func (a ClickAction) Node() *ax.Node {
	return a.Target
}

func (a ClickAction) Do() (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("click action: %w", err)
		}
	}()

	nodeIDs, err := dom.PushNodesByBackendIDsToFrontend([]cdp.BackendNodeID{
		cdp.BackendNodeID(a.Target.DomNodeId),
	}).
		Do(a.ctx)
	if err != nil {
		return
	}
	if len(nodeIDs) != 1 {
		err = fmt.Errorf("incorrect number of node IDs returned (%v)", nodeIDs)
		return
	}

	err = dom.SetAttributeValue(nodeIDs[0], action_node_key, fmt.Sprint(a.Target.DomNodeId)).
		Do(a.ctx)
	if err != nil {
		return
	}

	err = chromedp.Click(fmt.Sprintf(
		"[%s=%d]",
		action_node_key,
		a.Target.DomNodeId,
	)).
		Do(a.ctx)
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
			ctx:    c.tabctx,
			Target: node,
		}
	}
	return nil
}

func (c Collector) findActionsInner(node *ax.Node, out *[]Action) {
	found := c.getAction(node)
	if found != nil {
		*out = append(*out, found)
		return
	}
	child := node.FirstChild
	for child != nil {
		c.findActionsInner(child, out)
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
			if strings.Contains(err.Error(), "nodeId or backendNodeId must be specified") {
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

func (c Collector) findActions(tree *ax.Node) (actions []Action, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("find actions: %w", err)
		}
	}()

	t1 := time.Now()
	c.findActionsInner(tree, &actions)
	t2 := time.Now()

	slog.Info("[collect] finding actions", "time", t2.Sub(t1).String())

	t1 = time.Now()

	// debugActions := make([]chromedp.Action, len(actions))
	// for i, a := range actions {
	// 	debugActions[i] = c.debugNodeAction(a.Node(), a.Color())
	// }
	// err = chromedp.Run(c.tabctx, debugActions...)

	debugActions := make([]chromedp.Action, len(actions))
	for i, a := range actions {
		debugActions[i] = chromedp.ActionFunc(func(ctx context.Context) (err error) {
			nodeIds, err := dom.PushNodesByBackendIDsToFrontend([]cdp.BackendNodeID{
				cdp.BackendNodeID(a.Node().DomNodeId),
			}).Do(ctx)
			if err != nil {
				return
			}

			var res css.GetMatchedStylesForNodeReturns
			err = cdp.Execute(
				ctx,
				css.CommandGetMatchedStylesForNode,
				&css.GetMatchedStylesForNodeParams{NodeID: nodeIds[0]},
				&res,
			)
			if err != nil {
				return
			}
			styles := make(CSSStyles)
			styles.FromMatched(&res)
			slog.Info(fmt.Sprintf("[collect] get styles\n%s", styles.String()))
			return
		})
	}
	err = chromedp.Run(c.tabctx, debugActions...)
	if err != nil {
		return
	}

	t2 = time.Now()

	slog.Info("[collect] debugging took", "time", t2.Sub(t1).String())

	return
}
