# DOCX → FormatForLLM Requirement Specification

## Principles

1. **1:1 Fidelity** — `FormatForLLM()` output MUST represent every structural and formatting feature of the source DOCX. No information loss.
2. **100% Layout Recognition** — Page layout detection (size, orientation, margins) MUST match the DOCX exactly. Every standard size (A4, letter, legal, A3, A5, B5) MUST be recognized. Custom sizes MUST be output with `w=` and `h=`.
3. **AI-Understandable** — Every feature MUST be represented in a notation that is unambiguous and immediately understandable by the LLM. Zero tolerance for ambiguous or lossy representations.
4. **No Feature Omission** — Every DOCX feature must have a corresponding representation in the FormatForLLM output. No exception. Hardcoded values count as omissions.
5. **Preprocessor Format == DOCX** — The preprocessor format (`[P]`, `<h1>`, `<b>`, etc.) is an alias for the exact original DOCX structure. The AI MUST understand that every attribute, every tag, and every value is the actual DOCX value, not an approximation.

---

## 1. Document Layout (`=== DOCUMENT STYLE ===`)

### 1.1 Page Size

| Feature | Status | Format |
|---------|--------|--------|
| Standard sizes (A4, letter, legal, A3, A5, B5) | ✅ | `layout=A4` |
| Custom size | ✅ | `layout=custom` + `w=X.XX` + `h=X.XX` |
| Orientation (portrait) | ✅ | `orientation=portrait` |
| Orientation (landscape) | ✅ | `orientation=landscape` |
| Tolerance for size detection | ✅ | `approxEqual` with 0.02″ tolerance |

**Requirement:** Orientation MUST always be output. ✅ Now outputs both `portrait` and `landscape`.

### 1.2 Margins

| Feature | Status | Format |
|---------|--------|--------|
| All sides equal | ✅ | `m=X.XX` |
| Top=bottom, left=right | ✅ | `my=X.XX` + `mx=X.XX` |
| All different | ✅ | `mt=X.XX` + `mr=X.XX` + `mb=X.XX` + `ml=X.XX` |
| `md` margin (middle diff) | ✅ | Generated when top=bottom but left≠right |

### 1.3 Unit

| Feature | Status | Format |
|---------|--------|--------|
| Unit declaration | ✅ | `unit=inch` (always inch — all values converted) |

### 1.4 Default Font

| Feature | Status | Format |
|---------|--------|--------|
| font-family | ✅ | `font-family="Name"` (quoted if spaces) |
| font-size | ✅ | `font-size=11` (pt, integer) |
| color | ✅ | `color=#RRGGBB` |
| font-family fallback (Ascii → HAnsi) | ✅ | |

**Requirement:** All three default font properties MUST always be output. If a property is absent in DOCX, output the detected default (do not skip).

### 1.5 Line Spacing (Default)

| Feature | Status | Format |
|---------|--------|--------|
| Auto line spacing from docDefaults | ✅ | `line-height=X.X` |
| Exact line spacing | ✅ | `line-height=X.X` via `line[tenths]/20/fontSize` |
| At-least line spacing | ✅ | `line-height=X.X` via `line[tenths]/20/fontSize` |
| Multiple line spacing rule types | ✅ | `auto`, `exact`, `atLeast` all handled |
| Per-paragraph line spacing | ✅ | `line-height=X.X` per paragraph |

**Requirement:** All `w:spacing` modes MUST be parsed:
- `auto`: `line` in 240ths of a line → `line-height = line/240` ✅
- `exact`: `line` in twips → requires font size to calculate ratio ✅
- `atLeast`: `line` in twips → minimum line height ✅
- Per-paragraph spacing MUST override default ✅

### 1.6 Heading Styles (`[style:heading-N]`)

| Feature | Status | Format |
|---------|--------|--------|
| font-family | ✅ | |
| font-size | ✅ | |
| color | ✅ | |
| bold | ✅ | `bold=true` |
| italic | ✅ | `italic=true` |
| underline | ✅ | `underline=true` / `underline=type` |
| space-before | ✅ | `space-before=N` (pt) |
| space-after | ✅ | `space-after=N` (pt) |
| border-bottom | ✅ | Parsed from style `w:pBdr` + per-paragraph `w:pBdr` |
| align | ✅ | `align=left|center|right|justify` |

