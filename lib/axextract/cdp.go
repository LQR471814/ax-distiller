package axextract

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/chromedp/cdproto/accessibility"
	"github.com/chromedp/cdproto/cdp"
	"github.com/mailru/easyjson"
)

type cdpValue struct {
	Type  string `json:"type"`
	Value any    `json:"value"`
}

type cdpNode struct {
	NodeID     string    `json:"nodeId"`
	Ignored    bool      `json:"ignored"`
	Role       *cdpValue `json:"role,omitempty"`
	Name       *cdpValue `json:"name,omitempty"`
	Value      *cdpValue `json:"value,omitempty"`
	ChildNodes []string  `json:"childIds,omitempty"`
	DomNodeId  int64     `json:"backendDOMNodeId"`
}

type getNodesResult struct {
	Nodes []cdpNode `json:"nodes"`
}

type AXNode struct {
	Role      string   `json:"role"`
	Name      string   `json:"name"`
	Children  []AXNode `json:"children"`
	DomNodeId int64    `json:"-"`
}

const (
	TREE_ROOT_ROLE = "ROOT"
)

func getAccessibilityTree(pageCtx context.Context) (AXNode, error) {
	params := easyjson.RawMessage("{}")
	returns := easyjson.RawMessage("")
	err := cdp.Execute(pageCtx, accessibility.CommandGetFullAXTree, &params, &returns)
	if err != nil {
		return AXNode{}, err
	}

	root, err := accessibility.GetRootAXNode().Do(pageCtx)
	if err != nil {
		return AXNode{}, err
	}

	result := getNodesResult{}
	err = json.Unmarshal(returns, &result)
	if err != nil {
		return AXNode{}, err
	}

	mapping := map[string]cdpNode{}
	for _, node := range result.Nodes {
		mapping[node.NodeID] = node
	}

	roots := buildAXTree(mapping, mapping[root.NodeID.String()])

	return AXNode{
		Role:      TREE_ROOT_ROLE,
		Name:      "",
		Children:  roots,
		DomNodeId: -1,
	}, nil
}

func buildAXTree(mapping map[string]cdpNode, node cdpNode) []AXNode {
	var role string
	if node.Role != nil && reflect.TypeOf(node.Role.Value).Kind() == reflect.String {
		role = node.Role.Value.(string)
	}

	var name string
	if node.Name != nil && reflect.TypeOf(node.Role.Value).Kind() == reflect.String {
		name = node.Name.Value.(string)
	}

	children := []AXNode{}
	for _, childId := range node.ChildNodes {
		child, ok := mapping[childId]
		if ok {
			subNodes := buildAXTree(mapping, child)
			children = append(children, subNodes...)
		}
	}

	if node.Ignored || role == "generic"  {
		return children
	}

	return []AXNode{{
		Role:      role,
		Name:      name,
		Children:  children,
		DomNodeId: node.DomNodeId,
	}}
}
