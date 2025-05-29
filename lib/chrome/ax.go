package chrome

import (
	"context"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/chromedp/cdproto/accessibility"
	"github.com/chromedp/cdproto/cdp"
	easyjson "github.com/mailru/easyjson"
)

type AX struct {
	PageCtx    context.Context
	Nodes      map[uint64]*AXNode
	staleNodes map[uint64]struct{}
}

func NewAX(pageCtx context.Context) AX {
	return AX{
		PageCtx:    pageCtx,
		Nodes:      make(map[uint64]*AXNode),
		staleNodes: make(map[uint64]struct{}),
	}
}

func (ax AX) convertNodeList(allNodes map[string]cdpAXNode, nodeList []string) *AXNode {
	var nextNode *AXNode
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

		converted := &AXNode{}
		converted.metadataFromCDP(node)
		converted.NextSibling = nextNode
		converted.FirstChild = ax.convertNodeList(allNodes, node.ChildIds)
		ax.Nodes[converted.ID] = converted
		nextNode = converted
	}
	return nextNode
}

func (ax AX) FetchFullAXTree() (root *AXNode, err error) {
	params := easyjson.RawMessage("{}")
	// 512 kB
	resmsg := make(easyjson.RawMessage, 524288)

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

	ax.Nodes = make(map[uint64]*AXNode)
	allNodes := make(map[string]cdpAXNode)
	for _, node := range result.Nodes {
		allNodes[node.NodeID] = node
	}

	// it is assumed that nodes are defined in order
	root = ax.convertNodeList(allNodes, []string{result.Nodes[0].NodeID})
	return
}
