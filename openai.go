package dcdmaker

import (
	"context"
	"fmt"
	"strings"

	"github.com/rachmanzz/words-xml/words"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type openAIProvider struct {
	cfg          *openAIConfig
	client       openai.Client
	clientReady  bool
}

func newOpenAIProvider(cfg *openAIConfig) *openAIProvider {
	return &openAIProvider{cfg: cfg}
}

func (p *openAIProvider) Name() string {
	return fmt.Sprintf("openai:%s", p.cfg.Model)
}

func (p *openAIProvider) getClient() openai.Client {
	if !p.clientReady {
		opts := []option.RequestOption{
			option.WithAPIKey(p.cfg.APIKey),
		}
		if p.cfg.BaseURL != "" {
			opts = append(opts, option.WithBaseURL(p.cfg.BaseURL))
		}
		p.client = openai.NewClient(opts...)
		p.clientReady = true
	}
	return p.client
}

func (p *openAIProvider) GenerateWithFile(ctx context.Context, prompt string, _ string, data []byte) (string, error) {
	doc, err := words.ProcessDOCXBytes(data)
	if err != nil {
		return "", fmt.Errorf("openai: parse docx: %w", err)
	}

	cleanedContent := doc.WordsXML
	writeWordsXMLDebug(p.Name(), cleanedContent)

	var b strings.Builder
	b.WriteString(prompt)
	b.WriteString("\n\n=== SOURCE DOCUMENT ===\n")
	b.WriteString(cleanedContent)
	b.WriteString("\n\n")
	return p.chat(ctx, []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage(b.String()),
	})
}

func (p *openAIProvider) GenerateWithHistory(ctx context.Context, history []Message, prompt string) (string, error) {
	messages := make([]openai.ChatCompletionMessageParamUnion, 0, len(history)+1)
	for _, msg := range history {
		switch msg.Role {
		case "assistant":
			messages = append(messages, openai.AssistantMessage(msg.Content))
		default:
			messages = append(messages, openai.UserMessage(msg.Content))
		}
	}
	messages = append(messages, openai.UserMessage(prompt))
	return p.chat(ctx, messages)
}

func (p *openAIProvider) chat(ctx context.Context, messages []openai.ChatCompletionMessageParamUnion) (string, error) {
	if p.cfg.Stream {
		return p.chatStream(ctx, messages)
	}
	return p.chatSync(ctx, messages)
}

func (p *openAIProvider) chatSync(ctx context.Context, messages []openai.ChatCompletionMessageParamUnion) (string, error) {
	client := p.getClient()

	ctx, cancel := context.WithTimeout(ctx, p.cfg.Timeout)
	defer cancel()

	resp, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:       openai.ChatModel(p.cfg.Model),
		Temperature: openai.Float(p.cfg.Temperature),
		MaxTokens:   openai.Int(int64(p.cfg.MaxTokens)),
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

func (p *openAIProvider) chatStream(ctx context.Context, messages []openai.ChatCompletionMessageParamUnion) (string, error) {
	client := p.getClient()

	ctx, cancel := context.WithTimeout(ctx, p.cfg.Timeout)
	defer cancel()

	stream := client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
		Model:       openai.ChatModel(p.cfg.Model),
		Temperature: openai.Float(p.cfg.Temperature),
		MaxTokens:   openai.Int(int64(p.cfg.MaxTokens)),
		Messages:    messages,
	})

	var result strings.Builder
	for stream.Next() {
		event := stream.Current()
		if len(event.Choices) > 0 {
			result.WriteString(event.Choices[0].Delta.Content)
		}
	}
	if stream.Err() != nil {
		if result.Len() > 0 {
			return result.String(), nil
		}
		return "", fmt.Errorf("openai: stream: %w", stream.Err())
	}

	if result.Len() == 0 {
		return "", fmt.Errorf("openai: empty stream response")
	}

	return result.String(), nil
}
