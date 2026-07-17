# Plan: XML Cleaner for DCD Maker

## Status: PENDING APPROVAL

## Problem

Raw DOCX XML injected into the LLM is too large:

| Metric | Value |
|---|---|
| DOCX file (compressed) | 59,284 bytes |
| `word/document.xml` (uncompressed) | 1,471,576 bytes (1.4MB) |
| Estimated tokens from XML | ~367,872 tokens |
| Plain text tokens | ~13,352 tokens |
| **Multiplier** | **28x** |
| Model context window | 262,144 tokens |

XML alone exceeds the model context window. This causes context length to reach 262K tokens.

## Root Cause

DOCX XML contains 66,060 tags including:
- 7,252 `w:rPr` (run properties/formatting)
- 7,252 `w:rFonts` (font definitions)
- 5,698 `w:proofErr` (proofing errors)
- 2,429 `w:lang` (language tags)
- 1,940 `w:bCs` (complex script bold)
- 1,787 `w:tab` (tab characters)
- 413 `w:pPr` (paragraph properties)
- 413 `w:tabs` (tab stops)
- 390 `w:widowControl`, `w:autoSpaceDE/DN`, `w:adjustRightInd`

Each text character is wrapped hierarchically: `<w:p><w:r><w:rPr>...<w:rPr><w:t>text</w:t></w:r></w:p>` — 5-10x overhead per character.

## Proposed Solution: DOCX Parser in `docx.go`

### Approach

Parse DOCX XML + `styles.xml`, extract structured data needed by the LLM to generate DCD. Output cleaned structured text, not raw XML.

### What to Extract

| Info | Source | DCD Mapping |
|---|---|---|
| Page layout | `document.xml` → `w:pgSz` | `[style] layout=A4, w, h` |
| Margins | `document.xml` → `w:pgMar` | `[style] m-t, m-r, m-b, m-l` |
| Default font | `styles.xml` → `w:docDefaults` | `[style] font-family, font-size` |
| Heading level | `styles.xml` → `w:outlineLvl` | `<h1>`-`<h6>` |
| Paragraph style | `document.xml` → `w:pStyle` | Map to style properties |
| Alignment | `document.xml` → `w:jc` | `align=center/justify` |
| Indentation | `document.xml` → `w:ind` | `indent=X, hanging=Y` |
| Bold | `document.xml` → `w:b` | `<b>` |
| Text content | `document.xml` → `w:t` | Preserve exact |
| List/numbering | `document.xml` → `w:numPr` | `<ol>`/`<ul>` |

### What to Strip

| Tag | Count | Purpose | Action |
|---|---|---|---|
| `w:proofErr` | 5,698 | Proofing errors | Strip |
| `w:lang` | 2,429 | Language tags | Strip |
| `w:rFonts` | 7,252 | Font names | Strip (inherit from style) |
| `w:widowControl` | 390 | Widow control | Strip |
| `w:autoSpaceDE/DN` | 780 | Auto spacing | Strip |
| `w:adjustRightInd` | 390 | Right indent adjust | Strip |
| `w:tabs` | 413 | Tab stop definitions | Strip |
| `w:shd` | varies | Shading | Strip |
| `w:color` | varies | Text color | Strip |
| `w:spacing` | varies | Exact spacing | Strip |
| `w:cs` | varies | Complex script | Strip |

### Implementation Plan

#### 1. Add `ParsedDocument` struct to `docx.go`

```go
type ParsedDocument struct {
    PageLayout   PageLayout
    DefaultFont  FontDef
    Paragraphs   []ParsedParagraph
}

type PageLayout struct {
    Width  float64 // inches
    Height float64 // inches
    MarginTop    float64
    MarginRight  float64
    MarginBottom float64
    MarginLeft   float64
}

type FontDef struct {
    Family string
    SizePt float64
}

type ParsedParagraph struct {
    StyleID     string
    Text        string
    Align       string   // "left", "center", "right", "justify"
    IndentLeft  float64  // inches
    Hanging     float64  // inches
    Bold        bool
    Italic      bool
    FontSizePt  float64
    HeadingLevel int     // 0 = not heading, 1-6 = heading level
    IsList      bool
    ListType    string   // "ol" or "ul"
    ListLevel   int
}
```

