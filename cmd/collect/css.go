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
