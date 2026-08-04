# Source Document Format — words-XML v1.1.0

## What is words-XML?

words-XML is a structured representation of DOCX (OOXML) that has been deliberately simplified and restructured for LLM consumption. It is NOT raw DOCX XML.

### Design Philosophy

**Minimalist transformation from OOXML:**
- Strips verbose XML namespaces (`w:`, `r:`, `wp:`, etc.)
- Removes redundant formatting metadata (proofing errors, language tags, font inheritance chains)
- Collapses deep nesting into flat, readable structures
- Preserves only what matters: content, structure, and essential formatting

**HTML-like but NOT HTML:**
- Uses familiar element names: `<p>`, `<h1>`-`<h9>`, `<ul>`, `<ol>`, `<li>`, `<table>`, `<tr>`, `<td>`, `<th>`, `<b>`, `<i>`, `<u>`, `<s>`, `<a>`
- Attributes use standard XML syntax: `key="value"` pairs (e.g., `<p c="MyStyle" lang="en">`)
- Attribute names are abbreviated but meaningful: `c` (custom style), `at` (borders), `dir` (direction), `lang` (language)
- Self-closing tags use `/>` syntax

**Some attributes are unique to words-XML:**
- `at` — compact border representation (not found in HTML or OOXML)
- `c` — custom style name preservation (NOT CSS class)
- `s:page`, `s:line`, `s:gap`, `s:indent` — layout primitives in `<style>` block

### NOT HTML

Do NOT apply HTML/CSS rules. This is a purpose-built XML format for document representation.

**Critical differences from HTML:**
- `c` attribute is NOT a CSS class — it preserves the original Word style name (e.g., `<h1 c="Heading1Custom">`)
- Style elements use `s:` namespace prefix (e.g., `<s:page>`, `<s:gap>`), not `<style>` HTML tag
- No CSS selectors or cascade — formatting is explicit via attributes
- `<style>` block contains layout primitives (`<s:page>`, `<s:gap>`, etc.), not CSS rules

### NOT OOXML

Do NOT look for XML tags like `<w:p>`, `<w:r>`, `<w:t>`, `<w:pgMar>`, `<w:rPr>`, etc. All OOXML-specific elements have been replaced with simplified equivalents.

## Modes

The format operates in one of two modes (`mode` attribute on root `<words>`):

- **`mode="semantic"`** (default): stripped-down for AI consumption.
  - Whitespace normalized (collapsed spaces, trimmed newlines).
  - Tracked changes (`w:ins`/`w:del`) dropped entirely.
- **`mode="lossless`**: preserves additional metadata for round-tripping.
  - Whitespace NOT normalized (original spacing preserved).
  - Tracked changes emitted as `<ins>`/`<del>` with `id`/`author`/`date` attrs.

## Root Structure

```xml
<words xmlns="urn:words:v1" xmlns:s="urn:words:v1:style" version="1.1.0" mode="semantic">
  <meta>...</meta>          <!-- optional, before <style> -->
  <style>...</style>        <!-- required, before <write> -->
  <header id="n">...</header>  <!-- optional, per section -->
  <footer id="n">...</footer>  <!-- optional, per section -->
  <write>...</write>         <!-- required, main body -->
  <notes>...</notes>         <!-- optional, after </write> -->
</words>
```

## `<meta>` — Document Metadata (optional)

Contains properties from `docProps/core.xml`. Only non-empty fields emitted.

```xml
<meta>
  <title>Sample Document</title>
  <subject>Legal Document</subject>
  <author>John Doe</author>
  <created>2025-01-15T10:30:00Z</created>
  <modified>2025-06-20T14:22:00Z</modified>
  <keywords>contract, legal</keywords>
  <language>en-US</language>
</meta>
```

## `<style>` — Layout Block (required)

The `<style>` block is XML (not INI, not HTML). It uses the `s:` namespace prefix for style primitives.

**Namespace clarification:**
- All elements inside `<style>` use `s:` prefix: `<s:page>`, `<s:gap>`, `<s:line>`, `<s:indent>`, `<s:align>`, `<s:cols>`, `<s:col>`, `<s:tab>`, `<s:theme>`, `<s:custom>`
- The `<style>` element itself uses NO prefix
- All elements outside `<style>` (in `<write>`, `<header>`, `<footer>`, `<notes>`) use NO namespace prefix

