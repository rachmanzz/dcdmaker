package dcdmaker

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

const (
	twipsPerInch = 1440
	halfPtPerPt  = 2
)

type DocxContent struct {
	DocumentXML string
	HasHeader   bool
	HasFooter   bool
}

func extractDocxContent(data []byte) (*DocxContent, error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("open docx: %w", err)
	}

	var docXML strings.Builder
	content := &DocxContent{}

	for _, f := range r.File {
		switch {
		case f.Name == "word/document.xml":
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("read document.xml: %w", err)
			}
			buf := new(bytes.Buffer)
			_, err = buf.ReadFrom(rc)
			rc.Close()
			if err != nil {
				return nil, fmt.Errorf("read document.xml: %w", err)
			}
			docXML.WriteString(buf.String())

		case strings.HasPrefix(f.Name, "word/header") && strings.HasSuffix(f.Name, ".xml"):
			content.HasHeader = true

		case strings.HasPrefix(f.Name, "word/footer") && strings.HasSuffix(f.Name, ".xml"):
			content.HasFooter = true
		}
	}

	if docXML.Len() == 0 {
		return nil, fmt.Errorf("word/document.xml not found in DOCX")
	}

	content.DocumentXML = docXML.String()
	return content, nil
}

type ParsedDocument struct {
	PageLayout  PageLayout
	DefaultFont FontDef
	Paragraphs  []ParsedParagraph
}

type PageLayout struct {
	WidthInch  float64
	HeightInch float64
	MarginTop  float64
	MarginRight float64
	MarginBottom float64
	MarginLeft float64
}

type FontDef struct {
	Family string
	SizePt float64
	Color  string
}

type ParsedParagraph struct {
	StyleID      string
	Text         string
	Align        string
	IndentLeft   float64
	Hanging      float64
	Bold         bool
	Italic       bool
	FontFamily   string
	FontSizePt   float64
	FontColor    string
	HeadingLevel int
	IsList       bool
	ListType     string
}

type docStyles struct {
	XMLName   xml.Name     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main styles"`
	DocDefaults *docDefaults `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main docDefaults"`
	Styles     []styleDef  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main style"`
}

type docDefaults struct {
	RunPropsDefault *runPropsDefault `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main rPrDefault"`
}

type runPropsDefault struct {
	RunProps *runProps `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main rPr"`
}

type runProps struct {
	RFonts *rFonts `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main rFonts"`
	Sz     *szVal  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main sz"`
	Color  *colorVal `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main color"`
}

type rFonts struct {
	Ascii    string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main ascii,attr"`
	HAnsi    string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main hAnsi,attr"`
	EastAsia string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main eastAsia,attr"`
}

type szVal struct {
	Val int `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr"`
}

type colorVal struct {
	Val string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr"`
}

type styleDef struct {
	ID         string      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main id,attr"`
	Type       string      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main type,attr"`
	OutlineLvl *outlineLvl `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main outlineLvl"`
	RunProps    *runProps   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main rPr"`
}

type outlineLvl struct {
	Val int `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr"`
}

type docDocument struct {
	XMLName xml.Name     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main document"`
	Body    docBody      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main body"`
}

type docBody struct {
	XMLName xml.Name     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main body"`
	Content []docPara    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main p"`
	Sections []docSection `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main sectPr"`
}

type docSection struct {
	PageSz  *docPageSize `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main pgSz"`
	PageMar *docPageMar  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main pgMar"`
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
}

type docPara struct {
	ParaProps  *paraProps `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main pPr"`
	RunContent []docRun   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main r"`
}

type paraProps struct {
	Style   *styleRef    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main pStyle"`
	JC      *jcVal       `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main jc"`
	Ind     *indVal      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main ind"`
	NumPr   *numPr       `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main numPr"`
}

type styleRef struct {
	Val string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr"`
}

type jcVal struct {
	Val string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr"`
}

type indVal struct {
	Left   int `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main left,attr"`
	Hanging int `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main hanging,attr"`
}

