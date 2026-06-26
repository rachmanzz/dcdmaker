# Changelog

## v0.1.6 (2026-06-26)

### Changed
- **DOCX extraction**: replaced `extractDocxText` (plain text only) with `extractDocxContent` — sends raw `word/document.xml` to AI instead of stripped text. AI now has full access to margins, fonts, headings, tables, and lists from the source XML.
- **Prompt instructions**: rewritten in English with strict rules — "99% identical", "DO NOT assume", "Extract ALL values directly from SOURCE DOCUMENT XML". Header/footer generation is now conditional based on actual presence in the source.
- **Retry delay**: progressive backoff (5s → 10s → 15s) between attempts. Resets to 5s when switching providers.

### Removed
- `task/` directory (stale planning files)
- `extractDocxText` function (replaced by `extractDocxContent` in `docx.go`)

### Added
- `docx.go` — new `extractDocxContent()` function that parses DOCX ZIP and extracts raw `word/document.xml` plus header/footer presence detection.
- `LastProvider() string` — returns the provider and model name (e.g. `gemini:gemini-2.5-flash`) that successfully generated the DCD template.
