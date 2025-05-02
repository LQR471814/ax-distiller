package ax

import (
	"context"
	"encoding/xml"
	"fmt"
	"log"

	"github.com/chromedp/cdproto/accessibility"
	"github.com/chromedp/cdproto/cdp"
	"github.com/mailru/easyjson"
)

type cdpValue struct {
	Type  string `json:"type"`
	Value any    `json:"value"`
}

type cdpNodeProp struct {
	Name  string    `json:"name"`
	Value *cdpValue `json:"value"`
}

type cdpNode struct {
	NodeID      string        `json:"nodeId"`
	Ignored     bool          `json:"ignored"`
	Role        *cdpValue     `json:"role,omitempty"`
	Name        *cdpValue     `json:"name,omitempty"`
	Description *cdpValue     `json:"description,omitempty"`
	Properties  []cdpNodeProp `json:"properties,omitempty"`
	ChildIds    []string      `json:"childIds,omitempty"`
	DomNodeId   int64         `json:"backendDOMNodeId"`
}

type getNodesResult struct {
	Nodes []cdpNode `json:"nodes"`
}

type AXNodeProp struct {
	Name  string
	Value string
}

type AXNode struct {
	Role        string
	Name        string
	Description string
	Properties  []AXNodeProp
	Children    []AXNode
	DomNodeId   int64
}

func (n AXNode) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name.Local = n.Role
	start.Attr = make([]xml.Attr, 0, len(n.Properties)+1)
	if n.Name != "" {
		start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Local: "name"}, Value: n.Name})
	}
	for _, prop := range n.Properties {
		if prop.Value == "" {
			continue
		}
		start.Attr = append(start.Attr, xml.Attr{
			Name:  xml.Name{Local: prop.Name},
			Value: prop.Value,
		})
	}
	err := e.EncodeToken(start)
	if err != nil {
		return err
	}
	for _, child := range n.Children {
		err := e.Encode(child)
		if err != nil {
			return err
		}
	}
	return e.EncodeToken(xml.EndElement{Name: start.Name})
}

const (
	TREE_ROOT_ROLE = "ROOT"
)

func getAccessibilityTree(pageCtx context.Context) (AXNode, error) {
	params := easyjson.RawMessage("{}")
	result := &getNodesResult{}

	err := cdp.Execute(pageCtx, accessibility.CommandGetFullAXTree, &params, result)
	if err != nil {
		return AXNode{}, err
	}
	if len(result.Nodes) == 0 {
		return AXNode{}, fmt.Errorf("no result nodes returned")
	}

	mapping := map[string]cdpNode{}
	for _, node := range result.Nodes {
		mapping[node.NodeID] = node
	}

	// implicitly given that nodes are defined in order
	roots := buildAXTree(mapping, result.Nodes[0])

	return AXNode{
		Role:      TREE_ROOT_ROLE,
		Name:      "",
		Children:  roots,
		DomNodeId: -1,
	}, nil
}

func buildAXTree(mapping map[string]cdpNode, node cdpNode) []AXNode {
	var role string
	if node.Role != nil {
		role, _ = node.Role.Value.(string)
	}

	var name string
	if node.Name != nil {
		name, _ = node.Name.Value.(string)
	}

	var desc string
	if node.Description != nil {
		desc, _ = node.Description.Value.(string)
	}

	props := make([]AXNodeProp, len(node.Properties))
	for i, p := range node.Properties {
		props[i] = AXNodeProp{
			Name:  p.Name,
			Value: fmt.Sprint(p.Value.Value),
		}
	}

	children := make([]AXNode, 0, len(node.ChildIds))
	for _, childId := range node.ChildIds {
		child, ok := mapping[childId]
		if !ok {
			log.Printf("buildAXTree: child id '%v' could not be found\n", childId)
			continue
		}
		subNodes := buildAXTree(mapping, child)
		children = append(children, subNodes...)
	}

	if node.Ignored || role == "generic" {
		return children
	}

	return []AXNode{{
		Role:        role,
		Name:        name,
		Description: desc,
		Properties:  props,
		Children:    children,
		DomNodeId:   node.DomNodeId,
	}}
}
