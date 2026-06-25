package dcdmaker

import (
	"os"
	"time"
)

type geminiConfig struct {
	APIKey      string
	Model       string
	Temperature float64
	Timeout     time.Duration
}

type GeminiOption func(*geminiConfig)

func WithModel(m string) GeminiOption {
	return func(c *geminiConfig) { c.Model = m }
}

func WithTemperature(t float64) GeminiOption {
	return func(c *geminiConfig) { c.Temperature = t }
}

func WithAPIKey(k string) GeminiOption {
	return func(c *geminiConfig) { c.APIKey = k }
}

func WithTimeout(d time.Duration) GeminiOption {
	return func(c *geminiConfig) { c.Timeout = d }
}

func Gemini(opts ...GeminiOption) Provider {
	cfg := &geminiConfig{
		APIKey:      os.Getenv("GEMINI_API_KEY"),
		Model:       "gemini-2.5-flash",
		Temperature: 0.5,
		Timeout:     60 * time.Second,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return newGeminiProvider(cfg)
}

type openAIConfig struct {
	APIKey      string
	BaseURL     string
	Model       string
	Temperature float64
	MaxTokens   int
	Timeout     time.Duration
}

type OpenAIOption func(*openAIConfig)

func WithOpenAIModel(m string) OpenAIOption {
	return func(c *openAIConfig) { c.Model = m }
}

func WithOpenAITemperature(t float64) OpenAIOption {
	return func(c *openAIConfig) { c.Temperature = t }
}

func WithOpenAIAPIKey(k string) OpenAIOption {
	return func(c *openAIConfig) { c.APIKey = k }
}

func WithOpenAIMaxTokens(n int) OpenAIOption {
	return func(c *openAIConfig) { c.MaxTokens = n }
}

func WithOpenAIBaseURL(u string) OpenAIOption {
	return func(c *openAIConfig) { c.BaseURL = u }
}

func WithOpenAITimeout(d time.Duration) OpenAIOption {
	return func(c *openAIConfig) { c.Timeout = d }
}

func OpenAI(opts ...OpenAIOption) Provider {
	cfg := &openAIConfig{
		APIKey:      os.Getenv("OPENAI_API_KEY"),
		Model:       "gpt-4o",
		Temperature: 0.5,
		MaxTokens:   8192,
		Timeout:     60 * time.Second,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return newOpenAIProvider(cfg)
}
