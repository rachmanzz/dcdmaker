# DOCX Preprocessor

Converts raw DOCX XML into simplified structured text for LLM consumption.

## Source Document Format

The source document is **PRE-PROCESSED structured text**, NOT raw XML.

Do NOT look for XML tags like `<w:p>`, `<w:r>`, `<w:t>`, `<w:pgMar>`, etc.

The DOCX file has been parsed and simplified into the following format:

| Tag | Description | Example |
|---|---|---|
| `[PAGE]` | Page layout (width, height, margins in inches) | `[PAGE] 8.27x11.69 in, margins: top=0.71 right=0.24 bottom=0.87 left=2.44 in` |
| `[DEFAULT]` | Default font-family, font-size, color | `[DEFAULT] font-family=Times New Roman font-size=11pt` |
| `[P]` | Paragraph with optional attributes | `[P align=center font-family=Courier New]AKTA PENDIRIAN...` |
| `[LI]` | List item | `[LI]Perseroan didirikan untuk jangka waktu...` |
| `<h1>`-`<h6>` | Heading levels (h1 is most important) | `<h1>judul</h1>` |
| `<b>text</b>` | Bold text | `<b>nama Perseroan</b>` |
| `<i>text</i>` | Italic text | `<i>catatan</i>` |

### Paragraph Attributes

| Attribute | Description | Example |
|---|---|---|
| `align` | Text alignment: left, center, right, both/justify | `align=center` |
| `indent` | Left indentation in inches | `indent=0.32` |
| `hanging` | Hanging indent in inches | `hanging=0.32` |
| `font-family` | Font family name (only if different from DEFAULT) | `font-family=Courier New` |
| `font-size` | Font size in pt (only if different from DEFAULT) | `font-size=12pt` |
| `color` | Text color hex (only if different from DEFAULT) | `color=#FF0000` |

### Example

```
[PAGE] 8.27x11.69 in, margins: top=0.71 right=0.24 bottom=0.87 left=2.44 in
[DEFAULT] font-family=Times New Roman font-size=11pt
[P align=center font-family=Courier New]AKTA PENDIRIAN PERSEROAN TERBATAS
[P align=center font-family=Courier New]PT. ....
[P align=both indent=0.32 hanging=0.32 font-family=Courier New]1. .... .........., lahir di ....
[LI]Perseroan didirikan untuk jangka waktu tidak terbatas.
```

**Key Point:** Extract information from these structured tags, NOT from XML.
