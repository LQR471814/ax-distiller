package chrome

import (
	"context"
	"encoding/xml"
	"fmt"
	"log"
	"log/slog"

	"github.com/chromedp/cdproto/accessibility"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
	"github.com/mailru/easyjson"
)

type Prop struct {
	Name  string
	Value string
}

type AXNode struct {
	Role        string
	Name        string
	Description string
	Properties  []Prop
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

type AX struct {
	PageCtx context.Context
}

func (p AX) FetchFullTree() (AXNode, error) {
	params := easyjson.RawMessage("{}")
	result := &getAXNodesResult{}

	err := cdp.Execute(p.PageCtx, accessibility.CommandGetFullAXTree, &params, result)
	if err != nil {
		return AXNode{}, err
	}
	if len(result.Nodes) == 0 {
		return AXNode{}, fmt.Errorf("no result nodes returned")
	}

	mapping := map[string]cdpAXNode{}
	for _, node := range result.Nodes {
		mapping[node.NodeID] = node
	}

	// implicitly given that nodes are defined in order
	roots := buildAXTree(mapping, result.Nodes[0])

	return AXNode{
		Role:      "ROOT",
		Name:      "",
		Children:  roots,
		DomNodeId: -1,
	}, nil
}

func buildAXTree(mapping map[string]cdpAXNode, node cdpAXNode) []AXNode {
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

	props := make([]Prop, len(node.Properties))
	for i, p := range node.Properties {
		props[i] = Prop{
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

func (p AX) fetchSubtree(domNodeID int64) {

}

func (p AX) ListenChanges(handler func()) {
	chromedp.ListenTarget(p.PageCtx, func(ev any) {
		// fmt.Printf("%T\n", ev)
		switch typed := ev.(type) {
		// case *accessibility.EventLoadComplete:
		// 	slog.Info("[event] load complete", "id", typed.Root.NodeID)
		// case *accessibility.EventNodesUpdated:
		// 	roles := make([]string, len(typed.Nodes))
		// 	for i, n := range typed.Nodes {
		// 		roles[i] = n.Role.Value.String()
		// 	}
		// 	slog.Info("[event] nodes updated", "roles", roles)

		// case *dom.EventAttributeModified:
		// 	slog.Info("[event] attribute modified", "id", typed.NodeID, "attr", typed.Name)
		// case *dom.EventAttributeRemoved:
		// 	slog.Info("[event] attribute removed", "id", typed.NodeID, "attr", typed.Name)
		// case *dom.EventCharacterDataModified:
		// 	slog.Info("[event] character data modified", "id", typed.NodeID, "data", typed.CharacterData)
		case *dom.EventChildNodeCountUpdated:
			slog.Info("[event] DOM node count updated", "id", typed.NodeID, "count", typed.ChildNodeCount)
			go func() {
				params := easyjson.RawMessage(fmt.Sprintf(`{"nodeId":%d,"fetchRelatives":false}`, typed.NodeID.Int64()))
				result := &getAXNodesResult{}

				err := cdp.Execute(p.PageCtx, accessibility.CommandGetPartialAXTree, &params, result)
				if err != nil {
					slog.Error("[nav] get partial ax tree", "err", err)
					return
				}

				type node struct {
					Name string
					Role string
				}
				nodes := make([]node, len(result.Nodes))
				for i, r := range result.Nodes {
					if r.Name != nil {
						nodes[i].Name = r.Name.Value.(string)
					}
					nodes[i].Role = r.Role.Value.(string)
				}
				slog.Info("[nav] got partial ax tree", "nodes", nodes)
			}()

			// case *dom.EventChildNodeInserted:
			// 	slog.Info("[event] DOM node inserted", "parent_id", typed.ParentNodeID, "id", typed.Node.NodeID)
			// case *dom.EventChildNodeRemoved:
			// 	slog.Info("[event] DOM node removed", "parent_id", typed.ParentNodeID, "id", typed.NodeID)
		}
	})
}
