package ax

import (
	"encoding/xml"
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

type NodeProp struct {
	Name  string
	Value string
}

type Node struct {
	Role        string
	Name        string
	Description string
	Properties  []NodeProp
	Children    []Node
	DomNodeId   int64
}

func (n Node) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
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