type numPr struct {
	Ilvl *intVal `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main ilvl"`
	NumID *intVal `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main numId"`
}

type intVal struct {
	Val int `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr"`
}

type docRun struct {
	RunProps  *runProps `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main rPr"`
	RunText   []docText `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main t"`
	RunBreak  *struct{} `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main br"`
}

type docText struct {
	Text string `xml:",chardata"`
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

	var docXML, stylesXML []byte

	for _, f := range r.File {
		switch f.Name {
		case "word/document.xml":
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("read document.xml: %w", err)
			}
			docXML, err = readAll(rc)
			rc.Close()
			if err != nil {
				return nil, fmt.Errorf("read document.xml: %w", err)
			}
		case "word/styles.xml":
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("read styles.xml: %w", err)
			}
			stylesXML, err = readAll(rc)
			rc.Close()
			if err != nil {
				return nil, fmt.Errorf("read styles.xml: %w", err)
			}
		}
	}

	if len(docXML) == 0 {
		return nil, fmt.Errorf("word/document.xml not found in DOCX")
	}

	if len(stylesXML) > 0 {
		styleMap := buildStyleMap(stylesXML)
		if def, ok := styleMap["default"]; ok {
			if def.Family != "" {
				doc.DefaultFont.Family = def.Family
			}
			if def.SizePt > 0 {
				doc.DefaultFont.SizePt = def.SizePt
			}
			if def.Color != "" {
				doc.DefaultFont.Color = def.Color
			}
		}
	}

	var document docDocument
	if err := xml.Unmarshal(docXML, &document); err != nil {
		return nil, fmt.Errorf("parse document.xml: %w", err)
	}

	body := document.Body

	if len(body.Sections) > 0 {
		sec := body.Sections[len(body.Sections)-1]
		if sec.PageSz != nil {
			doc.PageLayout.WidthInch = parseTwipsToInches(sec.PageSz.W)
			doc.PageLayout.HeightInch = parseTwipsToInches(sec.PageSz.H)
		}
		if sec.PageMar != nil {
			doc.PageLayout.MarginTop = parseTwipsToInches(sec.PageMar.Top)
			doc.PageLayout.MarginRight = parseTwipsToInches(sec.PageMar.Right)
			doc.PageLayout.MarginBottom = parseTwipsToInches(sec.PageMar.Bottom)
			doc.PageLayout.MarginLeft = parseTwipsToInches(sec.PageMar.Left)
		}
	}

	styleMap := make(map[string]StyleDef)
	if len(stylesXML) > 0 {
		styleMap = buildStyleMap(stylesXML)
	}

	for _, p := range body.Content {
		pp := parseDocParagraph(p, styleMap)
		if pp.Text != "" {
			doc.Paragraphs = append(doc.Paragraphs, pp)
		}
	}

	return doc, nil
}

func readAll(r io.Reader) ([]byte, error) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r)
	return buf.Bytes(), err
}

type StyleDef struct {
	Family        string
	SizePt        float64
	Color         string
	Bold          bool
	HeadingLevel  int
}

func buildStyleMap(stylesXML []byte) map[string]StyleDef {
	var styles docStyles
	if err := xml.Unmarshal(stylesXML, &styles); err != nil {
		return nil
	}

	m := make(map[string]StyleDef)

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
		}
		if s.OutlineLvl != nil {
			sd.HeadingLevel = s.OutlineLvl.Val + 1
		}
		m[s.ID] = sd
	}

	return m
}

