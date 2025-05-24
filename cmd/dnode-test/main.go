package main

import (
	"ax-distiller/lib/chrome"
	"ax-distiller/lib/dnode"
	"encoding/xml"
	"strings"

	_ "embed"
)

// helper struct to use with decoder
type rawNode struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:",any,attr"`
	Content []byte     `xml:",innerxml"`
}

func parseNode(decoder *xml.Decoder, start xml.StartElement) (*chrome.AXNode, error) {
	node := &chrome.AXNode{
		Role: start.Name.Local,
	}

	for _, attr := range start.Attr {
		if attr.Value == "<nil>" {
			attr.Value = ""
		}
		node.Properties = append(node.Properties, chrome.Prop{
			Name:  attr.Name.Local,
			Value: attr.Value,
		})
	}

	var lastChild *chrome.AXNode
	for {
		tok, err := decoder.Token()
		if err != nil {
			return nil, err
		}

		switch elem := tok.(type) {
		case xml.StartElement:
			childNode, err := parseNode(decoder, elem)
			if err != nil {
				return nil, err
			}

			if node.FirstChild == nil {
				node.FirstChild = childNode
			} else {
				lastChild.NextSibling = childNode
			}
			lastChild = childNode

		case xml.EndElement:
			if elem.Name == start.Name {
				return node, nil
			}
		}
	}
}

func parse(xmlData string) *chrome.AXNode {
	decoder := xml.NewDecoder(strings.NewReader(xmlData))
	for {
		tok, err := decoder.Token()
		if err != nil {
			panic(err)
		}
		if startElem, ok := tok.(xml.StartElement); ok {
			node, err := parseNode(decoder, startElem)
			if err != nil {
				panic(err)
			}
			return node
		}
	}
}

//go:embed 1.xml
var file1 string

//go:embed 2.xml
var file2 string

func main() {
	n1 := parse(file1)
	n2 := parse(file2)

	dft := dnode.NewDiffTree(1024)

	debug := make(map[uint64]string, 512)
	dft.Register(dnode.NewAXTree(n1, debug).Root)
	dft.Register(dnode.NewAXTree(n2, debug).Root)

	// fmt.Println(dnode.Print(debug, root1))
	// fmt.Println(dnode.Print(debug, root2))
}