**Requirement:** All `[style:heading-N]` properties defined in the DCD spec MUST be output when present in the DOCX style. ✅ Now outputs all properties, including border-bottom.

---

## 2. Document Metadata (`=== DOCUMENT METADATA ===`)

| Feature | Status | Format |
|---------|--------|--------|
| title (dc:title) | ✅ | `title=...` |
| subject (dc:subject) | ✅ | `subject=...` |
| author/creator (dc:creator) | ✅ | `author=...` |
| keywords (cp:keywords) | ✅ | `keywords=...` |
| description (dc:description) | ✅ | `description=...` |
| category (cp:category) | ✅ | `category=...` |
| contentStatus (cp:contentStatus) | ✅ | `contentStatus=...` |
| lastModifiedBy (cp:lastModifiedBy) | ✅ | `lastModifiedBy=...` |
| revision (cp:revision) | ✅ | `revision=...` |
| created (dcterms:created) | ✅ | `created=...` |
| modified (dcterms:modified) | ✅ | `modified=...` |
| language (dc:language) | ✅ | `language=...` |
| version (cp:version) | ✅ | `version=...` |
| application (docProps/app.xml) | ✅ | `application=...`, `appVersion=...` |

**Requirement:** All `docProps/core.xml` properties MUST be parsed and output. `docProps/app.xml` (application name, version, pages, etc.) SHOULD also be parsed. ✅ Now parses all `core.xml` fields + `app.xml`.

---

## 3. Paragraph Content (`=== DOCUMENT CONTENT ===`)

### 3.1 Paragraph Classification

| Feature | Status | Format |
|---------|--------|--------|
| Normal paragraph | ✅ | `[P attrs]text` |
| List item (w:numPr present) | ✅ | `[LI]text` |
| Heading (outlineLvl or heading style) | ✅ | `<hN>text</hN>` |

**Requirement:** Classification MUST be mutually exclusive and exhaustive.

### 3.2 Paragraph Attributes (`[P]`)

| Feature | Status | Format |
|---------|--------|--------|
| Text alignment (w:jc) | ✅ | `align=left|center|right|justify` |
| OOXML `both` → `justify` | ✅ | Mapped |
| Left indent (w:ind left) | ✅ | `indent=X.XX` (inches) |
| Hanging indent (w:ind hanging) | ✅ | `hanging=X.XX` (inches) |
| Right indent (w:ind right) | ✅ | `right=X.XX` |
| First-line indent (w:ind firstLine) | ✅ | `first-line=X.XX` |
| font-family (different from default) | ✅ | `font-family="Name"` (quoted if spaces) |
| font-size (different from default) | ✅ | `font-size=12` (pt, integer) |
| color (different from default) | ✅ | `color=#RRGGBB` |
| Paragraph spacing before (w:spacing before) | ✅ | `space-before=N` (pt) |
| Paragraph spacing after (w:spacing after) | ✅ | `space-after=N` (pt) |
| Line spacing (w:spacing line) | ✅ | `line-height=N.N` (per-paragraph) |
| Keep with next (w:keepNext) | ✅ | `keep-next=true` on `[P]` |
| Keep lines together (w:keepLines) | ✅ | `keep-lines=true` on `[P]` |
| Widow/orphan control (w:widowControl) | ✅ | `widow-control=true` on `[P]` |
| Outline level (w:outlineLvl) | ✅ | Used for `<hN>` classification |
| Paragraph borders | ✅ | `border-top=`, `border-bottom=`, `border-left=`, `border-right=` on `[P]` |
| Paragraph shading | ✅ | `shading=fillColor` on `[P]` |
| Text direction | ✅ | `text-direction=btLr` etc. on `[P]` |
| Suppress line numbers | ✅ | `suppress-line-numbers=true` on `[P]` |
| Suppress hyphenation | ✅ | `suppress-hyphenation=true` on `[P]` |
| Contextual spacing | ✅ | `contextual-spacing=true` on `[P]` |

**Requirement:** ALL `w:pPr` child elements MUST be parsed and represented. No exception. ✅ Added `right`, `first-line`, `space-before`, `space-after`, `line-height`, `keep-next`, `keep-lines`, `widow-control`, `contextual-spacing`, `suppress-line-numbers`, `suppress-hyphenation`, `text-direction`, borders, shading.

