package dcdmaker

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"
)

const (
	twipsPerInch = 1440
	halfPtPerPt  = 2
)

type ContentType int

const (
	ContentParagraph ContentType = iota
	ContentTable
)

type ContentItem struct {
	Type      ContentType
	Paragraph *ParsedParagraph
	Table     *ParsedTable
}

type ParsedTable struct {
	ID           int
	Grid         []float64
	Rows         []ParsedTableRow
	HasBorders   bool
	BorderTop    *BorderInfo
	BorderBottom *BorderInfo
	BorderLeft   *BorderInfo
	BorderRight  *BorderInfo
	Width        float64
	Alignment    string
	Caption      string
	Summary      string
	Indent       float64
	CellSpacing  float64
	StyleName    string
}

type ParsedTableRow struct {
	IsHeader bool
	Cells    []ParsedTableCell
}

type ParsedTableCell struct {
	Content      []ContentItem
	GridSpan     int
	RowSpan      int
	VMerge       string
	ShadingFill  string
	BorderTop    *BorderInfo
	BorderBottom *BorderInfo
	BorderLeft   *BorderInfo
	BorderRight  *BorderInfo
	VAlign       string
	TextDirection string
	NoWrap       bool
}

type ParsedHdrFtr struct {
	Type    string        // "default", "first", "even"
	Content []ContentItem
}

type ParsedSection struct {
	Layout       PageLayout
	Headers      []ParsedHdrFtr
	Footers      []ParsedHdrFtr
	Content      []ContentItem
	BreakType    string
	ColCount     int
	ColSpace     float64
	PageNumFmt   string
	PageNumStart int
}

type ParsedDocument struct {
	PageLayout    PageLayout
	DefaultFont   FontDef
	LineHeight    float64
	Title         string
	Subject       string
	Author        string
	Keywords      string
	Description   string
	Category      string
	ContentStatus string
	LastModifiedBy string
	Revision      string
	Version       string
	Created       string
	Modified      string
	Language      string
	Application   string
	AppVersion    string
	HeadingStyles map[int]StyleDef
	Content       []ContentItem
	Sections      []ParsedSection
	AllTables     []ParsedTable
	Notes         []NoteItem
	CustomStyles  []CustomStyleDef
	Mode          string // "semantic" (default) or "lossless"
	Theme         *ThemeData
}

type CustomStyleDef struct {
	Name    string
	Type    string
	BasedOn string
	StyleDef
}

type BasedOnStyle struct {
	ID   string
	Name string
}

type NoteItem struct {
	Type   string // "footnote", "endnote", "bookmark", "comment"
	ID     int
	Name   string // bookmark name
	Author string // comment author
	Date   string // comment date
	Body   []ContentItem
}

type PageLayout struct {
	WidthInch    float64
	HeightInch   float64
	MarginTop    float64
	MarginRight  float64
	MarginBottom float64
	MarginLeft   float64
	HeaderMargin float64
	FooterMargin float64
	FromDocx     bool
}

type FontDef struct {
	Family  string
	SizePt  float64
	Color   string
	FromDocx bool
}

type TextRun struct {
	Text        string
	Bold        bool
	Italic      bool
	Underline   string
	Strike      bool
	DStrike     bool
	SuperScript bool
	SubScript   bool
	FontFamily  string
	FontSizePt  float64
	FontColor   string
	Highlight   string
	SmallCaps   bool
	AllCaps     bool
	Hidden      bool
	CharSpacing    int
	Position       int
	Language       string
	Emphasis       string
	RTL            bool
	NoProof        bool
	CS             bool
	SpecVanish     bool
	Emboss         bool
	Engrave        bool
	Shadow         bool
	Imprint        bool
	Border         string
	Effect         bool
	Animate        bool
	BCs            bool
	ICs            bool
	FontEA         string
	FontCS         string
	SizeCS         float64
	BreakType      string
	IsTab          bool
	IsLineBreak    bool
	IsPageBreak    bool
	IsImage        bool
	ImageSrc       string
	ImageWidth     float64
	ImageHeight    float64
	ImageAlt       string
	IsField        bool
	FieldType      string
	FieldFormat    string
	FieldName      string
	IsHyperlink    bool
	HyperlinkURL   string
	IsFootnoteRef  bool
	IsEndnoteRef   bool
	NoteID         int
	TextBoxContent []ContentItem
	IsIns          bool
	IsDel          bool
	InsID          int
	InsAuthor      string
	InsDate        string
}

type BorderInfo struct {
	Val   string
	Sz    int
	Space int
	Color string
}

type ParsedParagraph struct {
	Runs               []TextRun
	Align              string
	IndentLeft         float64
	IndentRight        float64
	FirstLineIndent    float64
	Hanging            float64
	SpacingBefore      float64
	SpacingAfter       float64
	LineHeight         float64
	Bold               bool
	Italic             bool
	FontFamily         string
	FontSizePt         float64
	FontColor          string
	HeadingLevel       int
	IsList             bool
	ListLevel          int
	ListFormat         string
	ListStartOverride  int
	StyleID            string
	StyleName          string
	KeepNext           bool
	KeepLines          bool
	WidowControl       bool
	ContextualSpacing  bool
	SuppressLineNumbers bool
	SuppressHyphenation bool
	TextDirection      string
	BorderTop          *BorderInfo
	BorderBottom       *BorderInfo
	BorderLeft         *BorderInfo
	BorderRight        *BorderInfo
	ShadingFill        string
	PageBreakBefore    bool
	Kinsoku            bool
	WordWrap           bool
	CharSpacingJust    int
	TwoLineOne         bool
	AutoSpaceDE        bool
	AutoSpaceDN        bool
	Bidi               bool
	IsCode             bool
	IsQuote            bool
	Lang               string
	TabStops           []TabStopDef
}

type TabStopDef struct {
	Pos    float64
	Align  string
	Leader string
}

type docStyles struct {
	XMLName   xml.Name     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main styles"`
	DocDefaults *docDefaults `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main docDefaults"`
	Styles     []styleDef  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main style"`
}

type docDefaults struct {
	RunPropsDefault *runPropsDefault   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main rPrDefault"`
	ParaPropsDefault *paraPropsDefault `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main pPrDefault"`
}

type paraPropsDefault struct {
	ParaProps *paraProps `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main pPr"`
}

type spacingVal struct {
	Before   int    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main before,attr"`
	After    int    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main after,attr"`
	Line     int    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main line,attr"`
	LineRule string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main lineRule,attr"`
}

type runPropsDefault struct {
	RunProps *runProps `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main rPr"`
}

type uVal struct {
	Val string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr"`
}

type vertAlignVal struct {
	Val string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr"`
}

type highlightVal struct {
	Val string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr"`
}

type langVal struct {
	Val string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr"`
}

type emVal struct {
	Val string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr"`
}

type brVal struct {
	Type string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main type,attr"`
}

type strVal struct {
	Val string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr"`
}

type shdVal struct {
	Val   string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr"`
	Color string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main color,attr"`
	Fill  string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main fill,attr"`
}

type pBdrProps struct {
	Top     *borderVal `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main top"`
	Bottom  *borderVal `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main bottom"`
	Left    *borderVal `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main left"`
	Right   *borderVal `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main right"`
	Between *borderVal `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main between"`
}

type borderVal struct {
	Val   string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr"`
	Color string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main color,attr"`
	Sz    int    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main sz,attr"`
	Space int    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main space,attr"`
}

type runProps struct {
	RFonts    *rFonts        `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main rFonts"`
	Sz        *szVal         `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main sz"`
	Color     *colorVal      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main color"`
	Bold      *struct{}      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main b"`
	Italic    *struct{}      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main i"`
	Uline     *uVal          `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main u"`
	Strike    *struct{}      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main strike"`
	DStrike   *struct{}      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main dstrike"`
	VertAlign *vertAlignVal  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main vertAlign"`
	Highlight *highlightVal  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main highlight"`
	SmallCaps *struct{}      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main smallCaps"`
	Caps      *struct{}      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main caps"`
	Vanish    *struct{}      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main vanish"`
	Spacing   *intVal        `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main spacing"`
	Position  *intVal        `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main position"`
	Kern      *intVal        `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main kern"`
	Lang      *langVal       `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main lang"`
	Em        *emVal         `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main em"`
	RTL       *struct{}      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main rtl"`
	NoProof   *struct{}      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main noProof"`
	CS        *struct{}      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main cs"`
	SpecVanish *struct{}     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main specVanish"`
	Emboss    *struct{}      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main emboss"`
	Engrave   *struct{}      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main engrave"`
	Shadow    *struct{}      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main shadow"`
	Imprint   *struct{}      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main imprint"`
	Bdr       *borderVal     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main bdr"`
	Effect    *struct{}      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main effect"`
	Animate   *struct{}      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main animate"`
	BCs       *struct{}      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main bCs"`
	ICs       *struct{}      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main iCs"`
	SzCs      *szVal         `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main szCs"`
}

type rFonts struct {
	Ascii    string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main ascii,attr"`
	HAnsi    string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main hAnsi,attr"`
	EastAsia string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main eastAsia,attr"`
	CS       string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main cs,attr"`
}

type szVal struct {
	Val int `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr"`
}

type colorVal struct {
	Val string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr"`
}

type styleDef struct {
	ID        string      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main styleId,attr"`
	Type      string      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main type,attr"`
	Name      *nameVal    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main name"`
	BasedOn   *basedOnVal `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main basedOn"`
	ParaProps *paraProps  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main pPr"`
	RunProps  *runProps   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main rPr"`
}

type basedOnVal struct {
	Val string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr"`
}

type nameVal struct {
	Val string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr"`
}

type outlineLvl struct {
	Val int `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr"`
}

type docDocument struct {
	XMLName xml.Name     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main document"`
	Body    docBody      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main body"`
}

type docBody struct {
	XMLName  xml.Name     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main body"`
	Paras    []docPara    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main p"`
	Tables   []docTbl     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tbl"`
	Sections []docSection `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main sectPr"`
}

type docSection struct {
	PageSz     *docPageSize `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main pgSz"`
	PageMar    *docPageMar  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main pgMar"`
	HdrRef     []hdrRef     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main headerReference"`
	FtrRef     []ftrRef     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main footerReference"`
	TitlePg    *struct{}    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main titlePg"`
	EvenAndOdd *struct{}    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main evenAndOddHeaders"`
	Cols       *docCols     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main cols"`
	PgNum      *docPgNum    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main pgNumType"`
	SectType   *strVal      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main type,attr"`
}

type docPageSize struct {
	W int `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main w,attr"`
	H int `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main h,attr"`
}

type docPageMar struct {
	Top    int `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main top,attr"`
	Right  int `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main right,attr"`
	Bottom int `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main bottom,attr"`
	Left   int `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main left,attr"`
	Header int `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main header,attr"`
	Footer int `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main footer,attr"`
}

type hdrRef struct {
	Type string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main type,attr"`
	ID   string `xml:"http://schemas.openxmlformats.org/officeDocument/2006/relationships id,attr"`
}

type ftrRef struct {
	Type string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main type,attr"`
	ID   string `xml:"http://schemas.openxmlformats.org/officeDocument/2006/relationships id,attr"`
}

type docCols struct {
	Num   int `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main num,attr"`
	Space int `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main space,attr"`
}

type docPgNum struct {
	Fmt   string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main fmt,attr"`
	Start int    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main start,attr"`
}

type docHeader struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main hdr"`
	Paras   []docPara `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main p"`
	Tables  []docTbl  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tbl"`
}

type docFooter struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main ftr"`
	Paras   []docPara `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main p"`
	Tables  []docTbl  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tbl"`
}

type docFootnotes struct {
	XMLName   xml.Name     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main footnotes"`
	Footnotes []docNote    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main footnote"`
}

type docEndnotes struct {
	XMLName  xml.Name    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main endnotes"`
	Endnotes []docNote   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main endnote"`
}

type docNote struct {
	ID     int      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main id,attr"`
	Type   string   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main type,attr"`
	Paras  []docPara `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main p"`
	Tables []docTbl  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tbl"`
}

type docComments struct {
	XMLName  xml.Name     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main comments"`
	Comments []docComment `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main comment"`
}

type docComment struct {
	ID     int      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main id,attr"`
	Author string   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main author,attr"`
	Date   string   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main date,attr"`
	Paras  []docPara `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main p"`
}

type relsDoc struct {
	XMLName xml.Name   `xml:"http://schemas.openxmlformats.org/package/2006/relationships Relationships"`
	Items   []relsItem `xml:"http://schemas.openxmlformats.org/package/2006/relationships Relationship"`
}

type relsItem struct {
	ID     string `xml:"Id,attr"`
	Type   string `xml:"Type,attr"`
	Target string `xml:"Target,attr"`
}

type docPara struct {
	ParaProps  *paraProps     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main pPr"`
	RunContent []docRun       `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main r"`
	Hyperlinks []docHyperlink `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main hyperlink"`
	Bookmarks  []docBookmark  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main bookmarkStart"`
	Ins        []docIns       `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main ins"`
	Del        []docDel       `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main del"`
}

type docBookmark struct {
	ID   int    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main id,attr"`
	Name string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main name,attr"`
}

type docTbl struct {
	TblPr   *tblPr    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tblPr"`
	TblGrid *tblGrid  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tblGrid"`
	Rows    []tblRow  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tr"`
}

type tblPr struct {
	TblW       *tblWidth    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tblW"`
	JC         *jcVal       `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main jc"`
	TblBorders *tblBorders  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tblBorders"`
	TblShd     *shdVal      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main shd"`
	TblStyle   *styleRef    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tblStyle"`
	TblCaption *strVal      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tblCaption"`
	TblDescription *strVal `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tblDescription"`
	TblIndent  *tblWidth    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tblInd"`
	TblCellSpacing *tblWidth `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tblCellSpacing"`
}

type tblWidth struct {
	W    int    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main w,attr"`
	Type string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main type,attr"`
}

type tblBorders struct {
	Top    *borderVal `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main top"`
	Bottom *borderVal `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main bottom"`
	Left   *borderVal `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main left"`
	Right  *borderVal `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main right"`
	InsideH *borderVal `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main insideH"`
	InsideV *borderVal `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main insideV"`
}

type tblGrid struct {
	GridCols []gridCol `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main gridCol"`
}

type gridCol struct {
	W int `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main w,attr"`
}

type tblRow struct {
	TrPr  *trPr     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main trPr"`
	Cells []tblCell `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tc"`
}

type trPr struct {
	TblHeader *struct{} `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tblHeader"`
}

type tblCell struct {
	TcPr   *tcPr     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tcPr"`
	Paras  []docPara `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main p"`
	Tables []docTbl  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tbl"`
}

type tcPr struct {
	TcW          *tblWidth   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tcW"`
	GridSpan     *intVal     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main gridSpan"`
	VMerge       *strVal     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main vMerge"`
	TcBorders    *tblBorders `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tcBorders"`
	Shd          *shdVal     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main shd"`
	VAlign       *strVal     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main vAlign"`
	TextDir      *strVal     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main textDirection"`
	NoWrap       *struct{}   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main noWrap"`
}

