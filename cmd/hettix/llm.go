package main

import (
	"os"

	"github.com/Vibe-Coding-Base/Hettix/pkg/llm"
)

// llmSettingsFromEnv reads LLM settings from HETTIX_LLM_* environment variables.
// They seed the stored settings on first run; afterwards the Settings page is
// the source of truth. A single OpenAI-compatible endpoint covers DeepSeek,
// OpenAI, Ollama and similar:
//
//	HETTIX_LLM_BASE_URL   e.g. https://api.deepseek.com/v1
//	HETTIX_LLM_API_KEY    provider API key (omit for local Ollama)
//	HETTIX_LLM_MODEL      e.g. deepseek-chat
//	HETTIX_LLM_PROVIDER   optional label (default "llm")
func llmSettingsFromEnv() llm.Settings {
	name := os.Getenv("HETTIX_LLM_PROVIDER")
	if name == "" {
		name = "llm"
	}

	return llm.Settings{
		Provider: name,
		BaseURL:  os.Getenv("HETTIX_LLM_BASE_URL"),
		APIKey:   os.Getenv("HETTIX_LLM_API_KEY"),
		Model:    os.Getenv("HETTIX_LLM_MODEL"),
		Enabled:  true,
	}
}
