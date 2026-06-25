package dcdmaker

import (
	"context"
	"fmt"
	"os"
	"strings"
)

type VarType int

const (
	VarObject VarType = iota
	VarArray
)

type KeyDef struct {
	Name   string
	Type   VarType
	Fields []string
}

func Object(name string, fields ...string) KeyDef {
	return KeyDef{Name: name, Type: VarObject, Fields: fields}
}

func Array(name string, fields ...string) KeyDef {
	return KeyDef{Name: name, Type: VarArray, Fields: fields}
}

type Maker struct {
	providers       []Provider
	source          string
	userPrompt      string
	resume          bool
	predictableKeys []KeyDef
}

func NewMaker(providers ...Provider) *Maker {
	return &Maker{
		providers: providers,
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

func (m *Maker) Resume(enabled bool) *Maker {
	m.resume = enabled
	return m
}

func (m *Maker) PredictableKeys(keys ...KeyDef) *Maker {
	m.predictableKeys = keys
	return m
}

func (m *Maker) Run(output string) error {
	if len(m.providers) == 0 {
		return fmt.Errorf("dcdmaker: at least one provider required")
	}

	if m.source == "" {
		return fmt.Errorf("dcdmaker: source document required")
	}

	if m.resume {
		session, err := loadSession(output)
		if err == nil && session.PartialOutput != "" {
			return m.resumeSession(session, output)
		}
	}

	data, err := os.ReadFile(m.source)
	if err != nil {
		return fmt.Errorf("dcdmaker: read source: %w", err)
	}

	prompt := buildPrompt(m.userPrompt, m.predictableKeys)

	ctx := context.Background()

	for pi, provider := range m.providers {
		var lastErr error

		for attempt := range 3 {
			result, err := provider.GenerateWithFile(ctx, prompt, m.source, data)
			if err != nil {
				lastErr = fmt.Errorf("%s attempt %d: %w", provider.Name(), attempt+1, err)
				continue
			}

			result = resolveChunks(ctx, provider, result)
			result = sanitizeDCD(result)

			if isDCDValid(result) {
				if m.resume {
					_ = clearSession(output)
				}
				return os.WriteFile(output, []byte(result), 0644)
			}

			lastErr = fmt.Errorf("%s attempt %d: invalid DCD output", provider.Name(), attempt+1)
			prompt = fmt.Sprintf(
				"The previous output was not valid DCD syntax. "+
					"Output ONLY valid DCD template, no explanations.\n\n"+
					"Invalid output:\n---\n%s\n---\n\nRegenerate:",
				result,
			)
		}

		if pi < len(m.providers)-1 {
			prompt = buildPrompt(m.userPrompt, m.predictableKeys)
			continue
		}

		return fmt.Errorf("dcdmaker: all providers failed: %w", lastErr)
	}

	return fmt.Errorf("dcdmaker: no providers configured")
}

func (m *Maker) resumeSession(session *Session, output string) error {
	if len(m.providers) == 0 {
		return fmt.Errorf("dcdmaker: at least one provider required")
	}

	provider := m.providers[0]
	for _, p := range m.providers {
		if p.Name() == session.ProviderName {
			provider = p
			break
		}
	}

	prompt := continuationPrompt(session.PartialOutput)

	ctx := context.Background()

	var full strings.Builder
	full.WriteString(session.PartialOutput)

	for attempt := range 3 {
		result, err := provider.GenerateWithHistory(ctx, session.History, prompt)
		if err != nil {
			if attempt < 2 {
				continue
			}
			return fmt.Errorf("dcdmaker: resume failed: %w", err)
		}

		full.WriteString(result)

		if isTruncated(full.String()) {
			prompt = continuationPrompt(full.String())
			continue
		}

		dcd := sanitizeDCD(full.String())
		if isDCDValid(dcd) {
			if m.resume {
				_ = clearSession(output)
			}
			return os.WriteFile(output, []byte(dcd), 0644)
		}

		return fmt.Errorf("dcdmaker: resume produced invalid DCD")
	}

	return fmt.Errorf("dcdmaker: resume exhausted retries")
}

func resolveChunks(ctx context.Context, provider Provider, result string) string {
	var full strings.Builder
	full.WriteString(result)

	for range 3 {
		if !isTruncated(full.String()) {
			break
		}

		clean := strings.TrimSuffix(full.String(), "\n\n<TRUNCATED/>")
		prompt := continuationPrompt(clean)

		chunk, err := provider.Generate(ctx, prompt)
		if err != nil {
			break
		}
		full.WriteString(chunk)
	}

	return full.String()
}
