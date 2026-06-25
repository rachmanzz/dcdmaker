package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/rachmanzz/dcdmaker"
)

func main() {
	fs := flag.NewFlagSet("dcdmaker", flag.ExitOnError)

	source := fs.String("source", "", "Path to source document (.docx)")
	output := fs.String("output", "", "Output path for .dcd template")
	prompt := fs.String("prompt", "", "Optional instruction for the AI")

	geminiKey := fs.String("gemini-key", "", "Gemini API key (default: $GEMINI_API_KEY)")
	geminiModel := fs.String("gemini-model", "gemini-2.5-flash", "Gemini model name")

	openaiKey := fs.String("openai-key", "", "OpenAI API key (default: $OPENAI_API_KEY)")
	openaiModel := fs.String("openai-model", "", "OpenAI model name")
	openaiBaseURL := fs.String("openai-base-url", "", "OpenAI-compatible base URL")

	resume := fs.Bool("resume", false, "Resume from previous session")
	noGemini := fs.Bool("no-gemini", false, "Disable Gemini provider")
	noOpenAI := fs.Bool("no-openai", false, "Disable OpenAI provider")
	version := fs.Bool("version", false, "Show version")

	if err := fs.Parse(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if *version {
		fmt.Println("dcdmaker v0.1.0")
		return
	}

	if *source == "" || *output == "" {
		fmt.Fprintln(os.Stderr, "Usage: dcdmaker -source invoice.docx -output template.dcd [options]")
		fs.PrintDefaults()
		os.Exit(1)
	}

	var providers []dcdmaker.Provider

	if !*noGemini {
		opts := []dcdmaker.GeminiOption{
			dcdmaker.WithModel(*geminiModel),
		}
		if *geminiKey != "" {
			opts = append(opts, dcdmaker.WithAPIKey(*geminiKey))
		}
		providers = append(providers, dcdmaker.Gemini(opts...))
	}

	if !*noOpenAI {
		if *openaiModel != "" || *openaiKey != "" || *openaiBaseURL != "" {
			opts := []dcdmaker.OpenAIOption{}
			if *openaiModel != "" {
				opts = append(opts, dcdmaker.WithOpenAIModel(*openaiModel))
			}
			if *openaiKey != "" {
				opts = append(opts, dcdmaker.WithOpenAIAPIKey(*openaiKey))
			}
			if *openaiBaseURL != "" {
				opts = append(opts, dcdmaker.WithOpenAIBaseURL(*openaiBaseURL))
			}
			providers = append(providers, dcdmaker.OpenAI(opts...))
		}
	}

	if len(providers) == 0 {
		fmt.Fprintln(os.Stderr, "error: no providers enabled (enable at least one)")
		os.Exit(1)
	}

	maker := dcdmaker.NewMaker(providers...)
	maker.Source(*source).OptionalPrompt(*prompt).Resume(*resume)

	if err := maker.Run(*output); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("DCD template generated:", *output)
}
