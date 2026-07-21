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
