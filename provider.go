package dcdmaker

import "context"

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Provider interface {
	Name() string
	Generate(ctx context.Context, prompt string) (string, error)
	GenerateWithFile(ctx context.Context, prompt string, filename string, data []byte) (string, error)
	GenerateWithHistory(ctx context.Context, history []Message, prompt string) (string, error)
}
