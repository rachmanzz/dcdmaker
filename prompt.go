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
				if k.FieldDefs != nil {
					b.WriteString(fmt.Sprintf("  %s []%s (array — for <loop>)\n", k.Name, renderFieldDefs(k.FieldDefs)))
				} else {
					b.WriteString(fmt.Sprintf("  %s []%s (array — for <loop>)\n", k.Name, strings.Join(k.Fields, ", ")))
				}
			case VarObject:
				if k.FieldDefs != nil {
					b.WriteString(fmt.Sprintf("  %s {%s}\n", k.Name, renderFieldDefs(k.FieldDefs)))
				} else {
					b.WriteString(fmt.Sprintf("  %s {%s}\n", k.Name, strings.Join(k.Fields, ", ")))
				}
			case VarKeys:
				if k.FieldDefs != nil {
					b.WriteString(fmt.Sprintf("  %s (keys)\n", renderFieldDefs(k.FieldDefs)))
				} else {
					b.WriteString(fmt.Sprintf("  %s (keys)\n", strings.Join(k.Fields, ", ")))
				}
			}
		}
		b.WriteString("\n")
	}

	b.WriteString("=== INSTRUCTION ===\n")
	b.WriteString("Generate a DCD template that is 99% identical to the source document.\n")
	b.WriteString("DO NOT use default values. DO NOT assume anything.\n")
	b.WriteString("Extract ALL values directly from the SOURCE DOCUMENT XML below.\n\n")

	b.WriteString("=== CRITICAL: NO DCD SYNTAX ASSUMPTIONS ===\n")
	b.WriteString("The DCD DSL SPECIFICATION above is the COMPLETE and AUTHORITATIVE reference.\n")
	b.WriteString("Do NOT invent, extrapolate, or assume any DCD syntax features beyond what is explicitly listed there.\n")
	b.WriteString("Every tag, attribute, and syntax pattern you use MUST be directly from the specification.\n")
	b.WriteString("If a DCD feature is not documented in the specification, it does not exist — do not use it.\n\n")

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
	b.WriteString("- Wrapper paragraph tags (<w:c>, <w:r>, <w:j>, <w:l>, <w:b>, <w:i>, <w:u>) contain ONLY pure text + variables. NO inline tags allowed. Use <w:*|flags> for combined properties (e.g., <w:c|b>, <w:r|i>). When paragraph has mixed formatting (normal + bold + italic), use rich paragraph <p align=center/right/justify> with inline tags instead.\n")
	b.WriteString("- Rich paragraph tags (<p>, <p align=center/right/justify>) can contain inline tags (<b>, <i>, <u>, <set:>, <tab>). Use rich paragraphs when text has mixed formatting within same paragraph.\n")
	b.WriteString("- Standalone tags (<pb>, <page-break>, <hr>) must NOT be nested inside <p>, <li>, <col>, or <w:*>. If source XML has page break mid-paragraph, split into two paragraphs with <pb> between them.\n")
	b.WriteString("- Tab characters from <w:tab/> in source XML must map to <tab> or <tab size=N> inline tag.\n")
	b.WriteString("- Table structure (<w:tbl>) must map to <table>/<row>/<col>.\n")
	b.WriteString("- Lists (<w:numPr>) must map to <ol>/<ul>/<li>.\n")
	b.WriteString("- Split sections by context/topic: [section 0] for header, [section 1] for parties, [section 2] for transaction, etc. Each section MUST have maximum 3 var and maximum 15 keys. Use [section:next-page N] to start new section on new page.\n")
	b.WriteString("- PREDICTED VARIABLES: use exact names as listed above — no changes. UNPREDICTABLE [keys-unpredictable] / [object-unpredictable]: infer names from document context (strong assumptions allowed, e.g. 'Invoice No:' → invoice_no).\n")
	b.WriteString("- `<w:*>`, `<p>`, `<li>`, `<col>` support size/font-size and color attributes (see SKILL.md for details).\n")
	b.WriteString("- Output ONLY raw DCD syntax, no markdown fences, no extra text.\n\n")

	b.WriteString("=== UNPREDICTABLE VARIABLES ===\n")
	b.WriteString("After the main template sections, include:\n\n")
	b.WriteString("[object-unpredictable]\n")
	b.WriteString("- varName=[]field1, field2     ← array of objects\n")
	b.WriteString("- varName=field1, field2       ← single object\n\n")
	b.WriteString("[keys-unpredictable]\n")
	b.WriteString("- field1, field2, field3       ← simple key mappings\n\n")
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
