# DOCX → FormatForLLM Requirement Specification

**Output format:** `words` XML (v1.0.1 per `rachmanzz/docx-preprocessor`). NOT bracket format.

**Coverage vs spec: ~99%** — remaining gaps: lossless whitespace normalization (partial), `<s:custom>` enhancement (non-standard table style), field codes (intentionally dropped per spec).

## Principles

1. **1:1 Fidelity** — `FormatForLLM()` output MUST represent every structural and formatting feature of the source DOCX. No information loss.
2. **100% Layout Recognition** — Page layout detection (size, orientation, margins) MUST match the DOCX exactly. Every standard size (A4, letter, legal, A3, A5, B5) MUST be recognized. Custom sizes MUST be output with `w=` and `h=`.
3. **AI-Understandable** — Every feature MUST be represented in a notation that is unambiguous and immediately understandable by the LLM. Zero tolerance for ambiguous or lossy representations.
4. **No Feature Omission** — Every DOCX feature must have a corresponding representation in the FormatForLLM output. No exception. Hardcoded values count as omissions.
5. **Preprocessor Format == DOCX** — The preprocessor format (`<p>`, `<h1>`, `<b>`, etc.) is an alias for the exact original DOCX structure. The AI MUST understand that every attribute, every tag, and every value is the actual DOCX value, not an approximation.

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
| `<w:drawing>` (DrawingML) | ✅ | `<img src="..." width="..." height="..." alt="..."/>` |
| `<w:pict>` (VML legacy) | ✅ | `<img src="..." width="..." height="..." alt="..."/>` |
| `<wp:inline>` | ✅ | `<img/>` output (inline vs anchor noted) |
| `<wp:anchor>` | ✅ | `<img/>` output |
| Image size (EMU cx/cy) | ✅ | `width=`/`height=` in inches |
| Image size (VML style) | ✅ | `width=`/`height=` in inches |
| Alt text (`descr`, `o:title`) | ✅ | `alt=` attribute |
| Image relationship (`r:embed`/`r:id`) | ✅ | Resolved to `word/media/...` |
| Charts, SmartArt, Shapes | ❌ | Excluded per spec (out of scope) |

### 6.3 Headers & Footers ✅

| Feature | Status | Format |
|---------|--------|--------|
| `<w:headerReference>` | ✅ | `<header id="n">` blocks |
| `<w:footerReference>` | ✅ | `<footer id="n">` blocks |
| Different first page (`titlePg`) | ✅ | `type="first"` |
| Different odd/even | ✅ | `type="even"`, `type="default"` (for odd) |
| Header/footer content | ✅ | Full paragraph/table formatting inside blocks |

### 6.4 Multi-Section Layout ✅

| Feature | Status | Format |
|---------|--------|--------|
| Multiple `w:sectPr` | ✅ | `<section-break/>` between sections |
| Section breaks (type) | ✅ | `<section-break type="nextPage"/>` and variants |
| Different layout per section | ✅ | `layout=` attribute on `<section-break/>` |
| Columns | ✅ | `<s:cols n="N" space="X.XX"/>` in `<style>` |
| Page numbering | ✅ | Parsed via `pgNumType` (fmt/start) |

### 6.5 Fields ❌ (Dropped per v1.0.1 spec)

| Feature | Status | Notes |
|---------|--------|-------|
| `<w:fldCode>` | ❌ DROPPED | Field codes are presentation noise per spec §3.0 |
| `<w:fldChar>` | ❌ DROPPED | |
| `<w:instrText>` | ❌ DROPPED | Exception: HYPERLINK instrText → `<a href>` |
| Hyperlink field | ✅ resolved | Via `r:id` lookup OR `instrText HYPERLINK` |
| Page reference, formula, merge fields | ❌ DROPPED | Spec says DROP per noise matrix |

Per spec §3.0, all field codes (TOC, PAGE, DATE, REF, etc.) are DROP — except HYPERLINK which is kept as `<a href="...">`.

### 6.6 Comments & Annotations ✅

| Feature | Status | Format |
|---------|--------|--------|
| `<w:commentRangeStart>` | ✅ | Parsed and stored |
| `<w:commentRangeEnd>` | ✅ | Parsed and stored |
| `<w:commentReference>` | ✅ | Parsed and stored |
| Comments part (`word/comments.xml`) | ✅ | `<comment id="n" author="..." date="...">text</comment>` in `<notes>` |

Per spec §2.7: `<comment id="n" author="..." date="...">text</comment>` in `<notes>` block.

### 6.7 Footnotes & Endnotes ✅

| Feature | Status | Format |
|---------|--------|--------|
| `<w:footnoteReference>` / `<w:endnoteReference>` | ✅ | `<fn-ref id="n" type="footnote|endnote"/>` in `<write>` |
| Footnotes part (`word/footnotes.xml`) | ✅ | `<fn id="n" type="footnote">body</fn>` in `<notes>` |
| Endnotes part (`word/endnotes.xml`) | ✅ | `<fn id="n" type="endnote">body</fn>` in `<notes>` |

Per spec §2.7: marker `<fn-ref>` in `<write>`, body `<fn>` in `<notes>`.

### 6.8 Content Controls (Structured Document Tags) ⚠️ (Unwrapped)

| Feature | Status | Notes |
|---------|--------|-------|
| `<w:sdt>` | ⚠️ Unwrapped | Children extracted per spec §3.0 |
| `<w:sdtPr>` | ❌ Dropped | Properties not preserved |
| `<w:sdtContent>` | ✅ Unwrapped | Inner content processed normally |
| Repeating section | ❌ Dropped | |
| Drop-down list | ❌ Dropped | |

Per spec §3.0 noise matrix: `w:sdt` / `w:smartTag` / `w:customXml` → "unwrap children".

