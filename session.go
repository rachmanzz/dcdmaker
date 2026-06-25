package dcdmaker

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Session struct {
	Source        string    `json:"source"`
	Output        string    `json:"output"`
	UserPrompt    string    `json:"user_prompt"`
	ProviderName  string    `json:"provider_name"`
	History       []Message `json:"history"`
	PartialOutput string    `json:"partial_output"`
	Step          int       `json:"step"`
}

func sessionPath(output string) string {
	dir := filepath.Dir(output)
	base := filepath.Base(output)
	name := base + ".session.json"
	return filepath.Join(dir, name)
}

func saveSession(path string, m *Maker, result string, history []Message) error {
	s := &Session{
		Source:        m.source,
		Output:        path,
		UserPrompt:    m.userPrompt,
		PartialOutput: result,
		Step:          1,
		History:       history,
	}
	if len(m.providers) > 0 {
		s.ProviderName = m.providers[0].Name()
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("session marshal: %w", err)
	}

	return os.WriteFile(sessionPath(path), data, 0644)
}

func loadSession(output string) (*Session, error) {
	data, err := os.ReadFile(sessionPath(output))
	if err != nil {
		return nil, err
	}

	var s Session
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("session unmarshal: %w", err)
	}

	return &s, nil
}

func clearSession(output string) error {
	path := sessionPath(output)
	if _, err := os.Stat(path); err == nil {
		return os.Remove(path)
	}
	return nil
}

func continuationPrompt(partial string) string {
	return fmt.Sprintf(
		"The DCD template was truncated. Continue exactly from where it stopped. "+
			"Do NOT repeat any part that was already generated. Start from the incomplete section.\n\n"+
			"Partial output so far:\n%s\n\nContinue now:",
		partial,
	)
}
