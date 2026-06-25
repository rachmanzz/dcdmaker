package dcdmaker

import (
	_ "embed"
	"strings"
)

//go:embed .agents/skills/dcd-documents/SKILL.md
var dcdSpec string

func buildPrompt(userPrompt string) string {
	var b strings.Builder

	b.WriteString("You are a DCD template generator. " +
		"Generate ONLY valid DCD template syntax, no explanations, no markdown wrapping.\n\n")

	b.WriteString("=== DCD DSL SPECIFICATION ===\n")
	b.WriteString(dcdSpec)
	b.WriteString("\n\n")

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
	b.WriteString("5. --- BODY --- with DCD tags\n\n")

	if strings.TrimSpace(userPrompt) != "" {
		b.WriteString("=== USER INSTRUCTION ===\n")
		b.WriteString(userPrompt)
		b.WriteString("\n\n")
	}

	b.WriteString("Generate the DCD template now:")

	return b.String()
}
