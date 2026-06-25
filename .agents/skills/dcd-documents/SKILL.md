---
name: dcd-documents
description: Complete reference for the DCD DSL — sections, variables, body tags, styles, headings, tables, lists, loops, images, links, breaks, metadata, and header/footer
---

# DCD Documents

## 1. Style Configuration

Page layout and margin configuration for DCD documents.

```
[style]
layout=A4
unit=inch
orientation=portrait
font-family="Times New Roman"
font-size=12
color=#000000
line-height=1.5
m=1
```

### Layout

| Value    | Description                |
|----------|----------------------------|
| `A4`     | 210 × 297 mm               |
| `letter` | 8.5 × 11 in                |
| `legal`  | 8.5 × 14 in                |
| `A3`     | 297 × 420 mm               |
| `A5`     | 148 × 210 mm               |
| `B5`     | 176 × 250 mm               |
| `custom` | Requires explicit w / h    |

### Unit

`inch`, `cm`, `mm`, `pt`, `pica`

### Orientation

| Value       | Description                         |
|-------------|-------------------------------------|
| `portrait`  | Default. Taller than wide.          |
| `landscape` | Wider than tall. Swap width/height. |

```
[style]
layout=A4
unit=inch
orientation=landscape
```

### Font

| Property      | Description       | Example                   |
|---------------|-------------------|---------------------------|
| `font-family` | Font family name  | "Times New Roman", Arial  |
| `font-size`   | Base font size    | 12pt                      |
| `color`  | Text color        | #000000, black            |
| `line-height` | Line spacing      | 1.5                       |

```
[style]
layout=A4
unit=inch
font-family="Times New Roman"
font-size=12
color=#000000
line-height=1.5
```

### Margins

All margin examples below assume:

```
[style]
layout=A4
unit=inch
orientation=portrait
font-family="Times New Roman"
font-size=12
line-height=1.5
```

**Uniform:**
```
m=1
```

**Axis:** `mx` = left & right, `my` = top & bottom.
```
mx=1
my=1
```

**Individual:** `mt` top, `mb` bottom, `ml` left, `mr` right.
```
mt=1
mb=1
ml=1
mr=1
```

**Default + Bottom:** `md` = margin default (all sides), `mb` = bottom (override).
```
md=1
mb=1
```

**Precedence (low → high):**
1. `m`
2. `mx` / `my`
3. `md`
4. `mt` / `mb` / `ml` / `mr`

## 2. Sections & Variables

Document content template with structured data.

```
[section 0]
name=userinfo
var=info, entries
keys=username, date_field, time_field
formats=[date_field:dd-MM-yyyy], [time_field:HH\:m]

--- BODY ---
<w:c|b>Center Bold</w:c|b>
<p>your username is <b>{{info.username}}</b> created on <i>{{info.date_field}}</i> at <u>{{info.time_field}}</u></p>
<loop x from entries>
   {{x.name}} lives at {{x.address}}
</loop>

[section 1]
name=address
var=addr
keys=street, city, zip

--- BODY ---
<p>{{addr.street}}, {{addr.city}} - {{addr.zip}}</p>

[section 2]
name=simple
keys=title, message

--- BODY ---
<p>{{title}}: {{message}}</p>
```

### Section Properties

| Property   | Description                            |
|------------|----------------------------------------|
| `name`       | Section identifier                     |
| `var`        | Comma-separated variable names, each is an **object/map/object-in-array (array)**. Pattern: `var=nameA, nameB, ...` — **first** `nameA` is prefix for `{{nameA.key}}` via `keys`. **Subsequent** `nameB` are data source names used by `<loop x from nameB>`. |
| `keys`       | Comma-separated field names for data variable resolution. For primary var: field names. For array fields: `source.field` (e.g. `items.date_field`). Optional — sections without `var`/`keys` pass `{{...}}` through as literals. |
| `formats`    | Per-key format: `[key:format]` or `[source.field:format]`. Defines the output format of a key. The key (or dotted path) must be listed in `keys`. For array fields in loops, use `[source.field:format]` (e.g. `[items.date_field:dd-MM-yyyy`). |

> Properties use `=` separator (e.g. `name=example`). The parser also accepts `:` (e.g. `name:example`).

### Var Usage

```
var=info, entries
```

