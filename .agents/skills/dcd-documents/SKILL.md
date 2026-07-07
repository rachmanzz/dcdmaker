# DCD DSL Specification

> This specification is the complete DCD DSL reference. Do not assume syntax features beyond what is explicitly documented here.

## 1. Style Configuration

```
[style]
layout=A4                    # choose one: A4, letter, legal, A3, A5, B5, custom
unit=inch                    # choose one: inch, cm, mm, pt, pica
orientation=portrait         # choose one: portrait, landscape
font-family="Times New Roman"
font-size=12                 # in pt (points)
color=#000000
line-height=1.5
```

### Margins (precedence low→high: m → mx/my → md → mt/mb/ml/mr)

| Code | Effect |
|------|--------|
| `m=1` | Uniform margin all sides |
| `mx=1 my=1` | Horizontal (left/right) + vertical (top/bottom) |
| `mt=1 mb=1 ml=1 mr=1` | Individual sides: top, bottom, left, right |
| `md=1 mb=1` | Default for all sides (md), then override bottom (mb) |

Custom layout requires `w` (width) and `h` (height):

```
[style]
layout=custom
unit=inch
w=8.5
h=11
```

`m`, `mx`, `my`, `md`, `mt`, `mb`, `ml`, `mr`, `w`, `h` — values in configured unit.

---

## 2. Sections & Variables

```
[section 0]
name=userinfo                 # section identifier
var=info, entries             # 1st = object prefix, 2nd+ = loop sources
keys=username, date_field     # field names or dotted paths (source.field)
formats=[date_field:dd-MM-yyyy]  # per-key: [key:format] or [source.field:format]

--- BODY ---
<w:c|b>Center Bold</w:c|b>
<p>{{info.username}} created <i>{{info.date_field}}</i></p>
<loop x from entries>
  {{x.name}} lives at {{x.address}}
</loop>
```

`var` first name → object prefix (`{{info.key}}`). Additional names → loop data sources (`<loop x from entries>`).

`name`, `var`, `keys`, `formats` accept both `=` and `:` separators (e.g. `name=header` or `name:header`).

Any `{{...}}` referencing data fields MUST be in `keys` + `var`. Sections without var/keys pass unresolvable `{{...}}` as literals. Built-in vars (`{{page}}`, `{{total}}`, `{{title}}`, `{{date}}`) auto-resolved.

### Format Specifiers

`dd`, `MM`, `yyyy`, `HH`, `mm`, `ss` — date/time formats.

Array fields use dotted path in keys/formats: `keys=items.date_field`, `formats=[items.date_field:dd-MM-yyyy]`. After loop expansion, `{{x.date_field}}` → `{{items.0.date_field}}` matched by stripping the index.

### Section Splitting Guidelines

Split document into multiple sections by **context/topic**, not by size.

**Rules:**
- Each section represents a logical part: header info, parties, transaction details, signatures, etc.
- Section `name` must describe the logical context (e.g., `name=header_info`, `name=seller`, `name=object`)
- Keep `var` count: **aim for ≤3** per section (split if more needed)
- Keep `keys` count: **aim for ≤15** per section (split if more needed)
- Use `[section:next-page N]` to start a new section on a new page

**Example: Complex Document Split**

```
[section 0]
name=ppat_header
var=ppat_info
keys=nama, kedudukan, sk_nomor, sk_tanggal, alamat, telepon, email

--- BODY ---
<h1>LAND DEED OFFICER</h1>
<p align=center><b>{{ppat_info.nama}}</b></p>

[section:next-page 1]
name=seller_info
var=seller, seller_spouse
keys=seller.name, seller.id, seller.address, seller.birthdate, seller_spouse.name, seller_spouse.id
formats=[seller.birthdate:dd-MM-yyyy], [seller_spouse.birthdate:dd-MM-yyyy]

--- BODY ---
<p>Seller: <b>{{seller.name}}</b>, ID {{seller.id}}</p>
<p>With spouse consent: {{seller_spouse.name}}</p>

[section 2]
name=buyer_info
var=buyer
keys=name, id, address, birthdate
formats=[birthdate:dd-MM-yyyy]

--- BODY ---
<p>Buyer: <b>{{buyer.name}}</b>, ID {{buyer.id}}</p>

[section:next-page 3]
name=transaction_object
var=object
keys=type, area, certificate_no, price, address
formats=[price:#,##0]

--- BODY ---
<p>Object: {{object.type}}, area {{object.area}} m²</p>
<p>Price: Rp {{object.price}}</p>
```

**Benefits:**
- Easier to read and maintain
- Clear separation of concerns
- Predictable variable scope

### Block Tags (Wrapper Paragraphs — Pure Text Only)

Wrapper paragraph tags (`<w:*>`) wrap entire paragraph with single property. CANNOT contain inline tags.

