package main

import (
	"encoding/xml"
	"reflect"
	"slices"
	"strings"
	"testing"
	"unique"
)

func TestCSSTree(t *testing.T) {
	type testCase func() (input []CSSStyles, expected *CSSNode)

	table := []testCase{
		func() (input []CSSStyles, expected *CSSNode) {
			A := cssStyle("A")
			input = append(input, A)

			AB := cssStyle("A", "B")
			input = append(input, AB)

			ABC := cssStyle("A", "B", "C")
			input = append(input, ABC)

			expected = &CSSNode{
				Styles: A,
				FirstChild: &CSSNode{
					Styles: AB,
					FirstChild: &CSSNode{
						Styles: ABC,
					},
				},
			}
			return
		},
		func() (input []CSSStyles, expected *CSSNode) {
			A := cssStyle("A")
			input = append(input, A)

			AB := cssStyle("A", "B")
			input = append(input, AB)

			AC := cssStyle("A", "C")
			input = append(input, AC)

			ABC := cssStyle("A", "B", "C")
			input = append(input, ABC)

			expected = &CSSNode{
				Styles: A,
				FirstChild: &CSSNode{
					Styles: AB,
					FirstChild: &CSSNode{
						Styles: ABC,
					},
					NextSibling: &CSSNode{
						Styles: AC,
						FirstChild: &CSSNode{
							Styles: ABC,
						},
					},
				},
			}
			return
		},
		func() (input []CSSStyles, expected *CSSNode) {
			AB := cssStyle("A", "B")
			input = append(input, AB)

			ABC := cssStyle("A", "B", "C")
			input = append(input, ABC)

			AC := cssStyle("A", "C")
			input = append(input, AC)

			A := cssStyle("A")
			input = append(input, A)

			expected = &CSSNode{
				Styles: A,
				FirstChild: &CSSNode{
					Styles: AB,
					FirstChild: &CSSNode{
						Styles: ABC,
					},
					NextSibling: &CSSNode{
						Styles: AC,
						FirstChild: &CSSNode{
							Styles: ABC,
						},
					},
				},
			}
			return
		},
	}

	for i, tc := range table {
		input, expected := tc()

		tree := NewCSSTree()
		for _, inp := range input {
			tree.Add(inp)
		}

		if !compareTrees(tree.Root, expected) {
			got, err := xml.MarshalIndent(tree.Root, "", "  ")
			if err != nil {
				t.Fatal(err)
			}
			expected, err := xml.MarshalIndent(expected, "", "  ")
			if err != nil {
				t.Fatal(err)
			}
			t.Fatalf(`test case %d failed:
[EXPECTED]

%s

[GOT]

%s`, i, expected, got)
		}
	}
}

func cssStyle(props ...string) CSSStyles {
	styles := make(CSSStyles)
	for _, p := range props {
		styles[CSSProp{Key: unique.Make(p)}] = struct{}{}
	}
	return styles
}

func (s CSSStyles) String() string {
	keys := make([]string, 0, len(s))
	for prop := range s {
		keys = append(keys, prop.Key.Value())
	}
	slices.Sort(keys)
	return strings.Join(keys, "")
}

func (n *CSSNode) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name.Local = n.Styles.String()
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

func compareTrees(a, b *CSSNode) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	var nextSiblingEqual bool
	if reflect.DeepEqual(a.Styles, b.Styles) {
		nextSiblingEqual = compareTrees(a.NextSibling, b.NextSibling)
	}
	firstChildEqual := compareTrees(a.FirstChild, b.FirstChild)

	return nextSiblingEqual && firstChildEqual
}
