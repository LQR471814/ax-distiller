package ax

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/bytedance/sonic"
	"github.com/chromedp/cdproto/accessibility"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
	easyjson "github.com/mailru/easyjson"
)

type API struct {
	PageCtx    context.Context
	Nodes      map[uint64]*Node
	staleNodes map[uint64]struct{}
}

func NewAPI(pageCtx context.Context) API {
	return API{
		PageCtx:    pageCtx,
		Nodes:      make(map[uint64]*Node),
		staleNodes: make(map[uint64]struct{}),
	}
}

func (ax API) convertNodeList(allNodes map[string]cdpAXNode, nodeList []string) *Node {
	var nextNode *Node
	for i := len(nodeList) - 1; i >= 0; i-- {
		node := allNodes[nodeList[i]]

		if node.Ignored || fmt.Sprint(node.Role.Value) == "generic" {
			child := ax.convertNodeList(allNodes, node.ChildIds)

			cur := child
			for cur != nil {
				if cur.NextSibling == nil {
					cur.NextSibling = nextNode
					break
				}
				cur = cur.NextSibling
			}
			if child != nil {
				nextNode = child
			}
			continue
		}

		converted := &Node{}
		converted.metadataFromCDP(node)
		converted.NextSibling = nextNode
		converted.FirstChild = ax.convertNodeList(allNodes, node.ChildIds)
		ax.Nodes[converted.ID] = converted
		nextNode = converted
	}
	return nextNode
}

func (ax API) SubscribeFullTree(rootID string) (err error) {
	queue := []string{rootID}

	for len(queue) > 0 {
		popped := queue[0]
		queue = queue[1:]

		slog.Info("[main] request", "node_id", popped)

		var result getAXNodesResult
		params := easyjson.RawMessage(`{"id":"` + popped + `"}`)
		err = cdp.Execute(ax.PageCtx, accessibility.CommandGetChildAXNodes, &params, &result)
		if err != nil {
			return
		}

		for _, c := range result.Nodes {
			queue = append(queue, c.NodeID)
		}
	}

	return

}

func (ax API) FetchFullTree() (root *Node, err error) {
	params := easyjson.RawMessage("{}")
	// 32 kB
	resmsg := make(easyjson.RawMessage, 32000)

	err = cdp.Execute(ax.PageCtx, accessibility.CommandGetFullAXTree, &params, &resmsg)
	if err != nil {
		return
	}

	result := getAXNodesResult{}
	err = sonic.Unmarshal([]byte(resmsg), &result)
	if err != nil {
		return
	}

	if len(result.Nodes) == 0 {
		err = fmt.Errorf("no result nodes returned")
		return
	}

	ax.Nodes = make(map[uint64]*Node)
	allNodes := make(map[string]cdpAXNode)
	for _, node := range result.Nodes {
		allNodes[node.NodeID] = node
	}

	// it is assumed that nodes are defined in order
	root = ax.convertNodeList(allNodes, []string{result.Nodes[0].NodeID})
	return
}

/*
DOM changes:
- attribute modified/removed
- character data modified
- child node count updated
- child node inserted/removed

AX changes:
-
*/

func (ax API) Listen() {
	// var timer *time.Timer

	chromedp.ListenTarget(ax.PageCtx, func(ev any) {
		// fmt.Printf("%T\n", ev)
		switch typed := ev.(type) {
		case *accessibility.EventLoadComplete:
			slog.Info("[event] accessibility.EventLoadComplete", "id", typed.Root.NodeID)
		case *accessibility.EventNodesUpdated:
			roles := make([]string, len(typed.Nodes))
			for i, n := range typed.Nodes {
				roles[i] = n.Role.Value.String()
			}
			slog.Info("[event] accessibility.EventNodesUpdated", "roles", roles)

		case *dom.EventAttributeModified:
			slog.Info("[event] dom.EventAttributeModified", "id", typed.NodeID, "attr", typed.Name)
		case *dom.EventAttributeRemoved:
			slog.Info("[event] dom.EventAttributeRemoved", "id", typed.NodeID, "attr", typed.Name)
		case *dom.EventCharacterDataModified:
			slog.Info("[event] dom.EventCharacterDataModified", "id", typed.NodeID, "data", typed.CharacterData)
		case *dom.EventChildNodeCountUpdated:
			slog.Info("[event] dom.EventChildNodeCountUpdated", "id", typed.NodeID, "count", typed.ChildNodeCount)

			// if timer != nil {
			// 	timer.Stop()
			// }
			// timer = time.NewTimer(time.Second)
			// go func() {
			// 	<-timer.C
			//
			// 	params := easyjson.RawMessage(fmt.Sprintf(`{"nodeId":%d}`, typed.NodeID.Int64()))
			// 	result := &getAXNodesResult{}
			//
			// 	err := cdp.Execute(ax.PageCtx, accessibility.CommandGetAXNodeAndAncestors, &params, result)
			// 	if err != nil {
			// 		slog.Error("[nav] get partial ax tree", "err", err)
			// 		return
			// 	}
			//
			// 	// ancestors is the list of ancestors following the closest significant AX node ancestor
			// 	var ancestors []accessibility.NodeID
			// 	for i, anc := range result.Nodes {
			// 		if anc.Ignored ||
			// 			anc.Role.Value.(string) == "generic" ||
			// 			anc.Role.Value.(string) == "none" {
			// 			continue
			// 		}
			// 		ancestors = make([]accessibility.NodeID, 0, len(result.Nodes)-i)
			// 		for j := i; j < len(result.Nodes); j++ {
			// 			ancestors = append(ancestors, accessibility.NodeID(result.Nodes[i].NodeID))
			// 		}
			// 		break
			// 	}
			// }()

		case *dom.EventChildNodeInserted:
			slog.Info("[event] dom.EventChildNodeInserted", "parent_id", typed.ParentNodeID, "id", typed.Node.NodeID)
		case *dom.EventChildNodeRemoved:
			slog.Info("[event] dom.EventChildNodeRemoved", "parent_id", typed.ParentNodeID, "id", typed.NodeID)
		}
	})
}
