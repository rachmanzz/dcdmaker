# Decision Logs

Records of deliberate architectural decisions vs accidental leftovers.

## Dead Code — Removed

| Item | Action | Date |
|---|---|---|
| `extractDocxContent()` + `DocxContent` struct | Removed | 2026-07-16 |
| `Provider.Generate()` | Removed from interface + implementations | 2026-07-16 |
| `Maker.LastProvider()` + `Maker.lastProvider` | Removed | 2026-07-16 |
| `ParsedParagraph.ListType` | Removed field + setter | 2026-07-16 |
| `ParsedParagraph.StyleID` | Refactored to local variable | 2026-07-16 |
| `sectionInfo.Body` / `sectionInfo.raw` | Removed | 2026-07-16 |
| `WithOpenAITemperature()`, `WithOpenAIMaxTokens()`, `WithOpenAITimeout()` | Removed | 2026-07-16 |

## Dead Code — Deliberately Kept

| Item | Reason | Date | Reference |
|---|---|---|---|---|
| `Provider.GenerateWithHistory()` | Reserved for potential future streaming/retry use | 2026-07-10 | Commit `31605c0` |
| `Maker.LastResult()` | Public API — consumers may read raw DCD output | 2026-06-29 | CHANGELOG v0.1.11 |
| `Maker.UnpredictableObjects()` / `Maker.UnpredictableKeys()` | Public API — consumers may parse unpredictable sections from output | 2026-06-29 | CHANGELOG v0.1.11 |
| `validateVarsAndKeys` + `parseSections`/`scanBody` | Unit-test utility — validates AI-generated var/key declarations | 2026-06-XX | `dcdmaker_test.go` |

## Design Decisions

### Deterministic `[style]` Generation

**Date:** 2026-07-16
**Decision:** Code generates `[style]` section deterministically from `ParsedDocument`; LLM sees it as reference but output is overridden.
**Reason:** Prevent LLM hallucination of layout/margin/font values. LLM focuses on body content only.
**Files:** `docx.go:GenerateStyleBlock()`, `maker.go` (prepend logic)

### Opsi A: Reference Style + Code Override

**Date:** 2026-07-16
**Decision:** `FormatForLLM()` outputs `=== DOCUMENT STYLE ===` with full `[style]` block for AI reference, then `=== DOCUMENT CONTENT ===` with body.
AI sees exact layout/margins/fonts → body alignment/indent/font overrides accurate.
Code strips any AI-generated `[style]` and prepends deterministic version.
**Files:** `docx.go:FormatForLLM()`, `generator.go:stripStyleBlock()`, `maker.go`

### XML Cleaner — Structured DOCX Parser

**Date:** 2026-07-16
**Decision:** Replace raw DOCX XML injection with structured `ParseDOCX()` parser. Extract layout, margins, fonts, headings, alignment, indentation, bold/italic into `ParsedDocument`. Output cleaned `FormatForLLM()` text instead of 1.4MB raw XML.
**Reason:** Raw XML exceeded model context window (367K tokens vs 262K limit). Parser reduces to ~50-60K tokens.
**Files:** `docx.go:ParseDOCX()`, `docx.go:FormatForLLM()`

### Resume Feature Removed

**Date:** 2026-07-10
**Decision:** Remove `Resume(true)`, session files, `-resume` CLI flag, chunk continuation (`resolveChunks`, `<TRUNCATED/>`), `saveSession`/`loadSession` entirely.
**Reason:** Added complexity with marginal benefit — retry+fallback between providers sufficient. `GenerateWithHistory` kept in interface for potential future use.
**Reference:** Commit `31605c0`

### CLI Removed — Library Only

**Date:** 2026-07-16
**Decision:** `cmd/dcdmaker/main.go` deleted. Project is library-only.
**Reason:** `dcdmaker` is a Go library (`import "github.com/rachmanzz/dcdmaker"`). CLI wrapper was unnecessary — users embed the library directly in their applications.
**Replaces:** Previous CLI-as-wrapper approach.

### Provider Env Var Policy

**Date:** 2026-06-XX
**Decision:** Library (`dcdmaker`) does NOT read environment variables. Callers must pass `APIKey` explicitly via functional options (`WithAPIKey`, `WithOpenAIAPIKey`).
**Reason:** Library should be environment-agnostic.