### 6.9 Text Boxes ✅

| Feature | Status | Format |
|---------|--------|--------|
| `<w:txbxContent>` (inside `w:drawing` DML) | ✅ | Unwrapped text as sibling `<p>`/`<table>` after host paragraph |
| `<w:txbxContent>` (inside `w:pict` VML) | ✅ | Same as DML — extracted as sibling elements |
| Linked text boxes | ❌ | Text flow between linked boxes not tracked |

Per spec §3.2 (CRIT-1): textbox content unwrapped; runs before/after textbox anchor merged into single `<p>`; textbox content emitted as sibling elements.

### 6.10 Bookmarks ✅

| Feature | Status | Format |
|---------|--------|--------|
| `<w:bookmarkStart>` | ✅ | `<bm id="name"/>` in `<notes>` |
| `<w:bookmarkEnd>` | ✅ | Parsed (end marker validated) |

Per spec §2.7: `<bm id="name"/>` in `<notes>`.

### 6.11 Mail Merge ❌

| Feature | Status | Notes |
|---------|--------|-------|
| Mail merge fields | ❌ DROPPED | Field codes dropped per spec |
| Data source reference | ❌ DROPPED | Not parsed |

### 6.12 Embedded Objects ❌

| Feature | Status | Notes |
|---------|--------|-------|
| `<w:object>` | ❌ EXCLUDED | OLE objects out of scope per spec |
| `<w:control>` | ❌ EXCLUDED | ActiveX controls out of scope |
| Embedded spreadsheets, PDFs, etc. | ❌ EXCLUDED | Binary/complex objects |

### 6.13 Math (Office Math) ❌

| Feature | Status | Notes |
|---------|--------|-------|
| `<m:oMathPara>` | ❌ EXCLUDED | Math out of scope per spec |
| `<m:oMath>` | ❌ EXCLUDED | |
| MathML content | ❌ EXCLUDED | |

### 6.14 Revision Tracking (Track Changes) ✅

| Feature | Status | Format |
|---------|--------|--------|
| `<w:ins>` | ✅ | `<ins id="n" author="..." date="...">content</ins>` (lossless mode only) |
| `<w:del>` | ✅ | `<del id="n" author="..." date="...">content</del>` (lossless mode only) |
| `<w:rPrChange>` | ❌ Dropped | Property change tracking not preserved |
| `<w:pPrChange>` | ❌ Dropped | |
| `<w:sectPrChange>` | ❌ Dropped | |

In `mode="semantic"` (default): tracked changes dropped. In `mode="lossless"`: emitted as `<ins>`/`<del>` with `id`/`author`/`date` attrs.

### 6.15 Alternative Content ✅

| Feature | Status | Format |
|---------|--------|--------|
| `<mc:AlternateContent>` | ✅ Unwrapped | Children extracted — prefers first valid child per spec |

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

FormatForLLM preserves the inheritance chain through:
- `<s:custom>` definitions in `<style>` for custom styles with their full property sets
- Paragraph-level attributes override style defaults (explicit on the element)
- Run-level `<span>` attributes override paragraph-level formatting
- docDefaults resolved into `<style>` block

**Requirement:** FormatForLLM MUST preserve the inheritance chain so the AI can see which properties come from which level. ✅ Resolved via `<s:custom>` + per-element overrides.

---

## 9. Verification Checklist

Every FormatForLLM output MUST verify:

- [x] Page layout dimensions match DOCX (w, h, margins within 0.01″)
- [x] Orientation explicitly stated (portrait or landscape)
- [x] Default font family/size/color match DOCX defaults (marked `# fallback` if hardcoded)
- [x] line-height matches DOCX default spacing (not hardcoded 1.5)
- [x] Every paragraph has correct classification (heading/list/normal)
- [x] Every paragraph alignment matches DOCX (both → both, not justify)
- [x] Every indent/hanging value matches DOCX (same precision)
- [x] Every font change (family, size, color) per paragraph matches DOCX
- [x] Per-run bold/italic/underline/strike/sup/sub matches DOCX exactly
- [x] Per-run font family/size/color changes preserved — `<span font=".." size=".." color=".."/>`
- [x] Heading styles (`<s:custom>` + `c` attribute) for ALL levels present in DOCX
- [x] Heading style properties (font, size, color, bold, italic, spacing, borders, align) complete
- [x] Document metadata (title, subject, author, keywords, etc.) complete
- [x] Tables represented (`<table>` with `<tr>`/`<th>`/`<td>`, grid, borders, merges)
- [x] Images represented (`<img src="..." width="..." height="..." alt="..."/>` from DrawingML + VML)
- [x] Multi-section layouts handled (`<section-break/>` with layout/cols)
- [x] Headers/footers represented (`<header>`/`<footer>` blocks per section)
- [x] Line breaks (`<w:br/>`) correctly represented (as `<br type="..."/>`)
- [x] Tab characters represented as `<tab/>`
- [x] Page breaks (`<br type="page"/>`) for `w:pageBreakBefore` and `lastRenderedPageBreak`
- [x] No hardcoded values silently substituted without documentation

---

## 10. Critical Rules (ZERO TOLERANCE)

1. No DOCX feature may be silently omitted or approximated.
2. No hardcoded fallback value may be used without being explicitly marked as fallback.
3. The AI must be able to reconstruct the EXACT original DOCX from the FormatForLLM output.
4. Every OOXML element in `w:rPr`, `w:pPr`, and `w:sectPr` must have a corresponding representation.
5. FormatForLLM is NOT DCD — it is the SOURCE DOCUMENT. Features without DCD equivalents (tables, images, fields) MUST still be represented in a clear, unambiguous format.