| Position | Name       | Source of data           | Access in body                      |
|----------|------------|--------------------------|-------------------------------------|
| 1st      | `info`     | Resolved via `keys`      | `{{info.username}}`                 |
| 2nd+     | `entries`  | Array data source        | `<loop x from entries>{{x.name}}</loop>` |

- **First name** (`info`): variable prefix. Fields listed in `keys`. Accessed as `{{info.key}}`.
- **Additional names** (`entries`, ...): data sources for loops. Accessed via `<loop x from entries>`, then `{{x.field}}` per item.

### Variables

`{{var.key}}` — e.g. `{{info.username}}`, `{{info.date_field}}`.

Built-in variables (resolved automatically, no registration needed):
- `{{date}}` — current compilation date
- `{{title}}` — document title (from `[title]` section)

### Format Specifiers

Format is defined as `[key:format]` in the `formats` property.

| Specifier | Description     |
|-----------|-----------------|
| `dd`      | Day (01–31)     |
| `MM`      | Month (01–12)   |
| `yyyy`    | Year (4 digit)  |
| `HH`      | Hour (00–23)    |
| `mm`      | Minute (00–59)  |
| `ss`      | Second (00–59)  |

Example: `[date_field:yyyy-MM-dd]` → `2026-06-24`

Besides the specifiers above, format also supports regex patterns like `\d`, `\w`, or other regex.

### Format for Array Fields

For fields inside array objects (used in `<loop x from source>`), use **dotted path** in both `formats` and `keys`:

```ini
[section 0]
var=info, items
keys=title, items.date_field
formats=[items.date_field:dd-MM-yyyy]

--- BODY ---
<h1>{{info.title}}</h1>
<loop x from items>
  <p>{{x.name}} — {{x.date_field}}</p>   ← formatted via dotted path match
</loop>
```

After loop expansion, `{{x.date_field}}` becomes `{{items.0.date_field}}` — the format engine matches it against `items.date_field` by stripping the array index.

### Variable Registration Rule

`{{...}}` variables that reference data fields should be registered in:
- **`keys`** — field names or dotted paths
- **`var`** — data source names

Sections without `var` or `keys` are allowed: unresolvable `{{...}}` variables pass through as **literal strings** (e.g. `{{unknown}}` appears as-is). Built-in variables (`{{date}}`, `{{title}}`) are resolved automatically regardless of registration.

### Block Tags (outside `<p>`)

| Tag                              | Description                     |
|----------------------------------|---------------------------------|
| `<w:c>...</w:c>`                 | Center                          |
| `<w:b>...</w:b>`                 | Bold                            |
| `<w:i>...</w:i>`                 | Italic                          |
| `<w:u>...</w:u>`                 | Underline                       |
| `<w:c\|b>...</w:c\|b>`           | Center + Bold                   |
| `<w:b\|i>...</w:b\|i>`           | Bold + Italic                   |
| `<w:b\|i\|u>...</w:b\|i\|u>`     | Bold + Italic + Underline       |
| `<p>`                            | Paragraph                       |
| `<br>`                           | Line break                      |
| `<loop x from var>...</loop>`     | Iterate array `var`, each item as `x` |
| `<loop:ol x from var>...</loop:ol>` | Iterate + wrap `<ol><li>`       |
| `<loop:ul x from var>...</loop:ul>` | Iterate + wrap `<ul><li>`       |

> Note: `\|` inside table cells is markdown escape for `|` — the actual tag is `<w:b|i>` etc.

### Inline Tags (inside `<p>`, `<li>`, `<col>`)

| Tag              | Description             |
|------------------|-------------------------|
| `<b>...</b>`     | Bold                    |
| `<i>...</i>`     | Italic                  |
| `<u>...</u>`     | Underline               |
| `<code>...</code>`| Monospace / code font  |
| `<set:flags>...</set:flags>` | Combined formatting |

Combined formatting with `<set:>`:

```
<p><set:b|i>Bold and Italic</set:b|i></p>
<p><set:b|u>Bold and Underline</set:b|u></p>
<p><set:i|code>Italic monospace</set:i|code></p>
<p><set:b|i|u>Bold, Italic, and Underline</set:b|i|u></p>
```

**Available flags:** `b` (bold), `i` (italic), `u` (underline), `code` (monospace)

**Closing tag:** Can be `</set:flags>` (matching) or `</set>` (simplified)

## 3. Headings