type paraProps struct {
	Style        *styleRef    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main pStyle"`
	JC           *jcVal       `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main jc"`
	Ind          *indVal      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main ind"`
	NumPr        *numPr       `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main numPr"`
	RunProps     *runProps    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main rPr"`
	OutlineLvl   *outlineLvl  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main outlineLvl"`
	Spacing      *spacingVal  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main spacing"`
	Tabs         *tabsDef     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tabs"`
	KeepNext     *struct{}    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main keepNext"`
	KeepLines    *struct{}    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main keepLines"`
	WidowControl *struct{}    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main widowControl"`
	ContextualSpacing *struct{} `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main contextualSpacing"`
	SuppressLineNumbers  *struct{} `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main suppressLineNumbers"`
	SuppressHyphenation  *struct{} `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main suppressAutoHyphens"`
	TextDirection *strVal     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main textDirection"`
	PBdr          *pBdrProps  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main pBdr"`
	Shd           *shdVal     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main shd"`
	PageBreakBefore *struct{} `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main pageBreakBefore"`
	Kinsoku         *struct{} `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main kinsoku"`
	WordWrap        *struct{} `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main wordWrap"`
	CharSpacingJust *intVal   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main charSpacing"`
	TwoLineOne      *struct{} `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main twoLineOne"`
	AutoSpaceDE     *struct{} `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main autoSpaceDE"`
	AutoSpaceDN     *struct{} `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main autoSpaceDN"`
	Bidi            *struct{}  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main bidi"`
	SectPr          *docSection `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main sectPr"`
}

type styleRef struct {
	Val string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr"`
}

type jcVal struct {
	Val string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr"`
}

type indVal struct {
	Left      int `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main left,attr"`
	Hanging   int `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main hanging,attr"`
	Right     int `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main right,attr"`
	FirstLine int `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main firstLine,attr"`
}

type numPr struct {
	Ilvl *intVal `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main ilvl"`
	NumID *intVal `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main numId"`
}

type intVal struct {
	Val int `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr"`
}

type tabsDef struct {
	Tabs []tabVal `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tab"`
}

type tabVal struct {
	Val    string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr"`
	Pos    int    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main pos,attr"`
	Leader string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main leader,attr"`
}

type docRun struct {
	RunProps              *runProps   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main rPr"`
	RunText               []docText   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main t"`
	DelText               []docText   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main delText"`
	RunBreak              *brVal      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main br"`
	RunTab                *struct{}   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tab"`
	LastRenderedPageBreak *struct{}   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main lastRenderedPageBreak"`
	Drawing               *docDrawing `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main drawing"`
	Pict                  *docPict    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main pict"`
	FldChar               *fldCharVal `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main fldChar"`
	InstrText             []docText   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main instrText"`
	FfData                *ffDataVal  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main ffData"`
	FootnoteRef           *noteRef    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main footnoteReference"`
	EndnoteRef            *noteRef    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main endnoteReference"`
}

type docIns struct {
	ID     int      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main id,attr"`
	Author string   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main author,attr"`
	Date   string   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main date,attr"`
	Runs   []docRun `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main r"`
}

type docDel struct {
	ID     int      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main id,attr"`
	Author string   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main author,attr"`
	Date   string   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main date,attr"`
	Runs   []docRun `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main r"`
}

type noteRef struct {
	ID int `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main id,attr"`
}

type fldCharVal struct {
	FldCharType string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main fldCharType,attr"`
}

type ffDataVal struct {
	Name *strVal `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main name"`
}

type docHyperlink struct {
	ID   string   `xml:"http://schemas.openxmlformats.org/officeDocument/2006/relationships id,attr"`
	Runs []docRun `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main r"`
}

type docText struct {
	Text string `xml:",chardata"`
}

type docDrawing struct {
	Inline *wpInline `xml:"http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing inline"`
	Anchor *wpAnchor `xml:"http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing anchor"`
}

type wpInline struct {
	Extent  *wpExtent `xml:"http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing extent"`
	Graphic *aGraphic `xml:"http://schemas.openxmlformats.org/drawingml/2006/main graphic"`
}

type wpAnchor struct {
	Extent  *wpExtent `xml:"http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing extent"`
	Graphic *aGraphic `xml:"http://schemas.openxmlformats.org/drawingml/2006/main graphic"`
}

type wpExtent struct {
	Cx int `xml:"cx,attr"`
	Cy int `xml:"cy,attr"`
}

type aGraphic struct {
	Data *aGraphicData `xml:"http://schemas.openxmlformats.org/drawingml/2006/main graphicData"`
}

type aGraphicData struct {
	Pic *picPic `xml:"http://schemas.openxmlformats.org/drawingml/2006/picture pic"`
	Wsp *wpsWsp `xml:"http://schemas.microsoft.com/office/word/2010/wordprocessingShape wsp"`
}

type wpsWsp struct {
	Txbx *wpsTxbx `xml:"http://schemas.microsoft.com/office/word/2010/wordprocessingShape txbx"`
}

type wpsTxbx struct {
	TxbxContent *docTxbxContent `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main txbxContent"`
}

type docTxbxContent struct {
	Paras  []docPara `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main p"`
	Tables []docTbl  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tbl"`
}

type ThemeData struct {
	Bg string
	Fg string
}

type aTheme struct {
	ThemeElements *aThemeElements `xml:"http://schemas.openxmlformats.org/drawingml/2006/main themeElements"`
}

type aThemeElements struct {
	ClrScheme *aClrScheme `xml:"http://schemas.openxmlformats.org/drawingml/2006/main clrScheme"`
}

type aClrScheme struct {
	Dk1 *aThemeClr `xml:"http://schemas.openxmlformats.org/drawingml/2006/main dk1"`
	Lt1 *aThemeClr `xml:"http://schemas.openxmlformats.org/drawingml/2006/main lt1"`
}

type aThemeClr struct {
	SrgbClr *aSrgbClr `xml:"http://schemas.openxmlformats.org/drawingml/2006/main srgbClr"`
}

type aSrgbClr struct {
	Val string `xml:"val,attr"`
}

type picPic struct {
	NvPicPr  *picNvPicPr  `xml:"http://schemas.openxmlformats.org/drawingml/2006/picture nvPicPr"`
	BlipFill *picBlipFill `xml:"http://schemas.openxmlformats.org/drawingml/2006/picture blipFill"`
}

type picNvPicPr struct {
	CNvPr *picCNvPr `xml:"http://schemas.openxmlformats.org/drawingml/2006/picture cNvPr"`
}

type picCNvPr struct {
	ID    int    `xml:"id,attr"`
	Name  string `xml:"name,attr"`
	Descr string `xml:"descr,attr"`
}

type picBlipFill struct {
	Blip *aBlip `xml:"http://schemas.openxmlformats.org/drawingml/2006/main blip"`
}

type aBlip struct {
	Embed string `xml:"http://schemas.openxmlformats.org/officeDocument/2006/relationships embed,attr"`
}

type docPict struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main pict"`
	Shapes  []vShape `xml:"urn:schemas-microsoft-com:vml shape"`
}

type vShape struct {
	Style     string          `xml:"style,attr"`
	ImageData *vImagedata     `xml:"urn:schemas-microsoft-com:vml imagedata"`
	Textbox   *vmlTextbox     `xml:"urn:schemas-microsoft-com:vml textbox"`
}

type vmlTextbox struct {
	TxbxContent *docTxbxContent `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main txbxContent"`
}

type vImagedata struct {
	RelID string `xml:"http://schemas.openxmlformats.org/officeDocument/2006/relationships id,attr"`
	Title string `xml:"urn:schemas-microsoft-com:office:office title,attr"`
}

func getGraphicPtr(inline *wpInline, anchor *wpAnchor) *aGraphic {
	if inline != nil {
		return inline.Graphic
	}
	if anchor != nil {
		return anchor.Graphic
	}
	return nil
}

func parseVmlStyle(style string) (width, height float64) {
	for _, p := range strings.Split(style, ";") {
		p = strings.TrimSpace(p)
		if strings.HasPrefix(p, "width:") {
			width = parseVmlLength(strings.TrimPrefix(p, "width:"))
		} else if strings.HasPrefix(p, "height:") {
			height = parseVmlLength(strings.TrimPrefix(p, "height:"))
		}
	}
	return
}

func parseVmlLength(s string) float64 {
	s = strings.TrimSpace(s)
	if strings.HasSuffix(s, "pt") {
		v, _ := strconv.ParseFloat(strings.TrimSuffix(s, "pt"), 64)
		return v / 72.0
	}
	if strings.HasSuffix(s, "in") {
		v, _ := strconv.ParseFloat(strings.TrimSuffix(s, "in"), 64)
		return v
	}
	if strings.HasSuffix(s, "mm") {
		v, _ := strconv.ParseFloat(strings.TrimSuffix(s, "mm"), 64)
		return v / 25.4
	}
	if strings.HasSuffix(s, "cm") {
		v, _ := strconv.ParseFloat(strings.TrimSuffix(s, "cm"), 64)
		return v / 2.54
	}
	if strings.HasSuffix(s, "px") {
		v, _ := strconv.ParseFloat(strings.TrimSuffix(s, "px"), 64)
		return v / 96.0
	}
	v, _ := strconv.ParseFloat(s, 64)
	return v / 72.0
}

type docNumbering struct {
	XMLName      xml.Name          `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main numbering"`
	Nums         []numDef          `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main num"`
	AbstractNums []abstractNumDef  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main abstractNum"`
}

type numDef struct {
	NumID        int           `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main numId,attr"`
	AbstractNumID *intVal       `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main abstractNumId"`
	LvlOverrides []lvlOverride `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main lvlOverride"`
}

type lvlOverride struct {
	Ilvl          int    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main ilvl,attr"`
	StartOverride *intVal `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main startOverride"`
}

type abstractNumDef struct {
	AbstractNumID int      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main abstractNumId,attr"`
	Levels        []lvlDef `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main lvl"`
}

type lvlDef struct {
	Ilvl    int    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main ilvl,attr"`
	NumFmt  string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main numFmt,attr"`
	LvlText string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main lvlText,attr"`
	Start   *intVal `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main start"`
}

type coreProperties struct {
	XMLName       xml.Name `xml:"http://schemas.openxmlformats.org/package/2006/metadata/core-properties coreProperties"`
	Title         string   `xml:"http://purl.org/dc/elements/1.1/ title"`
	Subject       string   `xml:"http://purl.org/dc/elements/1.1/ subject"`
	Creator       string   `xml:"http://purl.org/dc/elements/1.1/ creator"`
	Description   string   `xml:"http://purl.org/dc/elements/1.1/ description"`
	Language      string   `xml:"http://purl.org/dc/elements/1.1/ language"`
	Keywords      string   `xml:"http://schemas.openxmlformats.org/package/2006/metadata/keywords"`
	Category      string   `xml:"http://schemas.openxmlformats.org/package/2006/metadata/category"`
	ContentStatus string   `xml:"http://schemas.openxmlformats.org/package/2006/metadata/contentStatus"`
	LastModifiedBy string  `xml:"http://schemas.openxmlformats.org/package/2006/metadata/lastModifiedBy"`
	Revision      string   `xml:"http://schemas.openxmlformats.org/package/2006/metadata/revision"`
	Version       string   `xml:"http://schemas.openxmlformats.org/package/2006/metadata/version"`
	Created       string   `xml:"http://purl.org/dc/terms/ created"`
	Modified      string   `xml:"http://purl.org/dc/terms/ modified"`
}

type appProperties struct {
	XMLName     xml.Name `xml:"http://schemas.openxmlformats.org/officeDocument/2006/extended-properties Properties"`
	Application string   `xml:"http://schemas.openxmlformats.org/officeDocument/2006/extended-properties Application"`
	AppVersion  string   `xml:"http://schemas.openxmlformats.org/officeDocument/2006/extended-properties AppVersion"`
}

func paragraphHasText(runs []TextRun) bool {
	for _, r := range runs {
		if strings.TrimSpace(r.Text) != "" {
			return true
		}
	}
	return false
}

func parseTwipsToInches(twips int) float64 {
	return float64(twips) / twipsPerInch
}

func halfPtToPt(halfPt int) float64 {
	return float64(halfPt) / halfPtPerPt
}

