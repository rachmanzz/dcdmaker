package dcdmaker

import (
	_ "embed"
	"fmt"
	"strings"
)

//go:embed .agents/skills/dcd-documents/SKILL.md
var dcdSpec string

func buildPrompt(userPrompt string, predictableKeys []KeyDef) string {
	var b strings.Builder

	b.WriteString("You are a DCD template generator. " +
		"Generate ONLY valid DCD template syntax, no explanations, no markdown wrapping.\n\n")

	b.WriteString("=== DCD DSL SPECIFICATION ===\n")
	b.WriteString(dcdSpec)
	b.WriteString("\n\n")

	if len(predictableKeys) > 0 {
		b.WriteString("=== PREDICTED VARIABLES ===\n")
		b.WriteString("Use these exact variable names and fields:\n\n")
		for _, k := range predictableKeys {
			switch k.Type {
			case VarArray:
				fmt.Fprintf(&b, "  %s []%s (array — for <loop>)\n", k.Name, strings.Join(k.Fields, ", "))
			case VarObject:
				fmt.Fprintf(&b, "  %s {%s}\n", k.Name, strings.Join(k.Fields, ", "))
			}
		}
		b.WriteString("\n")
	}

	b.WriteString("=== INSTRUCTION ===\n")
	b.WriteString("Generate a DCD template that can render documents like the source. ")
	b.WriteString("Use appropriate [style], sections, variables, and loops. ")
	b.WriteString("Use [section:next-page] for logical page breaks. ")
	b.WriteString("Output ONLY raw DCD syntax, no markdown fences, no extra text.\n\n")

	b.WriteString("Template structure:\n")
	b.WriteString("1. [style] layout and font settings\n")
	b.WriteString("2. [title] metadata\n")
	b.WriteString("3. [header] and [footer]\n")
	b.WriteString("4. [section] definitions with var, keys, formats\n")
	b.WriteString("5. --- BODY --- with DCD tags\n")
	b.WriteString("6. [object-unpredictable] for additional object/array fields found in document\n")
	b.WriteString("7. [keys-unpredictable] for additional simple key mappings found in document\n\n")

	b.WriteString("=== UNPREDICTABLE VARIABLES ===\n")
	b.WriteString("After the main template sections, include:\n\n")
	b.WriteString("[object-unpredictable]\n")
	b.WriteString("- varName=[]field1, field2     ← array of objects\n")
	b.WriteString("- varName=field1, field2       ← single object\n\n")
	b.WriteString("[keys-unpredictable]\n")
	b.WriteString("- varName=field1, field2       ← simple key mappings\n\n")
	b.WriteString("These are for additional fields found in the document ")
	b.WriteString("that are not listed in PREDICTED VARIABLES above.\n\n")

	if strings.TrimSpace(userPrompt) != "" {
		b.WriteString("=== USER INSTRUCTION ===\n")
		b.WriteString(userPrompt)
		b.WriteString("\n\n")
	}

	b.WriteString("Generate the DCD template now:")

	return b.String()
}
