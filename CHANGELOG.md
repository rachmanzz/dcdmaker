# Changelog

## v0.1.12 (2026-06-29)

### Added
- **maker.go**: `VarKeys`, `FieldDef`, `Field()`, `Keys()`, `ObjectEx()`, `ArrayEx()` — typed variable definitions with type/format support
- **prompt.go**: Render type and format in PREDICTED VARIABLES section when `FieldDefs` are set; new `VarKeys` rendering (flat keys)

### Changed
- **README.md**: Updated PredictableKeys section with `Keys()`, `ObjectEx()`, `ArrayEx()`, `Field()` docs

## v0.1.11 (2026-06-29)

### Added
- **unpredictable.go**: `UnpredictableObject` struct, `UnpredictableObjects()` and `UnpredictableKeys()` methods on `*Maker` to parse `[object-unpredictable]` and `[keys-unpredictable]` sections from AI-generated DCD output
- **maker.go**: `LastResult() string` to retrieve raw DCD output, `UnpredictableObjects()`, `UnpredictableKeys()` methods

### Fixed
- **prompt.go**: `[keys-unpredictable]` format corrected — no `varName=` prefix (flat key mappings)

## v0.1.10 (2026-06-26)

### Added
- **SKILL.md**: `size`/`font-size` and `color` attributes on `<w:*>`, `<p>`, `<li>`, `<col>` tags with examples
- **SKILL.md**: Completeness notice at top — "do not assume syntax features beyond what is documented"
- **prompt.go**: `=== CRITICAL: NO DCD SYNTAX ASSUMPTIONS ===` block — forbids inventing tags/attributes not in SKILL.md

### Changed
- **SKILL.md**: `should be` → `MUST be` (data fields in keys+var), `should describe` → `must describe` (section name)
- **prompt.go**: `should have 1-3 var` → `MUST have maximum 3 var`
- **prompt.go**: Variable naming rule split — predictable (exact names only) vs unpredictable (strong assumptions allowed)
- **SKILL.md**: 417 → 422 lines

## v0.1.9 (2026-06-26)

### Added
- **SKILL.md**: Section Splitting Guidelines — rules for splitting by context/topic, max 1-3 var and 15 keys per section, with multi-section example
- **SKILL.md**: `<w:r>`, `<w:j>`, `<w:l>` alignment tags — right, justify, left wrapper paragraphs
- **SKILL.md**: `<p align=center/right/justify>` rich paragraph tags for mixed formatting
- **SKILL.md**: Tag Nesting Rules section — comprehensive allowed/forbidden nesting rules with examples
- **SKILL.md**: Wrapper vs Rich paragraph distinction — `<w:*>` for pure text, `<p align=*>` for mixed formatting
- **prompt.go**: Wrapper paragraph rules — pure text only, no inline tags allowed
- **prompt.go**: Rich paragraph rules — can contain inline tags for mixed formatting
- **prompt.go**: Standalone tag restrictions — `<pb>`, `<hr>` must not be nested, split paragraphs if mid-text
- **prompt.go**: Section splitting rules — by context, 1-3 var max, under 15 keys

### Fixed
- Invalid nesting: AI no longer generates inline tags inside wrapper tags (e.g., `<w:c>Text <u>underline</u></w:c>`)
- Page break placement: AI now correctly places `<pb>` standalone, not inside paragraphs
- Gendut sections: AI now splits sections by context instead of cramming 8+ vars and 60+ keys into single section

### Changed
- **SKILL.md**: 292 → 417 lines (+125 lines for nesting rules and section splitting)

## v0.1.8 (2026-06-26)

### Changed
- **SKILL.md**: Optimized from 934 → 292 lines (69% reduction, 66% token savings)
- **SKILL.md**: Fixed 18 audit issues — removed ambiguous syntax, unverified claims, added missing examples
- **README.md**: Replaced Indonesian examples with English

### Removed
- Unverified `:` separator claim (only `=` is supported)
- Redundant examples and verbose subsections (46 → 5 subsections)

## v0.1.7 (2026-06-26)

### Added
- `LastProvider() string` — returns the provider and model name (e.g. `gemini:gemini-2.5-flash`) that successfully generated the DCD template.

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