func ParseDOCX(data []byte) (*ParsedDocument, error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("open docx: %w", err)
	}

	doc := &ParsedDocument{
		PageLayout: PageLayout{
			WidthInch:  8.27,
			HeightInch: 11.69,
			MarginTop:  0.79,
			MarginRight: 0.79,
			MarginBottom: 0.79,
			MarginLeft: 0.79,
		},
		DefaultFont: FontDef{
			Family: "Times New Roman",
			SizePt: 11,
		},
	}

	var docXML, stylesXML, coreXML, appXML, numberingXML, relsXML, footnotesXML, endnotesXML, commentsXML, themeXML []byte
	headerFiles := make(map[string][]byte)
	footerFiles := make(map[string][]byte)

	for _, f := range r.File {
		switch {
		case f.Name == "word/document.xml":
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("read document.xml: %w", err)
			}
			docXML, err = readAll(rc)
			rc.Close()
			if err != nil {
				return nil, fmt.Errorf("read document.xml: %w", err)
			}
		case f.Name == "word/styles.xml":
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("read styles.xml: %w", err)
			}
			stylesXML, err = readAll(rc)
			rc.Close()
			if err != nil {
				return nil, fmt.Errorf("read styles.xml: %w", err)
			}
		case f.Name == "word/numbering.xml":
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("read numbering.xml: %w", err)
			}
			numberingXML, err = readAll(rc)
			rc.Close()
			if err != nil {
				return nil, fmt.Errorf("read numbering.xml: %w", err)
			}
		case f.Name == "docProps/core.xml":
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("read core.xml: %w", err)
			}
			coreXML, err = readAll(rc)
			rc.Close()
			if err != nil {
				return nil, fmt.Errorf("read core.xml: %w", err)
			}
		case f.Name == "docProps/app.xml":
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("read app.xml: %w", err)
			}
			appXML, err = readAll(rc)
			rc.Close()
			if err != nil {
				return nil, fmt.Errorf("read app.xml: %w", err)
			}
		case f.Name == "word/_rels/document.xml.rels":
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("read document.xml.rels: %w", err)
			}
			relsXML, err = readAll(rc)
			rc.Close()
			if err != nil {
				return nil, fmt.Errorf("read document.xml.rels: %w", err)
			}
		case strings.HasPrefix(f.Name, "word/header") && strings.HasSuffix(f.Name, ".xml"):
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("read %s: %w", f.Name, err)
			}
			headerFiles[f.Name], err = readAll(rc)
			rc.Close()
			if err != nil {
				return nil, fmt.Errorf("read %s: %w", f.Name, err)
			}
		case strings.HasPrefix(f.Name, "word/footer") && strings.HasSuffix(f.Name, ".xml"):
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("read %s: %w", f.Name, err)
			}
			footerFiles[f.Name], err = readAll(rc)
			rc.Close()
			if err != nil {
				return nil, fmt.Errorf("read %s: %w", f.Name, err)
			}
		case f.Name == "word/footnotes.xml":
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("read footnotes.xml: %w", err)
			}
			footnotesXML, err = readAll(rc)
			rc.Close()
			if err != nil {
				return nil, fmt.Errorf("read footnotes.xml: %w", err)
			}
		case f.Name == "word/endnotes.xml":
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("read endnotes.xml: %w", err)
			}
			endnotesXML, err = readAll(rc)
			rc.Close()
			if err != nil {
				return nil, fmt.Errorf("read endnotes.xml: %w", err)
			}
		case f.Name == "word/comments.xml":
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("read comments.xml: %w", err)
			}
			commentsXML, err = readAll(rc)
			rc.Close()
			if err != nil {
				return nil, fmt.Errorf("read comments.xml: %w", err)
			}
		case f.Name == "word/theme/theme1.xml":
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("read theme1.xml: %w", err)
			}
			themeXML, err = readAll(rc)
			rc.Close()
			if err != nil {
				return nil, fmt.Errorf("read theme1.xml: %w", err)
			}
		}
	}

	if len(docXML) == 0 {
		return nil, fmt.Errorf("word/document.xml not found in DOCX")
	}

	if len(stylesXML) > 0 {
		styleMap, _, _ := buildStyleMap(stylesXML)
		if def, ok := styleMap["default"]; ok {
			if def.Family != "" {
				doc.DefaultFont.Family = def.Family
				doc.DefaultFont.FromDocx = true
			}
			if def.SizePt > 0 {
				doc.DefaultFont.SizePt = def.SizePt
				doc.DefaultFont.FromDocx = true
			}
			if def.Color != "" {
				doc.DefaultFont.Color = def.Color
				doc.DefaultFont.FromDocx = true
			}
		}

		var styles docStyles
		if err := xml.Unmarshal(stylesXML, &styles); err == nil {
			if styles.DocDefaults != nil && styles.DocDefaults.ParaPropsDefault != nil {
				pp := styles.DocDefaults.ParaPropsDefault.ParaProps
				if pp != nil && pp.Spacing != nil && pp.Spacing.Line > 0 {
					lineInPt := float64(pp.Spacing.Line) / 20.0
					switch pp.Spacing.LineRule {
					case "auto":
						doc.LineHeight = float64(pp.Spacing.Line) / 240.0
					case "exact":
						if doc.DefaultFont.SizePt > 0 {
							doc.LineHeight = lineInPt / doc.DefaultFont.SizePt
						} else {
							doc.LineHeight = lineInPt / 11.0
						}
					case "atLeast":
						if doc.DefaultFont.SizePt > 0 {
							doc.LineHeight = lineInPt / doc.DefaultFont.SizePt
						} else {
							doc.LineHeight = lineInPt / 11.0
						}
					}
				}
			}
		}

		doc.HeadingStyles = make(map[int]StyleDef)
		for id, sd := range styleMap {
			if sd.HeadingLevel > 0 && sd.HeadingLevel <= 6 {
				doc.HeadingStyles[sd.HeadingLevel] = sd
			} else if strings.HasPrefix(id, "Heading") && sd.HeadingLevel == 0 {
				numStr := strings.TrimPrefix(id, "Heading")
				if num, err := strconv.Atoi(numStr); err == nil && num >= 1 && num <= 6 {
					sd.HeadingLevel = num
					doc.HeadingStyles[num] = sd
				}
			}
		}
	}

	if len(coreXML) > 0 {
		var cp coreProperties
		if err := xml.Unmarshal(coreXML, &cp); err == nil {
			doc.Title = cp.Title
			doc.Subject = cp.Subject
			doc.Author = cp.Creator
			doc.Keywords = cp.Keywords
			doc.Description = cp.Description
			doc.Category = cp.Category
			doc.ContentStatus = cp.ContentStatus
			doc.LastModifiedBy = cp.LastModifiedBy
			doc.Revision = cp.Revision
			doc.Version = cp.Version
			doc.Created = cp.Created
			doc.Modified = cp.Modified
			doc.Language = cp.Language
		}
	}

	if len(appXML) > 0 {
		var ap appProperties
		if err := xml.Unmarshal(appXML, &ap); err == nil {
			doc.Application = ap.Application
			doc.AppVersion = ap.AppVersion
		}
	}

	var document docDocument
	if err := xml.Unmarshal(docXML, &document); err != nil {
		return nil, fmt.Errorf("parse document.xml: %w", err)
	}

	body := document.Body

	styleMap := make(map[string]StyleDef)
	styleNameMap := make(map[string]string)
	if len(stylesXML) > 0 {
		styleMap, styleNameMap, doc.CustomStyles = buildStyleMap(stylesXML)
	}

	var numberingMap map[int]map[int]string
	var numberingStartMap map[int]map[int]int
	if len(numberingXML) > 0 {
		numberingMap, numberingStartMap = buildNumberingMap(numberingXML)
	}

	relMap := make(map[string]string)
	if len(relsXML) > 0 {
		var rels relsDoc
		if err := xml.Unmarshal(relsXML, &rels); err == nil {
			for _, item := range rels.Items {
				relMap[item.ID] = "word/" + item.Target
			}
		}
	}

	if len(footnotesXML) > 0 {
		var fndoc docFootnotes
		if err := xml.Unmarshal(footnotesXML, &fndoc); err == nil {
			for _, fn := range fndoc.Footnotes {
				if fn.Type == "normal" || fn.Type == "" {
					body := parseCellContent(fn.Paras, fn.Tables, styleMap, styleNameMap, numberingMap, numberingStartMap, relMap)
					doc.Notes = append(doc.Notes, NoteItem{Type: "footnote", ID: fn.ID, Body: body})
				}
			}
		}
	}
	if len(endnotesXML) > 0 {
		var endoc docEndnotes
		if err := xml.Unmarshal(endnotesXML, &endoc); err == nil {
			for _, en := range endoc.Endnotes {
				if en.Type == "normal" || en.Type == "" {
					body := parseCellContent(en.Paras, en.Tables, styleMap, styleNameMap, numberingMap, numberingStartMap, relMap)
					doc.Notes = append(doc.Notes, NoteItem{Type: "endnote", ID: en.ID, Body: body})
				}
			}
		}
	}
	if len(commentsXML) > 0 {
		var cmdoc docComments
		if err := xml.Unmarshal(commentsXML, &cmdoc); err == nil {
			for _, cm := range cmdoc.Comments {
				body := parseCellContent(cm.Paras, nil, styleMap, styleNameMap, numberingMap, numberingStartMap, relMap)
				doc.Notes = append(doc.Notes, NoteItem{Type: "comment", ID: cm.ID, Author: cm.Author, Date: cm.Date, Body: body})
			}
		}
	}

	if len(themeXML) > 0 {
		var theme aTheme
		if err := xml.Unmarshal(themeXML, &theme); err == nil {
			if theme.ThemeElements != nil && theme.ThemeElements.ClrScheme != nil {
				td := &ThemeData{}
				if cs := theme.ThemeElements.ClrScheme; cs != nil {
					if cs.Dk1 != nil && cs.Dk1.SrgbClr != nil {
						td.Fg = cs.Dk1.SrgbClr.Val
					}
					if cs.Lt1 != nil && cs.Lt1.SrgbClr != nil {
						td.Bg = cs.Lt1.SrgbClr.Val
					}
				}
				if td.Fg != "" || td.Bg != "" {
					doc.Theme = td
				}
			}
		}
	}

	headerContent := make(map[string][]ContentItem)
	for path, data := range headerFiles {
		var hdr docHeader
		if err := xml.Unmarshal(data, &hdr); err == nil {
			headerContent[path] = parseCellContent(hdr.Paras, hdr.Tables, styleMap, styleNameMap, numberingMap, numberingStartMap, relMap)
		}
	}
	footerContent := make(map[string][]ContentItem)
	for path, data := range footerFiles {
		var ftr docFooter
		if err := xml.Unmarshal(data, &ftr); err == nil {
			footerContent[path] = parseCellContent(ftr.Paras, ftr.Tables, styleMap, styleNameMap, numberingMap, numberingStartMap, relMap)
		}
	}

	doc.Sections, doc.Content = buildSections(body, docXML, relMap, headerContent, footerContent, styleMap, styleNameMap, numberingMap, numberingStartMap, doc)

	if len(body.Sections) > 0 {
		sec := body.Sections[len(body.Sections)-1]
		doc.PageLayout.FromDocx = sec.PageSz != nil || sec.PageMar != nil
		if sec.PageSz != nil {
			doc.PageLayout.WidthInch = parseTwipsToInches(sec.PageSz.W)
			doc.PageLayout.HeightInch = parseTwipsToInches(sec.PageSz.H)
		}
		if sec.PageMar != nil {
			doc.PageLayout.MarginTop = parseTwipsToInches(sec.PageMar.Top)
			doc.PageLayout.MarginRight = parseTwipsToInches(sec.PageMar.Right)
			doc.PageLayout.MarginBottom = parseTwipsToInches(sec.PageMar.Bottom)
			doc.PageLayout.MarginLeft = parseTwipsToInches(sec.PageMar.Left)
			doc.PageLayout.HeaderMargin = parseTwipsToInches(sec.PageMar.Header)
			doc.PageLayout.FooterMargin = parseTwipsToInches(sec.PageMar.Footer)
		}
	}

	assignTableIDs(doc)

	return doc, nil
}

// assignTableIDs walks all content items in the document and assigns
// sequential 1-based IDs to each table (pre-order traversal of nested tables).
// Also collects all tables into doc.AllTables for style block generation.
func assignTableIDs(doc *ParsedDocument) {
	counter := 0
	doc.AllTables = nil
	var walk func(items []ContentItem)
	walk = func(items []ContentItem) {
		for i := range items {
			if items[i].Type == ContentTable {
				counter++
				items[i].Table.ID = counter
				doc.AllTables = append(doc.AllTables, *items[i].Table)
				for _, row := range items[i].Table.Rows {
					for _, cell := range row.Cells {
						walk(cell.Content)
					}
				}
			}
		}
	}
	walk(doc.Content)
	for _, sec := range doc.Sections {
		walk(sec.Content)
		for _, hf := range sec.Headers {
			walk(hf.Content)
		}
		for _, hf := range sec.Footers {
			walk(hf.Content)
		}
	}
}

func buildSections(body docBody, docXML []byte, relMap map[string]string, headerContent, footerContent map[string][]ContentItem, styleMap map[string]StyleDef, styleNameMap map[string]string, numberingMap map[int]map[int]string, numberingStartMap map[int]map[int]int, doc *ParsedDocument) ([]ParsedSection, []ContentItem) {
	var bt struct {
		Body struct {
			Inner []xml.Token `xml:",any"`
		} `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main body"`
	}
	xml.Unmarshal(docXML, &bt)

	pi, ti, si := 0, 0, 0
	var allSections []ParsedSection
	var curContent []ContentItem

	for _, tok := range bt.Body.Inner {
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		switch se.Name.Local {
		case "p":
			if pi >= len(body.Paras) {
				continue
			}
			pp := parseDocParagraph(body.Paras[pi], styleMap, styleNameMap, numberingMap, numberingStartMap, relMap)
			if paragraphHasText(pp.Runs) {
				curContent = append(curContent, ContentItem{
					Type:      ContentParagraph,
					Paragraph: &pp,
				})
			}
			for _, bm := range body.Paras[pi].Bookmarks {
				if bm.Name != "" {
					doc.Notes = append(doc.Notes, NoteItem{
						Type: "bookmark",
						Name: bm.Name,
					})
				}
			}
			if body.Paras[pi].ParaProps != nil && body.Paras[pi].ParaProps.SectPr != nil {
				ps := buildParsedSection(body.Paras[pi].ParaProps.SectPr, relMap, headerContent, footerContent)
				ps.Content = curContent
				allSections = append(allSections, ps)
				curContent = nil
			}
			pi++
		case "tbl":
			if ti >= len(body.Tables) {
				continue
			}
			tbl := parseDocTable(body.Tables[ti], styleMap, styleNameMap, numberingMap, numberingStartMap, relMap)
			curContent = append(curContent, ContentItem{
				Type:  ContentTable,
				Table: &tbl,
			})
			ti++
		case "sectPr":
			if si >= len(body.Sections) {
				continue
			}
			ps := buildParsedSection(&body.Sections[si], relMap, headerContent, footerContent)
			ps.Content = curContent
			allSections = append(allSections, ps)
			curContent = nil
			si++
		case "bookmarkStart":
			name := ""
			id := 0
			for _, attr := range se.Attr {
				switch attr.Name.Local {
				case "name":
					name = attr.Value
				case "id":
					id, _ = strconv.Atoi(attr.Value)
				}
			}
			if name != "" {
				_ = id
				doc.Notes = append(doc.Notes, NoteItem{
					Type: "bookmark",
					Name: name,
				})
			}
		}
	}

	if len(curContent) > 0 || len(allSections) == 0 {
		ps := ParsedSection{Content: curContent}
		if len(body.Sections) > 0 {
			lastSec := body.Sections[min(si, len(body.Sections)-1)]
			ps = buildParsedSection(&lastSec, relMap, headerContent, footerContent)
			ps.Content = curContent
		}
		allSections = append(allSections, ps)
	}

	var flat []ContentItem
	for _, s := range allSections {
		flat = append(flat, s.Content...)
	}
	return allSections, flat
}

func buildParsedSection(sec *docSection, relMap map[string]string, headerContent, footerContent map[string][]ContentItem) ParsedSection {
	ps := ParsedSection{}
	if sec == nil {
		return ps
	}
	if sec.PageSz != nil {
		ps.Layout.WidthInch = parseTwipsToInches(sec.PageSz.W)
		ps.Layout.HeightInch = parseTwipsToInches(sec.PageSz.H)
		ps.Layout.FromDocx = true
	}
	if sec.PageMar != nil {
		ps.Layout.MarginTop = parseTwipsToInches(sec.PageMar.Top)
		ps.Layout.MarginRight = parseTwipsToInches(sec.PageMar.Right)
		ps.Layout.MarginBottom = parseTwipsToInches(sec.PageMar.Bottom)
		ps.Layout.MarginLeft = parseTwipsToInches(sec.PageMar.Left)
		ps.Layout.HeaderMargin = parseTwipsToInches(sec.PageMar.Header)
		ps.Layout.FooterMargin = parseTwipsToInches(sec.PageMar.Footer)
		ps.Layout.FromDocx = true
	}
	for _, h := range sec.HdrRef {
		if path, ok := relMap[h.ID]; ok {
			if content, ok := headerContent[path]; ok {
				hfType := h.Type
				if hfType == "" {
					hfType = "default"
				}
				ps.Headers = append(ps.Headers, ParsedHdrFtr{Type: hfType, Content: content})
			}
		}
	}
	for _, f := range sec.FtrRef {
		if path, ok := relMap[f.ID]; ok {
			if content, ok := footerContent[path]; ok {
				hfType := f.Type
				if hfType == "" {
					hfType = "default"
				}
				ps.Footers = append(ps.Footers, ParsedHdrFtr{Type: hfType, Content: content})
			}
		}
	}
	if sec.Cols != nil {
		ps.ColCount = sec.Cols.Num
		ps.ColSpace = float64(sec.Cols.Space) / twipsPerInch
	}
	if sec.PgNum != nil {
		ps.PageNumFmt = sec.PgNum.Fmt
		ps.PageNumStart = sec.PgNum.Start
	}
	if sec.SectType != nil {
		ps.BreakType = sec.SectType.Val
	}
	return ps
}