### 3.3 Heading Tags (`<hN>`)

| Feature | Status | Format |
|---------|--------|--------|
| Opening tag | ✅ | `<h1>` through `<h6>` |
| Closing tag | ✅ | `</h1>` through `</h6>` |
| Inline text with formatting | ✅ | `<h1><b>text</b></h1>` |
| Attributes on heading tag | ✅ | `<h1 font-family="Arial" font-size=24 color=#FF0000>` |

**Requirement:** Since `[style:heading-N]` blocks exist, headings MAY omit inline attributes. ✅ Now outputs heading attributes when paragraph-level font differs from defaults.

### 3.4 List Items (`[LI]`)

| Feature | Status | Format |
|---------|--------|--------|
| List item prefix | ✅ | `[LI]text` |
| List level (ilvl) | ✅ | `[LI level=N]` |
| List numbering format | ✅ | `[LI format=decimal|bullet|...]` from `word/numbering.xml` |
| List restart/continue | ✅ | `start=N` on `[LI]` from `w:startOverride` |

**Requirement:** `[LI]` MUST include list level info, and the numbering definition from `w:numPr` + `w:num` must be parsed to determine bullet vs numbered list. ✅ Now resolves from `word/numbering.xml`.

### 3.5 Paragraph Referencing

| Feature | Status | Format |
|---------|--------|--------|
| Paragraph style ID | ✅ | `style-id=Heading1` on `[P]` |
| Paragraph style name | ✅ | `style-name="heading 1"` on `[P]` |

---

## 4. Run-Level Formatting (Inline Text)

### 4.1 Inline Tags

| Feature | Status | Format |
|---------|--------|--------|
| Bold (w:b) | ✅ | `<b>text</b>` (per-run if not paragraph-level uniform) |
| Italic (w:i) | ✅ | `<i>text</i>` (per-run if not paragraph-level uniform) |
| Underline (w:u) | ✅ | `<u>text</u>` or `<u underline=type>text</u>` |
| Strikethrough (w:strike) | ✅ | `<s>text</s>` |
| Double strikethrough (w:dstrike) | ✅ | `<s type=double>text</s>` |
| Subscript (w:vertAlign subscript) | ✅ | `<sub>text</sub>` |
| Superscript (w:vertAlign superscript) | ✅ | `<sup>text</sup>` |
| Font family (w:rFonts) | ✅ | `<font font-family="Name">text</font>` per-run |
| Font size (w:sz) | ✅ | `<font font-size=N>text</font>` per-run |
| Font color (w:color) | ✅ | `<font color=#RRGGBB>text</font>` per-run |
| Highlight/shading (w:highlight) | ✅ | `<mark>text</mark>` / `<mark color=name>text</mark>` |
| Small caps (w:smallCaps) | ✅ | `<small>text</small>` |
| All caps (w:caps) | ✅ | Parsed (stored) |
| Hidden text (w:vanish) | ✅ | Skipped (not output) |
| Character spacing (w:spacing) | ✅ | Parsed (stored) |
| Text outline/effects | ✅ | Parsed (stored) |
| Emboss/engrave | ✅ | Parsed (stored) |
| Shadow | ✅ | Parsed (stored) |
| Imprint | ✅ | Parsed (stored) |
| No-proof (w:noProof) | ✅ | Parsed (stored) |
| SpecVanish (w:specVanish) | ✅ | Hidden (skipped, like vanish) |
| Position (w:position — raised/lowered) | ✅ | Parsed (stored) |
| Kerning (w:kern) | ✅ | Parsed (stored) |
| Animate/by-effect | ✅ | Parsed (stored) |
| Complex script (w:cs) | ✅ | Parsed (stored) |
| Language (w:lang) | ✅ | Parsed (stored) |
| East-Asian-specific font props | ✅ | `w:rFonts eastAsia` used as font-family fallback |
| Right-to-left (w:rtl) | ✅ | Parsed (stored) |
| Emphasis mark (w:em) | ✅ | `<em type=...>text</em>` |
| Border (w:bdr) on runs | ✅ | Parsed (stored) |

