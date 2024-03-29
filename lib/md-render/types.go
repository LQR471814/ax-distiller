package mdrender

type NodeFlag = uint8

const (
	FLAG_DECORABLE NodeFlag = 0b00000001
	FLAG_LIST_ITEM          = 0b00000010
)

const (
	NODE_FLAGS_HEADER          NodeFlag = 0
	NODE_FLAGS_HORIZONTAL_RULE          = 0
	NODE_FLAGS_LINE_BREAK               = 0
	NODE_FLAGS_LINK                     = FLAG_DECORABLE | FLAG_LIST_ITEM
	NODE_FLAGS_PLAIN_TEXT               = FLAG_DECORABLE | FLAG_LIST_ITEM
	NODE_FLAGS_DECORATED_TEXT           = FLAG_DECORABLE | FLAG_LIST_ITEM
	NODE_FLAGS_LIST                     = FLAG_LIST_ITEM
	NODE_FLAGS_INLINE_CODE              = FLAG_DECORABLE | FLAG_LIST_ITEM
	NODE_FLAGS_BLOCK_CODE               = 0
	NODE_FLAGS_TABLE                    = 0
)

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
