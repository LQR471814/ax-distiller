package mdrender

import (
	"testing"
)

func TestRender(t *testing.T) {
	table := []struct {
		tree     []Node
		expected string
	}{
		{
			tree: []Node{
				Header{
					Order: 1,
					Content: []Node{PlainText{
						Content: "Dog doc",
					}},
				},
				BlockCode{
					Language: "typescript",
					Content:  "console.log('pain')",
				},
				DecoratedText{
					Type: DECOR_BOLD,
					Content: []Node{
						PlainText{
							Content: "  Dogs are funny creatures.",
						},
						DecoratedText{
							Type: DECOR_ITALIC,
							Content: []Node{PlainText{
								Content: "    They love to swim.  ",
							}},
						},
						PlainText{
							Content: " But only sometimes.",
						},
					},
				},
				List{
					Type: LIST_ORDERED,
					Items: [][]Node{{
						DecoratedText{
							Type: DECOR_UNDERLINE,
							Content: []Node{Link{
								URL:   "https://google.com",
								Title: []Node{PlainText{"Let me google that for you."}},
							}},
						},
						InlineCode{
							Content: "$ echo \"that's not a good thing\"",
						},
						List{
							Type: LIST_UNORDERED,
							Items: [][]Node{{
								PlainText{
									Content: "these are plain list items",
								},
								List{
									Type: LIST_ORDERED,
									Items: [][]Node{{
										PlainText{
											Content: "Finally,",
										},
										PlainText{
											Content: "Another list between items.",
										},
									}},
								},
								Link{
									URL:   "https://www.google.com",
									Title: []Node{PlainText{"Sometimes you need to add www."}},
								},
							}},
						},
					}},
				},
				PlainText{
					Content: "Lorem ipsum, whatever, this is text.",
				},
				PlainText{
					Content: "Next to text, cause why not.",
				},
				DecoratedText{
					Content: []Node{
						Link{
							Title: []Node{DecoratedText{
								Content: []Node{InlineCode{
									Content: "wikipedia.org",
								}},
								Type: DECOR_BOLD,
							}},
							URL: "https://wikipedia.org",
						},
						LineBreak{},
						Link{
							Title: []Node{
								InlineCode{
									Content: "Painful",
								},
								LineBreak{},
								Link{
									Title: []Node{DecoratedText{
										Content: []Node{PlainText{"This is the only thing that should appear."}},
									}},
									URL:   "https://this.should.not.be.here",
									Image: true,
								},
							},
							URL:   "https://imgs.xkcd.com/comics/standards.png",
							Image: true,
						},
					},
				},
				Table{
					Rows: [][][]Node{
						{
							{PlainText{"header 1"}},
							{PlainText{"header 2"}},
							{PlainText{"header 3"}},
						},
						{
							{
								InlineCode{"wow i love coding"},
								LineBreak{},
								Link{
									URL:   "https://github.com",
									Title: []Node{PlainText{"Find the best codes here"}},
								},
							},
							{PlainText{"body 2"}},
							{
								PlainText{"body 4"},
								PlainText{"body 4"},
								PlainText{"body 4"},
								PlainText{"body 4"},
							},
						},
					},
				},
			},
		},
	}

	for i, testCase := range table {
		result := Render(testCase.tree)
		if result != testCase.expected {
			t.Errorf(
				"Got =====\n%sExpected =====\n%s\nFor case %d\n",
				result, testCase.expected, i,
			)
		}
	}
}
