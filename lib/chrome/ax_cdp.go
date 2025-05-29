package chrome

type cdpAXValue struct {
	Value any `json:"value"`
}

type cdpAXNodeProp struct {
	Name  string     `json:"name"`
	Value cdpAXValue `json:"value"`
}

type cdpAXNode struct {
	NodeID      string          `json:"nodeId"`
	Ignored     bool            `json:"ignored"`
	Role        cdpAXValue      `json:"role"`
	Name        cdpAXValue      `json:"name"`
	Description cdpAXValue      `json:"description"`
	Properties  []cdpAXNodeProp `json:"properties,omitempty"`
	ChildIds    []string        `json:"childIds,omitempty"`
	DomNodeId   int64           `json:"backendDOMNodeId"`
}

type getAXNodesResult struct {
	Nodes []cdpAXNode `json:"nodes"`
}
