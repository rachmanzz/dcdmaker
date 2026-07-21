package dcdmaker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rachmanzz/words-xml/words"
)

type VarType int

const (
	VarObject VarType = iota
	VarArray
	VarKeys
)

type FieldDef struct {
	Name   string
	Type   string // "string", "number", "boolean", "date-str"
	Format string // optional, used with date-str
}

type KeyDef struct {
	Name      string
	Type      VarType
	Fields    []string   // backwards compat, used when FieldDefs is nil
	FieldDefs []FieldDef // takes priority over Fields when set
}

func Object(name string, fields ...string) KeyDef {
	return KeyDef{Name: name, Type: VarObject, Fields: fields}
}

func Array(name string, fields ...string) KeyDef {
	return KeyDef{Name: name, Type: VarArray, Fields: fields}
}

func Keys(fields ...string) KeyDef {
	return KeyDef{Type: VarKeys, Fields: fields}
}

func KeysEx(fields ...FieldDef) KeyDef {
	return KeyDef{Type: VarKeys, FieldDefs: fields}
}

func Field(name string, typ string, format ...string) FieldDef {
	f := FieldDef{Name: name, Type: typ}
	if len(format) > 0 {
		f.Format = format[0]
	}
	return f
}

func ObjectEx(name string, fields ...FieldDef) KeyDef {
	return KeyDef{Name: name, Type: VarObject, FieldDefs: fields}
}

func ArrayEx(name string, fields ...FieldDef) KeyDef {
	return KeyDef{Name: name, Type: VarArray, FieldDefs: fields}
}

type Maker struct {
	providers       []Provider
	source          string
	userPrompt      string
	predictableKeys []KeyDef
	lastProvider    string
	lastResult      string
	maxRetries      int
}

func NewMaker(providers ...Provider) *Maker {
	return &Maker{
		providers:  providers,
		maxRetries: 3,
	}
}

func (m *Maker) Source(path string) *Maker {
	m.source = path
	return m
}

func (m *Maker) OptionalPrompt(p string) *Maker {
	m.userPrompt = p
	return m
}

func (m *Maker) PredictableKeys(keys ...KeyDef) *Maker {
	m.predictableKeys = keys
	return m
}

func (m *Maker) MaxRetries(n int) *Maker {
	if n < 1 {
		n = 1
	}
	m.maxRetries = n
	return m
}

func (m *Maker) LastProvider() string {
	return m.lastProvider
}

func (m *Maker) LastResult() string {
	return m.lastResult
}

func (m *Maker) UnpredictableObjects() []UnpredictableObject {
	return parseUnpredictableObjects(m.lastResult)
}

func (m *Maker) UnpredictableKeys() []string {
	return parseUnpredictableKeys(m.lastResult)
}

func (m *Maker) AddPredictableKeys(keys ...KeyDef) *Maker {
	m.predictableKeys = append(m.predictableKeys, keys...)
	return m
}

func (m *Maker) Generate() (string, error) {
	if len(m.providers) == 0 {
		return "", fmt.Errorf("dcdmaker: at least one provider required")
	}
	if m.source == "" {
		return "", fmt.Errorf("dcdmaker: source document required")
	}

	data, err := os.ReadFile(m.source)
	if err != nil {
		return "", fmt.Errorf("dcdmaker: read source: %w", err)
	}

	result, err := m.generate(data)
	if err != nil {
		return "", err
	}

	return result, nil
}

func (m *Maker) Run(output string) error {
	if len(m.providers) == 0 {
		return fmt.Errorf("dcdmaker: at least one provider required")
	}

	if m.source == "" {
		return fmt.Errorf("dcdmaker: source document required")
	}

	data, err := os.ReadFile(m.source)
	if err != nil {
		return fmt.Errorf("dcdmaker: read source: %w", err)
	}

	result, err := m.generate(data)
	if err != nil {
		return err
	}

	return os.WriteFile(output, []byte(result), 0644)
}

