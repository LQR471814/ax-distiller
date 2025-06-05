package main

import (
	"ax-distiller/lib/chrome/ax"
	"context"
	"fmt"
	"math/rand"

	"github.com/chromedp/cdproto/cdp"
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

	err = chromedp.Click([]cdp.NodeID{nodeID}, chromedp.ByNodeID).Do(ctx)
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