Minimum required `<style>` block:

```xml
<style unit="in">
  <s:page size="A4" mt="0.75" mb="0.75" ml="0.75" mr="0.75" mh="0.5" mf="0.5"/>
</style>
```

### Unit Declaration

**`unit` attribute on `<style>`** declares the default unit for all numeric layout values.

**Rules:**
- Allowed values: `in` (inch, **recommended default**), `pt` (point), `px` (pixel), `cm`, `mm`
- A bare number is interpreted in the declared unit (e.g., `ml="54"` with `unit="in"` means 54 inches)
- All OOXML twips-based physical lengths are converted to the declared unit before emitting

**Point-based values are NEVER converted:**
- Font sizes (`size`, `sizeCS`) are ALWAYS in `pt` with explicit `pt` suffix
- Example: `size="11pt"` — NOT converted to inches even when `unit="in"`
- These are inherently point-based (typography standard) and stay in `pt`

**Line spacing exception:**
- `lineSpacing` with `lineRule="auto"` is a dimensionless multiplier, NOT a physical length
- Example: `value="1.5"` means 1.5× line height — no unit conversion applies

### Style Primitives

| Primitive | Purpose | Key Attributes |
|-----------|---------|----------------|
| `<s:page>` | Page geometry | `size`, `w`, `h`, `mt`, `mb`, `ml`, `mr`, `mh`, `mf` |
| `<s:gap>` | Paragraph/heading spacing | `el`, `c`, `before`, `after` |
| `<s:line>` | Line spacing | `el`, `value`, `rule` |
| `<s:indent>` | Paragraph indentation | `el`, `left`, `right`, `firstLine`, `hanging` |
| `<s:align>` | Paragraph alignment | `el`, `value` |
| `<s:cols>` | Multi-column layout | `n`, `space` |
| `<s:col>` | Column/grid widths | `ref`, `width`, `unit` |
| `<s:tab>` | Tab stop definition | `el`, `pos`, `align`, `leader` |
| `<s:theme>` | Global font/color defaults | `font`, `fontEA`, `fontCS`, `fg`, `bg` |
| `<s:custom>` | Custom style definition | `name`, `basedOn`, `type`, formatting attrs |

### `<s:page>` — Page Geometry

`size` is a named preset (A3/A4/A5/A6/B5/Letter/Legal/Tabloid/Executive/Statement/Folio).
If `w`/`h` are also given, they override the preset. For custom sizes, emit explicit
`w="X.XX" h="X.XX"` with no `size` attribute.

Attributes: `size`, `w`, `h`, `mt`, `mb`, `ml`, `mr`, `mh`, `mf` (all in declared unit).

Multiple sections: `<s:page>` MAY appear more than once, once per document section.

### `<s:gap>` — Spacing Rules

```xml
<s:gap el="h" c="Heading1" before="0.25" after="0.17"/>
<s:gap el="p" before="0" after="0.11"/>
```

`el` = target element (`h`, `p`). `c` = optional style name. `before`/`after` in declared unit.

### `<s:line>` — Line Spacing

```xml
<s:line el="p" value="1.5" rule="auto"/>
```

`value` = multiplier. `rule` = `auto`|`exact`|`atLeast`.

### `<s:indent>` — Paragraph Indentation

```xml
<s:indent el="p" left="0.5" right="0" firstLine="0.25" hanging="0"/>
```

`left`/`right`/`firstLine`/`hanging` in declared unit. `firstLine` and `hanging` are
mutually exclusive (both positive; Word stores one as positive, other as negative/zero).

### `<s:align>` — Paragraph Alignment

```xml
<s:align el="p" value="both"/>
```

`value` = `left`|`center`|`right`|`both`.

### `<s:cols>` — Multi-Column Layout

```xml
<s:cols n="2" space="0.25"/>
```

`n` = number of columns. `space` = gutter space between columns (in declared unit).

### `<s:col>` — Column/Grid Widths

```xml
<s:col ref="1" width="2.50" unit="pt"/>
```

Links to `<table id="n">` via `ref` attribute (1-based document order).

