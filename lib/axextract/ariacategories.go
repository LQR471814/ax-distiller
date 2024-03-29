package axextract

type Category = int

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

var RoleCategoryMap = map[string][]Category{
	"math":                     {CATEGORY_SECTION},
	"slider":                   {CATEGORY_WIDGET, CATEGORY_STRUCTURE},
	"progressbar":              {CATEGORY_WIDGET, CATEGORY_STRUCTURE},
	"comment":                  {CATEGORY_DOCUMENT},
	"navigation":               {CATEGORY_LANDMARK},
	"sectionhead":              {CATEGORY_SECTIONHEAD},
	"subscript":                {CATEGORY_SECTION},
	"paragraph":                {CATEGORY_SECTION},
	"region":                   {CATEGORY_LANDMARK},
	"listbox":                  {CATEGORY_WIDGET, CATEGORY_SECTION},
	"form":                     {CATEGORY_LANDMARK},
	"window":                   {},
	"caption":                  {CATEGORY_SECTION},
	"mark":                     {CATEGORY_SECTION},
	"tree":                     {CATEGORY_WIDGET, CATEGORY_SECTION},
	"toolbar":                  {CATEGORY_SECTION},
	"application":              {CATEGORY_STRUCTURE},
	"complementary":            {CATEGORY_LANDMARK},
	"suggestion":               {CATEGORY_SECTION},
	"superscript":              {CATEGORY_SECTION},
	"generic":                  {CATEGORY_GENERIC},
	"listitem":                 {CATEGORY_SECTION},
	"deletion":                 {CATEGORY_SECTION},
	"menubar":                  {CATEGORY_WIDGET, CATEGORY_SECTION},
	"tablist":                  {CATEGORY_WIDGET},
	"list":                     {CATEGORY_SECTION},
	"directory":                {CATEGORY_SECTION},
	"composite":                {CATEGORY_WIDGET},
	"treeitem":                 {CATEGORY_WIDGET, CATEGORY_SECTION},
	"columnheader":             {CATEGORY_WIDGET, CATEGORY_SECTION, CATEGORY_SECTIONHEAD},
	"alertdialog":              {CATEGORY_DIALOG, CATEGORY_SECTION},
	"document":                 {CATEGORY_DOCUMENT},
	"separator":                {CATEGORY_WIDGET, CATEGORY_STRUCTURE},
	"status":                   {CATEGORY_SECTION},
	"select":                   {CATEGORY_WIDGET, CATEGORY_SECTION},
	"checkbox":                 {CATEGORY_WIDGET},
	"cell":                     {CATEGORY_SECTION},
	"tabpanel":                 {CATEGORY_SECTION},
	"note":                     {CATEGORY_SECTION},
	"associationlist":          {CATEGORY_SECTION},
	"roletype":                 {},
	"article":                  {CATEGORY_DOCUMENT},
	"heading":                  {CATEGORY_SECTIONHEAD},
	"rowgroup":                 {CATEGORY_STRUCTURE},
	"insertion":                {CATEGORY_SECTION},
	"range":                    {CATEGORY_STRUCTURE},
	"landmark":                 {CATEGORY_LANDMARK},
	"banner":                   {CATEGORY_LANDMARK},
	"input":                    {CATEGORY_WIDGET},
	"textbox":                  {CATEGORY_WIDGET},
	"menuitem":                 {CATEGORY_WIDGET},
	"associationlistitemvalue": {CATEGORY_SECTION},
	"scrollbar":                {CATEGORY_WIDGET, CATEGORY_STRUCTURE},
	"presentation":             {CATEGORY_STRUCTURE},
	"main":                     {CATEGORY_LANDMARK},
	"figure":                   {CATEGORY_SECTION},
	"combobox":                 {CATEGORY_WIDGET},
	"tab":                      {CATEGORY_WIDGET, CATEGORY_SECTIONHEAD},
	"search":                   {CATEGORY_LANDMARK},
	"row":                      {CATEGORY_WIDGET, CATEGORY_SECTION},
	"feed":                     {CATEGORY_SECTION},
	"code":                     {CATEGORY_SECTION},
	"option":                   {CATEGORY_WIDGET},
	"switch":                   {CATEGORY_WIDGET},
	"emphasis":                 {CATEGORY_SECTION},
	"tooltip":                  {CATEGORY_SECTION},
	"structure":                {CATEGORY_STRUCTURE},
	"contentinfo":              {CATEGORY_LANDMARK},
	"strong":                   {CATEGORY_SECTION},
	"term":                     {CATEGORY_SECTION},
	"menuitemcheckbox":         {CATEGORY_WIDGET},
	"menuitemradio":            {CATEGORY_WIDGET},
	"meter":                    {CATEGORY_STRUCTURE},
	"rowheader":                {CATEGORY_WIDGET, CATEGORY_SECTION, CATEGORY_SECTIONHEAD},
	"table":                    {CATEGORY_SECTION},
	"button":                   {CATEGORY_WIDGET},
	"section":                  {CATEGORY_SECTION},
	"treegrid":                 {CATEGORY_WIDGET, CATEGORY_SECTION},
	"group":                    {CATEGORY_SECTION},
	"widget":                   {CATEGORY_WIDGET},
	"searchbox":                {CATEGORY_WIDGET},
	"command":                  {CATEGORY_WIDGET},
	"radio":                    {CATEGORY_WIDGET},
	"link":                     {CATEGORY_WIDGET},
	"dialog":                   {CATEGORY_DIALOG},
	"associationlistitemkey":   {CATEGORY_SECTION},
	"grid":                     {CATEGORY_WIDGET, CATEGORY_SECTION},
	"timer":                    {CATEGORY_SECTION},
	"log":                      {CATEGORY_SECTION},
	"alert":                    {CATEGORY_SECTION},
	"img":                      {CATEGORY_SECTION},
	"radiogroup":               {CATEGORY_WIDGET, CATEGORY_SECTION},
	"menu":                     {CATEGORY_WIDGET, CATEGORY_SECTION},
	"definition":               {CATEGORY_SECTION},
	"blockquote":               {CATEGORY_SECTION},
	"spinbutton":               {CATEGORY_WIDGET, CATEGORY_STRUCTURE},
	"gridcell":                 {CATEGORY_WIDGET, CATEGORY_SECTION},
	"time":                     {CATEGORY_SECTION},
	"marquee":                  {CATEGORY_SECTION},
}