package main

import (
	"ax-distiller/lib/chrome"
	"ax-distiller/lib/dnode"
	"encoding/xml"
	"fmt"
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

	km := dnode.NewKeymap(1024)
	ht1 := dnode.NewHashTree(dnode.FromAXTree(n1, km), 1024)
	ht2 := dnode.NewHashTree(dnode.FromAXTree(n2, km), 1024)

	common := findCommon(km, ht2, ht1, ht2.Root)
	fmt.Println(dnode.Print(km, common))

	buff, err := xml.MarshalIndent(toAXTree(km, common), "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(buff))
}

func findCommon(km dnode.Keymap, self, other dnode.HashTree, hash uint64) *dnode.Node {
	node, exists := self.FromHash[hash]
	if !exists {
		panic("self node's hash does not exist in self")
	}

	var commonNS *dnode.Node
	if node.NextSiblingHash != 0 {
		commonNS = findCommon(km, self, other, node.NextSiblingHash)
	}

	_, nodeAndSubtreeEqual := other.FromHash[hash]
	if nodeAndSubtreeEqual {
		return &dnode.Node{
			FullKey:     node.Original.FullKey,
			NextSibling: commonNS,
			FirstChild:  node.Original.FirstChild,
		}
	}

	_, nodeEqual := other.FromFullKey[node.Original.FullKey]
	if !nodeEqual {
		return commonNS
	}

	out := &dnode.Node{
		FullKey:     node.Original.FullKey,
		NextSibling: commonNS,
	}
	if node.FirstChildHash != 0 {
		out.FirstChild = findCommon(km, self, other, node.FirstChildHash)
	}

	return out
}

// this assumes node is the AX_NODE container
func toAXTree(km dnode.Keymap, node *dnode.Node) *chrome.AXNode {
	out := &chrome.AXNode{}

	var lastChild *chrome.AXNode

	cur := node.FirstChild
	for cur != nil {
		text, ok := km.StringOf(cur.FullKey)
		if !ok {
			panic("unknown key")
		}

		if strings.HasPrefix(text, "role:") {
			out.Role = text[5:]
		} else if strings.HasPrefix(text, "attr:") {
			value := ""
			if cur.FirstChild != nil {
				value, _ = km.StringOf(cur.FirstChild.FullKey)
			}
			out.Properties = append(out.Properties, chrome.Prop{
				Name:  text[5:],
				Value: value,
			})
		} else if strings.HasPrefix(text, "AX_NODE") {
			child := toAXTree(km, cur)
			if lastChild != nil {
				lastChild.NextSibling = child
			} else {
				out.FirstChild = child
			}
			lastChild = child
		}

		cur = cur.NextSibling
	}

	if out.Role == "" {
		out.Role = "UNKNOWN"
	}

	return out
}
