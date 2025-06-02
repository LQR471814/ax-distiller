package ax

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"unique"
)

type Role = unique.Handle[string]

type Prop struct {
	Name  Role
	Value string
}

type Node struct {
	ID          uint64
	Role        Role
	Name        string
	Description string
	Properties  []Prop
	FirstChild  *Node
	NextSibling *Node
	DomNodeId   int64
}

func (n Node) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name.Local = n.Role.Value()
	start.Attr = make([]xml.Attr, 0, len(n.Properties)+1)
	if n.Name != "" {
		start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Local: "name"}, Value: n.Name})
	}
	for _, prop := range n.Properties {
		if prop.Value == "" {
			continue
		}
		start.Attr = append(start.Attr, xml.Attr{
			Name:  xml.Name{Local: prop.Name.Value()},
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

func (n *Node) metadataFromCDP(cn cdpAXNode) (err error) {
	n.ID = mustParseNodeID(cn.NodeID)

	n.Role = unique.Make(fmt.Sprint(cn.Role.Value))
	n.Name = fmt.Sprint(cn.Name.Value)
	n.Description = fmt.Sprint(cn.Description.Value)
	n.DomNodeId = cn.DomNodeId

	n.Properties = make([]Prop, len(cn.Properties))
	for i, p := range cn.Properties {
		n.Properties[i] = Prop{
			Name:  unique.Make(p.Name),
			Value: fmt.Sprint(p.Value.Value),
		}
	}
	return
}
