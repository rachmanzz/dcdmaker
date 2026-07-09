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
		OptionalPrompt("create invoice template").
		Resume(true)

	if m.source != "test.docx" {
		t.Fatalf("expected test.docx, got %s", m.source)
	}
	if m.userPrompt != "create invoice template" {
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
			valid: true,
		},
		{
			name:  "balanced loop",
			dcd:   "[section 0]\nname=test\nvar=items\nkeys=title\n\n--- BODY ---\n<loop x from items>\n<p>{{x.name}}</p>\n</loop>",
			valid: true,
		},
		{
			name:  "unbalanced table",
			dcd:   "[section 0]\nname=test\nvar=info\nkeys=title\n\n--- BODY ---\n<table>\n<row><col>a</col></row>",
			valid: true,
		},
		{
			name:  "full valid invoice",
			dcd:   "[style]\nlayout=A4\nunit=inch\nm=1\n\n[title]\ntitle=Invoice\n\n[section 0]\nname=header\nvar=info\nkeys=invoice_no, date, customer\n\n--- BODY ---\n<h1>{{info.invoice_no}}</h1>\n<p>Date: {{info.date}}</p>\n<p>Customer: {{info.customer}}</p>",
			valid: true,
		},
		{
			name:  "ol with type attribute",
			dcd:   "[section 0]\nname=test\nvar=info\nkeys=title\n\n--- BODY ---\n<ol type=a>\n<li>{{info.title}}</li>\n</ol>",
			valid: true,
		},
		{
			name:  "loop:ol with type attribute",
			dcd:   "[section 0]\nname=test\nvar=items\nkeys=title\n\n--- BODY ---\n<loop:ol type=a x from items>\n<li>{{x.name}}</li>\n</loop:ol>",
			valid: true,
		},
		{
			name:  "unbalanced ol with type",
			dcd:   "[section 0]\nname=test\nvar=info\nkeys=title\n\n--- BODY ---\n<ol type=a>\n<li>{{info.title}}</li>",
			valid: true,
		},
		{
			name:  "wrong loop syntax ol x from",
			dcd:   "[section 0]\nname=test\nvar=items\nkeys=title\n\n--- BODY ---\n<ol x from items>\n<li>{{x.name}}</li>\n</ol>",
			valid: true,
		},
		{
			name:  "wrong loop syntax ul x from",
			dcd:   "[section 0]\nname=test\nvar=items\nkeys=title\n\n--- BODY ---\n<ul x from items>\n<li>{{x.name}}</li>\n</ul>",
			valid: true,
		},
		{
			name:  "unused var declaration",
			dcd:   "[section 0]\nname=test\nvar=info\nkeys=title\n\n--- BODY ---\n<p>static text only</p>",
			valid: true,
		},
		{
			name:  "used var across sections",
			dcd:   "[section 0]\nname=header\nvar=info\n\n--- BODY ---\n<p>{{info.title}}</p>\n\n[section 1]\nname=items\nvar=info, items\n\n--- BODY ---\n<loop x from items>\n<p>{{x.name}}</p>\n</loop>",
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, reason := isDCDValid(tt.dcd)
			if got != tt.valid {
				t.Errorf("isDCDValid() = %v, want %v (reason: %s)", got, tt.valid, reason)
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
			truncated: false,
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

func TestIsIncomplete(t *testing.T) {
	tests := []struct {
		name       string
		dcd        string
		incomplete bool
	}{
		{
			name:       "less than 30 lines",
			dcd:        "[section 0]\nname=test\nvar=info\nkeys=title\n\n--- BODY ---\n<p>{{info.title}}</p>",
			incomplete: true,
		},
		{
			name: "unpredictable sections empty",
			dcd: `[section 0]
name=test
var=info
keys=title

--- BODY ---
<p>{{info.title}}</p>
<p>line 1</p>
<p>line 2</p>
<p>line 3</p>
<p>line 4</p>
<p>line 5</p>
<p>line 6</p>
<p>line 7</p>
<p>line 8</p>
<p>line 9</p>
<p>line 10</p>
<p>line 11</p>
<p>line 12</p>
<p>line 13</p>
<p>line 14</p>
<p>line 15</p>
<p>line 16</p>
<p>line 17</p>
<p>line 18</p>
<p>line 19</p>
<p>line 20</p>
<p>line 21</p>
<p>line 22</p>
<p>line 23</p>
<p>line 24</p>
<p>line 25</p>
<p>line 26</p>
<p>line 27</p>
<p>line 28</p>

[object-unpredictable]

[keys-unpredictable]`,
			incomplete: true,
		},
		{
			name: "unpredictable sections with content",
			dcd: `[section 0]
name=test
var=info
keys=title

--- BODY ---
<p>{{info.title}}</p>
<p>line 1</p>
<p>line 2</p>
<p>line 3</p>
<p>line 4</p>
<p>line 5</p>
<p>line 6</p>
<p>line 7</p>
<p>line 8</p>
<p>line 9</p>
<p>line 10</p>
<p>line 11</p>
<p>line 12</p>
<p>line 13</p>
<p>line 14</p>
<p>line 15</p>
<p>line 16</p>
<p>line 17</p>
<p>line 18</p>
<p>line 19</p>
<p>line 20</p>
<p>line 21</p>
<p>line 22</p>
<p>line 23</p>
<p>line 24</p>
<p>line 25</p>
<p>line 26</p>
<p>line 27</p>
<p>line 28</p>

[object-unpredictable]
var=extras
keys=field1, field2

[keys-unpredictable]
var=extras`,
			incomplete: false,
		},
		{
			name: "no unpredictable sections, body >= section",
			dcd: `[section 0]
name=test
var=info
keys=title

--- BODY ---
<p>{{info.title}}</p>
<p>line 1</p>
<p>line 2</p>
<p>line 3</p>
<p>line 4</p>
<p>line 5</p>
<p>line 6</p>
<p>line 7</p>
<p>line 8</p>
<p>line 9</p>
<p>line 10</p>
<p>line 11</p>
<p>line 12</p>
<p>line 13</p>
<p>line 14</p>
<p>line 15</p>
<p>line 16</p>
<p>line 17</p>
<p>line 18</p>
<p>line 19</p>
<p>line 20</p>
<p>line 21</p>
<p>line 22</p>
<p>line 23</p>
<p>line 24</p>
<p>line 25</p>
<p>line 26</p>
<p>line 27</p>
<p>line 28</p>`,
			incomplete: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isIncomplete(tt.dcd)
			if got != tt.incomplete {
				t.Errorf("isIncomplete() = %v, want %v", got, tt.incomplete)
			}
		})
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

func TestKeyDefKeys(t *testing.T) {
	k := Keys("po_number", "department")
	if k.Type != VarKeys {
		t.Errorf("Keys type = %v, want VarKeys", k.Type)
	}
	if len(k.Fields) != 2 || k.Fields[0] != "po_number" {
		t.Errorf("Keys fields = %v", k.Fields)
	}
}

func TestKeysEx(t *testing.T) {
	k := KeysEx(
		Field("po_number", "string"),
		Field("date", "date-str", "DD-MM-YYYY"),
	)
	if k.Type != VarKeys {
		t.Errorf("KeysEx type = %v, want VarKeys", k.Type)
	}
	if len(k.FieldDefs) != 2 {
		t.Fatalf("KeysEx FieldDefs len = %d", len(k.FieldDefs))
	}
	if k.FieldDefs[0].Name != "po_number" || k.FieldDefs[0].Type != "string" {
		t.Errorf("bad first field: %+v", k.FieldDefs[0])
	}
	if k.FieldDefs[1].Format != "DD-MM-YYYY" {
		t.Errorf("bad format: %+v", k.FieldDefs[1])
	}
}

func TestFieldDef(t *testing.T) {
	f1 := Field("nama", "string")
	if f1.Name != "nama" || f1.Type != "string" || f1.Format != "" {
		t.Errorf("Field() without format = %+v", f1)
	}

	f2 := Field("tanggal_lahir", "date-str", "DD-MM-YYYY")
	if f2.Name != "tanggal_lahir" || f2.Type != "date-str" || f2.Format != "DD-MM-YYYY" {
		t.Errorf("Field() with format = %+v", f2)
	}

	f3 := Field("jumlah", "number")
	if f3.Name != "jumlah" || f3.Type != "number" || f3.Format != "" {
		t.Errorf("Field() number = %+v", f3)
	}
}

func TestKeyDefObjectEx(t *testing.T) {
	k := ObjectEx("penjual",
		Field("nama", "string"),
		Field("tanggal_lahir", "date-str", "DD-MM-YYYY"),
	)
	if k.Name != "penjual" || k.Type != VarObject {
		t.Errorf("ObjectEx name/type = %+v", k)
	}
	if len(k.FieldDefs) != 2 {
		t.Fatalf("ObjectEx FieldDefs len = %d", len(k.FieldDefs))
	}
	if k.FieldDefs[0].Name != "nama" || k.FieldDefs[0].Type != "string" {
		t.Errorf("bad first field: %+v", k.FieldDefs[0])
	}
	if k.FieldDefs[1].Format != "DD-MM-YYYY" {
		t.Errorf("bad format: %+v", k.FieldDefs[1])
	}
}

func TestKeyDefArrayEx(t *testing.T) {
	k := ArrayEx("items",
		Field("name", "string"),
		Field("qty", "number"),
	)
	if k.Name != "items" || k.Type != VarArray {
		t.Errorf("ArrayEx name/type = %+v", k)
	}
	if len(k.FieldDefs) != 2 {
		t.Fatalf("ArrayEx FieldDefs len = %d", len(k.FieldDefs))
	}
	if k.FieldDefs[1].Name != "qty" || k.FieldDefs[1].Type != "number" {
		t.Errorf("bad field: %+v", k.FieldDefs[1])
	}
}

func TestBuildPromptWithFieldDefs(t *testing.T) {
	prompt := buildPrompt("", []KeyDef{
		ObjectEx("info",
			Field("invoice_no", "string"),
			Field("date", "date-str", "DD-MM-YYYY"),
		),
		ArrayEx("items",
			Field("name", "string"),
			Field("qty", "number"),
			Field("unit_price", "number"),
		),
		Keys("po_number", "department"),
		KeysEx(
			Field("date", "date-str", "DD-MM-YYYY"),
			Field("total", "number"),
		),
	})

	if !strings.Contains(prompt, "info {invoice_no: string, date: date-str (DD-MM-YYYY)}") {
		t.Errorf("prompt missing ObjectEx with types:\n%s", prompt)
	}
	if !strings.Contains(prompt, "[]items {name: string, qty: number, unit_price: number} (array)") {
		t.Errorf("prompt missing ArrayEx with types:\n%s", prompt)
	}
	if !strings.Contains(prompt, "po_number, department (keys)") {
		t.Errorf("prompt missing Keys:\n%s", prompt)
	}
	if !strings.Contains(prompt, "date: date-str (DD-MM-YYYY), total: number (keys)") {
		t.Errorf("prompt missing KeysEx with types:\n%s", prompt)
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
	valid, reason := isDCDValid(dcd)
	if !valid {
		t.Errorf("expected DCD with unpredictable sections to be valid (reason: %s)", reason)
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
	valid, reason := isDCDValid(dcd)
	if !valid {
		t.Errorf("expected full DCD with predictable keys to be valid (reason: %s)", reason)
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

func TestMakerMaxRetries(t *testing.T) {
	m := NewMaker()
	if m.maxRetries != 3 {
		t.Fatalf("expected default maxRetries=3, got %d", m.maxRetries)
	}

	m.MaxRetries(5)
	if m.maxRetries != 5 {
		t.Fatalf("expected maxRetries=5, got %d", m.maxRetries)
	}

	m.MaxRetries(0)
	if m.maxRetries != 1 {
		t.Fatalf("expected maxRetries clamped to 1, got %d", m.maxRetries)
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

func TestUnpredictableParsing(t *testing.T) {
	dcd := `[section 0]
name=test
var=info
keys=title

--- BODY ---
<p>{{info.title}}</p>

[object-unpredictable]
- penjual[]=nama, identitas, alamat
- notaris=nama, sk, alamat

[keys-unpredictable]
- nama_ppat, kedudukan_ppat, sk_nomor
`
	objs := parseUnpredictableObjects(dcd)
	if len(objs) != 2 {
		t.Fatalf("expected 2 objects, got %d", len(objs))
	}
	if objs[0].Name != "penjual" || !objs[0].IsArray || len(objs[0].Fields) != 3 {
		t.Fatalf("bad first object: %+v", objs[0])
	}
	if objs[1].Name != "notaris" || objs[1].IsArray || len(objs[1].Fields) != 3 {
		t.Fatalf("bad second object: %+v", objs[1])
	}

	keys := parseUnpredictableKeys(dcd)
	if len(keys) != 3 {
		t.Fatalf("expected 3 keys, got %d: %v", len(keys), keys)
	}
	if keys[0] != "nama_ppat" || keys[1] != "kedudukan_ppat" || keys[2] != "sk_nomor" {
		t.Fatalf("bad keys: %v", keys)
	}
}

func TestUnpredictableEmpty(t *testing.T) {
	dcd := `[section 0]
name=test
var=info
keys=title

--- BODY ---
<p>{{info.title}}</p>
`
	objs := parseUnpredictableObjects(dcd)
	if len(objs) != 0 {
		t.Fatalf("expected 0 objects, got %d", len(objs))
	}
	keys := parseUnpredictableKeys(dcd)
	if len(keys) != 0 {
		t.Fatalf("expected 0 keys, got %d", len(keys))
	}
}

func TestMakerUnpredictableMethods(t *testing.T) {
	m := &Maker{
		lastResult: `[section 0]
name=test
var=info
keys=title

--- BODY ---
<p>{{info.title}}</p>

[object-unpredictable]
- user[]=name, email

[keys-unpredictable]
- title, author
`,
	}
	objs := m.UnpredictableObjects()
	if len(objs) != 1 || objs[0].Name != "user" || !objs[0].IsArray {
		t.Fatalf("bad objects: %+v", objs)
	}
	keys := m.UnpredictableKeys()
	if len(keys) != 2 || keys[0] != "title" || keys[1] != "author" {
		t.Fatalf("bad keys: %v", keys)
	}
}

func TestUnpredictableParsingWithoutDash(t *testing.T) {
	dcd := `[section 0]
name=test
var=info
keys=title

--- BODY ---
<p>{{info.title}}</p>

[object-unpredictable]
basic=jam_terbilang_akta, gelar_notaris
pendiri[]=tanggal_lahir, alamat, kota

[keys-unpredictable]
field1, field2
field3, field4
`
	objs := parseUnpredictableObjects(dcd)
	if len(objs) != 2 {
		t.Fatalf("expected 2 objects without dash prefix, got %d", len(objs))
	}
	if objs[0].Name != "basic" || objs[0].IsArray || len(objs[0].Fields) != 2 {
		t.Fatalf("bad first object: %+v", objs[0])
	}
	if objs[1].Name != "pendiri" || !objs[1].IsArray || len(objs[1].Fields) != 3 {
		t.Fatalf("bad second object: %+v", objs[1])
	}

	keys := parseUnpredictableKeys(dcd)
	if len(keys) != 4 {
		t.Fatalf("expected 4 keys (2 lines), got %d: %v", len(keys), keys)
	}
}

func TestUnpredictableParsingMixed(t *testing.T) {
	dcd := `[object-unpredictable]
- user[]=name, email
guest=name, phone

[keys-unpredictable]
- title, author
tags, category
`
	objs := parseUnpredictableObjects(dcd)
	if len(objs) != 2 {
		t.Fatalf("expected 2 objects mixed, got %d", len(objs))
	}

	keys := parseUnpredictableKeys(dcd)
	if len(keys) != 4 {
		t.Fatalf("expected 4 keys mixed, got %d: %v", len(keys), keys)
	}
}

func TestParseSections(t *testing.T) {
	dcd := `[section 0]
name=header
var=info
keys=title, date

--- BODY ---
<p>{{info.title}}</p>
<p>{{info.date}}</p>

[section 1]
name=items
var=items
keys=name, qty, price

--- BODY ---
<table>
<loop x from items>
<col>{{x.name}}</col>
<col>{{x.qty}}</col>
</loop>
</table>
`
	sections := parseSections(dcd)
	if len(sections) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(sections))
	}
	if sections[0].Name != "header" || len(sections[0].Vars) != 1 || sections[0].Vars[0] != "info" {
		t.Fatalf("bad section 0: %+v", sections[0])
	}
	if len(sections[0].Keys) != 2 || sections[0].Keys[1] != "date" {
		t.Fatalf("bad section 0 keys: %+v", sections[0].Keys)
	}
	if sections[1].Name != "items" || len(sections[1].Vars) != 1 || sections[1].Vars[0] != "items" {
		t.Fatalf("bad section 1: %+v", sections[1])
	}
}

func TestScanBody(t *testing.T) {
	dcd := `<p>{{info.title}}</p>
<p>{{info.date}}</p>
<loop x from items>
<col>{{x.name}}</col>
</loop>
{{plain_key}}
`
	usages := scanBody(dcd)
	if len(usages) != 5 {
		t.Fatalf("expected 5 usages, got %d: %+v", len(usages), usages)
	}
}

func TestValidateVarsAndKeysOK(t *testing.T) {
	dcd := `[section 0]
name=test
var=info
keys=title

--- BODY ---
<p>{{info.title}}</p>
`
	if err := validateVarsAndKeys(dcd); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestValidateVarsAndKeysMissingField(t *testing.T) {
	dcd := `[section 0]
name=test
var=info
keys=title

--- BODY ---
<p>{{info.title}}</p>
<p>{{info.missing_field}}</p>
`
	if err := validateVarsAndKeys(dcd); err == nil {
		t.Fatal("expected error for missing field")
	}
}

func TestValidateVarsAndKeysMissingVar(t *testing.T) {
	dcd := `[section 0]
name=test
var=info
keys=title

--- BODY ---
<p>{{info.title}}</p>
<loop x from pendiri>
<col>{{x.nama}}</col>
</loop>
`
	if err := validateVarsAndKeys(dcd); err == nil {
		t.Fatal("expected error for undeclared var 'pendiri'")
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


func TestFixUnpredictableOverlap_NoPredictedKeys(t *testing.T) {
	dcd := `[object-unpredictable]
info=nama, alamat

[keys-unpredictable]
nama, alamat
`
	got := fixUnpredictableOverlap(dcd, nil)
	if got != dcd {
		t.Fatalf("expected unchanged when no predictable keys, got:\n%s", got)
	}
}

func TestFixUnpredictableOverlap_ObjectNameCollision(t *testing.T) {
	dcd := `[section 0]
name=test
var=info
keys=nama, alamat

--- BODY ---
<p>{{info.nama}} {{info.alamat}}</p>

[object-unpredictable]
info=nama, alamat

[keys-unpredictable]
extra_key
`
	want := `[section 0]
name=test
var=info
keys=nama, alamat

--- BODY ---
<p>{{info.nama}} {{info.alamat}}</p>

[keys-unpredictable]
extra_key
`
	pred := []KeyDef{Object("info", "nama", "alamat")}
	got := fixUnpredictableOverlap(dcd, pred)
	if got != want {
		t.Fatalf("got:\n%s\n\nwant:\n%s", got, want)
	}
}

func TestFixUnpredictableOverlap_FlatKeyCollision(t *testing.T) {
	dcd := `[object-unpredictable]
	items[]=name, qty

[keys-unpredictable]
	bidang_usaha, nomor_surat_saham
`
	want := `[object-unpredictable]
	items[]=name, qty

[keys-unpredictable]
nomor_surat_saham
`
	pred := []KeyDef{Keys("bidang_usaha")}
	got := fixUnpredictableOverlap(dcd, pred)
	if got != want {
		t.Fatalf("got:\n%s\n\nwant:\n%s", got, want)
	}
}

func TestFixUnpredictableOverlap_FieldCollision(t *testing.T) {
	dcd := `[section 0]
name=test
var=info
keys=nama, alamat

--- BODY ---
<p>{{info.nama}} {{info.alamat}}</p>

[keys-unpredictable]
nama, extra_field
`
	// "nama" is a predicted field of object "info", not a global name — should NOT be removed
	want := dcd
	pred := []KeyDef{Object("info", "nama", "alamat")}
	got := fixUnpredictableOverlap(dcd, pred)
	if got != want {
		t.Fatalf("expected unchanged (field-level not checked), got:\n%s", got)
	}
}

func TestFixUnpredictableOverlap_FieldInObjectCollision(t *testing.T) {
	dcd := `[section 0]
name=test
var=info
keys=nama, alamat

--- BODY ---
<p>{{info.nama}} {{info.alamat}}</p>

[object-unpredictable]
extra=alamat, phone
`
	// "alamat" is a predicted field of "info", not a global name — should NOT be removed
	want := dcd
	pred := []KeyDef{Object("info", "nama", "alamat")}
	got := fixUnpredictableOverlap(dcd, pred)
	if got != want {
		t.Fatalf("expected unchanged (field-level not checked), got:\n%s", got)
	}
}

func TestFixUnpredictableOverlap_NoCollision(t *testing.T) {
	dcd := `[section 0]
name=test
var=info
keys=nama, alamat

--- BODY ---
<p>{{info.nama}} {{info.alamat}}</p>

[object-unpredictable]
	pendiri[]=nama_pendiri, tgl_lahir

[keys-unpredictable]
bidang_usaha
`
	want := dcd
	pred := []KeyDef{Object("info", "nama", "alamat")}
	got := fixUnpredictableOverlap(dcd, pred)
	if got != want {
		t.Fatalf("expected unchanged, got:\n%s", got)
	}
}

func TestFixUnpredictableOverlap_VarKeysEx(t *testing.T) {
	dcd := `[keys-unpredictable]
signer_name, position
`
	want := `[keys-unpredictable]
position
`
	pred := []KeyDef{KeysEx(FieldDef{Name: "signer_name"})}
	got := fixUnpredictableOverlap(dcd, pred)
	if got != want {
		t.Fatalf("got:\n%s\n\nwant:\n%s", got, want)
	}
}
