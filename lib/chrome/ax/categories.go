package ax

import "unique"

type Role = unique.Handle[string]

type Category int

const (
	CATEGORY_DIALOG Category = iota
	CATEGORY_WIDGET
	CATEGORY_DOCUMENT
	CATEGORY_LANDMARK
	CATEGORY_STRUCTURE
	CATEGORY_SECTION
	CATEGORY_SECTIONHEAD
	CATEGORY_GENERIC
)

var Categories = make(map[Role][]Category)

func init() {
	Categories[unique.Make("math")] = []Category{CATEGORY_SECTION}
	Categories[unique.Make("slider")] = []Category{CATEGORY_WIDGET, CATEGORY_STRUCTURE}
	Categories[unique.Make("progressbar")] = []Category{CATEGORY_WIDGET, CATEGORY_STRUCTURE}
	Categories[unique.Make("comment")] = []Category{CATEGORY_DOCUMENT}
	Categories[unique.Make("navigation")] = []Category{CATEGORY_LANDMARK}
	Categories[unique.Make("sectionhead")] = []Category{CATEGORY_SECTIONHEAD}
	Categories[unique.Make("subscript")] = []Category{CATEGORY_SECTION}
	Categories[unique.Make("paragraph")] = []Category{CATEGORY_SECTION}
	Categories[unique.Make("region")] = []Category{CATEGORY_LANDMARK}
	Categories[unique.Make("listbox")] = []Category{CATEGORY_WIDGET, CATEGORY_SECTION}
	Categories[unique.Make("form")] = []Category{CATEGORY_LANDMARK}
	Categories[unique.Make("window")] = []Category{}
	Categories[unique.Make("caption")] = []Category{CATEGORY_SECTION}
	Categories[unique.Make("mark")] = []Category{CATEGORY_SECTION}
	Categories[unique.Make("tree")] = []Category{CATEGORY_WIDGET, CATEGORY_SECTION}
	Categories[unique.Make("toolbar")] = []Category{CATEGORY_SECTION}
	Categories[unique.Make("application")] = []Category{CATEGORY_STRUCTURE}
	Categories[unique.Make("complementary")] = []Category{CATEGORY_LANDMARK}
	Categories[unique.Make("suggestion")] = []Category{CATEGORY_SECTION}
	Categories[unique.Make("superscript")] = []Category{CATEGORY_SECTION}
	Categories[unique.Make("generic")] = []Category{CATEGORY_GENERIC}
	Categories[unique.Make("listitem")] = []Category{CATEGORY_SECTION}
	Categories[unique.Make("deletion")] = []Category{CATEGORY_SECTION}
	Categories[unique.Make("menubar")] = []Category{CATEGORY_WIDGET, CATEGORY_SECTION}
	Categories[unique.Make("tablist")] = []Category{CATEGORY_WIDGET}
	Categories[unique.Make("list")] = []Category{CATEGORY_SECTION}
	Categories[unique.Make("directory")] = []Category{CATEGORY_SECTION}
	Categories[unique.Make("composite")] = []Category{CATEGORY_WIDGET}
	Categories[unique.Make("treeitem")] = []Category{CATEGORY_WIDGET, CATEGORY_SECTION}
	Categories[unique.Make("columnheader")] = []Category{CATEGORY_WIDGET, CATEGORY_SECTION, CATEGORY_SECTIONHEAD}
	Categories[unique.Make("alertdialog")] = []Category{CATEGORY_DIALOG, CATEGORY_SECTION}
	Categories[unique.Make("document")] = []Category{CATEGORY_DOCUMENT}
	Categories[unique.Make("separator")] = []Category{CATEGORY_WIDGET, CATEGORY_STRUCTURE}
	Categories[unique.Make("status")] = []Category{CATEGORY_SECTION}
	Categories[unique.Make("select")] = []Category{CATEGORY_WIDGET, CATEGORY_SECTION}
	Categories[unique.Make("checkbox")] = []Category{CATEGORY_WIDGET}
	Categories[unique.Make("cell")] = []Category{CATEGORY_SECTION}
	Categories[unique.Make("tabpanel")] = []Category{CATEGORY_SECTION}
	Categories[unique.Make("note")] = []Category{CATEGORY_SECTION}
	Categories[unique.Make("associationlist")] = []Category{CATEGORY_SECTION}
	Categories[unique.Make("roletype")] = []Category{}
	Categories[unique.Make("article")] = []Category{CATEGORY_DOCUMENT}
	Categories[unique.Make("heading")] = []Category{CATEGORY_SECTIONHEAD}
	Categories[unique.Make("rowgroup")] = []Category{CATEGORY_STRUCTURE}
	Categories[unique.Make("insertion")] = []Category{CATEGORY_SECTION}
	Categories[unique.Make("range")] = []Category{CATEGORY_STRUCTURE}
	Categories[unique.Make("landmark")] = []Category{CATEGORY_LANDMARK}
	Categories[unique.Make("banner")] = []Category{CATEGORY_LANDMARK}
	Categories[unique.Make("input")] = []Category{CATEGORY_WIDGET}
	Categories[unique.Make("textbox")] = []Category{CATEGORY_WIDGET}
	Categories[unique.Make("menuitem")] = []Category{CATEGORY_WIDGET}
	Categories[unique.Make("associationlistitemvalue")] = []Category{CATEGORY_SECTION}
	Categories[unique.Make("scrollbar")] = []Category{CATEGORY_WIDGET, CATEGORY_STRUCTURE}
	Categories[unique.Make("presentation")] = []Category{CATEGORY_STRUCTURE}
	Categories[unique.Make("main")] = []Category{CATEGORY_LANDMARK}
	Categories[unique.Make("figure")] = []Category{CATEGORY_SECTION}
	Categories[unique.Make("combobox")] = []Category{CATEGORY_WIDGET}
	Categories[unique.Make("tab")] = []Category{CATEGORY_WIDGET, CATEGORY_SECTIONHEAD}
	Categories[unique.Make("search")] = []Category{CATEGORY_LANDMARK}
	Categories[unique.Make("row")] = []Category{CATEGORY_WIDGET, CATEGORY_SECTION}
	Categories[unique.Make("feed")] = []Category{CATEGORY_SECTION}
	Categories[unique.Make("code")] = []Category{CATEGORY_SECTION}
	Categories[unique.Make("option")] = []Category{CATEGORY_WIDGET}
	Categories[unique.Make("switch")] = []Category{CATEGORY_WIDGET}
	Categories[unique.Make("emphasis")] = []Category{CATEGORY_SECTION}
	Categories[unique.Make("tooltip")] = []Category{CATEGORY_SECTION}
	Categories[unique.Make("structure")] = []Category{CATEGORY_STRUCTURE}
	Categories[unique.Make("contentinfo")] = []Category{CATEGORY_LANDMARK}
	Categories[unique.Make("strong")] = []Category{CATEGORY_SECTION}
	Categories[unique.Make("term")] = []Category{CATEGORY_SECTION}
	Categories[unique.Make("menuitemcheckbox")] = []Category{CATEGORY_WIDGET}
	Categories[unique.Make("menuitemradio")] = []Category{CATEGORY_WIDGET}
	Categories[unique.Make("meter")] = []Category{CATEGORY_STRUCTURE}
	Categories[unique.Make("rowheader")] = []Category{CATEGORY_WIDGET, CATEGORY_SECTION, CATEGORY_SECTIONHEAD}
	Categories[unique.Make("table")] = []Category{CATEGORY_SECTION}
	Categories[unique.Make("button")] = []Category{CATEGORY_WIDGET}
	Categories[unique.Make("section")] = []Category{CATEGORY_SECTION}
	Categories[unique.Make("treegrid")] = []Category{CATEGORY_WIDGET, CATEGORY_SECTION}
	Categories[unique.Make("group")] = []Category{CATEGORY_SECTION}
	Categories[unique.Make("widget")] = []Category{CATEGORY_WIDGET}
	Categories[unique.Make("searchbox")] = []Category{CATEGORY_WIDGET}
	Categories[unique.Make("command")] = []Category{CATEGORY_WIDGET}
	Categories[unique.Make("radio")] = []Category{CATEGORY_WIDGET}
	Categories[unique.Make("link")] = []Category{CATEGORY_WIDGET}
	Categories[unique.Make("dialog")] = []Category{CATEGORY_DIALOG}
	Categories[unique.Make("associationlistitemkey")] = []Category{CATEGORY_SECTION}
	Categories[unique.Make("grid")] = []Category{CATEGORY_WIDGET, CATEGORY_SECTION}
	Categories[unique.Make("timer")] = []Category{CATEGORY_SECTION}
	Categories[unique.Make("log")] = []Category{CATEGORY_SECTION}
	Categories[unique.Make("alert")] = []Category{CATEGORY_SECTION}
	Categories[unique.Make("img")] = []Category{CATEGORY_SECTION}
	Categories[unique.Make("radiogroup")] = []Category{CATEGORY_WIDGET, CATEGORY_SECTION}
	Categories[unique.Make("menu")] = []Category{CATEGORY_WIDGET, CATEGORY_SECTION}
	Categories[unique.Make("definition")] = []Category{CATEGORY_SECTION}
	Categories[unique.Make("blockquote")] = []Category{CATEGORY_SECTION}
	Categories[unique.Make("spinbutton")] = []Category{CATEGORY_WIDGET, CATEGORY_STRUCTURE}
	Categories[unique.Make("gridcell")] = []Category{CATEGORY_WIDGET, CATEGORY_SECTION}
	Categories[unique.Make("time")] = []Category{CATEGORY_SECTION}
	Categories[unique.Make("marquee")] = []Category{CATEGORY_SECTION}
}