### `<s:tab>` — Tab Stop Definition

```xml
<s:tab el="p" pos="1.0" align="left" leader="none"/>
```

`el` = target element, `pos` = position (in declared unit), `align` = alignment,
`leader` = leader character style.

### `<s:theme>` — Global Defaults (Font + Color Tokens)

```xml
<s:theme font="Calibri" fontEA="SimSun" fontCS="Courier New" fg="000000" bg="FFFFFF"/>
```

Optional global defaults resolved from `w:docDefaults` + theme fontScheme in `theme/theme1.xml`.
`font` = Latin/ASCII font, `fontEA` = East Asian font, `fontCS` = Complex Script font.
`fg` = dk1 (text color), `bg` = lt1 (background color).
Priority: inline run > style definition > global default.

### `<s:custom>` — Custom Style Definition

```xml
<s:custom name="MyHeading" basedOn="Heading1" type="paragraph"
  font="Arial" size="24" color="2B5797" bold="true"
  alignment="left" spacingBefore="18" spacingAfter="12"/>
```

Only emitted for non-standard custom styles. Formatting attributes: font, fontEA,
fontCS, size, sizeCS, color, bold, italic, underline, strikethrough, smallCaps,
uppercase, alignment, spacingBefore, spacingAfter, lineSpacing, lineRule,
indentLeft, indentRight, indentFirst, indentHanging, borderWidth, borderColor,
borderStyle, cellSpacing, width.

## `<header>` / `<footer>` — Header & Footer Content

Per section, after `<style>` and before `<write>`:

```xml
<header id="1" type="default">
  <p>Header text</p>
</header>
<footer id="2" type="first">
  <p>Footer text</p>
</footer>
```

`id` = 1-based sequential number. `type` = omitted for `default`, otherwise `first`
or `even`. Content uses same block elements as `<write>`.

## `<write>` — Document Content

### Block Elements

| Tag | Description | Attributes |
|-----|-------------|------------|
| `<h1>`..`<h9>` | Heading levels 1-9 | `c`, `at`, `dir`, `lang`, font/size/color overrides |
| `<p>` | Paragraph | `c`, `at`, `dir`, `lang`, `valign`, font/size/color overrides |
| `<pre>` | Code/monospace block (whitespace preserved verbatim) | `c`, `lang` |
| `<blockquote>` | Quoted block | `c`, `lang` |
| `<ul>` | Unordered list | `type` |
| `<ol>` | Ordered list | `type`, `start` |
| `<li>` | List item | (inline children, may nest `<ul>`/`<ol>`) |
| `<table>` | Table | `id`, `cols`, `c`, `at`, `width`, `align`, `caption`, `summary`, `indent`, `cellSpacing` |
| `<tr>` | Table row | — |
| `<th>` | Table header cell | `colspan`, `rowspan`, `valign`, `textDir`, `noWrap`, `at` |
| `<td>` | Table data cell | `colspan`, `rowspan`, `valign`, `textDir`, `noWrap`, `at` |
| `<section-break/>` | Section boundary | `type`, `layout`, `columns` |
| `<img/>` | Image placeholder | `alt`, `src`, `width`, `height` |
| `<fn-ref/>` | Footnote/endnote marker | `id`, `type` (`footnote`/`endnote`) |

### Inline Formatting Tags

| Tag | Description |
|-----|-------------|
| `<b>text</b>` | Bold |
| `<i>text</i>` | Italic |
| `<u>text</u>` | Underline (optional `underline` attr for type: `single`, `double`, etc.) |
| `<s>text</s>` | Strikethrough (optional `type="double"` for double strikethrough) |
| `<smallcaps>text</smallcaps>` | Small caps |
| `<uppercase>text</uppercase>` | All caps |
| `<sub>text</sub>` | Subscript |
| `<sup>text</sup>` | Superscript |
| `<bcs>text</bcs>` | Complex Script bold |
| `<ics>text</ics>` | Complex Script italic |
| `<span font="Arial" size="12" color="FF0000" highlight="yellow" lang="en" hidden="true" fontEA="..." fontCS="..." sizeCS="...">text</span>` | Run formatting overrides |
| `<a href="url">text</a>` | Hyperlink |
| `<br/>` | Line break (optional `type` attr: `textWrapping`, `page`, `column`, `clear`) |
| `<tab/>` | Tab character |
| `<ins>` / `<del>` (lossless mode only) | Tracked changes with `id`/`author`/`date` attrs |

