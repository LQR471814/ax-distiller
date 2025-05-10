package chrome

type cdpAXValue struct {
	Type  string `json:"type"`
	Value any    `json:"value"`
}

type cdpAXNodeProp struct {
	Name  string      `json:"name"`
	Value *cdpAXValue `json:"value"`
}

type cdpAXNode struct {
	NodeID      string          `json:"nodeId"`
	Ignored     bool            `json:"ignored"`
	Role        *cdpAXValue     `json:"role,omitempty"`
	Name        *cdpAXValue     `json:"name,omitempty"`
	Description *cdpAXValue     `json:"description,omitempty"`
	Properties  []cdpAXNodeProp `json:"properties,omitempty"`
	ChildIds    []string        `json:"childIds,omitempty"`
	DomNodeId   int64           `json:"backendDOMNodeId"`
}

type getAXNodesResult struct {
	Nodes []cdpAXNode `json:"nodes"`
}
