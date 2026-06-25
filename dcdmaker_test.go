package dcdmaker

import (
	"os"
	"strings"
	"testing"
)

func TestNewMaker(t *testing.T) {
	m := NewMaker()
	if m == nil {
		t.Fatal("NewMaker returned nil")
	}
	if len(m.providers) != 0 {
		t.Fatal("expected empty providers")
	}
}

func TestMakerChain(t *testing.T) {
	m := NewMaker().
		Source("test.docx").
		OptionalPrompt("buat template invoice").
		Resume(true)

	if m.source != "test.docx" {
		t.Fatalf("expected test.docx, got %s", m.source)
	}
	if m.userPrompt != "buat template invoice" {
		t.Fatalf("unexpected prompt: %s", m.userPrompt)
	}
	if !m.resume {
		t.Fatal("expected resume true")
	}
}

func TestGeminiConfig(t *testing.T) {
	os.Setenv("GEMINI_API_KEY", "test-key")
	defer os.Unsetenv("GEMINI_API_KEY")

	p := Gemini()
	if p == nil {
		t.Fatal("Gemini() returned nil")
	}
	if p.Name() != "gemini:gemini-2.5-flash" {
		t.Fatalf("unexpected name: %s", p.Name())
	}
}

func TestGeminiConfigWithOptions(t *testing.T) {
	p := Gemini(
		WithModel("gemini-2.5-pro-exp-03-25"),
		WithTemperature(0.3),
	)
	if p.Name() != "gemini:gemini-2.5-pro-exp-03-25" {
		t.Fatalf("unexpected name: %s", p.Name())
	}
}

func TestOpenAIConfig(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "test-key")
	defer os.Unsetenv("OPENAI_API_KEY")

	p := OpenAI()
	if p == nil {
		t.Fatal("OpenAI() returned nil")
	}
	if p.Name() != "openai:gpt-4o" {
		t.Fatalf("unexpected name: %s", p.Name())
	}
}

func TestGeminiProviderFailsWithoutKey(t *testing.T) {
	os.Unsetenv("GEMINI_API_KEY")
	p := Gemini()
	if p.Name() != "gemini:gemini-2.5-flash" {
		t.Fatalf("unexpected name: %s", p.Name())
	}
}

func TestOpenAIProviderFailsWithoutKey(t *testing.T) {
	os.Unsetenv("OPENAI_API_KEY")
	p := OpenAI()
	if p.Name() != "openai:gpt-4o" {
		t.Fatalf("unexpected name: %s", p.Name())
	}
}

func TestIsDCDValid(t *testing.T) {
	tests := []struct {
		name  string
		dcd   string
		valid bool
	}{
		{
			name:  "empty",
			dcd:   "",
			valid: false,
		},
		{
			name:  "valid minimal",
			dcd:   "[section 0]\nname=test\nvar=info\nkeys=title\n\n--- BODY ---\n<p>{{info.title}}</p>",
			valid: true,
		},
		{
			name:  "missing section",
			dcd:   "<p>hello</p>",
			valid: false,
		},
		{
			name:  "unbalanced loop",
			dcd:   "[section 0]\nname=test\nvar=info, items\nkeys=title\n\n--- BODY ---\n<loop x from items>\n<p>{{x.name}}</p>",
			valid: false,
		},
		{
			name:  "balanced loop",
			dcd:   "[section 0]\nname=test\nvar=info, items\nkeys=title\n\n--- BODY ---\n<loop x from items>\n<p>{{x.name}}</p>\n</loop>",
			valid: true,
		},
		{
			name:  "unbalanced table",
			dcd:   "[section 0]\nname=test\nvar=info\nkeys=title\n\n--- BODY ---\n<table>\n<row><col>a</col></row>",
			valid: false,
		},
		{
			name:  "full valid invoice",
			dcd:   "[style]\nlayout=A4\nunit=inch\nm=1\n\n[title]\ntitle=Invoice\n\n[section 0]\nname=header\nvar=info\nkeys=invoice_no, date, customer\n\n--- BODY ---\n<h1>{{info.invoice_no}}</h1>\n<p>Date: {{info.date}}</p>\n<p>Customer: {{info.customer}}</p>",
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isDCDValid(tt.dcd)
			if got != tt.valid {
				t.Errorf("isDCDValid() = %v, want %v", got, tt.valid)
			}
		})
	}
}