func parseDocTable(tbl docTbl, styleMap map[string]StyleDef, styleNameMap map[string]string, numberingMap map[int]map[int]string, numberingStartMap map[int]map[int]int, relMap map[string]string) ParsedTable {
	pt := ParsedTable{}
	if tbl.TblGrid != nil {
		for _, gc := range tbl.TblGrid.GridCols {
			pt.Grid = append(pt.Grid, float64(gc.W)/twipsPerInch)
		}
	}
	if tbl.TblPr != nil {
		if tbl.TblPr.TblBorders != nil {
			pt.HasBorders = true
			b := tbl.TblPr.TblBorders
			if b.Top != nil {
				pt.BorderTop = &BorderInfo{Val: b.Top.Val, Sz: b.Top.Sz, Space: b.Top.Space, Color: b.Top.Color}
			}
			if b.Bottom != nil {
				pt.BorderBottom = &BorderInfo{Val: b.Bottom.Val, Sz: b.Bottom.Sz, Space: b.Bottom.Space, Color: b.Bottom.Color}
			}
			if b.Left != nil {
				pt.BorderLeft = &BorderInfo{Val: b.Left.Val, Sz: b.Left.Sz, Space: b.Left.Space, Color: b.Left.Color}
			}
			if b.Right != nil {
				pt.BorderRight = &BorderInfo{Val: b.Right.Val, Sz: b.Right.Sz, Space: b.Right.Space, Color: b.Right.Color}
			}
		}
		if tbl.TblPr.TblW != nil {
			switch tbl.TblPr.TblW.Type {
			case "dxa", "twips":
				pt.Width = float64(tbl.TblPr.TblW.W) / twipsPerInch
			default:
				pt.Width = float64(tbl.TblPr.TblW.W) / twipsPerInch
			}
		}
		if tbl.TblPr.JC != nil {
			pt.Alignment = tbl.TblPr.JC.Val
		}
		if tbl.TblPr.TblCaption != nil {
			pt.Caption = tbl.TblPr.TblCaption.Val
		}
		if tbl.TblPr.TblDescription != nil {
			pt.Summary = tbl.TblPr.TblDescription.Val
		}
		if tbl.TblPr.TblIndent != nil {
			pt.Indent = float64(tbl.TblPr.TblIndent.W) / twipsPerInch
		}
		if tbl.TblPr.TblCellSpacing != nil {
			pt.CellSpacing = float64(tbl.TblPr.TblCellSpacing.W) / twipsPerInch
		}
		if tbl.TblPr.TblStyle != nil {
			if styleNameMap != nil {
				pt.StyleName = styleNameMap[tbl.TblPr.TblStyle.Val]
			}
			if pt.StyleName == "" {
				pt.StyleName = tbl.TblPr.TblStyle.Val
			}
		}
	}

	// First pass: parse all cells, track tblHeader and vMerge
	for _, row := range tbl.Rows {
		pr := ParsedTableRow{}
		if row.TrPr != nil && row.TrPr.TblHeader != nil {
			pr.IsHeader = true
		}
		for _, cell := range row.Cells {
			pc := ParsedTableCell{}
			if cell.TcPr != nil {
				if cell.TcPr.GridSpan != nil {
					pc.GridSpan = cell.TcPr.GridSpan.Val
				}
				if cell.TcPr.VMerge != nil {
					pc.VMerge = cell.TcPr.VMerge.Val
				}
				if cell.TcPr.Shd != nil && cell.TcPr.Shd.Fill != "" {
					pc.ShadingFill = cell.TcPr.Shd.Fill
				}
				if cell.TcPr.VAlign != nil {
					pc.VAlign = cell.TcPr.VAlign.Val
				}
				if cell.TcPr.TextDir != nil {
					pc.TextDirection = cell.TcPr.TextDir.Val
				}
				if cell.TcPr.NoWrap != nil {
					pc.NoWrap = true
				}
				if cell.TcPr.TcBorders != nil {
					b := cell.TcPr.TcBorders
					if b.Top != nil {
						pc.BorderTop = &BorderInfo{Val: b.Top.Val, Sz: b.Top.Sz, Space: b.Top.Space, Color: b.Top.Color}
					}
					if b.Bottom != nil {
						pc.BorderBottom = &BorderInfo{Val: b.Bottom.Val, Sz: b.Bottom.Sz, Space: b.Bottom.Space, Color: b.Bottom.Color}
					}
					if b.Left != nil {
						pc.BorderLeft = &BorderInfo{Val: b.Left.Val, Sz: b.Left.Sz, Space: b.Left.Space, Color: b.Left.Color}
					}
					if b.Right != nil {
						pc.BorderRight = &BorderInfo{Val: b.Right.Val, Sz: b.Right.Sz, Space: b.Right.Space, Color: b.Right.Color}
					}
				}
			}
			pc.Content = parseCellContent(cell.Paras, cell.Tables, styleMap, styleNameMap, numberingMap, numberingStartMap, relMap)
			pr.Cells = append(pr.Cells, pc)
		}
		pt.Rows = append(pt.Rows, pr)
	}

	// Second pass: reconstruct vMerge grid for rowspan
	pt = reconstructVMerge(pt)

	return pt
}

func reconstructVMerge(pt ParsedTable) ParsedTable {
	if len(pt.Rows) == 0 {
		return pt
	}
	// Determine max columns
	maxCols := 0
	for _, row := range pt.Rows {
		n := 0
		for _, cell := range row.Cells {
			gs := cell.GridSpan
			if gs < 1 {
				gs = 1
			}
			n += gs
		}
		if n > maxCols {
			maxCols = n
		}
	}
	if maxCols == 0 {
		return pt
	}

	// Build virtual grid: grid[row][col] = cell index or -1 (merged/continue)
	type gridCell struct {
		rowIdx int
		cellIdx int
	}
	grid := make([][]gridCell, len(pt.Rows))
	for r := range grid {
		grid[r] = make([]gridCell, maxCols)
		for c := range grid[r] {
			grid[r][c] = gridCell{rowIdx: -1, cellIdx: -1}
		}
	}

	for r := 0; r < len(pt.Rows); r++ {
		col := 0
		for ci := 0; ci < len(pt.Rows[r].Cells); ci++ {
			cell := pt.Rows[r].Cells[ci]
			gs := cell.GridSpan
			if gs < 1 {
				gs = 1
			}
			// Skip already-filled columns (from vMerge continues)
			for col < maxCols && grid[r][col].cellIdx >= 0 {
				col++
			}
			if col >= maxCols {
				break
			}
			for c := 0; c < gs && col+c < maxCols; c++ {
				grid[r][col+c] = gridCell{rowIdx: r, cellIdx: ci}
			}
			col += gs
		}
	}

	// Now compute rowspan: for each cell that is a restart (vMerge=restart or vMerge=continue with no prior restart),
	// count consecutive continues below
	for r := 0; r < len(pt.Rows); r++ {
		for ci := 0; ci < len(pt.Rows[r].Cells); ci++ {
			cell := &pt.Rows[r].Cells[ci]
			if cell.VMerge == "" {
				continue
			}
			if cell.VMerge == "restart" || cell.VMerge == "continue" {
				// Find the first column of this cell
				startCol := -1
				for c := 0; c < maxCols; c++ {
					if grid[r][c].rowIdx == r && grid[r][c].cellIdx == ci {
						startCol = c
						break
					}
				}
				if startCol < 0 {
					continue
				}
				gs := cell.GridSpan
				if gs < 1 {
					gs = 1
				}

				// If this is "continue", look up for the restart cell above
				if cell.VMerge == "continue" {
					// Mark this cell for removal (rowspan 0 signals omission)
					cell.RowSpan = 0
					continue
				}

				// This is "restart": count continues below
				span := 1
				for r2 := r + 1; r2 < len(pt.Rows); r2++ {
					if startCol >= len(grid[r2]) {
						break
					}
					gc := grid[r2][startCol]
					if gc.rowIdx != r2 {
						break
					}
					below := pt.Rows[r2].Cells[gc.cellIdx]
					if below.VMerge != "continue" {
						break
					}
					span++
				}
				if span > 1 {
					cell.RowSpan = span
				}
			}
		}
	}

	// Remove continue cells (RowSpan == 0)
	for r := range pt.Rows {
		filtered := make([]ParsedTableCell, 0, len(pt.Rows[r].Cells))
		for _, cell := range pt.Rows[r].Cells {
			if cell.RowSpan == 0 {
				continue
			}
			filtered = append(filtered, cell)
		}
		pt.Rows[r].Cells = filtered
	}

	return pt
}

func parseCellContent(paras []docPara, tables []docTbl, styleMap map[string]StyleDef, styleNameMap map[string]string, numberingMap map[int]map[int]string, numberingStartMap map[int]map[int]int, relMap map[string]string) []ContentItem {
	var result []ContentItem
	pi, ti := 0, 0
	for pi < len(paras) || ti < len(tables) {
		if pi < len(paras) {
			pp := parseDocParagraph(paras[pi], styleMap, styleNameMap, numberingMap, numberingStartMap, relMap)
			if paragraphHasText(pp.Runs) {
				result = append(result, ContentItem{
					Type:      ContentParagraph,
					Paragraph: &pp,
				})
			}
			pi++
		}
		if ti < len(tables) {
			tbl := parseDocTable(tables[ti], styleMap, styleNameMap, numberingMap, numberingStartMap, relMap)
			result = append(result, ContentItem{
				Type:  ContentTable,
				Table: &tbl,
			})
			ti++
		}
	}
	return result
}

func readAll(r io.Reader) ([]byte, error) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r)
	return buf.Bytes(), err
}

func buildNumberingMap(data []byte) (map[int]map[int]string, map[int]map[int]int) {
	var n docNumbering
	if err := xml.Unmarshal(data, &n); err != nil {
		return nil, nil
	}

	abstractMap := make(map[int]map[int]string)
	abstractStartMap := make(map[int]map[int]int)
	for _, a := range n.AbstractNums {
		lvlMap := make(map[int]string)
		startMap := make(map[int]int)
		for _, l := range a.Levels {
			lvlMap[l.Ilvl] = l.NumFmt
			if l.Start != nil {
				startMap[l.Ilvl] = l.Start.Val
			}
		}
		abstractMap[a.AbstractNumID] = lvlMap
		abstractStartMap[a.AbstractNumID] = startMap
	}

	numMap := make(map[int]map[int]string)
	numStartMap := make(map[int]map[int]int)
	for _, num := range n.Nums {
		if num.AbstractNumID == nil {
			continue
		}
		if lvlMap, ok := abstractMap[num.AbstractNumID.Val]; ok {
			numMap[num.NumID] = lvlMap
		}
		if sm, ok := abstractStartMap[num.AbstractNumID.Val]; ok {
			// apply lvlOverride start values on top of abstract defaults
			merged := make(map[int]int)
			for k, v := range sm {
				merged[k] = v
			}
			for _, ov := range num.LvlOverrides {
				if ov.StartOverride != nil {
					merged[ov.Ilvl] = ov.StartOverride.Val
				}
			}
			numStartMap[num.NumID] = merged
		}
	}

	return numMap, numStartMap
}

type StyleDef struct {
	Family        string
	SizePt        float64
	Color         string
	Bold          bool
	Italic        bool
	Underline     string
	SpacingBefore float64
	SpacingAfter  float64
	Align         string
	BorderBottom  string
	HeadingLevel  int
}

func buildStyleMap(stylesXML []byte) (map[string]StyleDef, map[string]string, []CustomStyleDef) {
	var styles docStyles
	if err := xml.Unmarshal(stylesXML, &styles); err != nil {
		return nil, nil, nil
	}

	m := make(map[string]StyleDef)
	nameMap := make(map[string]string)
	var customStyles []CustomStyleDef

	if styles.DocDefaults != nil && styles.DocDefaults.RunPropsDefault != nil {
		rp := styles.DocDefaults.RunPropsDefault.RunProps
		if rp != nil {
			def := StyleDef{}
			if rp.RFonts != nil {
				if rp.RFonts.Ascii != "" {
					def.Family = rp.RFonts.Ascii
				} else if rp.RFonts.HAnsi != "" {
					def.Family = rp.RFonts.HAnsi
				}
			}
			if rp.Sz != nil {
				def.SizePt = halfPtToPt(rp.Sz.Val)
			}
			if rp.Color != nil && rp.Color.Val != "" {
				def.Color = rp.Color.Val
			}
			m["default"] = def
		}
	}

	for _, s := range styles.Styles {
		sd := StyleDef{}
		
		if s.Name != nil {
			nameMap[s.ID] = s.Name.Val
		}
		
		// Extract from RunProps (rPr)
		if s.RunProps != nil {
			if s.RunProps.RFonts != nil {
				if s.RunProps.RFonts.Ascii != "" {
					sd.Family = s.RunProps.RFonts.Ascii
				} else if s.RunProps.RFonts.HAnsi != "" {
					sd.Family = s.RunProps.RFonts.HAnsi
				}
			}
			if s.RunProps.Sz != nil {
				sd.SizePt = halfPtToPt(s.RunProps.Sz.Val)
			}
			if s.RunProps.Color != nil && s.RunProps.Color.Val != "" {
				sd.Color = s.RunProps.Color.Val
			}
			if s.RunProps.Bold != nil {
				sd.Bold = true
			}
			if s.RunProps.Italic != nil {
				sd.Italic = true
			}
			if s.RunProps.Uline != nil {
				sd.Underline = s.RunProps.Uline.Val
			}
		}
		
		// Extract from ParaProps (pPr) - outlineLvl, spacing, alignment, borders
		if s.ParaProps != nil {
			if s.ParaProps.OutlineLvl != nil {
				sd.HeadingLevel = s.ParaProps.OutlineLvl.Val + 1
			}
			if s.ParaProps.JC != nil {
				sd.Align = s.ParaProps.JC.Val
			}
			if s.ParaProps.Spacing != nil {
				// before/after in twips → pt (twips/20)
				if s.ParaProps.Spacing.Before > 0 {
					sd.SpacingBefore = float64(s.ParaProps.Spacing.Before) / 20.0
				}
				if s.ParaProps.Spacing.After > 0 {
					sd.SpacingAfter = float64(s.ParaProps.Spacing.After) / 20.0
				}
			}
			if s.ParaProps.PBdr != nil && s.ParaProps.PBdr.Bottom != nil {
				sd.BorderBottom = s.ParaProps.PBdr.Bottom.Val
			}
		}
		
		// Infer heading from style name if outlineLvl not found
		if sd.HeadingLevel == 0 {
			sd.HeadingLevel = inferHeadingFromStyleName(s.ID, s.Type)
		}

		m[s.ID] = sd

		// Collect custom (non-standard) styles for <s:custom>
		if s.Name != nil && !isStandardStyle(s.ID) {
			basedOn := ""
			if s.BasedOn != nil {
				basedOn = s.BasedOn.Val
			}
			customStyles = append(customStyles, CustomStyleDef{
				Name:     s.Name.Val,
				Type:     s.Type,
				BasedOn:  basedOn,
				StyleDef: sd,
			})
		}
	}

	return m, nameMap, customStyles
}

