package dcdmaker

import (
	"context"
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
	content, err := extractDocxContent(data)
	if err != nil {
		return "", fmt.Errorf("openai: extract docx: %w", err)
	}

	var b strings.Builder
	b.WriteString(prompt)
	b.WriteString("\n\n=== SOURCE DOCUMENT XML ===\n")
	b.WriteString(content.DocumentXML)
	b.WriteString("\n\n")
	return p.chat(ctx, []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: b.String()},
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
	if p.cfg.Stream {
		return p.chatStream(ctx, messages)
	}
	return p.chatSync(ctx, messages)
}

func (p *openAIProvider) chatSync(ctx context.Context, messages []openai.ChatCompletionMessage) (string, error) {
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

	return resp.Choices[0].Message.Content, nil
}

func (p *openAIProvider) chatStream(ctx context.Context, messages []openai.ChatCompletionMessage) (string, error) {
	client := p.getClient()

	ctx, cancel := context.WithTimeout(ctx, p.cfg.Timeout)
	defer cancel()

	req := openai.ChatCompletionRequest{
		Model:       p.cfg.Model,
		Temperature: float32(p.cfg.Temperature),
		MaxTokens:   p.cfg.MaxTokens,
		Messages:    messages,
		Stream:      true,
	}

	stream, err := client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return "", fmt.Errorf("openai: stream create: %w", err)
	}
	defer stream.Close()

	var result strings.Builder
	for {
		response, err := stream.Recv()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			if result.Len() > 0 {
				return result.String(), nil
			}
			return "", fmt.Errorf("openai: stream recv: %w", err)
		}
		if len(response.Choices) > 0 {
			result.WriteString(response.Choices[0].Delta.Content)
		}
	}

	if result.Len() == 0 {
		return "", fmt.Errorf("openai: empty stream response")
	}

	return result.String(), nil
}


