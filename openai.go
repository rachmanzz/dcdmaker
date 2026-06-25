package dcdmaker

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"
)

type openAIProvider struct {
	cfg    *openAIConfig
	client *openai.Client
}

func newOpenAIProvider(cfg *openAIConfig) *openAIProvider {
	return &openAIProvider{cfg: cfg}
}

func (p *openAIProvider) Name() string {
	return fmt.Sprintf("openai:%s", p.cfg.Model)
}

func (p *openAIProvider) getClient() *openai.Client {
	if p.client == nil {
		if p.cfg.BaseURL != "" {
			cfg := openai.DefaultConfig(p.cfg.APIKey)
			cfg.BaseURL = p.cfg.BaseURL
			p.client = openai.NewClientWithConfig(cfg)
		} else {
			p.client = openai.NewClient(p.cfg.APIKey)
		}
	}
	return p.client
}

func (p *openAIProvider) Generate(ctx context.Context, prompt string) (string, error) {
	return p.chat(ctx, []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: prompt},
	})
}

func (p *openAIProvider) GenerateWithFile(ctx context.Context, prompt string, _ string, data []byte) (string, error) {
	text := extractDocxText(data)
	full := prompt + "\n\n=== SOURCE DOCUMENT ===\n" + text
	return p.chat(ctx, []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: full},
	})
}

func (p *openAIProvider) GenerateWithHistory(ctx context.Context, history []Message, prompt string) (string, error) {
	messages := make([]openai.ChatCompletionMessage, 0, len(history)+1)
	for _, msg := range history {
		var role string
		switch msg.Role {
		case "assistant":
			role = openai.ChatMessageRoleAssistant
		default:
			role = openai.ChatMessageRoleUser
		}
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    role,
			Content: msg.Content,
		})
	}
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: prompt,
	})
	return p.chat(ctx, messages)
}

func (p *openAIProvider) chat(ctx context.Context, messages []openai.ChatCompletionMessage) (string, error) {
	client := p.getClient()

	ctx, cancel := context.WithTimeout(ctx, p.cfg.Timeout)
	defer cancel()

	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       p.cfg.Model,
		Temperature: float32(p.cfg.Temperature),
		MaxTokens:   p.cfg.MaxTokens,
		Messages:    messages,
	})
	if err != nil {
		return "", fmt.Errorf("openai: generate: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("openai: no choices")
	}

	result := resp.Choices[0].Message.Content
	if resp.Choices[0].FinishReason == "length" {
		result += "\n\n<TRUNCATED/>"
	}

	return result, nil
}

func extractDocxText(data []byte) string {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return ""
	}
	for _, f := range r.File {
		if f.Name != "word/document.xml" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return ""
		}
		defer rc.Close()

		dec := xml.NewDecoder(rc)
		var text []string
		inT := false
		for {
			tok, err := dec.Token()
			if err != nil {
				break
			}
			switch t := tok.(type) {
			case xml.StartElement:
				if t.Name.Local == "t" {
					inT = true
				}
			case xml.CharData:
				if inT {
					text = append(text, string(t))
				}
			case xml.EndElement:
				if t.Name.Local == "t" {
					inT = false
				}
			}
		}
		return strings.Join(text, "")
	}
	return ""
}