func isStandardStyle(id string) bool {
	if id == "Normal" || id == "Title" || id == "Subtitle" || id == "Quote" ||
		id == "IntenseQuote" || id == "BlockText" || id == "ListParagraph" ||
		id == "Code" || id == "HTML" || id == "XML" || id == "PlainText" ||
		id == "SourceCode" || id == "Example" || id == "Output" ||
		id == "NoSpacing" || id == "ListBullet" || id == "ListNumber" ||
		id == "List" || id == "ListContinue" || id == "ListContinueNumber" ||
		id == "Hyperlink" || id == "FollowedHyperlink" ||
		id == "BodyText" || id == "BodyTextIndent" ||
		id == "BodyTextIndent2" || id == "BodyTextIndent3" ||
		id == "MacroText" || id == "DefaultParagraphFont" ||
		id == "TableNormal" || id == "TableGrid" ||
		id == "CommentText" || id == "CommentReference" ||
		id == "CommentSubject" || id == "BalloonText" ||
		id == "Header" || id == "Footer" ||
		id == "PageNumber" || id == "EndnoteReference" ||
		id == "EndnoteText" || id == "FootnoteReference" ||
		id == "FootnoteText" || id == "DocDefaults" {
		return true
	}
	return strings.HasPrefix(id, "Heading") || strings.HasPrefix(id, "heading")
}

func inferHeadingFromStyleName(styleID, styleType string) int {
	if styleType != "paragraph" {
		return 0
	}
	
	// Normalize by stripping common suffixes
	id := styleID
	id = strings.TrimSuffix(id, "1")
	
	// Check for heading patterns
	if strings.EqualFold(id, "Heading1") || strings.EqualFold(id, "heading1") ||
		strings.EqualFold(id, "Heading") || strings.EqualFold(id, "heading") ||
		strings.EqualFold(styleID, "Title") || strings.EqualFold(styleID, "title") {
		return 1
	}
	if strings.EqualFold(id, "Heading2") || strings.EqualFold(id, "heading2") {
		return 2
	}
	if strings.EqualFold(id, "Heading3") || strings.EqualFold(id, "heading3") {
		return 3
	}
	if strings.EqualFold(id, "Heading4") || strings.EqualFold(id, "heading4") {
		return 4
	}
	if strings.EqualFold(id, "Heading5") || strings.EqualFold(id, "heading5") {
		return 5
	}
	if strings.EqualFold(id, "Heading6") || strings.EqualFold(id, "heading6") {
		return 6
	}
	if strings.EqualFold(id, "Heading7") || strings.EqualFold(id, "heading7") {
		return 7
	}
	if strings.EqualFold(id, "Heading8") || strings.EqualFold(id, "heading8") {
		return 8
	}
	if strings.EqualFold(id, "Heading9") || strings.EqualFold(id, "heading9") {
		return 9
	}
	
	// Check for Subtitle style - treat as heading 2
	if strings.EqualFold(styleID, "Subtitle") || strings.EqualFold(styleID, "subtitle") {
		return 2
	}
	
	return 0
}

func parseDocParagraph(p docPara, styleMap map[string]StyleDef, styleNameMap map[string]string, numberingMap map[int]map[int]string, numberingStartMap map[int]map[int]int, relMap map[string]string) ParsedParagraph {
	pp := ParsedParagraph{}

	if p.ParaProps != nil {
		if p.ParaProps.Style != nil {
			styleID := p.ParaProps.Style.Val
			pp.StyleID = styleID
			if styleNameMap != nil {
				pp.StyleName = styleNameMap[styleID]
			}
			if sd, ok := styleMap[styleID]; ok {
				pp.HeadingLevel = sd.HeadingLevel
				if sd.Family != "" {
					pp.FontFamily = sd.Family
				}
				if sd.SizePt > 0 {
					pp.FontSizePt = sd.SizePt
				}
				if sd.Color != "" {
					pp.FontColor = sd.Color
				}
				if sd.Bold {
					pp.Bold = true
				}
			}
		}
		if p.ParaProps.JC != nil {
			pp.Align = p.ParaProps.JC.Val
		}
		if p.ParaProps.Ind != nil {
			pp.IndentLeft = parseTwipsToInches(p.ParaProps.Ind.Left)
			pp.IndentRight = parseTwipsToInches(p.ParaProps.Ind.Right)
			pp.FirstLineIndent = parseTwipsToInches(p.ParaProps.Ind.FirstLine)
			pp.Hanging = parseTwipsToInches(p.ParaProps.Ind.Hanging)
		}
		if p.ParaProps.Spacing != nil {
			if p.ParaProps.Spacing.Line > 0 {
				lineInPt := float64(p.ParaProps.Spacing.Line) / 20.0
				switch p.ParaProps.Spacing.LineRule {
				case "auto":
					pp.LineHeight = float64(p.ParaProps.Spacing.Line) / 240.0
				case "exact":
					fontSize := pp.FontSizePt
					if fontSize <= 0 {
						fontSize = 11
					}
					pp.LineHeight = lineInPt / fontSize
				case "atLeast":
					fontSize := pp.FontSizePt
					if fontSize <= 0 {
						fontSize = 11
					}
					pp.LineHeight = lineInPt / fontSize
				}
			}
			if p.ParaProps.Spacing.Before > 0 {
				pp.SpacingBefore = float64(p.ParaProps.Spacing.Before) / 20.0
			}
			if p.ParaProps.Spacing.After > 0 {
				pp.SpacingAfter = float64(p.ParaProps.Spacing.After) / 20.0
			}
		}
		if p.ParaProps.NumPr != nil {
			pp.IsList = true
			if p.ParaProps.NumPr.Ilvl != nil {
				pp.ListLevel = p.ParaProps.NumPr.Ilvl.Val
			}
			if p.ParaProps.NumPr.NumID != nil && numberingMap != nil {
				if lvlMap, ok := numberingMap[p.ParaProps.NumPr.NumID.Val]; ok {
					if fmtStr, ok := lvlMap[pp.ListLevel]; ok {
						pp.ListFormat = fmtStr
					}
				}
			}
			if p.ParaProps.NumPr.NumID != nil && numberingStartMap != nil {
				if sm, ok := numberingStartMap[p.ParaProps.NumPr.NumID.Val]; ok {
					if start, ok := sm[pp.ListLevel]; ok {
						pp.ListStartOverride = start
					}
				}
			}
		}
		if p.ParaProps.KeepNext != nil {
			pp.KeepNext = true
		}
		if p.ParaProps.KeepLines != nil {
			pp.KeepLines = true
		}
		if p.ParaProps.WidowControl != nil {
			pp.WidowControl = true
		}
		if p.ParaProps.ContextualSpacing != nil {
			pp.ContextualSpacing = true
		}
		if p.ParaProps.SuppressLineNumbers != nil {
			pp.SuppressLineNumbers = true
		}
		if p.ParaProps.SuppressHyphenation != nil {
			pp.SuppressHyphenation = true
		}
		if p.ParaProps.TextDirection != nil {
			pp.TextDirection = p.ParaProps.TextDirection.Val
		}
		if p.ParaProps.PBdr != nil {
			if p.ParaProps.PBdr.Top != nil {
				pp.BorderTop = &BorderInfo{
					Val:   p.ParaProps.PBdr.Top.Val,
					Sz:    p.ParaProps.PBdr.Top.Sz,
					Space: p.ParaProps.PBdr.Top.Space,
					Color: p.ParaProps.PBdr.Top.Color,
				}
			}
			if p.ParaProps.PBdr.Bottom != nil {
				pp.BorderBottom = &BorderInfo{
					Val:   p.ParaProps.PBdr.Bottom.Val,
					Sz:    p.ParaProps.PBdr.Bottom.Sz,
					Space: p.ParaProps.PBdr.Bottom.Space,
					Color: p.ParaProps.PBdr.Bottom.Color,
				}
			}
			if p.ParaProps.PBdr.Left != nil {
				pp.BorderLeft = &BorderInfo{
					Val:   p.ParaProps.PBdr.Left.Val,
					Sz:    p.ParaProps.PBdr.Left.Sz,
					Space: p.ParaProps.PBdr.Left.Space,
					Color: p.ParaProps.PBdr.Left.Color,
				}
			}
			if p.ParaProps.PBdr.Right != nil {
				pp.BorderRight = &BorderInfo{
					Val:   p.ParaProps.PBdr.Right.Val,
					Sz:    p.ParaProps.PBdr.Right.Sz,
					Space: p.ParaProps.PBdr.Right.Space,
					Color: p.ParaProps.PBdr.Right.Color,
				}
			}
		}
		if p.ParaProps.Shd != nil && p.ParaProps.Shd.Fill != "" {
			pp.ShadingFill = p.ParaProps.Shd.Fill
		}
		if p.ParaProps.PageBreakBefore != nil {
			pp.PageBreakBefore = true
		}
		if p.ParaProps.Kinsoku != nil {
			pp.Kinsoku = true
		}
		if p.ParaProps.WordWrap != nil {
			pp.WordWrap = true
		}
		if p.ParaProps.CharSpacingJust != nil {
			pp.CharSpacingJust = p.ParaProps.CharSpacingJust.Val
		}
		if p.ParaProps.TwoLineOne != nil {
			pp.TwoLineOne = true
		}
		if p.ParaProps.AutoSpaceDE != nil {
			pp.AutoSpaceDE = true
		}
		if p.ParaProps.AutoSpaceDN != nil {
			pp.AutoSpaceDN = true
		}
		if p.ParaProps.Bidi != nil {
			pp.Bidi = true
		}
		if p.ParaProps.Tabs != nil {
			for _, t := range p.ParaProps.Tabs.Tabs {
				pp.TabStops = append(pp.TabStops, TabStopDef{
					Pos:    float64(t.Pos) / twipsPerInch,
					Align:  t.Val,
					Leader: t.Leader,
				})
			}
		}
		if p.ParaProps.RunProps != nil {
			if p.ParaProps.RunProps.Bold != nil {
				pp.Bold = true
			}
			if p.ParaProps.RunProps.Italic != nil {
				pp.Italic = true
			}
			if p.ParaProps.RunProps.RFonts != nil {
				if p.ParaProps.RunProps.RFonts.Ascii != "" {
					pp.FontFamily = p.ParaProps.RunProps.RFonts.Ascii
				} else if p.ParaProps.RunProps.RFonts.HAnsi != "" {
					pp.FontFamily = p.ParaProps.RunProps.RFonts.HAnsi
				} else if p.ParaProps.RunProps.RFonts.EastAsia != "" {
					pp.FontFamily = p.ParaProps.RunProps.RFonts.EastAsia
				}
			}
			if p.ParaProps.RunProps.Sz != nil {
				pp.FontSizePt = halfPtToPt(p.ParaProps.RunProps.Sz.Val)
			}
			if p.ParaProps.RunProps.Color != nil && p.ParaProps.RunProps.Color.Val != "" {
				pp.FontColor = p.ParaProps.RunProps.Color.Val
			}
		}
	}

	for _, r := range p.RunContent {
		if r.FldChar != nil {
			pp.Runs = append(pp.Runs, TextRun{Text: r.FldChar.FldCharType, IsField: true})
			continue
		}
		if len(r.InstrText) > 0 {
			var instr string
			for _, t := range r.InstrText {
				instr += t.Text
			}
			pp.Runs = append(pp.Runs, TextRun{Text: instr, IsField: true})
			continue
		}
		if r.RunTab != nil {
			pp.Runs = append(pp.Runs, TextRun{IsTab: true})
			continue
		}
		if r.LastRenderedPageBreak != nil {
			pp.Runs = append(pp.Runs, TextRun{IsPageBreak: true})
			continue
		}

		if r.Drawing != nil || r.Pict != nil {
			tr := TextRun{IsImage: true}
			if r.Drawing != nil {
				dwg := r.Drawing
				var ext *wpExtent
				if dwg.Inline != nil {
					ext = dwg.Inline.Extent
				} else if dwg.Anchor != nil {
					ext = dwg.Anchor.Extent
				}
				if ext != nil {
					tr.ImageWidth = float64(ext.Cx) / 914400.0
					tr.ImageHeight = float64(ext.Cy) / 914400.0
				}
				for _, g := range []*aGraphic{getGraphicPtr(dwg.Inline, dwg.Anchor)} {
					if g != nil && g.Data != nil {
						if g.Data.Pic != nil {
							pic := g.Data.Pic
							if pic.BlipFill != nil && pic.BlipFill.Blip != nil {
								if path, ok := relMap[pic.BlipFill.Blip.Embed]; ok {
									tr.ImageSrc = path
								}
							}
							if pic.NvPicPr != nil && pic.NvPicPr.CNvPr != nil {
								tr.ImageAlt = pic.NvPicPr.CNvPr.Descr
							}
						}
						if g.Data.Wsp != nil && g.Data.Wsp.Txbx != nil && g.Data.Wsp.Txbx.TxbxContent != nil {
							tc := g.Data.Wsp.Txbx.TxbxContent
							if len(tc.Paras) > 0 || len(tc.Tables) > 0 {
								extracted := parseCellContent(tc.Paras, tc.Tables, styleMap, styleNameMap, numberingMap, numberingStartMap, relMap)
								tr.TextBoxContent = extracted
							}
						}
					}
					break
				}
			}
			if r.Pict != nil {
				for _, sh := range r.Pict.Shapes {
					if sh.ImageData != nil {
						if path, ok := relMap[sh.ImageData.RelID]; ok {
							tr.ImageSrc = path
						}
						if sh.ImageData.Title != "" {
							tr.ImageAlt = sh.ImageData.Title
						}
					}
					if sh.Style != "" {
						w, h := parseVmlStyle(sh.Style)
						if w > 0 {
							tr.ImageWidth = w
						}
						if h > 0 {
							tr.ImageHeight = h
						}
					}
					if sh.Textbox != nil && sh.Textbox.TxbxContent != nil {
						tc := sh.Textbox.TxbxContent
						if len(tc.Paras) > 0 || len(tc.Tables) > 0 {
							extracted := parseCellContent(tc.Paras, tc.Tables, styleMap, styleNameMap, numberingMap, numberingStartMap, relMap)
							tr.TextBoxContent = extracted
						}
					}
				}
			}
			pp.Runs = append(pp.Runs, tr)
			continue
		}

		tr := TextRun{}
		tr = applyRunProps(r, tr)
		for _, t := range r.RunText {
			tr.Text += t.Text
		}
		if r.FootnoteRef != nil {
			tr.IsFootnoteRef = true
			tr.NoteID = r.FootnoteRef.ID
		}
		if r.EndnoteRef != nil {
			tr.IsEndnoteRef = true
			tr.NoteID = r.EndnoteRef.ID
		}
		if r.RunBreak != nil {
			tr.IsLineBreak = true
			tr.BreakType = r.RunBreak.Type
		}
		if tr.Text != "" || tr.IsFootnoteRef || tr.IsEndnoteRef {
			pp.Runs = append(pp.Runs, tr)
		}
	}

	for _, ins := range p.Ins {
		for _, r := range ins.Runs {
			tr := TextRun{IsIns: true, InsID: ins.ID, InsAuthor: ins.Author, InsDate: ins.Date}
			tr = applyRunProps(r, tr)
			for _, t := range r.RunText {
				tr.Text += t.Text
			}
			if tr.Text != "" {
				pp.Runs = append(pp.Runs, tr)
			}
		}
	}

	for _, del := range p.Del {
		for _, r := range del.Runs {
			tr := TextRun{IsDel: true, InsID: del.ID, InsAuthor: del.Author, InsDate: del.Date}
			tr = applyRunProps(r, tr)
			for _, t := range r.DelText {
				tr.Text += t.Text
			}
			if tr.Text != "" {
				pp.Runs = append(pp.Runs, tr)
			}
		}
	}

	collapseFields(&pp)

	for _, hl := range p.Hyperlinks {
		tr := TextRun{IsHyperlink: true}
		if path, ok := relMap[hl.ID]; ok {
			tr.HyperlinkURL = path
		}
		for _, r := range hl.Runs {
			for _, t := range r.RunText {
				tr.Text += t.Text
			}
		}
		if tr.Text != "" {
			pp.Runs = append(pp.Runs, tr)
		}
	}

	detectCodeBlock(&pp, styleMap, styleNameMap)
	detectBlockQuote(&pp, styleNameMap)
	detectParagraphLang(&pp)

	return pp
}

