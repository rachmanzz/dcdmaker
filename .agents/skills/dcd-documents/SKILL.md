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
font-size=12pt
color=#000000
line-spacing=1.5
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

`inch`, `cm`, `mm`, `pt`, `pica` — `inch` by default.

All length values (margins, page `w`/`h`, `space-before/after`, `indent-left`/`indent-right`/`first-line`/`indent-hanging`, `line-spacing` with `line-rule=exact`/`atLeast`, `tabs` positions, header/footer margins) use the declared unit. A bare number means the declared unit; any other unit must include a suffix (`12pt`, `2cm`, `5mm`). `font-size` and `letter-spacing` are always in `pt` and the suffix is MANDATORY (`font-size=12pt` — a bare `font-size=12` is an error). Image `width`/`height` are in inches.

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
| `line-spacing` | Line spacing. Bare number without `line-rule` = multiplier; with `line-rule=exact`/`atLeast` = absolute length | 1.5 |
| `line-rule` | Line spacing rule: `exact`, `atLeast`, or omit for `auto` | exact |

```
[style]
layout=A4
unit=inch
font-family="Times New Roman"
font-size=12pt
color=#000000
line-spacing=1.5
```

### Paragraph

Global default for paragraph indentation.

| Property  | Description                      | Example |
|-----------|----------------------------------|---------|
| `indent-left`  | Left indent (in document unit)   | 1       |
| `indent-hanging` | Hanging indent (in document unit)| 0.5     |

```
[style]
indent-left=0.5
indent-hanging=0.25
```

Inline `<p indent-left=X>` overrides this default.

### Margins

All margin examples below assume:

