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
					Content: PlainText{
						Content: "Dog doc",
					},
				},
				BlockCode{
					Language: "typescript",
					Content:  "console.log('pain')",
				},
				Paragraph{
					Elements: []ParagraphElement{DecoratedText{
						Type: DECOR_BOLD,
						Content: Paragraph{
							Elements: []ParagraphElement{
								PlainText{
									Content: "  Dogs are funny creatures.",
								},
								DecoratedText{
									Type: DECOR_ITALIC,
									Content: PlainText{
										Content: "    They love to swim.  ",
									},
								},
								PlainText{
									Content: " But only sometimes.",
								},
							},
						},
					}},
				},
				List{
					Type: LIST_ORDERED,
					Items: []ListItem{
						DecoratedText{
							Type: DECOR_UNDERLINE,
							Content: Link{
								URL:   "https://google.com",
								Title: PlainText{"Let me google that for you."},
							},
						},
						InlineCode{
							Content: "$ echo \"that's not a good thing\"",
						},
						List{
							Type: LIST_UNORDERED,
							Items: []ListItem{
								PlainText{
									Content: "these are plain list items",
								},
								List{
									Type: LIST_ORDERED,
									Items: []ListItem{
										PlainText{
											Content: "Finally,",
										},
										PlainText{
											Content: "Another list between items.",
										},
									},
								},
								Link{
									URL:   "https://www.google.com",
									Title: PlainText{"Sometimes you need to add www."},
								},
							},
						},
					},
				},
				Paragraph{
					Elements: []ParagraphElement{
						PlainText{
							Content: "Lorem ipsum, whatever, this is text.",
						},
						PlainText{
							Content: "Next to text, cause why not.",
						},
						DecoratedText{
							Content: Paragraph{
								Elements: []ParagraphElement{
									Link{
										Title: DecoratedText{
											Content: InlineCode{
												Content: "wikipedia.org",
											},
											Type: DECOR_BOLD,
										},
										URL: "https://wikipedia.org",
									},
									LineBreak{},
									Link{
										Title: Paragraph{
											Elements: []ParagraphElement{
												InlineCode{
													Content: "Painful",
												},
												LineBreak{},
												Link{
													Title: DecoratedText{
														Content: PlainText{"This is the only thing that should appear."},
													},
													URL:   "https://this.should.not.be.here",
													Image: true,
												},
											},
										},
										URL:   "https://imgs.xkcd.com/comics/standards.png",
										Image: true,
									},
								},
							},
						},
					},
				},
				Table{
					Rows: [][]CanBeDecorated{
						{
							PlainText{"header 1"},
							PlainText{"header 2"},
							PlainText{"header 3"},
						},
						{
							Paragraph{
								Elements: []ParagraphElement{
									InlineCode{"wow i love coding"},
									LineBreak{},
									Link{
										URL:   "https://github.com",
										Title: PlainText{"Find the best codes here"},
									},
								},
							},
							PlainText{"body 2"},
						},
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
