# DOCX Preprocessor

Converts raw DOCX XML into `words` XML for LLM consumption.

## Source Document Format

The source document is `words` XML — a structured XML format NOT raw DOCX XML.

Do NOT look for XML tags like `<w:p>`, `<w:r>`, `<w:t>`, `<w:pgMar>`, etc.

### Root Structure

```xml
<words xmlns="urn:words:v1" xmlns:s="urn:words:v1:style" version="1.0.1" mode="semantic">
  <meta>...</meta>
  <style>...</style>
  <write>...</write>
</words>
```

### Document Metadata (`<meta>`)

Contains document properties extracted from the DOCX:

```xml
<meta>
  <title>Sample Document</title>
  <subject>Legal Document</subject>
  <author>John Doe</author>
  <language>en-US</language>
</meta>
```

Only non-empty fields are emitted. Use this to populate the `[title]` block in the DCD output.

### Document Style (`<style>`)

Contains page layout, margins, line spacing, and heading gaps:

```xml
<style unit="in">
  <s:page size="A4" mt="0.79" mb="0.79" ml="0.79" mr="0.79" mh="0.50" mf="0.50"/>
  <s:line el="p" value="1.5" rule="auto"/>
  <s:gap el="h" c="Heading1" before="0.25" after="0.17"/>
</style>
```

Use this as the reference for page layout defaults. Body elements override when specified.

### Document Content (`<write>`)

Contains the parsed body text as XML:

| Tag | Description | Example |
|---|---|---|
| `<p>` | Paragraph | `<p>text</p>` |
| `<p dir="rtl">` | Right-to-left paragraph | `<p dir="rtl">نص</p>` |
| `<h1>`-`<h9>` | Heading levels (1-9) | `<h1>Chapter Title</h1>` |
| `<ul>` / `<ol>` | Unordered/ordered list container | `<ul><li>item</li></ul>` |
| `<li>` | List item (inside `<ul>`/`<ol>`) | `<li>text</li>` |
| `<table>` | Table with `cols` attribute | `<table cols="3">` |
| `<tr>` | Table row | `<tr><td>cell</td></tr>` |
| `<td>` | Table cell with optional `cols` for gridSpan | `<td cols="2">merged</td>` |
| `<s:col>` | Column width definition (inside `<table>`) | `<s:col width="2.50" unit="pt"/>` |
| `<header id="N" type="default">` | Header block | `<header id="1"><p>Header</p></header>` |
| `<footer id="N" type="first">` | Footer block | `<footer id="2"><p>Footer</p></footer>` |
| `<section-break/>` | Section boundary with optional attributes | `<section-break type="nextPage" layout="landscape"/>` |

### Inline Formatting Tags

| Tag | Description | Example |
|---|---|---|
| `<b>text</b>` | Bold | `<b>important</b>` |
| `<i>text</i>` | Italic | `<i>note</i>` |
| `<u>text</u>` / `<u underline="type">text</u>` | Underline | `<u>underline</u>` |
| `<s>text</s>` | Strikethrough | `<s>deleted</s>` |
| `<s type="double">text</s>` | Double strikethrough | `<s type="double">void</s>` |
| `<sup>text</sup>` | Superscript | m<sup>2</sup> |
| `<sub>text</sub>` | Subscript | H<sub>2</sub>O |
| `<span font="Arial" size="12" color="FF0000" highlight="yellow">` | Run formatting overrides | `<span font="Courier New" size="8">small code</span>` |
| `<smallcaps>text</smallcaps>` | Small caps | `<smallcaps>NOTE</smallcaps>` |
| `<uppercase>text</uppercase>` | All caps | `<uppercase>WARNING</uppercase>` |
| `<a href="url">text</a>` | Hyperlink | `<a href="http://example.com">click</a>` |
| `<br/>` | Line break (standalone) | `text<br/>more text` |
| `<tab/>` | Tab character | `<tab/>` |
| `<br type="page"/>` | Page break (standalone) | `<br type="page"/>` |
| `<img alt="description"/>` | Image placeholder (no src/width/height) | `<img alt="Logo"/>` |

### Example

```xml
<words xmlns="urn:words:v1" xmlns:s="urn:words:v1:style" version="1.0.1" mode="semantic">
  <meta>
    <title>Akta Pendirian</title>
    <author>Notaris John</author>
  </meta>
  <style unit="in">
    <s:page size="A4" mt="0.79" mb="0.79" ml="0.79" mr="0.79" mh="0.50" mf="0.50"/>
    <s:line el="p" value="1.5" rule="auto"/>
  </style>
  <write>
    <p>AKTA PENDIRIAN PERSEROAN TERBATAS</p>
    <p>1. .... .........., lahir di ....</p>
    <ul>
      <li>Perseroan didirikan untuk jangka waktu tidak terbatas.</li>
    </ul>
  </write>
</words>
```

**Key Point:** Extract information from these XML structures, NOT from raw DOCX XML or bracket format.
