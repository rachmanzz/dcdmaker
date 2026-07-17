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

	b.WriteString("=== DCD DSL SPECIFICATION ===\n")
	b.WriteString(dcdSpec)
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

	b.WriteString("=== INSTRUCTION ===\n")
	b.WriteString("DO NOT use default values. DO NOT assume anything.\n")
	b.WriteString("Extract ALL values directly from the SOURCE DOCUMENT below.\n\n")
	b.WriteString("=== SOURCE DOCUMENT FORMAT ===\n")
	b.WriteString(docxPreprocessorSpec)
	b.WriteString("\n\n")
	b.WriteString("You MUST scan the ENTIRE source document from first paragraph to last paragraph. Do NOT stop early, skip sections, or summarize.\n")
	b.WriteString("Every single paragraph, table, and list in the source MUST have a corresponding DCD entry in the output.\n\n")

	b.WriteString("=== CRITICAL: NO DCD SYNTAX ASSUMPTIONS ===\n")
	b.WriteString("The DCD DSL SPECIFICATION above is the COMPLETE and AUTHORITATIVE reference.\n")
	b.WriteString("Do NOT invent, extrapolate, or assume any DCD syntax features beyond what is explicitly listed there.\n")
	b.WriteString("Every tag, attribute, and syntax pattern you use MUST be directly from the specification.\n")
	b.WriteString("If a DCD feature is not documented in the specification, it does not exist — do not use it.\n\n")

	b.WriteString("Task:\n")
	b.WriteString("- Use the source document's `<style>` block (from `<words>` XML) as the reference for page layout, margins, fonts, and line-height.\n")
	b.WriteString("- Parse the `<write>` content for paragraphs (`<p>`), headings (`<h1>`-`<h9>`), lists (`<ul>`/`<ol>`/`<li>`), alignment, and formatting.\n")
	b.WriteString("- Match font-family, font-size, heading hierarchy EXACTLY from the source `<style>` block.\n")
	b.WriteString("- Every text line must be preserved in correct order.\n")
	b.WriteString("- CRITICAL: Do NOT skip transitional clauses, introductory phrases, exception clauses, or connecting text between sections. Pay EXTRA attention to the FINAL paragraphs. Every sentence must map to DCD.\n")
	b.WriteString("- CRITICAL: If source contains placeholder dots (e.g. ............) used as fill-in fields, replace them with `{{var.field}}`. Infer field name from context. Do NOT copy placeholder dots literally. This is the default — only keep literal dots if user provides explicit instruction via `-prompt`.\n")
	b.WriteString("- CRITICAL: Do NOT copy all predicted variables into every section. Distribute them based on actual usage per section.\n")
	b.WriteString("- CRITICAL: If source contains multiple distinct entries needing different formatting, use separate source arrays or separate sections.\n")
	b.WriteString("- Organize content into `[section N]` blocks with proper var/keys declarations. Use `[object-unpredictable]` and `[keys-unpredictable]` only for fields not in the predictable data.\n\n")

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
