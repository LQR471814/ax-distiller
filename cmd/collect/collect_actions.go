package main

import (
	"ax-distiller/lib/chrome/ax"
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"slices"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/css"
	"github.com/chromedp/cdproto/dom"
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

func (c Collector) fetchActionNodeIDs(actions []Action) (nodeIds []cdp.NodeID, err error) {
	backendToFrontendID := make(map[uint64]cdp.NodeID)
	actionBackendIDs := make([]cdp.BackendNodeID, len(actions))
	for i, elem := range actions {
		actionBackendIDs[i] = cdp.BackendNodeID(elem.Node().DomNodeId)
	}
	err = chromedp.Run(c.tabctx, chromedp.ActionFunc(func(ctx context.Context) (err error) {
		nodeIds, err = dom.PushNodesByBackendIDsToFrontend(actionBackendIDs).Do(ctx)
		return
	}))
	if err != nil {
		return
	}
	for i, elem := range nodeIds {
		backendToFrontendID[uint64(actionBackendIDs[i])] = elem
	}
	return
}

func (c Collector) fetchActionStyleHashes(actions []Action, nodeIDs []cdp.NodeID) (styles []uint64, err error) {
	styles = make([]uint64, len(nodeIDs))
	err = chromedp.Run(c.tabctx, chromedp.ActionFunc(func(ctx context.Context) (err error) {
		for i, id := range nodeIDs {
			var res css.GetMatchedStylesForNodeReturns
			err = cdp.Execute(
				ctx,
				css.CommandGetMatchedStylesForNode,
				&css.GetMatchedStylesForNodeParams{NodeID: id},
				&res,
			)
			if err != nil {
				return
			}
			style := make(CSSStyles)
			style.FromMatched(&res)
			styles[i] = style.Hash()
		}
		return
	}))
	return
}

type actionRank struct {
	Index       int
	StyleSeen   bool
	SubtreeSeen bool

	StyleHash   uint64
	SubtreeHash uint64
}

func (c Collector) rankActions(state WebsiteState, actions []Action, nodeIDs []cdp.NodeID, styleHashes []uint64) (ranks []actionRank) {
	ranks = make([]actionRank, len(actions))
	for i := range actions {
		styleHash := styleHashes[i]
		_, styleSeen := c.seenStyles[styleHash]

		dnode := state.AXToDNode[actions[i].Node().ID]
		hashed := state.HashTree.FromFullKey[dnode.FullKey]
		subtreeHash := hashed.Hash
		_, subtreeSeen := c.seenSubtrees[subtreeHash]

		ranks[i] = actionRank{
			Index:       i,
			StyleSeen:   styleSeen,
			SubtreeSeen: subtreeSeen,
			StyleHash:   styleHash,
			SubtreeHash: subtreeHash,
		}
	}

	slices.SortFunc(ranks, func(a, b actionRank) int {
		arank := 2
		if a.StyleSeen {
			arank--
		}
		if a.SubtreeSeen {
			arank--
		}
		brank := 2
		if b.StyleSeen {
			brank--
		}
		if b.SubtreeSeen {
			brank--
		}
		// negative since it is ascending by default
		return -(arank - brank)
	})

	return
}

func (c Collector) findAndTakeAction(state WebsiteState) (err error) {
	var actions []Action
	c.findActions(state.AXTree, &actions)

	nodeIDs, err := c.fetchActionNodeIDs(actions)
	if err != nil {
		return
	}

	styleHashes, err := c.fetchActionStyleHashes(actions, nodeIDs)
	if err != nil {
		return
	}

	ranks := c.rankActions(state, actions, nodeIDs, styleHashes)
	if len(ranks) == 0 {
		err = fmt.Errorf("no action to take!")
		return
	}

	topAction := ranks[0]
	actionToTake := actions[topAction.Index]

	c.seenStyles[topAction.StyleHash] = struct{}{}
	c.seenSubtrees[topAction.SubtreeHash] = struct{}{}

	err = chromedp.Run(c.tabctx, chromedp.ActionFunc(func(ctx context.Context) error {
		return actionToTake.Do(ctx, nodeIDs[topAction.Index])
	}))

	slog.Info("[collect] pick action", "actions", len(actions), "picked", actionToTake.Node())

	return
}
