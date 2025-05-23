package chrome

import (
	"encoding/xml"
	"fmt"
	"strconv"
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

	n.Role = cn.Role.Value.(string)
	if cn.Name != nil {
		n.Name = cn.Name.Value.(string)
	}
	if cn.Description != nil {
		n.Description = cn.Description.Value.(string)
	}
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