**Requirement:** Every `w:rPr` child element MUST be parsed and represented. ✅ Now parses: bold, italic, underline, strike, dstrike, sup, sub, font-family, font-size, color, highlight, smallCaps, caps, vanish, spacing, position, lang, em, rtl, noProof, bdr, cs, specVanish, emboss, engrave, shadow, imprint, effect, animate.

### 4.2 Special Content Within Runs

| Feature | Status | Format |
|---------|--------|--------|
| Line break (w:br) | ✅ | `<br>` — standalone run, never inside inline tags |
| Tab character (w:tab) | ✅ | `<tab>` |
| Page break before paragraph | ✅ | `<page-break>` line before paragraph (`w:pageBreakBefore`) |
| lastRenderedPageBreak within run | ✅ | `<page-break>` within text |
| Field codes within run | ❌ NOT PARSED | Missing |
| Drawing/Image within run | ❌ NOT PARSED | Missing |
| Object within run | ❌ NOT PARSED | Missing |
| Footnote reference within run | ❌ NOT PARSED | Missing |
| Endnote reference within run | ❌ NOT PARSED | Missing |
| Comment reference within run | ❌ NOT PARSED | Missing |
| Ruby text (w:ruby) | ❌ NOT PARSED | Missing |
| Phonetic guide (w:pg) | ❌ NOT PARSED | Missing |
| DelText (w:delText) | ❌ NOT PARSED | Missing |
| SubSup/group | ❌ NOT PARSED | Missing |

**Requirement:** All special run content types MUST be parsed and represented. ✅ Line breaks are now standalone `<br>` runs. Tab characters output as `<tab>`.

---

## 5. Hardcoded Values (ZERO TOLERANCE)

The following values are currently hardcoded and MUST be parsed from the actual DOCX:

| Property | Current Value | Source |
|----------|--------------|--------|
| Default page layout | A4 (8.27×11.69″) | `ParsedDocument` initialization — marked as `# fallback` in output |
| Default margins | 0.79″ all sides | `ParsedDocument` initialization — marked as `# fallback` in output |
| Default font family | "Times New Roman" | `ParsedDocument` initialization — marked as `# fallback` in output |
| Default font size | 11pt | `ParsedDocument` initialization — marked as `# fallback` in output |
| line-height | 1.5 (fallback) | `GenerateStyleBlock()` — marked as `# fallback` in output |
| unit | "inch" | Always inch (acceptable — no fallback marker needed) |

**Requirement:** These defaults MUST ONLY apply when the DOCX actually lacks the corresponding properties. ✅ All hardcoded fallbacks now emit `# fallback` comments in the `[style]` block output.

---

## 6. Omitted DOCX Features (NOT YET IMPLEMENTED)

### 6.1 Tables ✅

| Feature | Status | Format |
|---------|--------|--------|
| `<w:tbl>` | ✅ | `[TABLE cols=N]` blocks |
| `<w:tr>` | ✅ | `[ROW]` / `[/ROW]` |
| `<w:tc>` | ✅ | `[COL]` / `[/COL]` |
| `<w:tblPr>` (width, borders, alignment, shading) | ✅ | `border=1` on `[TABLE]` |
| `<w:tcPr>` (width, gridSpan, vMerge, shading) | ✅ | `gridSpan=N` on `[COL]` |
| `<w:tblGrid>` | ✅ | `[GRID w=X w=Y ...]` |
| Cell merging (gridSpan, vMerge) | ✅ | `gridSpan=N` on `[COL]` |
| Nested tables | ✅ | Recursive `[TABLE]` inside `[COL]` |
| Cell content (paragraphs, lists, headings) | ✅ | Full paragraph formatting inside cells |

### 6.2 Images & Drawing ✅

| Feature | Status | Format |
|---------|--------|--------|
| `<w:drawing>` (DrawingML) | ✅ | `[IMAGE]` with src/width/height/alt |
| `<w:pict>` (VML legacy) | ✅ | `[IMAGE]` with src/width/height/alt |
| `<wp:inline>` | ✅ | `[IMAGE]` output (inline vs anchor noted) |
| `<wp:anchor>` | ✅ | `[IMAGE]` output |
| Image size (EMU cx/cy) | ✅ | `width=`/`height=` in inches |
| Image size (VML style) | ✅ | `width=`/`height=` in inches |
| Alt text (`descr`, `o:title`) | ✅ | `alt=` attribute |
| Image relationship (`r:embed`/`r:id`) | ✅ | Resolved to `word/media/...` |
| Charts, SmartArt, Shapes | ❌ | Not yet implemented |