### Attributes

- **`c`** — preserves original custom style name (e.g., `<h1 c="MyCustomHeading">`).
  Not emitted for standard styles (Heading1-9, Normal, etc.).
  **NOT a CSS class** — do not apply CSS class logic. This is the Word style name.
- **`at`** — compact border representation (unique to words-XML).
  Format: `at="[side] [width] [style][space] [color]; ..."`
  Side: `bt`(top), `bb`(bottom), `bl`(left), `br`(right)
  Style: `s`(single), `d`(double), `ds`(dashed), `dt`(dotted), `n`(none)
  Example: `<p at="bb 12 s1 #000000"/>`
  **NOT CSS border syntax** — this is a compact, space-separated format specific to words-XML.
- **`dir`** — text direction: `rtl` or `ltr`.
- **`lang`** — language tag (BCP 47) on block elements and `<span>`.
- **`valign`** — vertical alignment: `top`, `center`, `baseline` on `<p>`, `<td>`, `<th>`.
- **`textDir`** — text direction in table cells.
- **`noWrap`** — no-wrap flag on table cells (`true`).

## `<notes>` — Notes Container (optional)

After `</write>`, before `</words>`:

```xml
<notes>
  <fn id="1" type="footnote">Footnote body text here.</fn>
  <fn id="2" type="endnote">Endnote body text here.</fn>
  <bm id="bookmark_name"/>
  <comment id="1" author="John Doe" date="2025-01-15T10:30:00Z">Comment text.</comment>
</notes>
```

## Transformation Rules Summary

- Paragraph → `<h1>`-`<h9>` / `<p>` / `<li>` / `<pre>` / `<blockquote>` based on style.
- Runs → inline tags for formatting (bold, italic, underline, strikethrough, etc.).
- Lists grouped by `numId` + `ilvl` + `abstractNumId` + restart state.
- Tables with `colspan`/`rowspan`, header rows as `<th>`.
- Textbox content unwrapped into `<write>` as sibling elements.
- Page size/margins in `<s:page>`; section breaks as `<section-break/>`.
- Headers/footers in `<header>`/`<footer>` blocks.
- Footnotes/endnotes: `<fn-ref/>` marker in `<write>`, `<fn>` body in `<notes>`.
- Bookmarks in `<notes>` as `<bm>`, comments as `<comment>`.
- Images: `<img>` placeholder with src/width/height/alt.
- `mode="semantic"`: whitespace normalized, tracked changes dropped.
- `mode="lossless"`: whitespace preserved, tracked changes as `<ins>`/`<del>`.
- Custom styles → `<s:custom>` in `<style>` + `c` attribute on element.
- `xml:space="preserve"` on `<w:t>` honored (whitespace not collapsed).
- All text and attributes XML-escaped ( `&` → `&amp;`, `<` → `&lt;`, etc.).
- Forbidden XML 1.0 control characters (0x00–0x08, 0x0B–0x0C, 0x0E–0x1F, 0x7F–0x84) stripped.

## Example

```xml
<words xmlns="urn:words:v1" xmlns:s="urn:words:v1:style" version="1.1.0" mode="semantic">
  <meta>
    <title>Sample Document</title>
    <author>John Smith</author>
  </meta>
  <style unit="in">
    <s:page size="A4" mt="0.79" mb="0.79" ml="0.79" mr="0.79" mh="0.50" mf="0.50"/>
    <s:line el="p" value="1.5" rule="auto"/>
    <s:gap el="h" c="Heading1" before="0.25" after="0.17"/>
  </style>
  <write>
    <h1>ARTICLES OF INCORPORATION</h1>
    <p>1. .... .........., born in ....</p>
    <ul type="bullet">
      <li>The company is established for an indefinite period.</li>
    </ul>
  </write>
</words>
```

**Key Point:** Extract information from these XML structures, NOT from raw DOCX XML.
The `words` format is the source document. Use its elements and attributes directly.