func parseDocParagraph(p docPara, styleMap map[string]StyleDef) ParsedParagraph {
	pp := ParsedParagraph{}

	if p.ParaProps != nil {
		if p.ParaProps.Style != nil {
			pp.StyleID = p.ParaProps.Style.Val
			if sd, ok := styleMap[pp.StyleID]; ok {
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
			}
		}
		if p.ParaProps.JC != nil {
			pp.Align = p.ParaProps.JC.Val
		}
		if p.ParaProps.Ind != nil {
			pp.IndentLeft = parseTwipsToInches(p.ParaProps.Ind.Left)
			pp.Hanging = parseTwipsToInches(p.ParaProps.Ind.Hanging)
		}
		if p.ParaProps.NumPr != nil {
			pp.IsList = true
			if p.ParaProps.NumPr.Ilvl != nil && p.ParaProps.NumPr.Ilvl.Val == 0 {
				pp.ListType = "ol"
			}
		}
	}

	for _, r := range p.RunContent {
		if r.RunProps != nil {
			if r.RunProps.RFonts != nil {
				if r.RunProps.RFonts.Ascii != "" {
					pp.FontFamily = r.RunProps.RFonts.Ascii
				} else if r.RunProps.RFonts.HAnsi != "" {
					pp.FontFamily = r.RunProps.RFonts.HAnsi
				}
			}
			if r.RunProps.Sz != nil {
				pp.FontSizePt = halfPtToPt(r.RunProps.Sz.Val)
			}
			if r.RunProps.Color != nil && r.RunProps.Color.Val != "" {
				pp.FontColor = r.RunProps.Color.Val
			}
		}
		for _, t := range r.RunText {
			pp.Text += t.Text
		}
		if r.RunBreak != nil {
			pp.Text += "\n"
		}
	}

	return pp
}

func (doc *ParsedDocument) FormatForLLM() string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("[PAGE] %.2fx%.2f in, margins: top=%.2f right=%.2f bottom=%.2f left=%.2f in\n",
		doc.PageLayout.WidthInch, doc.PageLayout.HeightInch,
		doc.PageLayout.MarginTop, doc.PageLayout.MarginRight,
		doc.PageLayout.MarginBottom, doc.PageLayout.MarginLeft))

	if doc.DefaultFont.Family != "" || doc.DefaultFont.SizePt > 0 {
		b.WriteString("[DEFAULT]")
		if doc.DefaultFont.Family != "" {
			b.WriteString(fmt.Sprintf(" font-family=%s", doc.DefaultFont.Family))
		}
		if doc.DefaultFont.SizePt > 0 {
			b.WriteString(fmt.Sprintf(" font-size=%.0fpt", doc.DefaultFont.SizePt))
		}
		if doc.DefaultFont.Color != "" {
			b.WriteString(fmt.Sprintf(" color=%s", doc.DefaultFont.Color))
		}
		b.WriteString("\n")
	}

	for _, p := range doc.Paragraphs {
		if p.HeadingLevel > 0 {
			b.WriteString(fmt.Sprintf("<h%d>", p.HeadingLevel))
		} else if p.IsList {
			b.WriteString("[LI]")
		} else {
			b.WriteString("[P")
			if p.Align != "" && p.Align != "left" {
				b.WriteString(fmt.Sprintf(" align=%s", p.Align))
			}
			if p.IndentLeft > 0 {
				b.WriteString(fmt.Sprintf(" indent=%.2f", p.IndentLeft))
			}
			if p.Hanging > 0 {
				b.WriteString(fmt.Sprintf(" hanging=%.2f", p.Hanging))
			}
			if p.FontFamily != "" && p.FontFamily != doc.DefaultFont.Family {
				b.WriteString(fmt.Sprintf(" font-family=%s", p.FontFamily))
			}
			if p.FontSizePt > 0 && p.FontSizePt != doc.DefaultFont.SizePt {
				b.WriteString(fmt.Sprintf(" font-size=%.0fpt", p.FontSizePt))
			}
			if p.FontColor != "" && p.FontColor != doc.DefaultFont.Color {
				b.WriteString(fmt.Sprintf(" color=%s", p.FontColor))
			}
			b.WriteString("]")
		}

		text := strings.TrimSpace(p.Text)
		if p.Bold {
			text = "<b>" + text + "</b>"
		}
		if p.Italic {
			text = "<i>" + text + "</i>"
		}
		b.WriteString(text)
		b.WriteString("\n")
	}

	return b.String()
}