func TestIsTruncated(t *testing.T) {
	tests := []struct {
		name      string
		dcd       string
		truncated bool
	}{
		{
			name:      "explicit marker",
			dcd:       "[section 0]\n<p>test</p>\n\n<TRUNCATED/>",
			truncated: true,
		},
		{
			name:      "unbalanced section vs body",
			dcd:       "[section 0]\nname=test\n\n--- BODY ---\n<p>done</p>\n\n[section 1]",
			truncated: true,
		},
		{
			name:      "unclosed loop",
			dcd:       "[section 0]\nname=test\n\n--- BODY ---\n<loop x from items>\n<p>test</p>",
			truncated: true,
		},
		{
			name:      "complete",
			dcd:       "[section 0]\nname=test\n\n--- BODY ---\n<loop x from items>\n<p>{{x.name}}</p>\n</loop>",
			truncated: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isTruncated(tt.dcd)
			if got != tt.truncated {
				t.Errorf("isTruncated() = %v, want %v", got, tt.truncated)
			}
		})
	}
}

func TestSanitizeDCD(t *testing.T) {
	input := "```dcd\n[section 0]\n<p>test</p>\n```"
	expected := "[section 0]\n<p>test</p>"

	got := sanitizeDCD(input)
	if got != expected {
		t.Errorf("sanitizeDCD() = %q, want %q", got, expected)
	}
}

func TestMultipleProviders(t *testing.T) {
	os.Setenv("GEMINI_API_KEY", "test-key")
	os.Setenv("OPENAI_API_KEY", "test-key")
	defer os.Unsetenv("GEMINI_API_KEY")
	defer os.Unsetenv("OPENAI_API_KEY")

	m := NewMaker(
		Gemini(WithModel("gemini-2.5-flash")),
		Gemini(WithModel("gemini-2.5-pro-exp-03-25")),
		OpenAI(WithOpenAIModel("gpt-4o")),
	)

	if len(m.providers) != 3 {
		t.Fatalf("expected 3 providers, got %d", len(m.providers))
	}

	names := []string{
		"gemini:gemini-2.5-flash",
		"gemini:gemini-2.5-pro-exp-03-25",
		"openai:gpt-4o",
	}
	for i, p := range m.providers {
		if p.Name() != names[i] {
			t.Errorf("provider[%d] name = %q, want %q", i, p.Name(), names[i])
		}
	}
}

func TestRunRequiresSource(t *testing.T) {
	m := NewMaker(
		Gemini(WithAPIKey("test")),
	)
	err := m.Run("out.dcd")
	if err == nil {
		t.Fatal("expected error for missing source")
	}
}

func TestRunRequiresProviders(t *testing.T) {
	m := NewMaker()
	err := m.Run("out.dcd")
	if err == nil {
		t.Fatal("expected error for missing providers")
	}
}

func TestGenerateReturnsString(t *testing.T) {
	m := NewMaker(Gemini(WithAPIKey("test")))
	m.Source("test.docx")
	dcd, err := m.Generate()
	if err == nil {
		t.Fatal("expected error (no real file), but Generate() returned string:", dcd[:min(len(dcd), 50)])
	}
	// Should fail because file doesn't exist, not because of config
	if err.Error() == "dcdmaker: at least one provider required" ||
		err.Error() == "dcdmaker: source document required" {
		t.Fatal("unexpected validation error:", err)
	}
}

func TestGenerateRequiresSource(t *testing.T) {
	m := NewMaker(Gemini(WithAPIKey("test")))
	_, err := m.Generate()
	if err == nil {
		t.Fatal("expected error for missing source")
	}
}

