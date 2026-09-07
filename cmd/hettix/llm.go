package main

import (
	"os"

	"github.com/Vibe-Coding-Base/Hettix/pkg/llm"
)

// llmProviderFromEnv builds an LLM provider from HETTIX_LLM_* environment
// variables, or returns nil when the assistant is not configured. A single
// OpenAI-compatible endpoint covers DeepSeek, OpenAI, Ollama and similar:
//
//	HETTIX_LLM_BASE_URL   e.g. https://api.deepseek.com/v1
//	HETTIX_LLM_API_KEY    provider API key (omit for local Ollama)
//	HETTIX_LLM_MODEL      e.g. deepseek-chat
//	HETTIX_LLM_PROVIDER   optional label shown in logs (default "llm")
func llmProviderFromEnv() llm.Provider {
	baseURL := os.Getenv("HETTIX_LLM_BASE_URL")
	model := os.Getenv("HETTIX_LLM_MODEL")

	if baseURL == "" || model == "" {
		return nil
	}

	name := os.Getenv("HETTIX_LLM_PROVIDER")
	if name == "" {
		name = "llm"
	}

	return llm.NewOpenAICompatible(name, llm.OpenAIConfig{
		BaseURL: baseURL,
		APIKey:  os.Getenv("HETTIX_LLM_API_KEY"),
		Model:   model,
	})
}
