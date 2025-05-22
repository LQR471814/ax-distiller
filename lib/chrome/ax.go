package chrome

import (
	"context"
	"encoding/xml"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/chromedp/cdproto/accessibility"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
	easyjson "github.com/mailru/easyjson"
)

type Prop struct {
	Name  string
	Value string
}

type AXNode struct {
	ID          uint64
	Role        string
	Name        string
	Description string
	Properties  []Prop
	FirstChild  *AXNode
	NextSibling *AXNode
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
	child := n.FirstChild
	for child != nil {
		err := e.Encode(child)
		if err != nil {
			return err
		}
		child = child.NextSibling
	}
	return e.EncodeToken(xml.EndElement{Name: start.Name})
}

func mustParseNodeID(id string) uint64 {
	var parsed int64
	parsed, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		panic(err)
	}
	return uint64(parsed)
}

func (n *AXNode) metadataFromCDP(cn cdpAXNode) (err error) {
	n.ID = mustParseNodeID(cn.NodeID)

	n.Name = cn.Name.Value.(string)
	n.Role = cn.Role.Value.(string)
	n.Description = cn.Description.Value.(string)
	n.DomNodeId = cn.DomNodeId

	n.Properties = make([]Prop, len(cn.Properties))
	for i, p := range cn.Properties {
		n.Properties[i] = Prop{
			Name:  p.Name,
			Value: fmt.Sprint(p.Value.Value),
		}
	}
	return
}

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
	var lastNode *AXNode
	for i := len(nodeList) - 1; i >= 0; i-- {
		cnode := allNodes[nodeList[i]]

		if cnode.Ignored {
			firstSubchild := ax.convertNodeList(allNodes, cnode.ChildIds)

			cur := firstSubchild
			for {
				if cur.NextSibling == nil {
					cur.NextSibling = lastNode
					break
				}
				cur = cur.NextSibling
			}

			lastNode = firstSubchild
			continue
		}

		node := &AXNode{}
		node.metadataFromCDP(cnode)
		node.NextSibling = lastNode
		node.FirstChild = ax.convertNodeList(allNodes, cnode.ChildIds)
		ax.Nodes[node.ID] = node

		lastNode = node
	}
	return lastNode
}

func (ax AX) FetchFullAXTree() (root *AXNode, err error) {
	params := easyjson.RawMessage("{}")
	result := &getAXNodesResult{}

	err = cdp.Execute(ax.PageCtx, accessibility.CommandGetFullAXTree, &params, result)
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

func (ax AX) FetchSubtree(id string) (root *AXNode, err error) {
	params := easyjson.RawMessage(fmt.Sprintf(`{"id":"%s"}`, id))
	result := &getAXNodesResult{}

	err = cdp.Execute(ax.PageCtx, accessibility.CommandGetChildAXNodes, &params, result)
	if err != nil {
		return
	}
	childList := result.Nodes

	var lastNode *AXNode
	for i := len(childList) - 1; i >= 0; i-- {
		child := childList[i]
		if child.Ignored {
			var firstSubchild *AXNode
			firstSubchild, err = ax.FetchSubtree(child.NodeID)
			if err != nil {
				return
			}

			cur := firstSubchild
			for {
				if cur.NextSibling == nil {
					cur.NextSibling = lastNode
					break
				}
				cur = cur.NextSibling
			}

			lastNode = firstSubchild
			continue
		}

		node := &AXNode{}
		node.metadataFromCDP(child)
		node.NextSibling = lastNode
		node.FirstChild = ax.FetchSubtree()
		ax.Nodes[node.ID] = node

		lastNode = node
	}

	return
}

func (ax AX) checkStale(id uint64) bool {
	_, exists := ax.Nodes[id]
	if !exists {
		return true
	}
	_, isStale := ax.staleNodes[id]
	return isStale
}

func (ax AX) refreshNode(id string) (err error) {
	uid := mustParseNodeID(id)
	if !ax.checkStale(uid) {
		return nil
	}
}

// ListenChanges fires an event whenever the subtree of an AX node changes. The
// ID of the closest significant AX node ancestor (and all the ancestors of
// that node) to the changed subtree is provided in onChange.
func (ax AX) Listen() {
	var timer *time.Timer

	chromedp.ListenTarget(ax.PageCtx, func(ev any) {
		// fmt.Printf("%T\n", ev)
		switch typed := ev.(type) {
		case *accessibility.EventLoadComplete:
			slog.Info("[event] load complete", "id", typed.Root.NodeID)
		case *accessibility.EventNodesUpdated:
			roles := make([]string, len(typed.Nodes))
			for i, n := range typed.Nodes {
				roles[i] = n.Role.Value.String()
			}
			slog.Info("[event] nodes updated", "roles", roles)

		case *dom.EventAttributeModified:
			slog.Info("[event] attribute modified", "id", typed.NodeID, "attr", typed.Name)
		case *dom.EventAttributeRemoved:
			slog.Info("[event] attribute removed", "id", typed.NodeID, "attr", typed.Name)
		case *dom.EventCharacterDataModified:
			slog.Info("[event] character data modified", "id", typed.NodeID, "data", typed.CharacterData)
		case *dom.EventChildNodeCountUpdated:
			slog.Info("[event] DOM node count updated", "id", typed.NodeID, "count", typed.ChildNodeCount)

			if timer != nil {
				timer.Stop()
			}
			timer = time.NewTimer(time.Second)
			go func() {
				<-timer.C

				params := easyjson.RawMessage(fmt.Sprintf(`{"nodeId":%d}`, typed.NodeID.Int64()))
				result := &getAXNodesResult{}

				err := cdp.Execute(ax.PageCtx, accessibility.CommandGetAXNodeAndAncestors, &params, result)
				if err != nil {
					slog.Error("[nav] get partial ax tree", "err", err)
					return
				}

				// ancestors is the list of ancestors following the closest significant AX node ancestor
				var ancestors []accessibility.NodeID
				for i, anc := range result.Nodes {
					if anc.Ignored ||
						anc.Role.Value.(string) == "generic" ||
						anc.Role.Value.(string) == "none" {
						continue
					}
					ancestors = make([]accessibility.NodeID, 0, len(result.Nodes)-i)
					for j := i; j < len(result.Nodes); j++ {
						ancestors = append(ancestors, accessibility.NodeID(result.Nodes[i].NodeID))
					}
					break
				}
			}()

		case *dom.EventChildNodeInserted:
			slog.Info("[event] DOM node inserted", "parent_id", typed.ParentNodeID, "id", typed.Node.NodeID)
		case *dom.EventChildNodeRemoved:
			slog.Info("[event] DOM node removed", "parent_id", typed.ParentNodeID, "id", typed.NodeID)
		}
	})
}
