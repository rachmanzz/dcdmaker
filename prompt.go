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
	b.WriteString("Generate a DCD template that is 99% identical to the source document.\n")
	b.WriteString("DO NOT use default values. DO NOT assume anything.\n")
	b.WriteString("Extract ALL values directly from the SOURCE DOCUMENT XML below.\n\n")

	b.WriteString("Template structure:\n")
	b.WriteString("1. [style] — extract exact layout, margins, font-family, font-size, line-height from source XML. Copy values directly — no guessing.\n")
	b.WriteString("2. [title] metadata\n")
	b.WriteString("3. [header] — ONLY if the source document has headers (see note at the end)\n")
	b.WriteString("4. [footer] — ONLY if the source document has footers (see note at the end)\n")
	b.WriteString("5. [section] definitions with var, keys, formats\n")
	b.WriteString("6. --- BODY --- with exact DCD tags matching source structure\n")
	b.WriteString("7. [object-unpredictable] for additional object/array fields found in document\n")
	b.WriteString("8. [keys-unpredictable] for additional simple key mappings found in document\n\n")

	b.WriteString("Rules:\n")
	b.WriteString("- Parse the SOURCE DOCUMENT XML for actual margins (<w:pgMar>), fonts (<w:rFonts>, <w:sz>), headings (<w:pStyle>, <w:b/> combined with size), tables (<w:tbl>), lists (<w:numPr>).\n")
	b.WriteString("- Match font-family, font-size, margins, page layout, heading hierarchy EXACTLY.\n")
	b.WriteString("- Every <w:t> text must be preserved in correct order.\n")
	b.WriteString("- Bold/italic/underline from <w:b/>, <w:i/>, <w:u/> must map to <b>, <i>, <u> in DCD.\n")
	b.WriteString("- Paragraph alignment from <w:jc> must map to <w:c> in DCD.\n")
	b.WriteString("- Table structure (<w:tbl>) must map to <table>/<row>/<col>.\n")
	b.WriteString("- Lists (<w:numPr>) must map to <ol>/<ul>/<li>.\n")
	b.WriteString("- Variable names should be inferred from document context (e.g. invoice number → invoice_no, date → date).\n")
	b.WriteString("- Use [section:next-page] only where source has explicit section/page breaks.\n")
	b.WriteString("- Output ONLY raw DCD syntax, no markdown fences, no extra text.\n\n")

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