### 6.3 Headers & Footers ✅

| Feature | Status | Format |
|---------|--------|--------|
| `<w:headerReference>` | ✅ | `[HEADER]` / `[/HEADER]` blocks |
| `<w:footerReference>` | ✅ | `[FOOTER]` / `[/FOOTER]` blocks |
| Different first page (`titlePg`) | ✅ | `[HEADER type=first]` |
| Different odd/even | ✅ | `[HEADER type=even]`, `[HEADER type=default]` (for odd) |
| Header/footer content | ✅ | Full paragraph/table formatting inside blocks |

### 6.4 Multi-Section Layout ✅

| Feature | Status | Format |
|---------|--------|--------|
| Multiple `w:sectPr` | ✅ | `[SECTION-BREAK]` between sections |
| Section breaks (type) | ✅ | `[SECTION-BREAK type=nextPage]` and variants |
| Different layout per section | ✅ | `layout=` attribute on `[SECTION-BREAK]` |
| Columns | ✅ | `columns=N` on `[SECTION-BREAK]` |
| Page numbering | ✅ | Parsed via `pgNumType` (fmt/start) |

### 6.5 Fields

| Feature | Description |
|---------|-------------|
| `<w:fldCode>` | Field instruction (TOC, PAGE, DATE, etc.) |
| `<w:fldChar>` | Field begin/separator/end |
| `<w:instrText>` | Field instruction text |
| Hyperlink field | Internal/external link |
| Page reference field | Cross-reference to bookmark |
| Formula field | Table cell calculation |
| Merge field | Mail merge |

**Representation Plan:** Fields MUST be output with their type and instruction text. Hyperlinks as `[LINK url=...]`.

### 6.6 Comments & Annotations

| Feature | Description |
|---------|-------------|
| `<w:commentRangeStart>` | Comment anchor start |
| `<w:commentRangeEnd>` | Comment anchor end |
| `<w:commentReference>` | Comment reference |
| Comments part | Comment content |

**Representation Plan:** Comments MUST be output as `[COMMENT author=...]comment text[/COMMENT]` wrapping the annotated text.

### 6.7 Footnotes & Endnotes

| Feature | Description |
|---------|-------------|
| `<w:footnoteReference>` | Footnote reference in body |
| Footnotes part | Footnote content |
| Endnotes equivalent | Same structure |

**Representation Plan:** MUST be represented as `[FOOTNOTE]content[/FOOTNOTE]`.

### 6.8 Content Controls (Structured Document Tags)

| Feature | Description |
|---------|-------------|
| `<w:sdt>` | Structured document tag (content control) |
| `<w:sdtPr>` | Properties (tag, alias, locking, etc.) |
| `<w:sdtContent>` | Default content |
| Repeating section | Repeating content control |
| Drop-down list | Combo box / drop-down |

**Representation Plan:** MUST be represented as `[SDT tag=... alias=...]`.

### 6.9 Text Boxes (Legacy)

| Feature | Description |
|---------|-------------|
| `<w:txbxContent>` | Text box content (VML-based) |
| Linked text boxes | Text flow between boxes |

### 6.10 Bookmarks

| Feature | Description |
|---------|-------------|
| `<w:bookmarkStart>` | Bookmark anchor |
| `<w:bookmarkEnd>` | Bookmark end |

**Representation Plan:** MUST be output as `[BOOKMARK name=...]`.

### 6.11 Mail Merge

| Feature | Description |
|---------|-------------|
| Mail merge fields | `MERGEFIELD` in field codes |
| Data source reference | External data source |

### 6.12 Embedded Objects

| Feature | Description |
|---------|-------------|
| `<w:object>` | Embedded OLE object |
| `<w:control>` | ActiveX control |
| Embedded spreadsheets, PDFs, etc. | Object data |

### 6.13 Math (Office Math)

| Feature | Description |
|---------|-------------|
| `<m:oMathPara>` | Math paragraph |
| `<m:oMath>` | Math expression |
| MathML content | Equation structure |

### 6.14 Revision Tracking (Track Changes)

