package mdrender

type NodeList = []Node

type Node interface {
	node()
}

type Header struct {
	Order   int
	Content NodeList
}

type HorizontalRule struct{}

type LineBreak struct{}

type Link struct {
	URL string
	// Note: this technically means a Link can go here, however
	// this will result in a silent conversion of the link to.
	// whatever is in the title of the Link.
	Title NodeList
	Image bool
}

type Paragraph struct {
	Content NodeList
}

type PlainText struct {
	Content string
}

type Decoration = uint8

const (
	DECOR_ITALIC Decoration = iota
	DECOR_BOLD
	DECOR_UNDERLINE
)

type DecoratedText struct {
	Content NodeList
	Type    Decoration
}

type ListType = uint8

const (
	LIST_ORDERED ListType = iota
	LIST_UNORDERED
)

type List struct {
	Items []NodeList
	Type  ListType
}

type InlineCode struct {
	Content string
}

type BlockCode struct {
	Language string
	Content  string
}

type Table struct {
	Rows [][]NodeList
}

func (Header) node()         {}
func (HorizontalRule) node() {}
func (LineBreak) node()      {}
func (Link) node()           {}
func (PlainText) node()      {}
func (DecoratedText) node()  {}
func (List) node()           {}
func (InlineCode) node()     {}
func (BlockCode) node()      {}
func (Table) node()          {}
func (Paragraph) node()      {}