func applyRunProps(r docRun, tr TextRun) TextRun {
	if r.RunProps == nil {
		return tr
	}
	if r.RunProps.Bold != nil {
		tr.Bold = true
	}
	if r.RunProps.Italic != nil {
		tr.Italic = true
	}
	if r.RunProps.Uline != nil {
		tr.Underline = r.RunProps.Uline.Val
	}
	if r.RunProps.Strike != nil {
		tr.Strike = true
	}
	if r.RunProps.DStrike != nil {
		tr.DStrike = true
	}
	if r.RunProps.VertAlign != nil {
		switch r.RunProps.VertAlign.Val {
		case "superscript":
			tr.SuperScript = true
		case "subscript":
			tr.SubScript = true
		}
	}
	if r.RunProps.RFonts != nil {
		if r.RunProps.RFonts.Ascii != "" {
			tr.FontFamily = r.RunProps.RFonts.Ascii
		} else if r.RunProps.RFonts.HAnsi != "" {
			tr.FontFamily = r.RunProps.RFonts.HAnsi
		} else if r.RunProps.RFonts.EastAsia != "" {
			tr.FontFamily = r.RunProps.RFonts.EastAsia
		}
		if r.RunProps.RFonts.EastAsia != "" {
			tr.FontEA = r.RunProps.RFonts.EastAsia
		}
		if r.RunProps.RFonts.CS != "" {
			tr.FontCS = r.RunProps.RFonts.CS
		}
	}
	if r.RunProps.Sz != nil {
		tr.FontSizePt = halfPtToPt(r.RunProps.Sz.Val)
	}
	if r.RunProps.Color != nil && r.RunProps.Color.Val != "" {
		tr.FontColor = r.RunProps.Color.Val
	}
	if r.RunProps.Highlight != nil && r.RunProps.Highlight.Val != "" {
		tr.Highlight = r.RunProps.Highlight.Val
	}
	if r.RunProps.SmallCaps != nil {
		tr.SmallCaps = true
	}
	if r.RunProps.Caps != nil {
		tr.AllCaps = true
	}
	if r.RunProps.Vanish != nil {
		tr.Hidden = true
	}
	if r.RunProps.Spacing != nil {
		tr.CharSpacing = r.RunProps.Spacing.Val
	}
	if r.RunProps.Position != nil {
		tr.Position = r.RunProps.Position.Val
	}
	if r.RunProps.Lang != nil {
		tr.Language = r.RunProps.Lang.Val
	}
	if r.RunProps.Em != nil {
		tr.Emphasis = r.RunProps.Em.Val
	}
	if r.RunProps.RTL != nil {
		tr.RTL = true
	}
	if r.RunProps.NoProof != nil {
		tr.NoProof = true
	}
	if r.RunProps.CS != nil {
		tr.CS = true
	}
	if r.RunProps.SpecVanish != nil {
		tr.SpecVanish = true
	}
	if r.RunProps.Emboss != nil {
		tr.Emboss = true
	}
	if r.RunProps.Engrave != nil {
		tr.Engrave = true
	}
	if r.RunProps.Shadow != nil {
		tr.Shadow = true
	}
	if r.RunProps.Imprint != nil {
		tr.Imprint = true
	}
	if r.RunProps.Bdr != nil {
		tr.Border = r.RunProps.Bdr.Val
	}
	if r.RunProps.Effect != nil {
		tr.Effect = true
	}
	if r.RunProps.Animate != nil {
		tr.Animate = true
	}
	if r.RunProps.BCs != nil {
		tr.BCs = true
	}
	if r.RunProps.ICs != nil {
		tr.ICs = true
	}
	if r.RunProps.SzCs != nil {
		tr.SizeCS = halfPtToPt(r.RunProps.SzCs.Val)
	}
	return tr
}

func detectCodeBlock(pp *ParsedParagraph, styleMap map[string]StyleDef, styleNameMap map[string]string) {
	sn := strings.ToLower(pp.StyleName)
	if sn == "code" || sn == "html" || sn == "xml" || sn == "plaintext" ||
		sn == "sourcecode" || sn == "example" || sn == "output" ||
		strings.Contains(sn, "code") || strings.Contains(sn, "source") ||
		strings.Contains(sn, "output") {
		pp.IsCode = true
		return
	}
	if pp.StyleID != "" && styleMap != nil {
		resolved := walkBasedOn(pp.StyleID, styleMap, styleNameMap)
		resolvedSn := strings.ToLower(resolved)
		if resolvedSn == "code" || resolvedSn == "html" || resolvedSn == "xml" ||
			resolvedSn == "plaintext" || resolvedSn == "sourcecode" ||
			resolvedSn == "example" || resolvedSn == "output" ||
			strings.Contains(resolvedSn, "code") || strings.Contains(resolvedSn, "source") ||
			strings.Contains(resolvedSn, "output") {
			pp.IsCode = true
			return
		}
	}
	font := strings.ToLower(pp.FontFamily)
	if font == "courier new" || font == "consolas" || font == "lucida console" ||
		font == "menlo" || font == "monaco" || font == "monospace" ||
		strings.Contains(font, "mono") || strings.Contains(font, "courier") {
		pp.IsCode = true
	}
}

func walkBasedOn(styleID string, styleMap map[string]StyleDef, styleNameMap map[string]string) string {
	seen := make(map[string]bool)
	current := styleID
	for current != "" && !seen[current] {
		seen[current] = true
		if name, ok := styleNameMap[current]; ok {
			_ = name
			// Check if this style has a heading level or known type
			if sd, ok := styleMap[current]; ok {
				if sd.HeadingLevel > 0 {
					return name
				}
				// Check by name
				lower := strings.ToLower(name)
				if lower == "code" || lower == "html" || lower == "xml" ||
					lower == "plaintext" || lower == "sourcecode" ||
					lower == "example" || lower == "output" ||
					lower == "quote" || lower == "intensequote" || lower == "blocktext" ||
					lower == "normal" || lower == "heading1" || lower == "heading2" ||
					lower == "heading3" || lower == "heading4" || lower == "heading5" ||
					lower == "heading6" || lower == "title" || lower == "subtitle" ||
					lower == "listparagraph" {
					return name
				}
			}
		}
		break
	}
	return ""
}

func detectBlockQuote(pp *ParsedParagraph, styleNameMap map[string]string) {
	sn := strings.ToLower(pp.StyleName)
	if sn == "quote" || sn == "intensequote" || sn == "blocktext" {
		pp.IsQuote = true
	}
}

func detectParagraphLang(pp *ParsedParagraph) {
	for _, r := range pp.Runs {
		if r.Language != "" {
			pp.Lang = r.Language
			return
		}
	}
}

// collapseFields replaces field marker runs (begin/instrText/separate/end) with a single
// [FIELD type=...] TextRun. Unknown field codes are left as plain text.
func collapseFields(pp *ParsedParagraph) {
	var out []TextRun
	i := 0
	for i < len(pp.Runs) {
		r := pp.Runs[i]
		if !r.IsField || r.Text != "begin" {
			out = append(out, r)
			i++
			continue
		}
		i++ // skip begin

		var instrBuf string
		var resultBuf string

		for i < len(pp.Runs) && pp.Runs[i].IsField && pp.Runs[i].Text != "separate" && pp.Runs[i].Text != "end" {
			instrBuf += pp.Runs[i].Text
			i++
		}

		if i < len(pp.Runs) && pp.Runs[i].IsField && pp.Runs[i].Text == "separate" {
			i++ // skip separate
		}

		for i < len(pp.Runs) && !(pp.Runs[i].IsField && pp.Runs[i].Text == "end") {
			resultBuf += pp.Runs[i].Text
			i++
		}

		if i < len(pp.Runs) && pp.Runs[i].IsField && pp.Runs[i].Text == "end" {
			i++ // skip end
		}

		ft, ff, fn := resolveField(instrBuf)
		if ft == "" {
			if resultBuf != "" {
				out = append(out, TextRun{Text: resultBuf})
			}
			continue
		}

		out = append(out, TextRun{
			IsField:    true,
			FieldType:  ft,
			FieldFormat: ff,
			FieldName:  fn,
		})
		if resultBuf != "" {
			out = append(out, TextRun{Text: resultBuf})
		}
	}
	pp.Runs = out
}

func resolveField(instr string) (ft, ff, fn string) {
	code := strings.TrimSpace(instr)
	if code == "" {
		return "", "", ""
	}
	switch {
	case strings.HasPrefix(code, "PAGE"):
		return "PAGE", "", ""
	case strings.HasPrefix(code, "NUMPAGES"):
		return "NUMPAGES", "", ""
	case strings.HasPrefix(code, "DATE"):
		ft = "DATE"
		if m := reDateFormat.FindStringSubmatch(code); len(m) > 1 {
			ff = m[1]
		}
		return
	case strings.HasPrefix(code, "MERGEFIELD"):
		ft = "MERGE"
		fn = strings.TrimSpace(strings.TrimPrefix(code, "MERGEFIELD"))
		if fn != "" && fn[0] == '"' && fn[len(fn)-1] == '"' {
			fn = fn[1 : len(fn)-1]
		}
		return
	case strings.HasPrefix(code, "TOC"):
		return "TOC", "", ""
	default:
		return "", "", ""
	}
}

var reDateFormat = regexp.MustCompile(`\\@\s*"([^"]+)"`)

type pageSize struct {
	Name string
	W, H float64
}

var standardSizes = []pageSize{
	{"A4", 8.27, 11.69},
	{"letter", 8.5, 11},
	{"legal", 8.5, 14},
	{"A3", 11.69, 16.54},
	{"A5", 5.83, 8.27},
	{"B5", 6.93, 9.84},
}

func approxEqual(a, b float64) bool {
	return math.Abs(a-b) < 0.02
}

func detectLayout(w, h float64) string {
	pw := math.Min(w, h)
	ph := math.Max(w, h)
	for _, s := range standardSizes {
		if approxEqual(pw, s.W) && approxEqual(ph, s.H) {
			return s.Name
		}
	}
	return "custom"
}

func marginValue(v float64) string {
	v = math.Round(v*100) / 100
	s := fmt.Sprintf("%.2f", v)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return s
}