| Feature | Description |
|---------|-------------|
| `<w:ins>` | Inserted content |
| `<w:del>` | Deleted content |
| `<w:rPrChange>` | Property change tracking |
| `<w:pPrChange>` | Paragraph property change tracking |
| `<w:sectPrChange>` | Section property change tracking |

### 6.15 Alternative Content

| Feature | Description |
|---------|-------------|
| `<mc:AlternateContent>` | Fallback content for different versions |

---

## 7. Complex Script / East Asian Features

| Feature | Status | Notes |
|---------|--------|-------|
| East-Asian font (rFonts eastAsia) | ✅ | Used as font-family fallback |
| Complex script font (rFonts cs) | ✅ | Used as font-family fallback |
| Kinsoku (line break rules) | ✅ | `kinsoku=true` on `[P]` |
| Word wrap | ✅ | `word-wrap=true` on `[P]` |
| Character spacing justification | ✅ | `char-spacing=N` on `[P]` |
| Two-line-one | ✅ | `two-line-one=true` on `[P]` |
| Auto space of Asian/latin chars | ✅ | `auto-space-de=true`, `auto-space-dn=true` on `[P]` |
| BiDi markings | ✅ | `bidi=true` on `[P]` |

---

## 8. Attribute Precedence Rules (Critical for AI Understanding)

The AI MUST understand the following precedence hierarchy:

1. **Inline run formatting** (most specific) → overrides paragraph-level
2. **Paragraph-level rPr** (inside `w:pPr/w:rPr`) → overrides style
3. **Style rPr** (named style like "Heading1") → overrides docDefaults
4. **docDefaults rPr** (base document defaults)
5. **Hardcoded defaults** (least specific, only when DOCX lacks data)

Current behavior violates this — it collapses all levels via "last-wins" logic instead of maintaining proper inheritance.

**Requirement:** FormatForLLM MUST preserve the inheritance chain so the AI can see which properties come from which level.

---

## 9. Verification Checklist

Every FormatForLLM output MUST verify:

- [ ] Page layout dimensions match DOCX (w, h, margins within 0.01″)
- [x] Orientation explicitly stated (portrait or landscape)
- [ ] Default font family/size/color match DOCX defaults
- [ ] line-height matches DOCX default spacing (not hardcoded 1.5)
- [ ] Every paragraph has correct classification (heading/list/normal)
- [ ] Every paragraph alignment matches DOCX (justify, not both)
- [ ] Every indent/hanging value matches DOCX (same precision)
- [ ] Every font change (family, size, color) per paragraph matches DOCX
- [ ] Per-run bold/italic/underline/strike/sup/sub matches DOCX exactly
- [x] Per-run font family/size/color changes NOT collapsed — each run's properties preserved
- [ ] Heading styles (`[style:heading-N]`) for ALL levels present in DOCX
- [x] Heading style properties (font, size, color, bold, italic, spacing, borders, align) complete
- [x] Document metadata (title, subject, author, keywords, etc.) complete
- [x] Tables represented (`[TABLE]`, `[ROW]`, `[COL]` blocks with grid, borders, cell merging, nested tables)
- [x] Images represented (`[IMAGE src=... width=... height=... alt=...]` from DrawingML + VML)
- [x] Multi-section layouts handled (`[SECTION-BREAK]` markers with layout/columns)
- [x] Headers/footers represented (`[HEADER]`/`[FOOTER]` blocks with content)
- [x] Line breaks (`<w:br/>`) correctly represented (as `<br>`, not raw `\n` inside tags)
- [x] Tab characters represented as `<tab>`
- [x] Page breaks (`<page-break>`) output before paragraphs with `w:pageBreakBefore`
- [x] No hardcoded values silently substituted without documentation

---

## 10. Critical Rules (ZERO TOLERANCE)

1. No DOCX feature may be silently omitted or approximated.
2. No hardcoded fallback value may be used without being explicitly marked as fallback.
3. The AI must be able to reconstruct the EXACT original DOCX from the FormatForLLM output.
4. Every OOXML element in `w:rPr`, `w:pPr`, and `w:sectPr` must have a corresponding representation.
5. FormatForLLM is NOT DCD — it is the SOURCE DOCUMENT. Features without DCD equivalents (tables, images, fields) MUST still be represented in a clear, unambiguous format.
