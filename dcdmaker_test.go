package dcdmaker

import (
	"os"
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

func TestSessionPath(t *testing.T) {
	path := sessionPath("templates/invoice.dcd")
	expected := "templates/invoice.dcd.session.json"
	if path != expected {
		t.Errorf("sessionPath() = %q, want %q", path, expected)
	}
}
