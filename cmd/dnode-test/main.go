package main

import (
	"ax-distiller/lib/chrome"
	"ax-distiller/lib/dnode"
	"encoding/xml"
	"fmt"
	"strings"
	"unique"

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
		Role: unique.Make(start.Name.Local),
	}

	for _, attr := range start.Attr {
		if attr.Value == "<nil>" {
			attr.Value = ""
		}
		node.Properties = append(node.Properties, chrome.Prop{
			Name:  unique.Make(attr.Name.Local),
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

	km := dnode.NewKeymap()
	ht1 := dnode.NewHashTree(dnode.FromAXTree(n1, km))
	ht2 := dnode.NewHashTree(dnode.FromAXTree(n2, km))

	common := dnode.FindCommon(km, ht2, ht1, ht2.Root)
	fmt.Println(dnode.Print(km, common))

	buff, err := xml.MarshalIndent(dnode.ToAXTree(km, common), "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(buff))
}