Heading `<h1>`–`<h6>` with global style in `[style:heading-N]`.

```
[style]
layout=A4
unit=inch
m=1

[style:heading-1]
font-family="Arial"
font-size=24
color=#2b5797
bold=true
space-before=18
space-after=12
border-bottom=1pt

[style:heading-2]
font-family="Arial"
font-size=18
color=#444444
bold=true
space-before=12
space-after=6

[style:heading-3]
font-family="Arial"
font-size=14
color=#444444
bold=false
space-before=6
space-after=3
```

Body:

```
--- BODY ---
<h1>Chapter 1: Introduction</h1>
<p>lorem ipsum...</p>
<h2>1.1 Background</h2>
<p>lorem ipsum...</p>
<h3>1.1.1 Sub Section</h3>
<p>lorem ipsum...</p>
```

Local override (higher priority):

```
<h1 color=red font-size=28>Chapter with local style</h1>
```

### Style Properties

| Property      | Description             |
|---------------|-------------------------|
| `font-family` | Heading font            |
| `font-size`   | Font size (pt)          |
| `color`  | Text color              |
| `bold`        | `true` / `false`        |
| `italic`      | `true` / `false`        |
| `underline`   | `true` / `false`        |
| `align`       | `left`, `center`, `right` |
| `space-before`| Space before (pt)       |
| `space-after` | Space after (pt)        |
| `border-bottom` | Bottom border line    |

### Precedence

1. Local attribute on tag `<h1 color=red>`
2. `[style:heading-N]` global
3. `[style]` font default

## 4. Tables

### Dynamic Table

```
<table border=1 width=100%>
<loop:row x from headers>
   <col>{{x}}</col>
</loop:row>
<loop:row x from entries>
   <col>{{x.field1}}</col>
   <col>{{x.field2}}</col>
</loop:row>
</table>
```

### Static Table

```
<table border=1>
  <row bg=#f0f0f0>
    <col align=center width=30%>Name</col>
    <col align=center width=30%>City</col>
    <col align=center width=40%>Age</col>
  </row>
  <row>
    <col align=left>John</col>
    <col align=left>Jakarta</col>
    <col align=center>25</col>
  </row>
</table>
```

### Tags

| Tag                              | Description                  |
|----------------------------------|------------------------------|
| `<table>...</table>`             | Table wrapper                |
| `<row>...</row>`                 | Row                          |
| `<col>...</col>`                 | Cell                         |
| `<loop:row x from var>...</loop:row>` | Loop data into rows    |

### Table Properties

| Property  | Example   | Description          |
|-----------|-----------|----------------------|
| `border`  | `1`       | Border width         |
| `width`   | `100%`    | Table width¹         |

¹ Not supported in v0.1.5 (library limitation).

### Row Properties

| Property  | Example       | Description          |
|-----------|---------------|----------------------|
| `bg`      | `#f0f0f0`     | Row background       |
| `style`   | `header`      | Named table-style    |

### Col Properties

| Property  | Example       | Description          |
|-----------|---------------|----------------------|
| `align`   | `center`      | Text alignment       |
| `width`   | `30%`         | Column width¹        |
| `bg`      | `#e0e0e0`     | Cell background      |
| `colspan` | `2`           | Merge columns¹       |
| `rowspan` | `2`           | Merge rows¹          |

¹ Not supported in v0.1.5 (library limitation).

### Named Table Style

```
[style:table header]
bg=#2b5797
color=white
font-weight=bold
align=center
border-bottom=2pt

[style:table alt]
bg=#f5f5f5
```

Usage:

```
<table border=1>
  <row style=header>
    <col>Name</col>
    <col>City</col>
  </row>
  <row style=alt>
    <col>John</col>
    <col>Jakarta</col>
  </row>
</table>
```

### Loop with style.first

Apply style to first row only:

```
<table border=1>
  <loop:row x from items style.first=header>
    <col>{{x.name}}</col>
    <col>{{x.value}}</col>
  </loop:row>
</table>
```

### Dynamic Row Style

Use variable for style name:

```
<row style={{myStyle}}>
  <col>Data</col>
</row>
```

## 5. Lists

Standalone lists (not from loop).

```
<ul>
  <li>item one</li>
  <li>item two</li>
  <li>item three</li>
</ul>

<ol>
  <li>first</li>
  <li>second</li>
  <li>third</li>
</ol>
```

Nested:

