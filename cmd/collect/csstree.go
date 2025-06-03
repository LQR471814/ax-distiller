package main

import (
	"bytes"
	"slices"
	"strings"
	"unique"

	"github.com/chromedp/cdproto/css"
	"github.com/zeebo/xxh3"
)

type CSSProp struct {
	Key   unique.Handle[string]
	Value unique.Handle[string]
}

type CSSStyles map[CSSProp]struct{}

func (styles CSSStyles) Set(prop *css.Property) {
	key := CSSProp{
		Key:   unique.Make(prop.Name),
		Value: unique.Make(prop.Value),
	}
	styles[key] = struct{}{}
}

func (styles CSSStyles) String() string {
	builder := strings.Builder{}
	for p := range styles {
		builder.WriteString(p.Key.Value())
		builder.WriteString(": ")
		builder.WriteString(p.Value.Value())
		builder.WriteString("\n")
	}
	return builder.String()
}

func (styles CSSStyles) FromMatched(matched *css.GetMatchedStylesForNodeReturns) {
	// inherited styles
	for _, inherited := range matched.Inherited {
		for _, rule := range inherited.MatchedCSSRules {
			if rule.Rule.Origin == css.StyleSheetOriginUserAgent {
				continue
			}
			for _, prop := range rule.Rule.Style.CSSProperties {
				styles.Set(prop)
			}
		}
		for _, prop := range matched.InlineStyle.CSSProperties {
			styles.Set(prop)
		}
	}

	/*
		the following types of styles are excluded (from least priority to highest priority) because:
		- inherited pseudo elements rules : minimal impact on element's actual style
		- property registration rules     : don't affect styling
		- position try rules              : don't exist officially yet
		- font palette rules              : rarely have any impact
		- function rules                  : don't exist officially yet
		- keyframes rules                 : only affect elements when animations are running
		- property rules                  : only deal with animations and custom properties
		- pseudo element styles           : have a minimal impact on the element's actual style
	*/

	// directly matched styles
	for _, rule := range matched.MatchedCSSRules {
		if rule.Rule.Origin == css.StyleSheetOriginUserAgent {
			continue
		}
		for _, prop := range rule.Rule.Style.CSSProperties {
			styles.Set(prop)
		}
	}

	// attribute set style (ex. <svg width="20px" ... />)
	if matched.AttributesStyle != nil {
		for _, prop := range matched.AttributesStyle.CSSProperties {
			styles.Set(prop)
		}
	}

	// inline styles
	if matched.InlineStyle != nil {
		for _, prop := range matched.InlineStyle.CSSProperties {
			styles.Set(prop)
		}
	}
}

func (styles CSSStyles) Hash() uint64 {
	entries := make([]CSSProp, 0, len(styles))
	for k := range styles {
		entries = append(entries, k)
	}

	slices.SortFunc(entries, func(a, b CSSProp) int {
		ak := a.Key.Value()
		bk := a.Key.Value()
		if ak > bk {
			return 1
		}
		if ak < bk {
			return -1
		}
		return 0
	})

	buff := bytes.NewBuffer(nil)
	for _, e := range entries {
		buff.WriteString(e.Key.Value())
		buff.WriteString(e.Value.Value())
	}
	return xxh3.Hash(buff.Bytes())
}

type CSSTree struct {
	// FromHash map[uint64]*CSSNode
	Root *CSSNode
}

func NewCSSTree() *CSSTree {
	return &CSSTree{}
}

type CSSNode struct {
	Styles      CSSStyles
	NextSibling *CSSNode
	FirstChild  *CSSNode
}

