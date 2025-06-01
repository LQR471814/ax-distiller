package ax

import (
	"context"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/chromedp/cdproto/accessibility"
	"github.com/chromedp/cdproto/cdp"
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