func (doc *ParsedDocument) GenerateStyleBlock() string {
	var b strings.Builder
	b.WriteString("<style unit=\"in\">\n")

	sectionLayouts := doc.Sections
	if len(sectionLayouts) == 0 {
		sectionLayouts = []ParsedSection{{Layout: doc.PageLayout}}
	}
	for _, sec := range sectionLayouts {
		emitPageStyle(&b, sec.Layout, doc.PageLayout)
	}

	lh := doc.LineHeight
	if lh <= 0 {
		lh = 1.5
	}
	b.WriteString(fmt.Sprintf(`  <s:line el="p" value="%.1f" rule="auto"/>`, lh))
	b.WriteString("\n")

	for level := 1; level <= 6; level++ {
		hs, ok := doc.HeadingStyles[level]
		if !ok {
			continue
		}
		b.WriteString(fmt.Sprintf(`  <s:gap el="h" c="Heading%d"`, level))
		if hs.SpacingBefore > 0 {
			b.WriteString(fmt.Sprintf(` before="%.2f"`, hs.SpacingBefore/72.0))
		}
		if hs.SpacingAfter > 0 {
			b.WriteString(fmt.Sprintf(` after="%.2f"`, hs.SpacingAfter/72.0))
		}
		b.WriteString("/>\n")
	}

	secHasGap := make(map[int]bool)
	for _, sec := range sectionLayouts {
		if sec.ColCount > 1 {
			if secHasGap[sec.ColCount] {
				continue
			}
			secHasGap[sec.ColCount] = true
			space := ""
			if sec.ColSpace > 0 {
				space = fmt.Sprintf(` space="%.2f"`, sec.ColSpace/72.0)
			}
			b.WriteString(fmt.Sprintf(`  <s:cols n="%d"%s/>`+"\n", sec.ColCount, space))
		}
	}

	if doc.Theme != nil {
		b.WriteString(`  <s:theme`)
		if doc.Theme.Fg != "" {
			b.WriteString(fmt.Sprintf(` fg="%s"`, doc.Theme.Fg))
		}
		if doc.Theme.Bg != "" {
			b.WriteString(fmt.Sprintf(` bg="%s"`, doc.Theme.Bg))
		}
		b.WriteString("/>\n")
	}

	colIDs := make(map[int]bool)
	for _, tbl := range doc.AllTables {
		if len(tbl.Grid) == 0 {
			continue
		}
		if colIDs[tbl.ID] {
			continue
		}
		colIDs[tbl.ID] = true
		for _, w := range tbl.Grid {
			b.WriteString(fmt.Sprintf(`  <s:col ref="%d" width="%.2f" unit="pt"/>`+"\n", tbl.ID, w))
		}
	}

	indentSeen := make(map[string]bool)
	alignSeen := make(map[string]bool)
	tabSeen := make(map[string]bool)

	walkContent(doc.Content, func(p *ParsedParagraph) {
		if p.IndentLeft != 0 || p.IndentRight != 0 || p.FirstLineIndent != 0 || p.Hanging != 0 {
			key := fmt.Sprintf("%.4f|%.4f|%.4f|%.4f", p.IndentLeft, p.IndentRight, p.FirstLineIndent, p.Hanging)
			if indentSeen[key] {
				return
			}
			indentSeen[key] = true
			el := "p"
			if p.HeadingLevel > 0 {
				el = fmt.Sprintf("h%d", p.HeadingLevel)
			}
			var attrs []string
			if p.IndentLeft != 0 {
				attrs = append(attrs, fmt.Sprintf(` left="%s"`, marginValue(p.IndentLeft)))
			}
			if p.IndentRight != 0 {
				attrs = append(attrs, fmt.Sprintf(` right="%s"`, marginValue(p.IndentRight)))
			}
			if p.FirstLineIndent != 0 {
				attrs = append(attrs, fmt.Sprintf(` firstLine="%s"`, marginValue(p.FirstLineIndent)))
			}
			if p.Hanging != 0 {
				attrs = append(attrs, fmt.Sprintf(` hanging="%s"`, marginValue(p.Hanging)))
			}
			b.WriteString(fmt.Sprintf("  <s:indent el=\"%s\"", el))
			for _, a := range attrs {
				b.WriteString(a)
			}
			b.WriteString("/>\n")
		}
		if p.Align != "" && p.Align != "left" {
			key := p.Align
			if alignSeen[key] {
				return
			}
			alignSeen[key] = true
			el := "p"
			if p.HeadingLevel > 0 {
				el = fmt.Sprintf("h%d", p.HeadingLevel)
			}
			b.WriteString(fmt.Sprintf(`  <s:align el="%s" value="%s"/>`+"\n", el, p.Align))
		}
		for _, ts := range p.TabStops {
			key := fmt.Sprintf("%.4f|%s|%s", ts.Pos, ts.Align, ts.Leader)
			if tabSeen[key] {
				continue
			}
			tabSeen[key] = true
			el := "p"
			if p.HeadingLevel > 0 {
				el = fmt.Sprintf("h%d", p.HeadingLevel)
			}
			fmt.Fprintf(&b, `  <s:tab el="%s" pos="%s" align="%s"`, el, marginValue(ts.Pos), ts.Align)
			if ts.Leader != "" {
				fmt.Fprintf(&b, ` leader="%s"`, ts.Leader)
			}
			b.WriteString("/>\n")
		}
	})

	for _, def := range doc.CustomStyles {
		b.WriteString(fmt.Sprintf(`  <s:custom name="%s"`, xmlEscape(def.Name)))
		if def.BasedOn != "" {
			if name, ok := resolveBasedOnName(def.BasedOn, doc.CustomStyles); ok {
				b.WriteString(fmt.Sprintf(` basedOn="%s"`, xmlEscape(name)))
			}
		}
		if def.Family != "" {
			b.WriteString(fmt.Sprintf(` font="%s"`, def.Family))
		}
		if def.SizePt > 0 {
			b.WriteString(fmt.Sprintf(` size="%.0f"`, def.SizePt))
		}
		if def.Color != "" {
			b.WriteString(fmt.Sprintf(` color="%s"`, strings.TrimPrefix(def.Color, "#")))
		}
		if def.Bold {
			b.WriteString(` bold="true"`)
		}
		if def.Italic {
			b.WriteString(` italic="true"`)
		}
		if def.Underline != "" {
			b.WriteString(fmt.Sprintf(` underline="%s"`, def.Underline))
		}
		if def.Align != "" {
			b.WriteString(fmt.Sprintf(` alignment="%s"`, def.Align))
		}
		if def.SpacingBefore > 0 {
			b.WriteString(fmt.Sprintf(` spacingBefore="%.2f"`, def.SpacingBefore))
		}
		if def.SpacingAfter > 0 {
			b.WriteString(fmt.Sprintf(` spacingAfter="%.2f"`, def.SpacingAfter))
		}
		b.WriteString("/>\n")
	}

	b.WriteString("</style>\n")
	return b.String()
}

func emitPageStyle(b *strings.Builder, sec PageLayout, defaultLayout PageLayout) {
	if !sec.FromDocx {
		sec = defaultLayout
	}
	w := sec.WidthInch
	h := sec.HeightInch
	layout := detectLayout(w, h)

	b.WriteString(fmt.Sprintf("  <s:page size=%s", layout))
	if layout == "custom" {
		b.WriteString(fmt.Sprintf(` w="%.2f" h="%.2f"`, w, h))
	}

	mt := marginValue(sec.MarginTop)
	mr := marginValue(sec.MarginRight)
	mb := marginValue(sec.MarginBottom)
	ml := marginValue(sec.MarginLeft)

	b.WriteString(fmt.Sprintf(` mt=%s mb=%s ml=%s mr=%s`, mt, mb, ml, mr))

	mh := marginValue(sec.HeaderMargin)
	mf := marginValue(sec.FooterMargin)
	if mh == "0" {
		mh = "0.50"
	}
	if mf == "0" {
		mf = "0.50"
	}
	fmt.Fprintf(b, ` mh=%s mf=%s`, mh, mf)
	b.WriteString("/>\n")
}

func resolveBasedOnName(id string, customStyles []CustomStyleDef) (string, bool) {
	for _, cs := range customStyles {
		if cs.BasedOn == id || cs.Name == id {
			return cs.Name, true
		}
	}
	return id, false
}

func walkContent(content []ContentItem, fn func(p *ParsedParagraph)) {
	for _, ci := range content {
		switch ci.Type {
		case ContentParagraph:
			if ci.Paragraph != nil {
				fn(ci.Paragraph)
			}
		case ContentTable:
			if ci.Table != nil {
				for _, row := range ci.Table.Rows {
					for _, cell := range row.Cells {
						walkContent(cell.Content, fn)
					}
				}
			}
		}
	}
}

func (doc *ParsedDocument) FormatForLLM() string {
	var b strings.Builder

	modeAttr := "semantic"
	if doc.Mode == "lossless" {
		modeAttr = "lossless"
	}
	b.WriteString(fmt.Sprintf(`<words xmlns="urn:words:v1" xmlns:s="urn:words:v1:style" version="1.0.1" mode="%s">`+"\n", modeAttr))

	b.WriteString("<meta>\n")
	if doc.Title != "" {
		b.WriteString(fmt.Sprintf("<title>%s</title>\n", xmlEscape(doc.Title)))
	}
	if doc.Subject != "" {
		b.WriteString(fmt.Sprintf("<subject>%s</subject>\n", xmlEscape(doc.Subject)))
	}
	if doc.Author != "" {
		b.WriteString(fmt.Sprintf("<author>%s</author>\n", xmlEscape(doc.Author)))
	}
	if doc.Keywords != "" {
		b.WriteString(fmt.Sprintf("<keywords>%s</keywords>\n", xmlEscape(doc.Keywords)))
	}
	if doc.Description != "" {
		b.WriteString(fmt.Sprintf("<description>%s</description>\n", xmlEscape(doc.Description)))
	}
	if doc.Category != "" {
		b.WriteString(fmt.Sprintf("<category>%s</category>\n", xmlEscape(doc.Category)))
	}
	if doc.ContentStatus != "" {
		b.WriteString(fmt.Sprintf("<contentStatus>%s</contentStatus>\n", xmlEscape(doc.ContentStatus)))
	}
	if doc.LastModifiedBy != "" {
		b.WriteString(fmt.Sprintf("<lastModifiedBy>%s</lastModifiedBy>\n", xmlEscape(doc.LastModifiedBy)))
	}
	if doc.Revision != "" {
		b.WriteString(fmt.Sprintf("<revision>%s</revision>\n", xmlEscape(doc.Revision)))
	}
	if doc.Version != "" {
		b.WriteString(fmt.Sprintf("<version>%s</version>\n", xmlEscape(doc.Version)))
	}
	if doc.Created != "" {
		b.WriteString(fmt.Sprintf("<created>%s</created>\n", xmlEscape(doc.Created)))
	}
	if doc.Modified != "" {
		b.WriteString(fmt.Sprintf("<modified>%s</modified>\n", xmlEscape(doc.Modified)))
	}
	if doc.Language != "" {
		b.WriteString(fmt.Sprintf("<language>%s</language>\n", xmlEscape(doc.Language)))
	}
	if doc.Application != "" {
		b.WriteString(fmt.Sprintf("<application>%s</application>\n", xmlEscape(doc.Application)))
	}
	if doc.AppVersion != "" {
		b.WriteString(fmt.Sprintf("<appVersion>%s</appVersion>\n", xmlEscape(doc.AppVersion)))
	}
	b.WriteString("</meta>\n")

	b.WriteString(doc.GenerateStyleBlock())

	b.WriteString("<write>\n")

	sections := doc.Sections
	if len(sections) == 0 {
		sections = []ParsedSection{{Content: doc.Content}}
	}
	var hdrFtrID int
	for i, sec := range sections {
		emitHdrFtrBlock(&b, "header", sec.Headers, doc, &hdrFtrID)
		emitHdrFtrBlock(&b, "footer", sec.Footers, doc, &hdrFtrID)
		if i > 0 {
			b.WriteString(`<section-break`)
			if sec.BreakType != "" {
				b.WriteString(fmt.Sprintf(` type="%s"`, sec.BreakType))
			}
			if sec.Layout.FromDocx {
				layout := detectLayout(sec.Layout.WidthInch, sec.Layout.HeightInch)
				b.WriteString(fmt.Sprintf(` layout="%s"`, layout))
			}
			if sec.ColCount > 1 {
				b.WriteString(fmt.Sprintf(` columns="%d"`, sec.ColCount))
			}
			b.WriteString("/>\n")
		}
		formatContentItems(&b, sec.Content, doc)
	}

	b.WriteString("</write>\n")

	if len(doc.Notes) > 0 {
		b.WriteString("<notes>\n")
		for _, n := range doc.Notes {
			switch n.Type {
			case "footnote", "endnote":
				fmt.Fprintf(&b, `<fn id="%d" type="%s"`, n.ID, n.Type)
				if len(n.Body) == 0 {
					b.WriteString("/>\n")
				} else {
					b.WriteString(">\n")
					for _, ci := range n.Body {
						if ci.Type == ContentParagraph {
							formatParagraph(&b, ci.Paragraph, doc)
							b.WriteString("\n")
						} else if ci.Type == ContentTable {
							writeTable(&b, ci.Table, doc)
							b.WriteString("\n")
						}
					}
					fmt.Fprintf(&b, "</fn>\n")
				}
			case "bookmark":
				fmt.Fprintf(&b, `<bm id="%s"/>\n`, n.Name)
			case "comment":
				fmt.Fprintf(&b, `<comment id="%d"`, n.ID)
				if n.Author != "" {
					fmt.Fprintf(&b, ` author="%s"`, xmlEscape(n.Author))
				}
				if n.Date != "" {
					fmt.Fprintf(&b, ` date="%s"`, xmlEscape(n.Date))
				}
				if len(n.Body) == 0 {
					b.WriteString("/>\n")
				} else {
					b.WriteString(">\n")
					for _, ci := range n.Body {
						if ci.Type == ContentParagraph {
							formatParagraph(&b, ci.Paragraph, doc)
							b.WriteString("\n")
						}
					}
					b.WriteString("</comment>\n")
				}
			}
		}
		b.WriteString("</notes>\n")
	}

	b.WriteString("</words>\n")

	return b.String()
}

func formatParagraph(b *strings.Builder, p *ParsedParagraph, doc *ParsedDocument) {
	var textboxItems []ContentItem
	for _, r := range p.Runs {
		if len(r.TextBoxContent) > 0 {
			textboxItems = append(textboxItems, r.TextBoxContent...)
		}
	}

	if p.HeadingLevel > 0 && p.HeadingLevel <= 9 {
		content := buildInlineText(*p, doc.Mode)
		fmt.Fprintf(b, "<h%d", p.HeadingLevel)
		writeParagraphAttrs(b, p)
		b.WriteString(">")
		b.WriteString(content)
		fmt.Fprintf(b, "</h%d>", p.HeadingLevel)
		emitTextBoxSiblings(b, textboxItems, doc)
		return
	}

	if p.IsCode {
		content := buildPlainText(*p)
		b.WriteString("<pre")
		writeParagraphAttrs(b, p)
		b.WriteString(">")
		b.WriteString(content)
		b.WriteString("</pre>")
		emitTextBoxSiblings(b, textboxItems, doc)
		return
	}

	if p.IsQuote {
		content := buildInlineText(*p, doc.Mode)
		b.WriteString("<blockquote")
		writeParagraphAttrs(b, p)
		b.WriteString(">")
		b.WriteString(content)
		b.WriteString("</blockquote>")
		emitTextBoxSiblings(b, textboxItems, doc)
		return
	}

	content := buildInlineText(*p, doc.Mode)
	b.WriteString("<p")
	writeParagraphAttrs(b, p)
	b.WriteString(">")
	b.WriteString(content)
	b.WriteString("</p>")
	emitTextBoxSiblings(b, textboxItems, doc)
}

func emitTextBoxSiblings(b *strings.Builder, items []ContentItem, doc *ParsedDocument) {
	if len(items) == 0 {
		return
	}
	b.WriteString("\n")
	for _, ci := range items {
		switch ci.Type {
		case ContentParagraph:
			formatParagraph(b, ci.Paragraph, doc)
		case ContentTable:
			writeTable(b, ci.Table, doc)
		}
		b.WriteString("\n")
	}
}

func writeParagraphAttrs(b *strings.Builder, p *ParsedParagraph) {
	if p.Bidi {
		b.WriteString(` dir="rtl"`)
	}
	if p.Lang != "" {
		fmt.Fprintf(b, ` lang="%s"`, p.Lang)
	}
	if p.StyleName != "" {
		fmt.Fprintf(b, ` c="%s"`, xmlEscape(p.StyleName))
	}
	at := buildBorderAttr(p)
	if at != "" {
		b.WriteString(at)
	}
}

func buildPlainText(p ParsedParagraph) string {
	var b strings.Builder
	for _, r := range p.Runs {
		if r.IsLineBreak {
			b.WriteString("\n")
			continue
		}
		if r.IsTab {
			b.WriteString("\t")
			continue
		}
		if r.IsField || r.Hidden {
			continue
		}
		if r.Text != "" {
			b.WriteString(xmlEscape(r.Text))
		}
	}
	return strings.TrimSpace(b.String())
}

