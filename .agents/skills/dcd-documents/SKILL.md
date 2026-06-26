# DCD DSL Specification

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

Any `{{...}}` referencing data fields should be in `keys` + `var`. Sections without var/keys pass unresolvable `{{...}}` as literals. Built-in vars (`{{page}}`, `{{total}}`, `{{title}}`, `{{date}}`) auto-resolved.

### Format Specifiers

`dd`, `MM`, `yyyy`, `HH`, `mm`, `ss` — date/time formats.

Array fields use dotted path in keys/formats: `keys=items.date_field`, `formats=[items.date_field:dd-MM-yyyy]`. After loop expansion, `{{x.date_field}}` → `{{items.0.date_field}}` matched by stripping the index.

### Block Tags (standalone, outside `<p>`)

| Tag | Description |
|-----|-------------|
| `<w:c>...</w:c>` | Center |
| `<w:b>...</w:b>` | Bold |
| `<w:i>...</w:i>` | Italic |
| `<w:u>...</w:u>` | Underline |
| `<w:c\|b>...</w:c\|b>` | Combined with pipe (e.g. `<w:b\|i\|u>`) |
| `<p>` | Paragraph |
| `<br>` | Line break |

> Note: In markdown tables above, `|` appears as `\|` due to markdown escaping. Actual DCD syntax uses plain `|` without backslash.

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

### Variables

| Variable | Description |
|----------|-------------|
| `{{page}}` | Page number |
| `{{total}}` | Total pages |
| `{{title}}` | Document title |
| `{{date}}` | Compilation date |