func (m *Maker) generate(data []byte) (string, error) {
	var styleBlock string
	wordsResult, parseErr := words.ProcessDOCXBytes(data)
	if parseErr == nil {
		styleBlock = extractStyleBlock(wordsResult.WordsXML)
	}

	originalPrompt := buildPrompt(m.userPrompt, m.predictableKeys)
	prompt := originalPrompt
	ctx := context.Background()

	delays := []time.Duration{5 * time.Second, 10 * time.Second, 15 * time.Second}
	debug := os.Getenv("DCD_DEBUG") == "true"

	for pi, provider := range m.providers {
		var lastErr error

		for attempt := range m.maxRetries {
			if attempt > 0 {
				select {
				case <-ctx.Done():
					return "", ctx.Err()
				case <-time.After(delays[min(attempt-1, len(delays)-1)]):
				}
			}

			result, err := provider.GenerateWithFile(ctx, prompt, m.source, data)
			if err != nil {
				lastErr = fmt.Errorf("%s attempt %d: %w", provider.Name(), attempt+1, err)
				continue
			}

			if debug {
				os.MkdirAll("dcd_temp", 0755)
				rawPath := filepath.Join("dcd_temp", fmt.Sprintf("dcd_debug_%s_attempt_%d_raw.dcd", provider.Name(), attempt+1))
				os.WriteFile(rawPath, []byte(result), 0644)
				fmt.Fprintf(os.Stderr, "[dcd-debug] %s attempt %d: raw output saved to %s (%d bytes)\n",
					provider.Name(), attempt+1, rawPath, len(result))
			}

			result = sanitizeDCD(result)
			if styleBlock != "" {
				result = stripStyleBlock(result)
			}
			result = fixUnpredictableOverlap(result, m.predictableKeys)

			if styleBlock != "" {
				result = styleBlock + "\n\n" + result
			}

			if debug {
				fmt.Fprintf(os.Stderr, "[dcd-debug] %s attempt %d: validating...\n",
					provider.Name(), attempt+1)
			}

			valid, reason := isDCDValid(result)
			if valid {
				warnings, err := validateVarsAndKeys(result)
				if debug && len(warnings) > 0 {
					for _, w := range warnings {
						fmt.Fprintf(os.Stderr, "[dcd-debug] %s attempt %d: warning: %s\n",
							provider.Name(), attempt+1, w)
					}
				}
				if err != nil {
					lastErr = fmt.Errorf("%s attempt %d: %w", provider.Name(), attempt+1, err)
					if debug {
						fmt.Fprintf(os.Stderr, "[dcd-debug] %s attempt %d: validation failed: %s\n",
							provider.Name(), attempt+1, err)
					}
					continue
				}
				m.lastProvider = provider.Name()
				m.lastResult = result
				return result, nil
			}

			lastErr = fmt.Errorf("%s attempt %d: %s", provider.Name(), attempt+1, reason)
			if debug {
				fmt.Fprintf(os.Stderr, "[dcd-debug] %s attempt %d: %s\n",
					provider.Name(), attempt+1, reason)
				sanPath := filepath.Join("dcd_temp", fmt.Sprintf("dcd_debug_%s_attempt_%d_sanitized.dcd", provider.Name(), attempt+1))
				os.WriteFile(sanPath, []byte(result), 0644)
			}

			if debug {
				fmt.Fprintf(os.Stderr, "[dcd-debug] %s attempt %d: invalid — %s. Retrying...\n",
					provider.Name(), attempt+1, reason)
			}

			prompt = originalPrompt + fmt.Sprintf(
				"\n\nThe previous attempt failed: %s.\n"+
					"Invalid output:\n---\n%s\n---\n\n"+
					"Regenerate a valid DCD template, fixing the issue above:\n",
				reason, result,
			)
		}

		if pi < len(m.providers)-1 {
			prompt = originalPrompt
			continue
		}

		return "", fmt.Errorf("dcdmaker: all providers failed: %w", lastErr)
	}

	return "", fmt.Errorf("dcdmaker: no providers configured")
}

func writeWordsXMLDebug(name string, content string) {
	if os.Getenv("WORDS_DEBUG") != "true" {
		return
	}
	os.MkdirAll("dcd_temp", 0755)
	fpath := filepath.Join("dcd_temp", fmt.Sprintf("words_xml_%s.xml", name))
	os.WriteFile(fpath, []byte(content), 0644)
	fmt.Fprintf(os.Stderr, "[words-debug] saved words XML to %s (%d bytes)\n", fpath, len(content))
}

func extractStyleBlock(wordsXML string) string {
	start := strings.Index(wordsXML, "<style")
	if start < 0 {
		return ""
	}
	end := strings.Index(wordsXML[start:], "</style>")
	if end < 0 {
		return ""
	}
	return wordsXML[start : start+end+len("</style>")]
}