#### 2. Add `ParseDocument()` function

```go
func ParseDocument(data []byte) (*ParsedDocument, error) {
    // 1. Open DOCX zip
    // 2. Read word/document.xml
    // 3. Read word/styles.xml
    // 4. Build style map from styles.xml
    // 5. Parse document.xml paragraphs
    // 6. For each paragraph, apply style defaults + overrides
    // 7. Return ParsedDocument
}
```

#### 3. Add style map builder

```go
func buildStyleMap(stylesXML string) map[string]StyleDef {
    // Parse styles.xml
    // Extract: styleID → {fontFamily, fontSize, bold, headingLevel, indent, etc}
    // Include docDefaults as base
}
```

#### 4. Add paragraph parser

```go
func parseParagraph(pXML string, styleMap map[string]StyleDef) ParsedParagraph {
    // Extract w:pStyle → lookup in styleMap
    // Extract w:jc → alignment
    // Extract w:ind → indentation
    // Extract w:b → bold
    // Extract w:t → text content
    // Apply style defaults, then override with inline values
}
```

#### 5. Add `FormatForLLM()` method

```go
func (doc *ParsedDocument) FormatForLLM() string {
    // Output structured text for LLM
    // Format:
    //   [PAGE] A4, margins: top=0.71in right=0.24in ...
    //   [FONT] Times New Roman, 12pt
    //   [H1] text content
    //   [P align=justify indent=0.3 hanging=0.3] text content
    //   [B] bold text [/B]
    //   [OL] list items
}
```

#### 6. Update `GenerateWithFile()` in `gemini.go` and `openai.go`

```go
func (p *geminiProvider) GenerateWithFile(ctx, prompt, filename, data) (string, error) {
    doc, err := ParseDocument(data)
    if err != nil {
        return "", err
    }
    cleanedContent := doc.FormatForLLM()
    
    var b bytes.Buffer
    b.WriteString(prompt)
    b.WriteString("\n\n=== SOURCE DOCUMENT ===\n")
    b.WriteString(cleanedContent)
    b.WriteString("\n\n")
    return p.generateContent(ctx, []*genai.Part{{Text: b.String()}})
}
```

### Token Savings Estimate

| Component | Before | After |
|---|---|---|
| Raw XML | ~367K tokens | - |
| Cleaned output | - | ~50-60K tokens |
| **Savings** | - | **83-86%** |

### Files to Modify

| File | Change |
|---|---|
| `docx.go` | Add `ParseDocument()`, `buildStyleMap()`, `parseParagraph()`, `FormatForLLM()` |
| `gemini.go` | Update `GenerateWithFile()` to use `ParseDocument()` |
| `openai.go` | Update `GenerateWithFile()` to use `ParseDocument()` |
| `prompt.go` | Update task instructions to reference cleaned format |

### Risk Assessment

| Risk | Impact | Mitigation |
|---|---|---|
| Style info lost | LLM misinterprets headings | Preserve style map in output |
| Indentation values wrong | DCD indent attributes incorrect | Use twips-to-inches conversion |
| List detection fails | Lists become paragraphs | Parse `w:numPr` carefully |
| Text content altered | DCD output has wrong text | Preserve `<w:t>` exact content |

### Testing Strategy

1. Test `ParseDocument()` with all 3 source DOCX files
2. Compare cleaned output vs raw XML content
3. Verify all text content is preserved
4. Verify heading detection works
5. Verify indentation values are correct
6. Run `dcdmaker run` with cleaned output and compare DCD quality
