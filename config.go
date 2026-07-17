package dcdmaker

import "time"

type geminiConfig struct {
	APIKey      string
	Model       string
	Temperature float64
	Timeout     time.Duration
	Stream      bool
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

func WithStream(s bool) GeminiOption {
	return func(c *geminiConfig) { c.Stream = s }
}

func Gemini(opts ...GeminiOption) Provider {
	cfg := &geminiConfig{
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
	Stream      bool
}

type OpenAIOption func(*openAIConfig)

func WithOpenAIModel(m string) OpenAIOption {
	return func(c *openAIConfig) { c.Model = m }
}

func WithOpenAIAPIKey(k string) OpenAIOption {
	return func(c *openAIConfig) { c.APIKey = k }
}

func WithOpenAIBaseURL(u string) OpenAIOption {
	return func(c *openAIConfig) { c.BaseURL = u }
}

func WithOpenAIStream(s bool) OpenAIOption {
	return func(c *openAIConfig) { c.Stream = s }
}

func OpenAI(opts ...OpenAIOption) Provider {
	cfg := &openAIConfig{
		Model:       "gpt-4o",
		Temperature: 0.5,
		MaxTokens:   16384,
		Timeout:     60 * time.Second,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return newOpenAIProvider(cfg)
}