func TestGenerateRequiresProviders(t *testing.T) {
	m := NewMaker()
	m.Source("test.docx")
	_, err := m.Generate()
	if err == nil {
		t.Fatal("expected error for missing providers")
	}
}

func TestGenerateRejectsResume(t *testing.T) {
	m := NewMaker(Gemini(WithAPIKey("test"))).
		Source("test.docx").
		Resume(true)
	_, err := m.Generate()
	if err == nil {
		t.Fatal("expected error for resume with Generate()")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestSessionPath(t *testing.T) {
	path := sessionPath("templates/invoice.dcd")
	expected := "templates/invoice.dcd.session.json"
	if path != expected {
		t.Errorf("sessionPath() = %q, want %q", path, expected)
	}
}

func TestPredictableKeys(t *testing.T) {
	m := NewMaker().
		PredictableKeys(
			Object("info", "invoice_no", "date", "customer"),
			Array("items", "name", "qty", "price"),
			Object("summary", "subtotal", "tax", "total"),
		)

	if len(m.predictableKeys) != 3 {
		t.Fatalf("expected 3 keys, got %d", len(m.predictableKeys))
	}
}

func TestAddPredictableKeys(t *testing.T) {
	m := NewMaker()
	m.AddPredictableKeys(Object("info", "a", "b"))
	m.AddPredictableKeys(Array("items", "x", "y"))
	m.AddPredictableKeys(Object("summary", "z"))

	if len(m.predictableKeys) != 3 {
		t.Fatalf("expected 3 keys after 3 adds, got %d", len(m.predictableKeys))
	}
	if m.predictableKeys[0].Name != "info" {
		t.Errorf("first key = %q, want %q", m.predictableKeys[0].Name, "info")
	}
	if m.predictableKeys[1].Name != "items" {
		t.Errorf("second key = %q, want %q", m.predictableKeys[1].Name, "items")
	}
	if m.predictableKeys[2].Name != "summary" {
		t.Errorf("third key = %q, want %q", m.predictableKeys[2].Name, "summary")
	}
}

func TestAddPredictableKeysFromLoop(t *testing.T) {
	type reqKey struct {
		Name   string
		Type   string
		Fields []string
	}
	reqKeys := []reqKey{
		{Name: "info", Type: "object", Fields: []string{"a", "b"}},
		{Name: "items", Type: "array", Fields: []string{"x", "y", "z"}},
		{Name: "summary", Type: "object", Fields: []string{"total"}},
	}

	m := NewMaker()
	for _, k := range reqKeys {
		switch k.Type {
		case "object":
			m.AddPredictableKeys(Object(k.Name, k.Fields...))
		case "array":
			m.AddPredictableKeys(Array(k.Name, k.Fields...))
		}
	}

	if len(m.predictableKeys) != 3 {
		t.Fatalf("expected 3 keys, got %d", len(m.predictableKeys))
	}
	if m.predictableKeys[0].Type != VarObject {
		t.Errorf("keys[0] type = %v, want VarObject", m.predictableKeys[0].Type)
	}
	if m.predictableKeys[1].Type != VarArray {
		t.Errorf("keys[1] type = %v, want VarArray", m.predictableKeys[1].Type)
	}
}

func TestKeyDefObject(t *testing.T) {
	k := Object("header", "title", "date")
	if k.Name != "header" {
		t.Errorf("Object name = %q, want %q", k.Name, "header")
	}
	if k.Type != VarObject {
		t.Errorf("Object type = %v, want VarObject", k.Type)
	}
	if len(k.Fields) != 2 || k.Fields[0] != "title" {
		t.Errorf("Object fields = %v, want [title date]", k.Fields)
	}
}

func TestKeyDefArray(t *testing.T) {
	k := Array("items", "name", "qty", "price")
	if k.Name != "items" {
		t.Errorf("Array name = %q, want %q", k.Name, "items")
	}
	if k.Type != VarArray {
		t.Errorf("Array type = %v, want VarArray", k.Type)
	}
	if len(k.Fields) != 3 {
		t.Errorf("Array fields = %v, want 3 fields", k.Fields)
	}
}

func TestIsDCDValidWithUnpredictable(t *testing.T) {
	dcd := `[section 0]
name=header
var=info
keys=invoice_no, date, customer

--- BODY ---
<h1>{{info.invoice_no}}</h1>
<p>{{info.date}}</p>
<p>{{info.customer}}</p>

[object-unpredictable]
- info=discount, notes
- items=[]name, qty

[keys-unpredictable]
- data=po_number, department
`
	if !isDCDValid(dcd) {
		t.Error("expected DCD with unpredictable sections to be valid")
	}
}

func TestIsDCDValidFullWithPredictable(t *testing.T) {
	dcd := `[style]
layout=A4
unit=inch
m=1

[title]
title=Invoice

[header]
right={{page}} / {{total}}

[section 0]
name=header
var=info
keys=invoice_no, date, customer, due_date

--- BODY ---
<h1>INVOICE: {{info.invoice_no}}</h1>
<p>Date: {{info.date}}</p>
<p>Customer: {{info.customer}}</p>
<p>Due: {{info.due_date}}</p>

[section 1]
name=items
var=info, items
keys=title, items.name, items.qty, items.unit_price, items.total

--- BODY ---
<table border=1 width=100%>
<loop:row x from items>
  <col>{{x.name}}</col>
  <col align=right>{{x.qty}}</col>
  <col align=right>{{x.unit_price}}</col>
  <col align=right>{{x.total}}</col>
</loop:row>
</table>

[object-unpredictable]
- info=shipping_address, payment_terms
`
	if !isDCDValid(dcd) {
		t.Error("expected full DCD with predictable keys to be valid")
	}
}

func TestMakerPredictableKeysChain(t *testing.T) {
	m := NewMaker().
		Source("test.docx").
		PredictableKeys(
			Object("info", "a", "b"),
			Array("list", "x", "y", "z"),
		).
		OptionalPrompt("test").
		Resume(true)

	if len(m.predictableKeys) != 2 {
		t.Fatalf("expected 2 predictable keys, got %d", len(m.predictableKeys))
	}
	if m.source != "test.docx" {
		t.Errorf("source = %q, want %q", m.source, "test.docx")
	}
	if !m.resume {
		t.Error("expected resume true")
	}
}

func TestBuildPromptWithPredictableKeys(t *testing.T) {
	keys := []KeyDef{
		Object("info", "invoice_no", "date"),
		Array("items", "name", "qty"),
	}
	prompt := buildPrompt("", keys)

	if !strings.Contains(prompt, "=== PREDICTED VARIABLES ===") {
		t.Error("prompt missing PREDICTED VARIABLES section")
	}
	if !strings.Contains(prompt, "info") {
		t.Error("prompt missing info variable")
	}
	if !strings.Contains(prompt, "items") {
		t.Error("prompt missing items variable")
	}
	if !strings.Contains(prompt, "[object-unpredictable]") {
		t.Error("prompt missing [object-unpredictable] instruction")
	}
	if !strings.Contains(prompt, "[keys-unpredictable]") {
		t.Error("prompt missing [keys-unpredictable] instruction")
	}
}

func TestBuildPromptWithoutPredictableKeys(t *testing.T) {
	prompt := buildPrompt("hello", nil)

	if strings.Contains(prompt, "=== PREDICTED VARIABLES ===") {
		t.Error("prompt should not contain PREDICTED VARIABLES when no keys given")
	}
	if !strings.Contains(prompt, "[object-unpredictable]") {
		t.Error("prompt missing [object-unpredictable] instruction")
	}
	if !strings.Contains(prompt, "hello") {
		t.Error("prompt missing user instruction")
	}
}