| Tag | Attributes | Description |
|-----|------------|-------------|
| `<w:c>...</w:c>` | `size` / `font-size`, `color` | Center alignment |
| `<w:r>...</w:r>` | `size` / `font-size`, `color` | Right alignment |
| `<w:j>...</w:j>` | `size` / `font-size`, `color` | Justify alignment |
| `<w:l>...</w:l>` | `size` / `font-size`, `color` | Left alignment (default) |
| `<w:b>...</w:b>` | `size` / `font-size`, `color` | Bold (entire paragraph) |
| `<w:i>...</w:i>` | `size` / `font-size`, `color` | Italic (entire paragraph) |
| `<w:u>...</w:u>` | `size` / `font-size`, `color` | Underline (entire paragraph) |
| `<w:c\|b>...</w:c\|b>` | `size` / `font-size`, `color` | Combined (e.g., `<w:r\|b>`, `<w:j\|i>`, `<w:c\|b\|i\|u>`) |

**Pure text + variables only:**

```
<w:c>Centered text {{var}}</w:c>                    <!-- VALID -->
<w:r|b>Right bold {{var}}</w:r|b>                   <!-- VALID -->
<w:c font-size=16>Big centered title</w:c>          <!-- VALID: size attribute -->
<w:r|b size=14 color=#333>Right bold subtitle</w:r|b>  <!-- VALID: size + color -->
<w:c>Text <u>underline</u></w:c>                    <!-- INVALID: no tags inside -->
```

**Rich Paragraph Tag (Can Contain Inline Tags):**

| Tag | Attributes | Description |
|-----|------------|-------------|
| `<p>` | `align`, `size` / `font-size`, `color` | Rich paragraph (default left align) |
| `<p align=center>` | `size` / `font-size`, `color` | Rich paragraph centered |
| `<p align=right>` | `size` / `font-size`, `color` | Rich paragraph right |
| `<p align=justify>` | `size` / `font-size`, `color` | Rich paragraph justified |
| `<br>` | — | Line break |

**Rich paragraphs can contain inline tags:**

```
<p align=center>Centered with <u>underline</u> and <b>bold</b></p>
<p>Normal <set:b|i>bold italic</set:b|i> text</p>
<p align=center font-size=16 color=#2b5797>Centered with color and size</p>
```

**Guideline:** Use `<w:*>` for pure text. Use `<p align=*>` when paragraph has mixed formatting (normal + bold + italic, etc.).

> Note: In markdown tables above, `|` appears as `\|` due to markdown escaping. Actual DCD syntax uses plain `|` without backslash.

> `size` and `font-size` are synonyms — either works, value in pt (points). `color` accepts hex (`#ff0000`) or named colors (`red`). Available on `<w:*>`, `<p>`, `<li>`, `<col>`.

### Inline Tags (inside `<p>`, `<li>`, `<col>`)

| Tag | Description |
|-----|-------------|
| `<b>...</b>` | Bold |
| `<i>...</i>` | Italic |
| `<u>...</u>` | Underline |
| `<code>...</code>` | Monospace |
| `<set:b\|i>...</set>` | Combined inline formatting (`b`, `i`, `u`, `code`) |
| `<tab>` / `<tab size=N>` | Tab character, optional N-space width |

> Note: Actual DCD syntax uses plain `|` without backslash (e.g., `<set:b|i>`).

### Tag Nesting Rules

**Allowed:**
- `<p>` and `<p align=*>` can contain: `<b>`, `<i>`, `<u>`, `<code>`, `<set:>`, `<tab>`, `<a>`, `{{...}}`
- `<li>` can contain: same as `<p>`
- `<col>` can contain: same as `<p>`
- `<loop>` variants can contain: any body tags

**Forbidden:**
- `<w:*>` wrapper tags (e.g., `<w:c>`, `<w:r>`, `<w:b>`) CANNOT contain any tags — pure text + variables only
- `<pb>`, `<page-break>`, `<hr>` MUST be standalone — NOT inside `<p>`, `<li>`, `<col>`, or `<w:*>`

**Examples:**

Valid:
```
<w:c>Center text {{var}}</w:c>
<p align=center>Center with <b>bold</b> and <u>underline</u></p>
<pb>
<p>New page text</p>
```

Invalid:
```
<w:c>Text <u>underlined</u></w:c>         <!-- WRONG: tag inside wrapper -->
<p><pb/>Text after break</p>              <!-- WRONG: pb inside p -->
<p align=center><w:c>Nested</w:c></p>     <!-- WRONG: wrapper inside p -->
```

**Page Break Mid-Paragraph:**
If source has page break in middle of paragraph text, split into two paragraphs:
```
<p>Text before break</p>
<pb>
<p>Text after break</p>
```

---

## 3. Headings

```
[style:heading-1]
font-family="Arial"
font-size=24                 # in pt
color=#2b5797
bold=true
space-before=18              # in pt
space-after=12               # in pt
border-bottom=1pt
```

| Property | Description |
|----------|-------------|
| `font-family` | Heading font |
| `font-size` | Size in pt |
| `color` | Text color |
| `bold`, `italic`, `underline` | `true`/`false` |
| `align` | `left`, `center`, `right` |
| `space-before`, `space-after` | Spacing in pt |
| `border-bottom` | Bottom border line (e.g., `1pt`) |

