# dcdmaker

AI-powered DCD template generator. Parses source documents (DOCX) and generates reusable `.dcd` templates using Gemini, OpenAI, or any OpenAI-compatible API.

## Installation

```bash
go get github.com/rachmanzz/dcdmaker
```

## CLI Usage

```bash
# Generate template from invoice
dcdmaker -source invoice.docx -output templates/invoice.dcd \
    -prompt "Buat template invoice dengan nomor, tanggal, customer, items, total"

# Custom Gemini model
dcdmaker -source contract.docx -output contract.dcd \
    -gemini-model gemini-2.5-pro-exp-03-25

# OpenAI-compatible (Ollama, vLLM, LocalAI, Groq, etc.)
dcdmaker -source doc.docx -output doc.dcd \
    -openai-model llama3 \
    -openai-base-url http://localhost:11434/v1 \
    -openai-key ollama \
    -no-gemini

# Multi-provider fallback (Gemini → OpenAI)
dcdmaker -source invoice.docx -output invoice.dcd \
    -gemini-model gemini-2.5-flash \
    -openai-model gpt-4o

# Resume interrupted session
dcdmaker -source invoice.docx -output invoice.dcd -resume

# Install CLI
go install github.com/rachmanzz/dcdmaker/cmd/dcdmaker@latest
```

## Library Usage

```go
package main

import "github.com/rachmanzz/dcdmaker"

func main() {
    maker := dcdmaker.NewMaker(
        dcdmaker.Gemini(
            dcdmaker.WithModel("gemini-2.5-flash"),
            dcdmaker.WithTemperature(0.3),
        ),
        dcdmaker.Gemini(
            dcdmaker.WithModel("gemini-2.5-pro-exp-03-25"),
        ),
        dcdmaker.OpenAI(
            dcdmaker.WithOpenAIModel("gpt-4o"),
        ),
    )

    // Write to file (supports resume)
    maker.
        Source("invoice.docx").
        OptionalPrompt("Buat template invoice dengan nomor, tanggal, customer, items, total").
        Resume(true).
        Run("templates/invoice.dcd")

    // Or get raw string (no file write, no resume)
    dcd, err := maker.Generate()
}
```

### PredictableKeys

Control variable names and structure — AI forced to use exact keys:

```go
maker.
    Source("invoice.docx").
    PredictableKeys(
        dcdmaker.Object("info", "invoice_no", "date", "customer", "due_date"),
        dcdmaker.Array("items", "name", "qty", "unit_price", "total"),
        dcdmaker.Object("summary", "subtotal", "tax", "grand_total"),
    ).
    Run("templates/invoice.dcd")
```

- `Object(name, fields...)` — singleton object, accessed as `{{name.field}}`
- `Array(name, fields...)` — array of objects for `<loop>`, accessed as `{{x.field}}`

Additional fields found in the document are written to `[object-unpredictable]` and `[keys-unpredictable]` sections.

### Fallback by model

```go
maker := dcdmaker.NewMaker(
    // Gemini Pro → retry 3x → Gemini Flash → retry 3x → OpenAI
    dcdmaker.Gemini(dcdmaker.WithModel("gemini-2.5-pro-exp-03-25")),
    dcdmaker.Gemini(dcdmaker.WithModel("gemini-2.5-flash")),
    dcdmaker.OpenAI(dcdmaker.WithOpenAIModel("gpt-4o")),
)
```

## Configuration

### Gemini

| Option | Default | Description |
|--------|---------|-------------|
| `WithModel` | `gemini-2.5-flash` | Model name |
| `WithTemperature` | `0.5` | Temperature (0–1) |
| `WithAPIKey` | `$GEMINI_API_KEY` | API key |
| `WithTimeout` | `60s` | Request timeout |

### OpenAI

| Option | Default | Description |
|--------|---------|-------------|
| `WithOpenAIModel` | `gpt-4o` | Model name |
| `WithOpenAIBaseURL` | — | Custom base URL (for Ollama, vLLM, etc.) |
| `WithOpenAITemperature` | `0.5` | Temperature (0–1) |
| `WithOpenAIMaxTokens` | `8192` | Max output tokens |
| `WithOpenAIAPIKey` | `$OPENAI_API_KEY` | API key |
| `WithOpenAITimeout` | `60s` | Request timeout |

## How It Works

```
DOCX ──> dcdmaker ──> AI (Gemini / OpenAI) ──> .dcd template
```

1. **Read** — Source document loaded from disk
2. **Send** — Raw DOCX sent as inline data (Gemini) or text-extracted (OpenAI)
3. **Generate** — AI produces DCD template based on document structure
4. **Validate** — Output checked for balanced tags, valid DCD syntax
5. **Retry** — 3 attempts per provider, fallback to next provider
6. **Resume** — Session state persisted for recovery on truncation

## Output: `.dcd` Template

```ini
[style]
layout=A4
unit=inch
m=1

[title]
title=Invoice

[header]
right={{page}} / {{total}}

[section 0]
name=header
var=info
keys=invoice_no, date, customer

--- BODY ---
<h1>{{info.invoice_no}}</h1>
<p>Date: {{info.date}}</p>
<p>Customer: {{info.customer}}</p>

[section 1]
name=items
var=info, items
keys=title, items.name, items.qty, items.price

--- BODY ---
<table border=1 width=100%>
<loop:row x from items>
  <col>{{x.name}}</col>
  <col align=right>{{x.qty}}</col>
  <col align=right>{{x.price}}</col>
</loop:row>
</table>

[object-unpredictable]
- info=discount, shipping_address

[keys-unpredictable]
- data=po_number, department
```

## Development

```bash
go test ./...
go vet ./...
go build ./...
```