```
[style]
layout=A4
unit=inch
orientation=portrait
font-family="Times New Roman"
font-size=12pt
line-spacing=1.5
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
var=info, []entries
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
| `name`       | Section identifier — **REQUIRED** in every `[section N]`. Must be declared before `var=` and `keys=`. |
| `var`        | Comma-separated variable names. **Objects:** plain name (e.g. `info`). **Arrays (loop sources):** prefix with `[]` (e.g. `[]entries`). Pattern: `var=info, []entries` — **first** `info` is prefix for `{{info.key}}` via `keys`. **Subsequent** `[]entries` is a data source for `<loop x from entries>`. |
| `keys`       | Comma-separated field names for data variable resolution. For primary var: field names. For array fields requiring formatting: `source.field` (e.g. `items.date_field`). **CONDITIONAL DOT-NOTATION RULE:** Dotted paths (object/array fields) MUST NOT be registered in `keys=` UNLESS they are explicitly formatted in `formats=`. Optional — sections without `var`/`keys` pass `{{...}}` through as literals. |
| `formats`    | Per-key format: `[key:format]` or `[source.field:format]`. Defines the output format of a key. **EXCLUSIVE REGISTRATION RULE:** Any key or dotted path targeted in `formats=` MUST be explicitly listed in `keys=`. For array fields in loops, use `[source.field:format]` (e.g. `[items.date_field:dd-MM-yyyy`). |

> Properties use `=` separator (e.g. `name=example`).

### Var Usage

- **Section Limits:** Aim for ≤ 5 `var` entries and ≤ 15 `keys` entries per section.
- **Splitting Rule:** If you exceed these limits, split into a new logical section (e.g. `[section 1]`, `[section 2]`).

```
var=info, []entries
```

| Position | Name         | Source of data           | Access in body                                |
|----------|--------------|--------------------------|-----------------------------------------------|
| 1st      | `info`       | Resolved via `keys`      | `{{info.username}}`                           |
| 2nd+     | `[]entries`  | Array data source        | `<loop x from entries>{{x.name}}</loop>`      |

- **First name** (`info`): variable prefix. Fields listed in `keys`. Accessed as `{{info.key}}`.
- **Additional names** (`[]entries`, ...): data sources for loops. Accessed via `<loop x from entries>`, then `{{x.field}}` per item.

### Variables

`{{var.key}}` — e.g. `{{info.username}}`, `{{info.date_field}}`.

Built-in variables (resolved automatically, no registration needed):
- `{{date}}` — current compilation date
- `{{title}}` — document title (from `[title]` section)
- `{{page}}` — page number (works in header/footer; passed as literal in body)
- `{{total}}` — total pages (works in header/footer; passed as literal in body)

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

For fields inside array objects (used in `<loop x from source>`) that need formatting, use **dotted path** in both `keys` and `formats`. Fields that do NOT need formatting must NOT appear in `keys=`:

```ini
[section 0]
var=info, []items
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

`{{...}}` variables that reference data fields must be registered in:
- **`keys`** — field names or dotted paths
- **`var`** — data source names

**Strict Usage:** Every variable in `var=` and every key in `keys=` MUST be used at least once in `--- BODY ---`. Do NOT declare unused variables or keys.

Sections without `var` or `keys` are allowed: unresolvable `{{...}}` variables pass through as **literal strings** (e.g. `{{unknown}}` appears as-is). Built-in variables (`{{date}}`, `{{title}}`, `{{page}}`, `{{total}}`) are resolved automatically regardless of registration.

### Block Tags (outside `<p>`)

> **TAG BALANCING:** Every opened tag must be closed exactly. `<loop:ol>` MUST close with `</loop:ol>`, NOT `</loop>`.
>
> **NO `<w:*>` NESTING:** `<w:*>` tags MUST NOT contain other `<w:*>` tags. `<w:c><w:b>text</w:b></w:c>` is an error. Use OR logic in a single tag: `<w:c|b>text</w:c|b>`.
>
> **TAB INSIDE `<w:*>`:** `<tab>`, `<tab/>`, and `<tab size=N>` ARE allowed inside `<w:*>` tags. `<br>` is also allowed.
>
> **NO HEADING NESTING:** Heading tags `<h1>`–`<h9>` MUST NOT appear inside `<p>`, `<w:*>`, or other heading tags. `<p><h2>text</h2></p>` and `<h1><h2>text</h2></h1>` are errors.

| Tag                              | Description                     |
|----------------------------------|---------------------------------|
| `<w:c>...</w:c>`                 | Center                          |
| `<w:b>...</w:b>`                 | Bold                            |
| `<w:i>...</w:i>`                 | Italic                          |
| `<w:u>...</w:u>`                 | Underline                       |
| `<w:s>...</w:s>`                 | Strikethrough                   |
| `<w:c\|b>...</w:c\|b>`           | Center + Bold                   |
| `<w:b\|i>...</w:b\|i>`           | Bold + Italic                   |
| `<w:b\|i\|u>...</w:b\|i\|u>`     | Bold + Italic + Underline       |
| `<w:c\|s>...</w:c\|s>`           | Center + Strikethrough           |
| `<p>`                            | Paragraph                       |
| `<br>`                           | Line break                      |
| `<tab>`                          | Tab character (inside `<p>`)    |
| `<tab size=N>`                   | Tab with N spaces               |
| `<loop x from var>...</loop>`     | Iterate array `var`, each item as `x` |
| `<loop:ol x from var>...</loop:ol>` | Iterate + wrap `<ol>`; each iteration MUST include `<li>...</li>` or `<p>...</p>` |
| `<loop:ul x from var>...</loop:ul>` | Iterate + wrap `<ul>`; each iteration MUST include `<li>...</li>` or `<p>...</p>` |

> Note: `\|` inside table cells is markdown escape for `|` — the actual tag is `<w:b|i>` etc.

### Paragraph Properties

`<p>` and `<h1>`–`<h9>` tags accept paragraph-level formatting attributes (`<li>` does NOT accept attributes — set properties on inner `<p>` tags instead):

> **KEBAB-CASE MANDATORY:** All attribute names are lowercase kebab-case — `indent-left`, `indent-hanging`, `line-spacing`, `line-rule`, `space-before`, `first-line`. camelCase forms (`indentLeft`, `indentHanging`, `lineSpacing`, `lineRule`) and any other casing are INVALID and silently ignored. Write exactly the kebab-case form documented below.

> **QUOTED VALUES:** Any attribute value may be wrapped in double quotes — this is REQUIRED when the value contains spaces (e.g. `style="Body Text Indent 3"`). Quoted values are stripped of their surrounding quotes; bare values must not contain spaces. Unquoted single-word values are unchanged and remain the common form.

#### Indent

| Property         | Example                  | Description                                      |
|------------------|--------------------------|--------------------------------------------------|
| `indent-left`    | `indent-left=0.5`        | Left indent (in document unit)                   |
| `indent-right`   | `indent-right=0.5`       | Right indent (in document unit)                  |
| `indent-hanging` | `indent-hanging=0.25`    | Hanging indent (removed from first line)         |
| `first-line`     | `first-line=0.5`         | Extra indent for first line only                 |

`indent-hanging` and `first-line` are mutually exclusive — if both are set, `first-line` takes precedence.

#### Spacing

| Property       | Example              | Description                                          |
|----------------|----------------------|------------------------------------------------------|
| `space-before` | `space-before=12pt`  | Space before paragraph (layout unit)                   |
| `space-after`  | `space-after=6pt`    | Space after paragraph (layout unit)                    |
| `line-spacing` | `line-spacing=1.5`   | Line spacing: multiplier (bare number without `line-rule`) or absolute length with `line-rule=exact`/`atLeast` (`12pt`, `0.17in`, `6mm`) |
| `line-rule`    | `line-rule=exact`    | Line spacing rule: `auto` (default), `exact`, `atLeast`. Only applied when `line-spacing` is also set. |

**Line Spacing:**

`line-spacing` follows docx semantics:

1. **Multiplier** — bare number (no unit suffix), no `line-rule`: `1.5`, `2.0`, `1.2`
   - `line-spacing=1.5` → 1.5× normal line spacing (`lineRule=auto`)

2. **Exact / At-least** — absolute length with `line-rule=exact` or `line-rule=atLeast`: `12pt`, `0.17in`, `6mm`, `1cm`
   - `line-spacing=12pt line-rule=exact` → exactly 12pt line height
   - A bare number with `line-rule=exact`/`atLeast` is also an absolute length in the document unit

```
<p line-spacing=1.5>1.5× normal spacing</p>
<p line-spacing=12pt line-rule=exact>exactly 12pt line height</p>
<p line-spacing=0.17in line-rule=exact>exactly 0.17 inch</p>
<p line-spacing=6mm line-rule=atLeast>minimum 6mm line height</p>
```

#### Borders

| Property         | Example                 | Description                                                                  |
|------------------|-------------------------|------------------------------------------------------------------------------|
| `border`         | `border=1pt`            | Border on all sides (single line)                                            |
| `border-bottom`  | `border-bottom=single`  | Border on one side — supported: `border-bottom`, `border-top`, `border-left`, `border-right` |

Border color uses `border-color` (default `auto`).

#### Pagination & Flow

| Property        | Example                   | Description                                |
|-----------------|---------------------------|--------------------------------------------|
| `keep-next`     | `keep-next=true`          | Keep with next paragraph                   |
| `keep-lines`    | `keep-lines=true`         | Keep lines together on same page           |
| `section-break` | `section-break=true`      | Force section break before this paragraph  |
| `widow-control` | `widow-control=false`     | Widow/orphan control (`true` / `false`)    |
| `contextual-spacing`| `contextual-spacing=true` | Remove space between same-style paragraphs |

#### Appearance

| Property       | Example                  | Description                                        |
|----------------|--------------------------|----------------------------------------------------|
| `shading`      | `shading=#f0f0f0`        | Paragraph background color                         |
| `outline-level`| `outline-level=1`        | Outline level 0–8 (used for TOC)                   |
| `dir`          | `dir=rtl` / `dir=ltr`    | Text direction                                    |
| `valign`       | `valign=center`          | Vertical alignment within line: `auto`, `top`, `center`, `baseline`, `bottom` |

#### Tab Stops

| Property | Example                                      | Description                             |
|----------|----------------------------------------------|-----------------------------------------|
| `tabs`   | `tabs="0.32 0.63 0.95 1.26"`                 | Space-separated tab stop positions      |
| `tabs`   | `tabs="6.10@right:hyphen"`                   | Tab stop with alignment and leader      |

Applies to `<p>` and `<h1>`–`<h9>`. Tab stop format:
- **position**: length in the document unit (1 inch = 1440 twips); a unit suffix (`in`, `cm`, `mm`, `pt`, `pica`) is allowed
- **`@alignment`** (optional): `L` left, `C` center, `R` right, `D` decimal
- **`:leader`** (optional): `none`, `dot`, `hyphen`, `underscore`, `middleDot`

Plain positions default to `L` / `none`. Positions with and without `@align:leader` can be mixed:

```
<p tabs="0.32 0.63 0.95 1.26">quarter-inch tabs</p>

<p tabs="6.10@right:hyphen">right-aligned with hyphen leader</p>

<p tabs="0.5 1in@right:dot 2.5cm@center:hyphen">mixed formats</p>
```

```
<p indent-left=1 indent-hanging=0.5>
  First line starts 0.5 from left margin,
  rest of paragraph indented 1 from margin.
</p>

<li>
  <p indent-left=0.5 indent-hanging=0.25>list item with custom indent</p>
</li>

<p space-before=12pt space-after=6pt line-spacing=1.5>spaced paragraph</p>

<p border=1pt border-color=#2b5797>framed paragraph</p>

<li>
  <p shading=#ffffcc>shaded list item</p>
</li>

<p section-break=true>starts on new section</p>

<p tabs="0.5 2@right:dot">content	right-aligned with dot leader</p>
```

**Precedence (high → low):**
1. Inline attribute on tag
2. `[style:paragraph <name>]` applied via `style=`
3. `[style:heading-N]` heading style (headings only)
4. `[style]` global default

### Named Paragraph Styles

Define reusable paragraph styles with `[style:paragraph <name>]` and apply them via `style=<name>` on `<p>` and `<h1>`–`<h9>`.

```
[style:paragraph quote]
font-size=14pt
color=#666666
indent-left=1
border=1pt

[style:paragraph code-block]
font-family="Consolas"
font-size=10pt
line-spacing=1.2
space-before=6pt
space-after=6pt
```

Usage:

```
<p style=quote>This is a quoted paragraph.</p>
<p style=code-block>func main() { fmt.Println("hi") }</p>
```

**Rules:**
1. Section name format: `[style:paragraph <name>]`
2. Supports all [Paragraph Properties](#paragraph-properties)
3. Apply via `style=<name>` on `<p>` or `<h1>`–`<h9>`
4. Inline attributes on the tag override the named style
5. Named style overrides the global `[style]` default
6. Style names may contain spaces. When applying, wrap the name in double quotes: `<p style="Body Text Indent 2">`. In the section definition the name may be bare (`[style:paragraph Body Text Indent 3]`) or, for multi-word names, wrapped in quotes (`[style:paragraph "Body Text Indent 3"]`) — both register the same `Body Text Indent 3`.

**In lists:**

```
[style:paragraph highlight]
shading=#ffffcc
indent-left=0.5

<ul>
<li>
<p style=highlight>Highlighted list item</p>
</li>
</ul>
```

### Inline Tags (inside `<p>`, `<li>`, `<col>`)

| Tag              | Description             |
|------------------|-------------------------|
| `<b>...</b>`     | Bold                    |
| `<i>...</i>`     | Italic                  |
| `<u>...</u>`     | Underline               |
| `<s>...</s>`     | Strikethrough           |
| `<code>...</code>`| Monospace / code font  |
| `<mark>...</mark>`| Highlight (default yellow, optional: `<mark color=green>`) |
| `<sub>...</sub>`  | Subscript               |
| `<sup>...</sup>`  | Superscript             |
| `<span attrs>...</span>` | Run formatting (color, bg, font-size, etc.) |
| `<set:flags>...</set:flags>` | Combined formatting |

**`<span>` run formatting:** Apply run-level attributes via `<span attrs>`. Only attributes with **no dedicated inline tag** belong here — `bold`/`italic`/`strike`/`underline` use `<b>`/`<i>`/`<s>`/`<u=style>` instead. Nested `<span>` inside `<span>` is NOT allowed (parse error); it composes with the other inline tags.

| Attribute      | Example                      | Description                          |
|----------------|------------------------------|--------------------------------------|
| `color`        | `color=#FF0000`              | Text color (hex or name)             |
| `bg`           | `bg=#ffffcc`                 | Background shading                   |
| `font-size`    | `font-size=18pt`             | Font size                            |
| `font-family`  | `font-family="Courier New"`  | Font family                          |
| `caps`         | `caps=true`                  | All caps                             |
| `small-caps`   | `small-caps=true`            | Small caps                           |
| `letter-spacing` | `letter-spacing=1pt`       | Letter spacing                       |

```
<p><span color=#FF0000>red text</span> and <span bg=#ffffcc>highlighted</span></p>
<p><span font-family="Courier New" font-size=10pt>small monospace</span></p>
<p><span color=#0000FF>blue <b>and bold</b> rest</span></p>
```

Combined formatting with `<set:>`:

```
<p><set:b|i>Bold and Italic</set:b|i></p>
<p><set:b|u>Bold and Underline</set:b|u></p>
<p><set:i|code>Italic monospace</set:i|code></p>
<p><set:b|i|u>Bold, Italic, and Underline</set:b|i|u></p>
<p><set:s|b>Strikethrough and Bold</set:s|b></p>
```

**Available flags:** `b` (bold), `i` (italic), `u` (underline), `s` (strikethrough), `code` (monospace)

**Closing tag:** Must match opening flags: `<set:u>text</set:u>`, `<set:b|i>text</set:b|i>`

**Attributes:** Pass additional formatting via attributes on `<set:flags>`:

```
<p><set:u underline=double>double underline</set:u></p>
<p><set:b|u underline=dash>bold with dashed underline</set:b|u></p>
```

| Attribute    | Values                    | Description        |
|--------------|---------------------------|--------------------|
| `underline`  | `single`, `double`, `dotted`, `dash`, `wavy` | Underline style |

### Block Tags with Attributes

Block `<w:>` tags also accept attributes for additional formatting:

```
<w:u underline=double>double underline paragraph</w:u>
<w:u underline=dash>dashed underline paragraph</w:u>
```

| Tag                     | Attribute               | Values                    | Description        |
|-------------------------|-------------------------|---------------------------|--------------------|
| `<w:u>`                 | `underline`             | `single`, `double`, `dotted`, `dash`, `wavy` | Underline style |

### Tab Inside W Block Tags

`<tab>`, `<tab/>`, and `<tab size=N>` are allowed inside `<w:*>` tags:

```
<w:c|b>Name:<tab>John Doe</w:c|b>
<w:c>City:<tab size=4>Jakarta</w:c>
<w:b>Phone:<tab/>+62-812-3456-7890</w:b>
```

This enables formatted key-value layouts with tab stops inside centered, bold, or other styled blocks.

## 3. Headings

Heading `<h1>`–`<h9>` with global style in `[style:heading-N]`.

**RESTRICTION:** `<h1>` through `<h9>` MUST contain ONLY plain text and `{{vars}}`. Nested tags (`<b>`, `<i>`, `<u>`, `<code>`, etc.) are STRICTLY FORBIDDEN inside headings.

```
[style]
layout=A4
unit=inch
m=1

[style:heading-1]
font-family="Arial"
font-size=24pt
color=#2b5797
bold=true
space-before=18pt
space-after=12pt
border=1pt

[style:heading-2]
font-family="Arial"
font-size=18pt
color=#444444
bold=true
space-before=12pt
space-after=6pt

[style:heading-3]
font-family="Arial"
font-size=14pt
color=#444444
bold=false
space-before=6pt
space-after=3pt
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
<h1 color=red font-size=28pt>Chapter with local style</h1>
```

### Style Properties

| Property            | Description                                                                 |
|---------------------|-----------------------------------------------------------------------------|
| `font-family`       | Heading font                                                                |
| `font-size`         | Font size (pt)                                                              |
| `color`             | Text color                                                                  |
| `bold`              | `true` / `false`                                                            |
| `italic`            | `true` / `false`                                                            |
| `strike`            | `true` / `false`                                                            |
| `underline`         | `true`, `single`, `double`, `dotted`, `dash`, `wavy`                       |
| `caps`              | `true` / `false` — all capitals                                            |
| `small-caps`        | `true` / `false` — small capitals                                          |
| `letter-spacing`    | Letter spacing (pt)                                                         |
| `align`             | `left`, `center`, `right`, `justify`                                        |
| `indent-left`       | Left indent (in document unit)                                             |
| `indent-right`      | Right indent (in document unit)                                            |
| `indent-hanging`    | Hanging indent (in document unit)                                          |
| `first-line`        | First line extra indent (in document unit)                                 |
| `space-before`      | Space before (layout unit)                                                 |
| `space-after`       | Space after (layout unit)                                                  |
| `line-spacing`      | Line spacing: multiplier (bare number without `line-rule`) or absolute length with `line-rule=exact`/`atLeast` |
| `line-rule`         | Line spacing rule: `auto`, `exact`, `atLeast`                              |
| `border`           | Border on all sides; `border-bottom`/`border-top`/`border-left`/`border-right` for a single side |
| `keep-next`         | `true` / `false` — keep with next paragraph                            |
| `keep-lines`        | `true` / `false` — keep lines together                                 |
| `section-break`     | `true` — force section break before heading                             |
| `widow-control`     | `true` / `false` — widow/orphan control                                |
| `contextual-spacing`| `true` — remove space between same-style paragraphs                    |
| `shading`           | Background color (hex or named)                                        |
| `outline-level`     | Outline level 0–8 (for TOC generation)                                 |
| `dir`               | `ltr` / `rtl`                                                          |
| `valign`            | Vertical: `auto`, `top`, `center`, `baseline`, `bottom`               |
| `tabs`              | Space-separated tab stops in document unit; per-stop `@L\|C\|R\|D:leader` optional (`none`, `dot`, `hyphen`, `underscore`, `middleDot`) |

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
| `border`  | `1`       | Enables table grid borders (flag; not a length) |
| `width`   | `100%`    | Table width¹         |

¹ Not yet implemented (roadmap item).

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

¹ `docx.Cell.ct` is unexported — `GridSpan`/`VMerge`/`CellWidth` cannot be set. Library patch required.

### Named Table Style

```
[style:table header]
bg=#2b5797
color=white
font-weight=bold
align=center
border=1pt

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
  <li>
    <p>item one</p>
  </li>
  <li>
    <p>item two</p>
  </li>
  <li>
    <p>item three</p>
  </li>
</ul>

<ol>
  <li>
    <p>first</p>
  </li>
  <li>
    <p>second</p>
  </li>
  <li>
    <p>third</p>
  </li>
</ol>
```

Nested:

```
<ul>
  <li>
    <p>fruit</p>
    <ul>
      <li>
        <p>apple</p>
      </li>
      <li>
        <p>mango</p>
      </li>
    </ul>
  </li>
  <li>
    <p>vegetable</p>
  </li>
</ul>
```

### Tags

| Tag       | Description                                    |
|-----------|------------------------------------------------|
| `<ol>`    | Ordered list (supports `type=a/A/1/i/I`, `start=N`) |
| `<ul>`    | Unordered list (supports `bullet=circle/square/check/dash`) |
| `<li>`    | List item container — MUST contain `<p>` tags. Does NOT accept attributes. All paragraph properties are set on the inner `<p>` tags. |

### Multi-Paragraph List Items

A `<li>` item MUST contain one or more `<p>...</p>` blocks. The first `<p>` is rendered with the list bullet/number; subsequent `<p>` blocks render as unnumbered continuation paragraphs aligned with the item (using the `ListContinue` style family).

**Container Rule:** `<li>` acts as a container — it MUST contain `<p>` tags. Plain text directly inside `<li>` is NOT allowed.

**Attribute Rule:** `<li>` does NOT accept attributes. All paragraph properties (`indent`, `hanging`, `space-before`, `align`, etc.) MUST be set on the `<p>` tags inside `<li>`.

```
<ul>
  <li>
    <p>Main heading of item</p>
    <p>Detail paragraph — no bullet/number shown.</p>
  </li>
</ul>
```

```
<ol>
  <li>
    <p>first block para</p>
    <p>second block para</p>
  </li>
</ol>
```

- `<li>` MUST wrap content in `<p>` tags — no bare text allowed.
- `<li>` does NOT accept attributes — use `<p>` attributes instead.
- `<p>` accepts all [Paragraph Properties](#paragraph-properties) attributes (`align`, `indent-left`, `color`, `space-before`, etc.).
- `<p>` blocks can span multiple lines.

```
<ul>
  <li>
    <p indent-left=0.5 indent-hanging=0.25>Item with custom indent</p>
    <p space-before=6pt>Continuation paragraph with spacing</p>
  </li>
</ul>
```

Supported `type` values for `<ol>`: `1` (default, numbers), `a` (lowercase letters), `A` (uppercase letters), `i` (lowercase roman), `I` (uppercase roman).

Supported `start` values for `<ol>`: any integer `>= 1` (default `1`). `<ol start=5>` numbers the first item as `5`; with `type=a` the numbering continues from the matching letter (`<ol type=a start=3>` starts at `c`). Only the first level honors `start`; it applies per-`<ol>` block.

### Bullet types

`<ul>` accepts a `bullet=` attribute to choose the bullet glyph (like `<ol type=>`). Defaults to `circle` (the `•` glyph) when `bullet` is omitted.

```
<ul bullet=circle>
<ul bullet=square>
<ul bullet=check>
<ul bullet=dash>
```

| Value    | Glyph |
|----------|-------|
| `circle` | `•` (default) |
| `square` | `▪` |
| `check`  | `✓` |
| `dash`   | `–` |

`bullet=` is also accepted on `<loop:ul>` (see [Loop with Unordered List](#loop-with-unordered-list)).

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

¹ Not yet implemented (roadmap item).

## 6. Loops

Iterate over array data sources declared in `var`.

The data source name must be listed with `[]` prefix in `var=`. See [Var Usage](#var-usage).

### Critical Loop Constraints

1. **Strict Syntax Order:** The iteration action (`x from source`) MUST come BEFORE any attributes (`type=...`). `<loop:ol x from items type=A>` ✅, `<loop:ol type=A x from items>` ❌.
2. **Source Matching:** The array source MUST be listed with a `[]` prefix in `var=`.
3. **Variable Access:** Inside the loop, access fields using the alias (e.g. `{{x.field}}`).
4. **Closing Tag Rule:** The closing tag MUST EXACTLY MATCH the opening variant prefix, but MUST OMIT the action and attributes. Opening: `<loop:ol x from items type=A>` ➔ Closing: `</loop:ol>` (NOT `</loop>` and NOT `</loop:ol type=A>`).
5. **List Loop Prohibition:** NEVER wrap a standard `<loop>` inside static `<ol>` or `<ul>` tags. Use the native `<loop:ol>` or `<loop:ul>` tags instead.
6. **Iteration Item Required:** `<loop:ol>` and `<loop:ul>` MUST wrap each iteration in `<li>...</li>` or `<p>...</p>` inside the loop body. `<p>` is auto-wrapped into `<li>`. Other content is a compile error. `<li>` is a pure container: it MUST contain `<p>` blocks and does NOT accept attributes.
7. **Silent Index:** Use `{index+N}` (single braces) to insert a 1-based counter. Index starts at 0, so `{index+1}` = 1, 2, 3...

```
[section 0]
name=example
var=info, []entries
keys=title

--- BODY ---
<loop x from entries>
  {index+1}. {{x.field}}
</loop>
```

Here `entries` is an array source declared as `[]entries` in `var=info, []entries`.

### Tags

| Tag                                              | Description                            |
|--------------------------------------------------|----------------------------------------|
| `<loop x from name>...</loop>`                   | Iterate array `name`, each item as `x` |
| `<loop:ol x from name type=1>...</loop:ol>`      | Iterate + wrap `<ol>`; each iteration is `<li>` (explicit or auto-wrapped from `<p>`) |
| `<loop:ul x from name>...</loop:ul>`             | Iterate + wrap `<ul>`; each iteration is `<li>` (explicit or auto-wrapped from `<p>`) |
| `<loop:row x from name>...</loop:row>`           | Iterate into table rows                |

> Closing tag MUST EXACTLY MATCH the opening variant: `<loop:ol>` closes with `</loop:ol>` (NOT `</loop>`).

### Basic Loop

```
<loop x from entries>
  <p>{{x.name}} — {{x.value}}</p>
</loop>
```

- `x` — loop variable alias (any name)
- `entries` — must match a name with `[]` prefix in `var=` (e.g. `var=info, []entries`)
- Inside: `{{x.field}}` accesses a field on each array element

### Loop with Ordered List

Each iteration MUST be `<li>...</li>` (explicit) or `<p>...</p>` (auto-wrapped into `<li>`). `<li>` is a pure container — it holds `<p>` blocks and carries NO attributes; all attributes go on the inner `<p>`.

**Explicit `<li>`:**

```
<loop:ol x from items type=A>
  <li>
    <p>{{x.label}}</p>
  </li>
</loop:ol>
```

**Direct `<p>` (auto-wrapped into `<li>`):**

```
<loop:ol x from items type=A>
  <p>{{x.label}}</p>
</loop:ol>
```

Both render as `<ol type=A><li><p>value</p></li><li><p>value</p></li></ol>`. Default `type` is `1` (numeric). Supported: `1`, `a`, `A`, `i`, `I`.

### Loop with Unordered List

Each iteration MUST be `<li>...</li>` (explicit) or `<p>...</p>` (auto-wrapped into `<li>`). Accepts `bullet=` to choose the glyph (`circle` default, `square`, `check`, `dash`).

**Explicit `<li>`:**

```
<loop:ul x from items bullet=square>
  <li>
    <p>{{x.label}}</p>
  </li>
</loop:ul>
```

**Direct `<p>` (auto-wrapped into `<li>`):**

```
<loop:ul x from founders bullet=dash>
  <p>text</p>
  <p>text</p>
</loop:ul>
```

Both render as `<ul><li><p>value</p></li><li><p>value</p></li></ul>`.

### Loop with Multi-Paragraph List Items

Each `<li>` MUST contain one or more `<p>` blocks — the first is numbered/bulleted, the rest are unnumbered continuation paragraphs.

```
<loop:ul x from items>
  <li>
    <p><b>{{x.label}}</b></p>
    <p>{{x.desc}}</p>
  </li>
</loop:ul>
```

A multi-paragraph item requires the explicit `<li>` form (direct `<p>` wraps each paragraph as its own item).

### Loop Index Counter

Use `{index+N}` (single braces) inside any loop variant to insert an auto-incrementing counter. Index starts at **0**, so `{index+1}` produces 1, 2, 3...

| Pattern     | Result                |
|-------------|-----------------------|
| `{index+1}` | 1, 2, 3, 4, ...      |
| `{index+0}` | 0, 1, 2, 3, ...      |
| `{index+10}`| 10, 11, 12, 13, ...  |

Works in `<loop>`, `<loop:ol>`, `<loop:ul>`, and `<loop:row>`:

```
<loop:ol x from items>
  <li>
    <p>{index+1}. {{x.label}}</p>
  </li>
</loop:ol>
```

Renders as `<ol><li><p>1. Item A</p></li><li><p>2. Item B</p></li><li><p>3. Item C</p></li></ol>`.

#### `{{loop.index}}` — Double-Brace Variant

Use `{{loop.index}}` (double braces) inside any loop for a 1-based counter. This variant works with double-brace syntax and takes priority over variable resolution.

| Pattern           | Result                |
|-------------------|-----------------------|
| `{{loop.index}}`  | 1, 2, 3, 4, ...      |

```
<loop x from items>
  <p>Item {{loop.index}}: {{x.name}}</p>
</loop>
```

Multiple indexes in one template:

```
<loop x from entries>
  <p>Entry #{index+1} of {{x.total}}: {{x.name}}</p>
</loop>
```

#### Index & Total Variables

All index/total variables below are available inside any loop variant, both in body text and in condition attributes (see [Conditionals](#7-conditionals)):

| Variable         | Meaning                                    |
|------------------|--------------------------------------------|
| `{{index}}`      | 1-based position (alias of `{{loop.index}}`) |
| `{{loop.index}}` | 1-based position                          |
| `{index}`        | 1-based position (single braces)          |
| `{index+N}`      | Position offset by `N`                    |
| `{{lastIndex}}`  | Last position (total count)               |
| `{{totalIndex}}` | Total count (same value as `{{lastIndex}}`) |
| `{lastIndex}`    | Single-brace form of `{{lastIndex}}`      |
| `{totalIndex}`   | Single-brace form of `{{totalIndex}}`     |

All forms respect `indexType=` when set: `a`/`A` produce lowercase/uppercase letters, `i`/`I` produce lowercase/uppercase roman numerals, `1` (default) produces numbers.

```
<loop:ul x from items>
  <li>
    <p>{{x.label}} ({{index}}/{{totalIndex}})</p>
  </li>
</loop:ul>
```

Renders each item with its position and the total count, e.g. `(1/3)`, `(2/3)`, `(3/3)`.

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
var=info, []items
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

## 7. Conditionals

Conditionally render content based on data values using `<if>`, `<elif>`, and `<else>`.

### Tags

| Tag                                  | Description                                                                  |
|--------------------------------------|------------------------------------------------------------------------------|
| `<if var="expr">...</if>`            | Render content only when `expr` is true. Standalone (simple) form.           |
| `<w-if use=[...]>...</w-if>`         | Wrapper for an if/elif/else chain. MUST list every variable used in `use=[...]`. |
| `<elif var="expr">...</elif>`        | Additional branch inside `<w-if>` (optional, any number).                    |
| `<else>...</else>`                   | Fallback branch inside `<w-if>` (optional, at most one, and MUST be last).   |

### Simple `<if>` (standalone)

```
<if var="{{info.status}} == 'active'">
<p>Visible only when status is active.</p>
</if>
```

Rules:

- Standalone `<if>` renders its content only when the condition is true; otherwise the content is dropped.
- Both `{{path}}` and bare paths are supported: `var="{{info.status}} == 'active'"` or `var="info.status == 'active'"`.
- A bareword resolves to data when present; otherwise it is treated as a **literal string**.
- A `{{path}}` left unresolved after variable resolution means the variable is missing and is treated as **empty** (use with `is empty`).
- `<elif>` and `<else>` are NOT allowed outside a `<w-if>` wrapper.

### `<w-if>` chain

```
<w-if use=[info.role]>
<if var="info.role == 'admin'">
<p>Admin panel</p>
</if>
<elif var="info.role == 'editor'">
<p>Editor workspace</p>
</elif>
<else>
<p>Read-only view</p>
</else>
</w-if>
```

Rules:

- `<w-if>` MUST contain exactly one `<if>` first, then zero or more `<elif>`, then at most one `<else>` (last).
- Branches close themselves: `<if>...</if>`, `<elif>...</elif>`, `<else>...</else>`.
- No other content is allowed between the branch tags inside `<w-if>` (compile error).
- The first branch whose condition is true is rendered; the rest are dropped. If no branch matches and there is no `<else>`, nothing is rendered.
- Nested `<if>`/`<w-if>` blocks are allowed inside any branch body.

### `use=` — declared variables

`use=[...]` MUST list every variable referenced by the branch conditions:

- Root keys and dotted paths: `info`, `info.role`, `founder.name`
- Loop aliases and fields: `x`, `x.done`
- Loop index variables: `index`, `lastIndex`, `totalIndex`

Any reference in a branch `var=` that is not covered by a `use=` entry is a **compile error**. A reference is covered when it equals a `use=` entry or is an ancestor/descendant of one (e.g. `use=[info]` covers `info.role`).

Literals inside `<w-if>` conditions MUST be quoted — barewords are always variables: `var="info.role == 'admin'"`.

### Operators

| Operator       | Meaning                                                          |
|----------------|------------------------------------------------------------------|
| `==`, `!=`     | Equality / inequality                                            |
| `>`, `>=`, `<`, `<=` | Comparison (numeric when both sides are numbers, else string) |
| `is empty`     | Operand is an empty or missing value                             |
| `is not empty` | Operand has a value                                              |
| `and`          | Logical AND (binds tighter than `or`)                            |
| `or`           | Logical OR                                                       |
| `( ... )`      | Grouping                                                         |

Examples:

```
<if var="info.total > 100 and (info.limit == 0 or info.limit >= info.total)">
<if var="{{missing}} is empty">placeholder</if>
<if var="info.name is not empty">{{info.name}}</if>
<if var="count >= 5 and count <= 10">in range</if>
```

### Conditions inside loops

Inside a loop, bare `x.field` references in `var=` and `use=[...]` resolve per item, and the loop index variables are available:

```
<loop:ul x from items>
<li>
<if var="x.done == 'yes'">
<p><b>{{x.label}}</b></p>
</if>
<if var="index == lastIndex">
<p><i>Last item: {{x.label}}</i></p>
</if>
</li>
</loop:ul>
```

- `x.field` — the current item's field
- `index` — current position (1-based, respects `indexType`)
- `lastIndex` / `totalIndex` — last position / total count

The `{{...}}` forms work too (`{{x.field}}`, `{{index}}`, `{{lastIndex}}`, `{{totalIndex}}`). To test a possibly missing field, prefer `{{x.field}} is empty`.

### Evaluation order

Conditions are evaluated after variable resolution (`{{...}}`) and loop expansion, so values are final before any branch is chosen.

## 8. Images

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
<img=./assets/photo.jpg width=3in>
```

### Properties

| Property   | Example        | Description                 |
|------------|----------------|-----------------------------|
| `width`    | `80%`, `3in`   | Width (% of page width, or inches) |
| `height`   | `2in`          | Height (inches)             |
| `align`    | `center`       | `left`, `center`, `right`   |
| `alt`      | "photo"        | Alternative text            |
| `border`   | `1`            | Single border on all sides (flag; value not converted to a length) |
| `bg`  | `#f0f0f0`      | Background container        |

## 9. Links

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

> **Limitation:** Hyperlinks render as blue underlined text but are **not clickable** in the DOCX output. The `godocx v0.1.5` library's `Hyperlink` struct lacks proper OOXML serialization (`<Children>` wrapper instead of raw `<w:r>`), which Word ignores. This affects all `<a=>` usage (inline and standalone). See [`KNOWN-LIMITATIONS.md`](/KNOWN-LIMITATIONS.md) for details.

### Bookmark

```
<a=#chapter1>see Chapter 1</a>
```

## 10. Page & Section Breaks

### Page Break

```
--- BODY ---
<p>page 1</p>
<pb>
<p>page 2</p>
```

| Tag              | Description       |
|------------------|-------------------|
| `<pb>`           | Page break        |
| `<page-break>`   | Alias for `<pb>`  |
| `<tab>`          | Tab character     |
| `<tab size=N>`   | Tab with N spaces |

`<tab>` can appear inside `<p>` paragraphs to insert a tab stop. The optional `size=N` attribute sets
an explicit number of spaces (defaults to one standard tab).

```
<p>Name:<tab>John Doe</p>
<p>Age:<tab size=4>25</p>
```

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

## 11. Metadata

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
font-size=8pt
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
- Properties are visible in document properties dialog (DOCX)
- The `{{title}}` variable only works in headers/footers and body content
- Use `{{date}}` for current date, `{{page}}` for page numbers

## 12. Header & Footer

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
| `justify_between` | 2 or 3 comma-separated items spread evenly via tab stops. Use `\,` for literal comma |
| `font-family` | Header/footer font override            |
| `font-size`   | Font size (pt suffix required, e.g. `9pt`) |
| `color`  | Text color                             |
| `border`      | `top`, `bottom`, `none`                |
| `margin`      | Distance from header/footer to content |
| `first-page`  | `true` / `false` — show on page 1     |
| `mirror`      | `true` / `false` — swap left↔right    |

### justify_between

Replaces `left`/`center`/`right` with evenly-spaced columns using OOXML tab stops.

```
[header]
justify_between={{title}}, {{page}} / {{total}}

[footer]
justify_between=Dept. A\, B\, and C, {{date}}, Page {{page}}
```

| Items | Behavior |
|---|---|
| 2 items | Left-aligned + right-aligned |
| 3 items | Left + center + right |

**Comma escaping:** Use `\,` for a literal comma inside a column value (e.g. `Dept. A\, B\, and C`).

Works with all header/footer variables and font styling properties.

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
font-size=10pt
color=#999999
border=bottom
margin=0.3

[footer]
center={{date}}
font-size=9pt
color=#666666
border=top
margin=0.2
first-page=false
```

### justify_between Example

```
[style]
layout=A4
unit=inch
m=1

[header]
justify_between={{title}}, {{page}} / {{total}}
font-size=10pt
color=#999999
border=bottom
margin=0.3

[footer]
justify_between=Dept. A\, B\, and C, {{date}}, Page {{page}}
font-size=9pt
color=#666666
border=top
margin=0.2
```

## See Also

- `dcd-cli` — CLI usage and options
- `golang-programming` — Go library API
- `dcd-guide` — Project overview and patterns