Body: `<h1>`–`<h6>` with local override: `<h1 color=red font-size=28>`. Precedence: local attr → `[style:heading-N]` → `[style]`.

---

## 4. Tables

```
<table border=1 width=100%>
  <row bg=#f0f0f0 style=header>
    <col align=center width=30%>Name</col>
    <col align=center>City</col>
  </row>
  <loop:row x from entries>
    <col>{{x.name}}</col>
    <col>{{x.city}}</col>
  </loop:row>
</table>
```

| Tag | Description |
|-----|-------------|
| `<table>...</table>` | Table wrapper; properties: `border`, `width` |
| `<row>...</row>` | Row; properties: `bg`, `style` (named table-style) |
| `<col>...</col>` | Cell; properties: `align`, `width`, `bg`, `colspan`, `rowspan` |
| `<loop:row x from var>` | Loop data into rows; property: `style.first=name` (applies style `name` to first row only) |

Named table style example:

```
[style:table header]
bg=#f0f0f0
color=#000000
font-weight=bold
align=center
border-bottom=1pt
```

Dynamic style: `<row style={{myStyle}}>` where `{{myStyle}}` is a variable from section var/keys containing a style name.

---

## 5. Lists

```
<ul>
  <li>one
    <ul>
      <li>nested</li>
    </ul>
  </li>
</ul>
<ol>
  <li>first</li>
</ol>
```

Tags: `<ol>`, `<ul>`, `<li>`.

Horizontal rule: `<hr>` with optional `width`, `color`, `thickness`.

```
<hr width=50% color=#cccccc thickness=2pt>
```

---

## 6. Loops

```
<loop x from entries>        # basic
  <p>{{x.field}}</p>
</loop>
<loop:ol x from items>       # ordered list
  {{x.label}}
</loop:ol>
<loop:ul x from items>       # unordered list
  {{x.label}}
</loop:ul>
<loop:row x from items>      # table rows
  <col>{{x.name}}</col>
</loop:row>
```

All loop variants: `<loop>`, `<loop:ol>`, `<loop:ul>`, `<loop:row>`. Closing tag must match opening variant. `x` = loop variable alias, source must be listed in `var` (2nd+ position). Fields accessed as `{{x.field}}`.

---

## 7. Images

```
<img={{source.img}} width=80% align=center>     # from data
<img=./assets/photo.jpg width=400>               # static path
```

Properties: `width`, `height`, `align` (left/center/right), `alt`, `border`, `bg` (background color for image container).

---

## 8. Links

```
<a={{source.url}}>{{source.label}}</a>           # from data
<a=https://example.com>visit</a>                  # static
<a=#chapter1>see Chapter 1</a>                    # bookmark
<p>click <a={{url}} target=_blank>here</a></p>    # inline
```

Properties: `target` (`_blank`), `color`, `underline` (`true`/`false`).

---

## 9. Page & Section Breaks

| Tag / Syntax | Description |
|---|---|
| `<pb>` or `<page-break>` | Page break |
| `[section:next-page N]` | Section break + page break, N = sequence number |

---

## 10. Metadata

```
[title]
title=Document Title
subject=Document Subject
author=Author Name
```

| Property | Description |
|----------|-------------|
| `title` | Document title (accessible as `{{title}}`) |
| `subject` | Document description (accessible as `{{subject}}`) |
| `author` | Document creator (accessible as `{{author}}`) |

---

## 11. Header & Footer

```
[header]
left={{title}}
right={{page}} / {{total}}
font-size=10
color=#999999
border=bottom
margin=0.3
first-page=true

[footer]
center={{date}}
font-size=9
border=top
margin=0.2
first-page=false
```

| Property | Description |
|----------|-------------|
| `left`, `center`, `right` | Column content |
| `justify_between` | 2-3 comma-separated items spread via tab stops (use `\,` for literal comma). 2 items = left+right, 3 items = left+center+right. Examples: `justify_between={{company}}, {{date}}` or `justify_between=Acme\, Inc., {{title}}, {{date}}` |
| `font-family`, `font-size`, `color` | Font override |
| `border` | `top`, `bottom`, `none` |
| `margin` | Distance to content |
| `first-page` | `true`/`false` — show on page 1 |
| `mirror` | `true`/`false` — swap left↔right (for odd/even pages in duplex) |

### Unpredictable Objects & Keys

Objects/keys whose count/structure cannot be known from the source document alone (e.g. dynamic form fields, repeated signatures, variable-numbered annexes).

**Syntax:**

```
[object-unpredictable]
signatures=signer_name, position, date          ← single object
items=[]name, qty, price                         ← array of objects

[keys-unpredictable]
signer_name, position, date
```

**Rules:**
- `[object-unpredictable]` declares objects/arrays whose keys are not known ahead. Use `name=field, field` for single object, `name=[]field, field` for array of objects.
- `[keys-unpredictable]` declares flat key mappings across sections.
- Both are required in DCD output when the source document contains such variability.

### Variables

| Variable | Description |
|----------|-------------|
| `{{page}}` | Page number |
| `{{total}}` | Total pages |
| `{{title}}` | Document title |
| `{{date}}` | Compilation date |
