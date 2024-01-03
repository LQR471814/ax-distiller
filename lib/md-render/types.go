package mdrender

type Node interface {
	node()
}

type CanBeDecorated interface {
	Node
	canBeDecorated()
}

type ParagraphElement interface {
	Node
	paragraphElement()
}

type ListItem interface {
	Node
	listItem()
}

type Header struct {
	Order   int
	Content CanBeDecorated
}

func (Header) node() {}

type LineBreak struct{}

func (LineBreak) canBeDecorated()   {}
func (LineBreak) paragraphElement() {}
func (LineBreak) node()             {}

type Paragraph struct {
	Elements []ParagraphElement
}

func (Paragraph) node()           {}
func (Paragraph) canBeDecorated() {}
func (Paragraph) listItem()       {}

type Link struct {
	URL string
	// Note: this technically means a Link can go here, however
	// this will result in a silent conversion of the link to.
	// whatever is in the title of the Link.
	Title CanBeDecorated
	Image bool
}

func (Link) node()             {}
func (Link) canBeDecorated()   {}
func (Link) paragraphElement() {}
func (Link) listItem()         {}

type PlainText struct {
	Content string
}

func (PlainText) node()             {}
func (PlainText) canBeDecorated()   {}
func (PlainText) paragraphElement() {}
func (PlainText) listItem()         {}

type Decoration = int

const (
	DECOR_ITALIC Decoration = iota
	DECOR_BOLD
	DECOR_UNDERLINE
)

type DecoratedText struct {
	Content CanBeDecorated
	Type    Decoration
}

func (DecoratedText) node()             {}
func (DecoratedText) canBeDecorated()   {}
func (DecoratedText) listItem()         {}
func (DecoratedText) paragraphElement() {}

type ListType = int

const (
	LIST_ORDERED ListType = iota
	LIST_UNORDERED
)

type List struct {
	Items []ListItem
	Type  ListType
}

func (List) node()     {}
func (List) listItem() {}

type InlineCode struct {
	Content string
}

func (InlineCode) node()             {}
func (InlineCode) paragraphElement() {}
func (InlineCode) canBeDecorated()   {}
func (InlineCode) listItem()         {}

type BlockCode struct {
	Language string
	Content  string
}

func (BlockCode) node() {}

type HorizontalRule struct{}

func (HorizontalRule) node() {}

type Table struct {
	Rows [][]CanBeDecorated
}

func (Table) node() {}
