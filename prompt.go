package dcdmaker

import (
	_ "embed"
	"fmt"
	"strings"
)

//go:embed .agents/skills/dcd-documents/SKILL.md
var dcdSpec string

//go:embed .agents/skills/docx-preprosesor/SKILL.md
var docxPreprocessorSpec string

func buildPrompt(userPrompt string, predictableKeys []KeyDef) string {
	var b strings.Builder

	b.WriteString("Output ONLY raw DCD template syntax, no explanations, no markdown wrapping.\n")
	b.WriteString("CRITICAL: Do NOT include, repeat, or echo back any part of the SOURCE DOCUMENT (<words> XML) below.\n")
	b.WriteString("The source document is input only — never copy it into your response.\n\n")
	b.WriteString("COMPLETENESS: Translate the source document fully. Classify every piece of source text as STATIC or DYNAMIC:\n")
	b.WriteString("- STATIC text (paragraphs, headings, labels, fixed wording) MUST appear in the DCD template verbatim, exactly once, in the original order.\n")
	b.WriteString("- DYNAMIC content (data values, names, dates, amounts — anything that will be filled by a variable) MUST be replaced by its {{var}} / {{var.field}} placeholder. Never write the source value literally for dynamic content; never drop it either.\n")
	b.WriteString("Never drop, skip, omit, merge, truncate, reorder, or summarize any source text. The full document must be representable from your template.\n\n")
	b.WriteString("INLINE ATTRIBUTES: Preserve inline tag attributes verbatim when the DCD equivalent tag supports them. Examples: <u underline=\"double\"> → <set:u underline=double>; <span highlight=\"yellow\"> → <mark color=yellow>; <a href=...> keeps target/color/underline. Do NOT strip, rename, or flatten an inline attribute into a [style] block or a paragraph attribute when the inline tag accepts it. Only fall back when DCD has no inline equivalent.\n\n")
	b.WriteString("PRESERVE FONT-FAMILY: Never miss a font-family declaration that exists in the source. Carry it over where the source declares it — as the [style] / [style:heading-N] default or on the paragraph or inline tag that carries it. Never invent a font-family declaration where the source has none, and never wrap individual words in <span> just to carry it.\n")
	b.WriteString("PRESERVE <span> AND <tab>: Every <span> (with all its attributes) and every <tab>, <tab/>, or <tab size=N> that EXISTS in the source MUST appear in the output — never drop, flatten, or merge spans that carry different formatting. Conversely, NEVER invent <span> tags: wrap exactly the same text range the source wraps — never individual words — and never add a span where the source has none.\n")
	b.WriteString("PRESERVE tabs ATTRIBUTE: When the source document defines a tabs attribute on a paragraph, the DCD output MUST include the corresponding tabs attribute. Do not omit it.\n")
	b.WriteString("STYLE BLOCK (REQUIRED): The DCD output MUST always include a [style] block. [style] is NOT optional — it defines the page geometry (layout, unit, w/h, margins), font, and line spacing. Map every value from the source document's <style> block (s:page, s:theme, s:line, s:indent, etc.) to its DCD equivalent. [style:heading-N] blocks are OPTIONAL — emit them only when the source defines heading styles.\n")
	b.WriteString("UNPREDICTABLE BLOCKS: During analysis you MAY discover objects, arrays, or keys that the document references but that are NOT listed in PREDICTED VARIABLES. Classify those as unpredictable and declare them in the dedicated blocks: objects and arrays → [object-unpredictable]; standalone keys → [keys-unpredictable]. These blocks only extend the schema — every object/array must STILL be registered in var= and every standalone key in keys= of its owning section. Predicted variables (listed in PREDICTED VARIABLES) must NEVER be re-declared in unpredictable blocks.\n\n")

	b.WriteString("=== DCD DSL SPECIFICATION ===\n")
	b.WriteString(dcdSpec)
	b.WriteString("\n\n")

	b.WriteString("=== SOURCE DOCUMENT FORMAT ===\n")
	b.WriteString(docxPreprocessorSpec)
	b.WriteString("\n\n")

	if len(predictableKeys) > 0 {
		b.WriteString("=== PREDICTED VARIABLES ===\n\n")
		for _, k := range predictableKeys {
			switch k.Type {
			case VarArray:
				if k.FieldDefs != nil {
					fmt.Fprintf(&b, "  []%s {%s} (array)\n", k.Name, renderFieldDefs(k.FieldDefs))
				} else {
					fmt.Fprintf(&b, "  []%s {%s} (array)\n", k.Name, strings.Join(k.Fields, ", "))
				}
			case VarObject:
				if k.FieldDefs != nil {
					fmt.Fprintf(&b, "  %s {%s}\n", k.Name, renderFieldDefs(k.FieldDefs))
				} else {
					fmt.Fprintf(&b, "  %s {%s}\n", k.Name, strings.Join(k.Fields, ", "))
				}
			case VarKeys:
				if k.FieldDefs != nil {
					fmt.Fprintf(&b, "  %s (keys)\n", renderFieldDefs(k.FieldDefs))
				} else {
					fmt.Fprintf(&b, "  %s (keys)\n", strings.Join(k.Fields, ", "))
				}
			}
		}
		b.WriteString("\n")
	}

	if len(predictableKeys) > 0 {
		b.WriteString("=== FORBIDDEN IN UNPREDICTABLE ===\n")
		b.WriteString("These predicted variables must NOT appear in [object-unpredictable] or [keys-unpredictable]:\n")
		for _, k := range predictableKeys {
			switch k.Type {
			case VarObject, VarArray:
				b.WriteString(fmt.Sprintf("  • %s", k.Name))
				if k.FieldDefs != nil {
					for _, f := range k.FieldDefs {
						fmt.Fprintf(&b, ", %s", f.Name)
					}
				} else if len(k.Fields) > 0 {
					fmt.Fprintf(&b, ", %s", strings.Join(k.Fields, ", "))
				}
				b.WriteString("\n")
			case VarKeys:
				for _, f := range k.FieldDefs {
					fmt.Fprintf(&b, "  • %s\n", f.Name)
				}
				for _, f := range k.Fields {
					fmt.Fprintf(&b, "  • %s\n", f)
				}
			}
		}
		b.WriteString("\n")
	}

	if strings.TrimSpace(userPrompt) != "" {
		b.WriteString("=== USER INSTRUCTION ===\n")
		b.WriteString(userPrompt)
		b.WriteString("\n\n")
	}

	b.WriteString("Generate the DCD template now:")

	return b.String()
}

func renderFieldDefs(fields []FieldDef) string {
	parts := make([]string, len(fields))
	for i, f := range fields {
		if f.Format != "" {
			parts[i] = fmt.Sprintf("%s: %s (%s)", f.Name, f.Type, f.Format)
		} else {
			parts[i] = fmt.Sprintf("%s: %s", f.Name, f.Type)
		}
	}
	return strings.Join(parts, ", ")
}
