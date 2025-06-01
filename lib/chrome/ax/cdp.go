package ax

type cdpValue struct {
	Value any `json:"value"`
}

type cdpAXNodeProp struct {
	Name  string   `json:"name"`
	Value cdpValue `json:"value"`
}

type cdpAXNode struct {
	NodeID      string          `json:"nodeId"`
	Ignored     bool            `json:"ignored"`
	Role        cdpValue        `json:"role"`
	Name        cdpValue        `json:"name"`
	Description cdpValue        `json:"description"`
	Properties  []cdpAXNodeProp `json:"properties,omitempty"`
	ChildIds    []string        `json:"childIds,omitempty"`
	DomNodeId   int64           `json:"backendDOMNodeId"`
}

type getAXNodesResult struct {
	Nodes []cdpAXNode `json:"nodes"`
}
