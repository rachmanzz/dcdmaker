package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/rachmanzz/dcdmaker"
)

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	fs := flag.NewFlagSet("dcdmaker", flag.ExitOnError)

	source := fs.String("source", "", "Path to source document (.docx)")
	output := fs.String("output", "", "Output path for .dcd template")
	prompt := fs.String("prompt", "", "Optional instruction for the AI")

	geminiKey := fs.String("gemini-key", "", "Gemini API key")
	geminiModel := fs.String("gemini-model", "", "Gemini model name (default: gemini-2.5-flash)")

	openaiKey := fs.String("openai-key", "", "OpenAI API key")
	openaiModel := fs.String("openai-model", "", "OpenAI model name")
	openaiBaseURL := fs.String("openai-base-url", "", "OpenAI-compatible base URL")

	noGemini := fs.Bool("no-gemini", false, "Disable Gemini provider")
	maxRetries := fs.Int("max-retries", 3, "Max retries per provider (env: DCD_MAX_RETRIES)")
	noOpenAI := fs.Bool("no-openai", false, "Disable OpenAI provider")
	stream := fs.Bool("stream", false, "Enable streaming (env: AI_STREAM, AI_OPENAPI_STREAM, AI_GEMINI_STREAM)")
	version := fs.Bool("version", false, "Show version")

	if err := fs.Parse(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if *version {
		fmt.Println("dcdmaker v0.1.5")
		return
	}

	if *source == "" || *output == "" {
		fmt.Fprintln(os.Stderr, "Usage: dcdmaker -source invoice.docx -output template.dcd [options]")
		fs.PrintDefaults()
		os.Exit(1)
	}

	var providers []dcdmaker.Provider

	if !*noGemini {
		key := *geminiKey
		if key == "" {
			key = envOr("GEMINI_API_KEY", envOr("AI_GEMINI_API_KEY", ""))
		}
		model := *geminiModel
		if model == "" {
			model = envOr("AI_GEMINI_MODEL", "gemini-2.5-flash")
		}
		opts := []dcdmaker.GeminiOption{
			dcdmaker.WithAPIKey(key),
			dcdmaker.WithModel(model),
		}
		geminiStream := *stream || envOr("AI_STREAM", "") == "true" || envOr("AI_GEMINI_STREAM", "") == "true"
		if geminiStream {
			opts = append(opts, dcdmaker.WithStream(true))
		}
		providers = append(providers, dcdmaker.Gemini(opts...))
	}

	if !*noOpenAI {
		key := *openaiKey
		if key == "" {
			key = envOr("OPENAI_API_KEY", "")
		}
		model := *openaiModel
		baseURL := *openaiBaseURL

		if key != "" || model != "" || baseURL != "" {
			opts := []dcdmaker.OpenAIOption{}
			opts = append(opts, dcdmaker.WithOpenAIAPIKey(key))
			if model != "" {
				opts = append(opts, dcdmaker.WithOpenAIModel(model))
			}
			if baseURL != "" {
				opts = append(opts, dcdmaker.WithOpenAIBaseURL(baseURL))
			}
			openaiStream := *stream || envOr("AI_STREAM", "") == "true" || envOr("AI_OPENAPI_STREAM", "") == "true"
			if openaiStream {
				opts = append(opts, dcdmaker.WithOpenAIStream(true))
			}
			providers = append(providers, dcdmaker.OpenAI(opts...))
		}
	}

	if len(providers) == 0 {
		fmt.Fprintln(os.Stderr, "error: no providers enabled (enable at least one)")
		os.Exit(1)
	}

	retries := *maxRetries
	if env := os.Getenv("DCD_MAX_RETRIES"); env != "" {
		if n, err := strconv.Atoi(env); err == nil {
			retries = n
		}
	}

	maker := dcdmaker.NewMaker(providers...)
	maker.Source(*source).OptionalPrompt(*prompt).MaxRetries(retries)

	if err := maker.Run(*output); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("DCD template generated:", *output)
}
