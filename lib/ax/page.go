package ax

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/url"
	"strings"

	"github.com/chromedp/cdproto/accessibility"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/runtime"
	"github.com/mailru/easyjson"
)

type Page struct {
	url    *url.URL
	ctx    context.Context
	cancel func()

	mapping map[string]cdpNode
	root    Node

	Tree Node
}

func (p Page) FetchAXTree() (Node, error) {
	params := easyjson.RawMessage("{}")
	result := &getNodesResult{}

	err := cdp.Execute(p.ctx, accessibility.CommandGetFullAXTree, &params, result)
	if err != nil {
		return Node{}, err
	}
	if len(result.Nodes) == 0 {
		return Node{}, fmt.Errorf("no result nodes returned")
	}

	mapping := map[string]cdpNode{}
	for _, node := range result.Nodes {
		mapping[node.NodeID] = node
	}

	// implicitly given that nodes are defined in order
	roots := buildAXTree(mapping, result.Nodes[0])

	return Node{
		Role:      "ROOT",
		Name:      "",
		Children:  roots,
		DomNodeId: -1,
	}, nil
}

func buildAXTree(mapping map[string]cdpNode, node cdpNode) []Node {
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

	props := make([]NodeProp, len(node.Properties))
	for i, p := range node.Properties {
		props[i] = NodeProp{
			Name:  p.Name,
			Value: fmt.Sprint(p.Value.Value),
		}
	}

	children := make([]Node, 0, len(node.ChildIds))
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

	return []Node{{
		Role:        role,
		Name:        name,
		Description: desc,
		Properties:  props,
		Children:    children,
		DomNodeId:   node.DomNodeId,
	}}
}

func (p Page) GetDomInfo(nodeId int64) (*cdp.Node, error) {
	return dom.DescribeNode().
		WithBackendNodeID(cdp.BackendNodeID(nodeId)).
		Do(p.ctx)
}

func (p Page) GetHTML(nodeId int64) (string, error) {
	return dom.GetOuterHTML().
		WithBackendNodeID(cdp.BackendNodeID(nodeId)).
		Do(p.ctx)
}

func (p Page) URL() *url.URL {
	return p.url
}

func (p Page) Close() {
	p.cancel()
}

func (p Page) ShowDebugInfo() error {
	err := dom.Enable().
		WithIncludeWhitespace(dom.EnableIncludeWhitespaceNone).
		Do(p.ctx)
	if err != nil {
		return err
	}
	return p.recurseDebug(p.Tree)
}

func (p Page) recurseDebug(node Node) error {
	err := p.debugDomNode(node)
	if err != nil {
		return err
	}

	for _, c := range node.Children {
		err := p.recurseDebug(c)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p Page) debugDomNode(node Node) error {
	if node.DomNodeId < 0 {
		return nil
	}

	obj, err := dom.ResolveNode().
		WithBackendNodeID(cdp.BackendNodeID(node.DomNodeId)).
		Do(p.ctx)
	if err != nil {
		if strings.Contains(err.Error(), "node with given id found") {
			slog.Warn("Node not found", "id", node.DomNodeId)
			return nil
		}
		if strings.Contains(err.Error(), "nodeId or backendNodeId must be specified") {
			slog.Warn("Unspecified backendNodeId", "id", node.DomNodeId)
			return nil
		}
		return err
	}

	color := "pink"
	category := RoleCategoryMap[node.Role]
	if len(category) > 0 {
		switch category[0] {
		case CATEGORY_STRUCTURE:
			color = "black"
		case CATEGORY_GENERIC:
			color = ""
		case CATEGORY_SECTIONHEAD:
			color = "aqua"
		case CATEGORY_SECTION:
			color = "blue"
		case CATEGORY_LANDMARK:
			color = "green"
		case CATEGORY_DIALOG:
			color = "purple"
		case CATEGORY_WIDGET:
			color = "red"
		case CATEGORY_DOCUMENT:
			color = "yellow"
		}
	}

	if color == "" {
		return nil
	}

	_, exception, err := runtime.CallFunctionOn(fmt.Sprintf(script, color)).
		WithObjectID(obj.ObjectID).
		WithArguments([]*runtime.CallArgument{
			{ObjectID: obj.ObjectID},
		}).
		Do(p.ctx)
	if err != nil {
		return err
	}
	if exception != nil {
		return exception
	}

	return nil
}

const script = `(e) => {
	if (!window.objs) {
		window.objs = []
	}
	window.objs.push(e)
	if (e.style) {
		e.style.border = "solid %s 1px"
	}
}`
