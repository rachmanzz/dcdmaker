package dcdmaker

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"google.golang.org/genai"
)

type geminiProvider struct {
	cfg    *geminiConfig
	client *genai.Client
}

func newGeminiProvider(cfg *geminiConfig) *geminiProvider {
	return &geminiProvider{cfg: cfg}
}

func (p *geminiProvider) Name() string {
	return fmt.Sprintf("gemini:%s", p.cfg.Model)
}

func (p *geminiProvider) getClient(ctx context.Context) (*genai.Client, error) {
	if p.client != nil {
		return p.client, nil
	}
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  p.cfg.APIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("gemini: create client: %w", err)
	}
	p.client = client
	return client, nil
}

func (p *geminiProvider) generateContent(ctx context.Context, parts []*genai.Part) (string, error) {
	if p.cfg.Stream {
		return p.generateContentStream(ctx, parts)
	}
	return p.generateContentSync(ctx, parts)
}

func (p *geminiProvider) generateContentSync(ctx context.Context, parts []*genai.Part) (string, error) {
	client, err := p.getClient(ctx)
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(ctx, p.cfg.Timeout)
	defer cancel()

	config := &genai.GenerateContentConfig{
		Temperature: genai.Ptr(float32(p.cfg.Temperature)),
	}

	contents := []*genai.Content{{Parts: parts}}
	result, err := client.Models.GenerateContent(ctx, p.cfg.Model, contents, config)
	if err != nil {
		return "", fmt.Errorf("gemini: generate: %w", err)
	}

	if len(result.Candidates) == 0 {
		return "", fmt.Errorf("gemini: no candidates")
	}

	cand := result.Candidates[0]
	if cand.Content == nil {
		return "", fmt.Errorf("gemini: empty content")
	}

	var out []string
	for _, part := range cand.Content.Parts {
		out = append(out, part.Text)
	}

	output := strings.Join(out, "")

	return output, nil
}

func (p *geminiProvider) generateContentStream(ctx context.Context, parts []*genai.Part) (string, error) {
	client, err := p.getClient(ctx)
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(ctx, p.cfg.Timeout)
	defer cancel()

	config := &genai.GenerateContentConfig{
		Temperature: genai.Ptr(float32(p.cfg.Temperature)),
	}

	contents := []*genai.Content{{Parts: parts}}
	iter := client.Models.GenerateContentStream(ctx, p.cfg.Model, contents, config)

	var result strings.Builder
	for resp, err := range iter {
		if err != nil {
			if result.Len() > 0 {
				return result.String(), nil
			}
			return "", fmt.Errorf("gemini: stream: %w", err)
		}
		if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
			for _, part := range resp.Candidates[0].Content.Parts {
				result.WriteString(part.Text)
			}
		}
	}

	if result.Len() == 0 {
		return "", fmt.Errorf("gemini: empty stream response")
	}

	return result.String(), nil
}

func (p *geminiProvider) Generate(ctx context.Context, prompt string) (string, error) {
	return p.generateContent(ctx, []*genai.Part{{Text: prompt}})
}

func (p *geminiProvider) GenerateWithFile(ctx context.Context, prompt string, _ string, data []byte) (string, error) {
	doc, err := ParseDOCX(data)
	if err != nil {
		return "", fmt.Errorf("gemini: parse docx: %w", err)
	}

	cleanedContent := doc.FormatForLLM()

	var b bytes.Buffer
	b.WriteString(prompt)
	b.WriteString("\n\n=== SOURCE DOCUMENT ===\n")
	b.WriteString(cleanedContent)
	b.WriteString("\n\n")
	return p.generateContent(ctx, []*genai.Part{{Text: b.String()}})
}

func (p *geminiProvider) GenerateWithHistory(ctx context.Context, history []Message, prompt string) (string, error) {
	client, err := p.getClient(ctx)
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(ctx, p.cfg.Timeout)
	defer cancel()

	config := &genai.GenerateContentConfig{
		Temperature: genai.Ptr(float32(p.cfg.Temperature)),
	}

	contents := make([]*genai.Content, 0, len(history)+1)
	for _, msg := range history {
		var role string
		switch msg.Role {
		case "assistant":
			role = "model"
		default:
			role = "user"
		}
		contents = append(contents, &genai.Content{
			Parts: []*genai.Part{{Text: msg.Content}},
			Role:  role,
		})
	}
	contents = append(contents, &genai.Content{
		Parts: []*genai.Part{{Text: prompt}},
		Role:  "user",
	})

	result, err := client.Models.GenerateContent(ctx, p.cfg.Model, contents, config)
	if err != nil {
		return "", fmt.Errorf("gemini: chat: %w", err)
	}

	if len(result.Candidates) == 0 {
		return "", fmt.Errorf("gemini: no candidates")
	}

	cand := result.Candidates[0]
	if cand.Content == nil {
		return "", fmt.Errorf("gemini: empty content")
	}

	var out []string
	for _, part := range cand.Content.Parts {
		out = append(out, part.Text)
	}

	output := strings.Join(out, "")

	return output, nil
}