/*
note:
	in this context, "disjoint" does not mean (no props in common), just that
	each contains props not present in the other

3 cases:
- new node is child of existing node:
	call recursively with first child, go to next sibling and check if child or
	disjoint
- new node is disjoint with existing node:
	[check for children of new node inside children of existing node] and move
	to the next sibling
- new node is parent of existing node:
	gather all the children of the new node on this layer, put them under the
	new node (while removing them from the old parent) then put parent as new
	sibling to disjoint nodes to the new node, or first child of original
	parent of children

theorem:
	siblings must be disjoint with one another.

theorem:
	there is no case where a new node is parent of some nodes and child of some
	other nodes in the same layer.

3 cases:
- disjoint with every node in the current layer:
	add as sibling to the current layer
- has some children in the current layer:
	move children under new node and add new node as sibling to current layer
- is child of some node(s) in the current layer:
	call recursively on each node
*/

type relationship uint8

const (
	rel_equal      relationship = iota // if a == b (exact same properties)
	rel_parent_of                      // if a > b (a has all the props of b, and then some)
	rel_child_of                       // if a < b (b has all the props of a, and then some)
	rel_sibling_of                     // if a != b (a and b both have props that are not in the other)
)

// relationshipOf returns the relationship of other to target.
func (t CSSTree) relationshipOf(a, b CSSStyles) relationship {
	allBInA := true
	for aprop := range a {
		_, existsInB := b[aprop]
		if !existsInB {
			allBInA = false
			break
		}
	}

	allAInB := true
	for bprop := range b {
		_, existsInB := a[bprop]
		if !existsInB {
			allAInB = false
			break
		}
	}

	if allBInA && allAInB {
		return rel_equal
	}
	if allAInB {
		return rel_child_of
	}
	if allBInA {
		return rel_parent_of
	}
	return rel_sibling_of
}

func (t CSSTree) addInner(layer, newNode *CSSNode) *CSSNode {
	if layer == nil {
		return newNode
	}

	var prev *CSSNode
	isChildOf := false

	cur := layer
	for cur != nil {
		rel := t.relationshipOf(newNode.Styles, cur.Styles)

		switch rel {
		case rel_sibling_of:
			prev = cur
			cur = cur.NextSibling
			continue
		case rel_parent_of:
			next := cur.NextSibling

			// remove cur from layer
			if prev != nil {
				prev.NextSibling = next
			} else {
				// we must make layer always point to the first node in the
				// layer, in this case if prev == nil and we remove cur, that
				// means we must move layer up one to cur's next sibling
				layer = next
			}

			// add cur as new node's child
			if newNode.FirstChild == nil {
				cur.NextSibling = nil
				newNode.FirstChild = cur
			} else {
				cur.NextSibling = newNode.FirstChild
				newNode.FirstChild = cur
			}

			cur = next
			continue
		case rel_child_of:
			cur.FirstChild = t.addInner(cur.FirstChild, newNode)
			isChildOf = true
		case rel_equal:
			return layer
		}

		prev = cur
		cur = cur.NextSibling
	}

	// if newNode is a parent of some nodes in the current layer this means it
	// must be the sibling of the other nodes in the current layer.
	if newNode.FirstChild != nil {
		if layer == nil {
			return newNode
		}
		next := layer.NextSibling
		layer.NextSibling = newNode
		newNode.NextSibling = next
		return layer
	}

	// if newNode is a child of some nodes and a sibling to all the other
	// nodes, simply return the layer as-is, as no nodes in the current layer
	// will have been added/removed
	if isChildOf {
		return layer
	}

	// if newNode is not a child of any nodes but a sibling to all the other
	// nodes, add the new node to the current layer.
	next := layer.NextSibling
	layer.NextSibling = newNode
	newNode.NextSibling = next

	return layer
}

func (t *CSSTree) Add(styles CSSStyles) {
	newNode := &CSSNode{
		Styles: styles,
	}

	// hash := newNode.Styles.Hash()
	//
	// _, exists := t.FromHash[hash]
	// if exists {
	// 	return
	// }

	// if t.Root == nil {
	// 	t.FromHash[hash] = newNode
	// 	t.Root = newNode
	// 	return
	// }

	t.Root = t.addInner(t.Root, newNode)
}