func borderValToStyle(val string) string {
	switch val {
	case "single":
		return "s"
	case "double":
		return "d"
	case "dashed", "dashSmallGap":
		return "ds"
	case "dotted":
		return "dt"
	case "none", "nil":
		return "n"
	case "triple":
		return "t"
	case "thinThickThinSmall", "thickThinThickSmall":
		return "ts"
	case "thinThickThinMedium", "thickThinThickMedium":
		return "tm"
	default:
		return val
	}
}

func formatBorderSide(side string, bi *BorderInfo) string {
	style := borderValToStyle(bi.Val)
	return fmt.Sprintf("%s %d %s%d #%s", side, bi.Sz, style, bi.Space, bi.Color)
}

func buildBorderAttrFromBorders(bt, bb, bl, br *BorderInfo) string {
	parts := make([]string, 0, 4)
	if bt != nil {
		parts = append(parts, formatBorderSide("bt", bt))
	}
	if bb != nil {
		parts = append(parts, formatBorderSide("bb", bb))
	}
	if bl != nil {
		parts = append(parts, formatBorderSide("bl", bl))
	}
	if br != nil {
		parts = append(parts, formatBorderSide("br", br))
	}
	if len(parts) == 0 {
		return ""
	}
	return fmt.Sprintf(` at="%s"`, strings.Join(parts, "; "))
}

func buildBorderAttr(p *ParsedParagraph) string {
	return buildBorderAttrFromBorders(p.BorderTop, p.BorderBottom, p.BorderLeft, p.BorderRight)
}

func listFormatToTag(fmt string) string {
	switch fmt {
	case "decimal", "lowerRoman", "upperRoman", "lowerLetter", "upperLetter", "ordinal", "decimalZero":
		return "ol"
	default:
		return "ul"
	}
}

func formatListGroup(b *strings.Builder, content []ContentItem, start int, doc *ParsedDocument) int {
	end := start
	for end < len(content) && content[end].Type == ContentParagraph && content[end].Paragraph.IsList {
		end++
	}
	if end == start {
		return start
	}

	type frame struct {
		tag    string
		liOpen bool
	}
	stack := make([]frame, 0, 4)

	for i := start; i < end; i++ {
		p := content[i].Paragraph
		lvl := p.ListLevel

		for len(stack) > lvl+1 {
			if stack[len(stack)-1].liOpen {
				b.WriteString("</li>\n")
			}
			b.WriteString(fmt.Sprintf("</%s>\n", stack[len(stack)-1].tag))
			stack = stack[:len(stack)-1]
		}

		if lvl+1 > len(stack) {
			for k := len(stack); k <= lvl; k++ {
				tag := listFormatToTag(p.ListFormat)
				stack = append(stack, frame{tag: tag})
				b.WriteString(fmt.Sprintf("<%s>\n", tag))
				if k < lvl {
					b.WriteString("<li>")
					stack[k].liOpen = true
				}
			}
			if stack[lvl].liOpen {
				b.WriteString("</li>\n")
			}
			b.WriteString("<li>")
			stack[lvl].liOpen = true
		} else if lvl+1 == len(stack) {
			if stack[lvl].liOpen {
				b.WriteString("</li>\n")
			}
			b.WriteString("<li>")
			stack[lvl].liOpen = true
		} else {
			if stack[lvl].liOpen {
				b.WriteString("</li>\n")
			}
			b.WriteString("<li>")
			stack[lvl].liOpen = true
		}

		b.WriteString(buildInlineText(*p, doc.Mode))
	}

	for i := len(stack) - 1; i >= 0; i-- {
		if stack[i].liOpen {
			b.WriteString("</li>\n")
		}
		b.WriteString(fmt.Sprintf("</%s>\n", stack[i].tag))
	}

	return end
}

func formatContentItems(b *strings.Builder, content []ContentItem, doc *ParsedDocument) {
	i := 0
	for i < len(content) {
		c := content[i]
		if c.Type == ContentParagraph && c.Paragraph.IsList && !c.Paragraph.IsCode && !c.Paragraph.IsQuote {
			i = formatListGroup(b, content, i, doc)
			continue
		}
		switch c.Type {
		case ContentParagraph:
			p := c.Paragraph
			if p.PageBreakBefore {
				b.WriteString(`<br type="page"/>` + "\n")
			}
			formatParagraph(b, p, doc)
			b.WriteString("\n")
		case ContentTable:
			writeTable(b, c.Table, doc)
			b.WriteString("\n")
		}
		i++
	}
}

func writeTable(b *strings.Builder, t *ParsedTable, doc *ParsedDocument) {
	nc := len(t.Grid)
	if nc == 0 && len(t.Rows) > 0 {
		nc = len(t.Rows[0].Cells)
	}
	fmt.Fprintf(b, `<table id="%d" cols="%d"`, t.ID, nc)
	if t.StyleName != "" {
		fmt.Fprintf(b, ` c="%s"`, xmlEscape(t.StyleName))
	}
	if t.Width > 0 {
		fmt.Fprintf(b, ` width="%.2f"`, t.Width)
	}
	if t.Alignment != "" {
		fmt.Fprintf(b, ` align="%s"`, t.Alignment)
	}
	if t.Caption != "" {
		fmt.Fprintf(b, ` caption="%s"`, xmlEscape(t.Caption))
	}
	if t.Summary != "" {
		fmt.Fprintf(b, ` summary="%s"`, xmlEscape(t.Summary))
	}
	if t.Indent > 0 {
		fmt.Fprintf(b, ` indent="%.2f"`, t.Indent)
	}
	if t.CellSpacing > 0 {
		fmt.Fprintf(b, ` cellSpacing="%.2f"`, t.CellSpacing)
	}
	at := buildBorderAttrFromBorders(t.BorderTop, t.BorderBottom, t.BorderLeft, t.BorderRight)
	if at != "" {
		b.WriteString(at)
	}
	b.WriteString(">\n")
	for _, row := range t.Rows {
		if row.IsHeader {
			b.WriteString("<tr>\n")
			for _, cell := range row.Cells {
				writeTableCell(b, "th", &cell, doc)
			}
			b.WriteString("\n</tr>\n")
		} else {
			b.WriteString("<tr>\n")
			for _, cell := range row.Cells {
				writeTableCell(b, "td", &cell, doc)
			}
			b.WriteString("\n</tr>\n")
		}
	}
	b.WriteString("</table>")
}

func writeTableCell(b *strings.Builder, tag string, cell *ParsedTableCell, doc *ParsedDocument) {
	fmt.Fprintf(b, "<%s", tag)
	if cell.GridSpan > 1 {
		fmt.Fprintf(b, ` colspan="%d"`, cell.GridSpan)
	}
	if cell.RowSpan > 1 {
		fmt.Fprintf(b, ` rowspan="%d"`, cell.RowSpan)
	}
	if cell.VAlign != "" {
		fmt.Fprintf(b, ` valign="%s"`, cell.VAlign)
	}
	if cell.TextDirection != "" {
		fmt.Fprintf(b, ` textDir="%s"`, cell.TextDirection)
	}
	if cell.NoWrap {
		b.WriteString(` noWrap="true"`)
	}
	at := buildBorderAttrFromBorders(cell.BorderTop, cell.BorderBottom, cell.BorderLeft, cell.BorderRight)
	if at != "" {
		b.WriteString(at)
	}
	b.WriteString(">")
	formatCellContent(b, cell.Content, doc)
	fmt.Fprintf(b, "</%s>", tag)
}

func formatCellContent(b *strings.Builder, content []ContentItem, doc *ParsedDocument) {
	i := 0
	for i < len(content) {
		c := content[i]
		if c.Type == ContentParagraph && c.Paragraph.IsList && !c.Paragraph.IsCode && !c.Paragraph.IsQuote {
			i = formatListGroup(b, content, i, doc)
			continue
		}
		switch c.Type {
		case ContentParagraph:
			formatParagraph(b, c.Paragraph, doc)
		case ContentTable:
			writeTable(b, c.Table, doc)
		}
		i++
	}
}

func emitHdrFtrBlock(b *strings.Builder, xmlTag string, items []ParsedHdrFtr, doc *ParsedDocument, idCounter *int) {
	for _, hf := range items {
		*idCounter++
		fmt.Fprintf(b, `<%s id="%d"`, xmlTag, *idCounter)
		if hf.Type != "" && hf.Type != "default" {
			fmt.Fprintf(b, ` type="%s"`, hf.Type)
		}
		b.WriteString(">\n")
		formatContentItems(b, hf.Content, doc)
		fmt.Fprintf(b, "</%s>\n", xmlTag)
	}
}

func buildInlineText(p ParsedParagraph, mode string) string {
	if len(p.Runs) == 0 {
		return ""
	}

	paraBold := p.Bold
	paraItalic := p.Italic

	anyBold := false
	anyItalic := false
	for _, r := range p.Runs {
		if r.Bold {
			anyBold = true
		}
		if r.Italic {
			anyItalic = true
		}
	}

	uniformBold := paraBold && !anyBold
	uniformItalic := paraItalic && !anyItalic

	var b strings.Builder
	for _, r := range p.Runs {
		if r.IsLineBreak {
			if r.BreakType != "" && r.BreakType != "textWrapping" {
				fmt.Fprintf(&b, `<br type="%s"/>`, r.BreakType)
			} else {
				b.WriteString("<br/>")
			}
			continue
		}
		if r.IsTab {
			b.WriteString("<tab/>")
			continue
		}
		if r.IsPageBreak {
			b.WriteString("<br type=\"page\"/>")
			continue
		}
		if r.IsImage {
			alt := r.ImageAlt
			if alt == "" {
				alt = ""
			}
			b.WriteString(`<img`)
			if alt != "" {
				fmt.Fprintf(&b, ` alt="%s"`, xmlEscape(alt))
			}
			if r.ImageSrc != "" {
				fmt.Fprintf(&b, ` src="%s"`, xmlEscape(r.ImageSrc))
			}
			if r.ImageWidth > 0 {
				fmt.Fprintf(&b, ` width="%.2f"`, r.ImageWidth)
			}
			if r.ImageHeight > 0 {
				fmt.Fprintf(&b, ` height="%.2f"`, r.ImageHeight)
			}
			b.WriteString("/>")
			continue
		}
		if r.IsField {
			continue
		}
		if r.IsHyperlink {
			b.WriteString(fmt.Sprintf(`<a href="%s">%s</a>`, xmlEscape(r.HyperlinkURL), xmlEscape(r.Text)))
			continue
		}
		if r.Hidden {
			continue
		}
		if r.IsFootnoteRef {
			b.WriteString(fmt.Sprintf(`<fn-ref id="%d" type="footnote"/>`, r.NoteID))
			continue
		}
		if r.IsEndnoteRef {
			b.WriteString(fmt.Sprintf(`<fn-ref id="%d" type="endnote"/>`, r.NoteID))
			continue
		}
		if len(r.TextBoxContent) > 0 {
			continue
		}
		if r.IsDel && mode != "lossless" {
			continue
		}
		t := r.Text
		if t == "" {
			continue
		}

		hasSpan := false
		spanAttrs := ""
		if r.FontFamily != "" && r.FontFamily != p.FontFamily {
			spanAttrs += fmt.Sprintf(` font="%s"`, r.FontFamily)
			hasSpan = true
		}
		if r.FontEA != "" {
			spanAttrs += fmt.Sprintf(` fontEA="%s"`, r.FontEA)
			hasSpan = true
		}
		if r.FontCS != "" {
			spanAttrs += fmt.Sprintf(` fontCS="%s"`, r.FontCS)
			hasSpan = true
		}
		if r.FontSizePt > 0 && r.FontSizePt != p.FontSizePt {
			spanAttrs += fmt.Sprintf(` size="%.0f"`, r.FontSizePt)
			hasSpan = true
		}
		if r.SizeCS > 0 && r.SizeCS != p.FontSizePt {
			spanAttrs += fmt.Sprintf(` sizeCS="%.0f"`, r.SizeCS)
			hasSpan = true
		}
		if r.FontColor != "" && r.FontColor != p.FontColor {
			color := strings.TrimPrefix(r.FontColor, "#")
			spanAttrs += fmt.Sprintf(` color="%s"`, color)
			hasSpan = true
		}
		if r.Highlight != "" && r.Highlight != "none" {
			spanAttrs += fmt.Sprintf(` highlight="%s"`, r.Highlight)
			hasSpan = true
		}

		if hasSpan {
			t = fmt.Sprintf(`<span%s>%s</span>`, spanAttrs, t)
		}
		if r.SmallCaps {
			t = "<smallcaps>" + t + "</smallcaps>"
		}
		if r.AllCaps {
			t = "<uppercase>" + t + "</uppercase>"
		}
		if r.DStrike {
			t = `<s type="double">` + t + "</s>"
		} else if r.Strike {
			t = "<s>" + t + "</s>"
		}
		if r.SuperScript {
			t = "<sup>" + t + "</sup>"
		}
		if r.SubScript {
			t = "<sub>" + t + "</sub>"
		}
		if r.Underline != "" {
			if r.Underline == "single" || r.Underline == "" {
				t = "<u>" + t + "</u>"
			} else {
				t = fmt.Sprintf(`<u underline="%s">%s</u>`, r.Underline, t)
			}
		}
		if r.Bold && !paraBold {
			t = "<b>" + t + "</b>"
		}
		if r.Italic && !paraItalic {
			t = "<i>" + t + "</i>"
		}
		if r.BCs {
			t = "<bcs>" + t + "</bcs>"
		}
		if r.ICs {
			t = "<ics>" + t + "</ics>"
		}

		if r.IsIns && mode == "lossless" {
			t = `<ins` + buildInsAttrs(r) + `>` + t + "</ins>"
		}
		if r.IsDel && mode == "lossless" {
			t = `<del` + buildInsAttrs(r) + `>` + t + "</del>"
		}

		b.WriteString(t)
	}

	text := strings.TrimSpace(b.String())

	if uniformBold {
		text = "<b>" + text + "</b>"
	}
	if uniformItalic {
		text = "<i>" + text + "</i>"
	}

	return text
}

func buildInsAttrs(r TextRun) string {
	var attrs string
	if r.InsID > 0 {
		attrs += fmt.Sprintf(` id="%d"`, r.InsID)
	}
	if r.InsAuthor != "" {
		attrs += fmt.Sprintf(` author="%s"`, xmlEscape(r.InsAuthor))
	}
	if r.InsDate != "" {
		attrs += fmt.Sprintf(` date="%s"`, xmlEscape(r.InsDate))
	}
	return attrs
}

func xmlEscape(s string) string {
	s = sanitizeForbiddenXMLChars(s)
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	return s
}

func sanitizeForbiddenXMLChars(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r == 0x09 || r == 0x0A || r == 0x0D {
			b.WriteRune(r)
			continue
		}
		if r >= 0x20 && r != 0x7F {
			b.WriteRune(r)
			continue
		}
	}
	return b.String()
}