```
<ul>
  <li>fruit
    <ul>
      <li>apple</li>
      <li>mango</li>
    </ul>
  </li>
  <li>vegetable</li>
</ul>
```

### Tags

| Tag       | Description              |
|-----------|--------------------------|
| `<ol>`    | Ordered list             |
| `<ul>`    | Unordered list           |
| `<li>`    | List item                |

### Horizontal Rule

```
<hr>
```

| Tag       | Description              |
|-----------|--------------------------|
| `<hr>`    | Horizontal rule          |

Properties:

| Property | Example   | Description     |
|----------|-----------|-----------------|
| `width`  | `50%`     | Line width      |
| `color`  | `#cccccc` | Line color      |
| `thick`  | `2`       | Thickness (pt)¹ |

¹ Not supported in v0.1.5 (library limitation).

## 6. Loops

Iterate over array data sources declared in `var`.

The data source name must be listed in `var` (after the first name). See [Var Usage](#var-usage).

```
[section 0]
name=example
var=info, entries
keys=title

--- BODY ---
<loop x from entries>
  {{x.field}}
</loop>
```

Here `entries` is the 2nd name in `var=info, entries` — an array data source for the loop.

### Tags

| Tag                                    | Description                            |
|----------------------------------------|----------------------------------------|
| `<loop x from name>...</loop>`         | Iterate array `name`, each item as `x` |
| `<loop:ol x from name>...</loop:ol>`   | Iterate + wrap each in `<ol><li>`      |
| `<loop:ul x from name>...</loop:ul>`   | Iterate + wrap each in `<ul><li>`      |
| `<loop:row x from name>...</loop:row>` | Iterate into table rows                |

> Closing tag must match the opening variant: `<loop:ol>` closes with `</loop:ol>`, `<loop:row>` with `</loop:row>`, etc.

### Basic Loop

```
<loop x from entries>
  <p>{{x.name}} — {{x.value}}</p>
</loop>
```

- `x` — loop variable alias (any name)
- `entries` — must match a name in `var` (2nd position or later)
- Inside: `{{x.field}}` accesses a field on each array element

### Loop with Ordered List

```
<loop:ol x from items>
  {{x.label}}
</loop:ol>
```

Renders as `<ol><li>value</li><li>value</li></ol>`.

### Loop with Unordered List

```
<loop:ul x from items>
  {{x.label}}
</loop:ul>
```

Renders as `<ul><li>value</li><li>value</li></ul>`.

### Loop into Table Rows

```
<table border=1>
<loop:row x from headers>
  <col>{{x}}</col>
</loop:row>
<loop:row x from entries>
  <col>{{x.field1}}</col>
  <col>{{x.field2}}</col>
</loop:row>
</table>
```

- First `loop:row` iterates `headers` — each item is a cell value (`{{x}}`)
- Second `loop:row` iterates `entries` — each item is an object (`{{x.field}}`)
- Each iteration produces a `<row>` with `<col>` cells

### Full Example

```
[section 0]
name=products
var=info, items
keys=title, items.date, items.price
formats=[items.date:dd-MM-yyyy], [items.price:#,##0.00]

--- BODY ---
<h1>{{info.title}}</h1>
<table border=1 width=100%>
  <loop:row x from items>
    <col>{{x.name}}</col>
    <col align=right>{{x.price}}</col>
    <col>{{x.date}}</col>
  </loop:row>
</table>
```

Fields from array objects (`items.date`, `items.price`) use dotted path notation in `keys` and `formats`. See [Format for Array Fields](#format-for-array-fields).

## 7. Images

From data section:

```
[section 0]
name=gallery
var=source
keys=img, caption

--- BODY ---
<img={{source.img}} width=80% align=center>
<p><i>{{source.caption}}</i></p>
```

Static path:

```
<img=./assets/photo.jpg width=400>
```

### Properties

| Property   | Example        | Description                 |
|------------|----------------|-----------------------------|
| `width`    | `100%`, `400`  | Width (px or %)             |
| `height`   | `300`          | Height (px)                 |
| `align`    | `center`       | `left`, `center`, `right`   |
| `alt`      | "photo"        | Alternative text            |
| `border`   | `1`            | Border width                |
| `bg`  | `#f0f0f0`      | Background container        |

## 8. Links

Internal and external hyperlinks.

From data section:

```
<section 0>
var=source
keys=url, label

--- BODY ---
<a={{source.url}}>{{source.label}}</a>
```

Static:

```
<a=https://example.com>visit website</a>
```

Inline:

```
<p>click <a={{source.url}} target=_blank>here</a> for more info</p>
```

### Properties

| Property    | Example         | Description          |
|-------------|-----------------|----------------------|
| `target`    | `_blank`        | Open in new tab (DOCX always opens external links in new window) |
| `color`     | `#0055cc`       | Link color           |
| `underline` | `true`          | Underline            |

### Bookmark

```
<a=#chapter1>see Chapter 1</a>
```

## 9. Page & Section Breaks

### Page Break

```
--- BODY ---
<p>page 1</p>
<pb>
<p>page 2</p>
```

| Tag             | Description       |
|-----------------|-------------------|
| `<pb>`          | Page break        |
| `<page-break>`  | Alias for `<pb>`  |

### Section Break

```
[section 0]
name=cover
var=info
keys=title, author

--- BODY ---
<h1>{{info.title}}</h1>
<p>{{info.author}}</p>

[section:next-page 1]

--- BODY ---
<p>new section after page break</p>
```

| Syntax                           | Description                           |
|----------------------------------|---------------------------------------|
| `[section:next-page N]`          | Section break + page break            |

`N` = section sequence number.

## 10. Metadata

Set document properties like title, subject, and author using the `[title]` section.

```
[title]
title=Document Title
subject=Document Subject
author=Author Name
```

### Properties

| Property  | Description                          | Example                    |
|-----------|--------------------------------------|----------------------------|
| `title`   | Document title                       | Annual Report 2025         |
| `subject` | Document subject/description         | Financial Summary          |
| `author`  | Document author/creator              | Finance Team               |

These properties are written to:
- **DOCX:** Document properties (`docProps/core.xml`)
- **PDF:** Document metadata (Title, Subject, Author fields)

### Built-in Variable: `{{title}}`

The `title` property can be referenced in headers and footers using the `{{title}}` variable.

```
[title]
title=My Report

[header]
left={{title}}
right={{date}}

[footer]
center={{title}} - Page {{page}}
```

### Full Example

```
[style]
layout=A4

[title]
title=Quarterly Business Review
subject=Q4 2024 Performance Report
author=Executive Team

[header]
left={{title}}
right={{date}}
font-size=8
color=#666666

[section 0]

--- BODY ---
<h1>{{title}}</h1>
<p>Prepared by: Finance Department</p>
```

The document will have:
- Title property set to "Quarterly Business Review"
- Subject set to "Q4 2024 Performance Report"
- Author set to "Executive Team"
- Header showing the title and current date
- Body displaying the title as heading

### Notes

- All properties are optional
- Properties are visible in document properties dialog (DOCX) or PDF metadata
- The `{{title}}` variable only works in headers/footers and body content
- Use `{{date}}` for current date, `{{page}}` for page numbers

## 11. Header & Footer

Header and footer for document pages.

```
[header]
left={{title}}
right={{page}} / {{total}}

[footer]
center={{date}}
```

### Properties

| Property      | Description                            |
|---------------|----------------------------------------|
| `left`        | Left column content                    |
| `center`      | Center column content                  |
| `right`       | Right column content                   |
| `font-family` | Header/footer font override            |
| `font-size`   | Font size                              |
| `color`  | Text color                             |
| `border`      | `top`, `bottom`, `none`                |
| `margin`      | Distance from header/footer to content |
| `first-page`  | `true` / `false` — show on page 1     |
| `mirror`      | `true` / `false` — swap left↔right¹   |

¹ Not fully supported in v0.1.5 (library limitation).

### Variables

| Variable      | Description          |
|---------------|----------------------|
| `{{page}}`    | Page number          |
| `{{total}}`   | Total pages          |
| `{{title}}`   | Document title       |
| `{{date}}`    | Compilation date     |

### Full Example

```
[style]
layout=A4
unit=inch
m=1

[header]
left={{title}}
right={{page}} / {{total}}
font-size=10
color=#999999
border=bottom
margin=0.3

[footer]
center={{date}}
font-size=9
color=#666666
border=top
margin=0.2
first-page=false
```

## See Also

- `dcd-cli` — CLI usage and options
- `golang-programming` — Go library API
- `dcd-guide` — Project overview and patterns
