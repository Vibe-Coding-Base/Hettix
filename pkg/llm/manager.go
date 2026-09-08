package llm

import (
	"context"
	"strings"
	"sync"
)

// Settings holds the operator-configurable LLM connection. A single
// OpenAI-compatible endpoint covers DeepSeek, OpenAI, Ollama and similar.
type Settings struct {
	Provider string // label, e.g. "deepseek", "openai", "ollama"
	BaseURL  string
	APIKey   string
	Model    string
	Enabled  bool
}

// Configured reports whether the settings are complete enough to build a
// provider.
func (s Settings) Configured() bool {
	return s.Enabled && s.BaseURL != "" && s.Model != ""
}

// SettingsRepository persists LLM settings.
type SettingsRepository interface {
	LoadLLMSettings(ctx context.Context) (Settings, bool, error)
	SaveLLMSettings(ctx context.Context, settings Settings) error
}

// Manager holds the active LLM settings and the provider built from them,
// rebuilding the provider whenever the settings change. It is safe for
// concurrent use.
type Manager struct {
	mu       sync.RWMutex
	settings Settings
	provider Provider
	repo     SettingsRepository
}

// NewManager returns a Manager. repo may be nil for an unpersisted manager.
func NewManager(repo SettingsRepository) *Manager {
	return &Manager{repo: repo}
}

// Load reads persisted settings (if any) and builds the provider from them.
func (m *Manager) Load(ctx context.Context) error {
	if m.repo == nil {
		return nil
	}

	settings, ok, err := m.repo.LoadLLMSettings(ctx)
	if err != nil {
		return err
	}

	if ok {
		m.set(settings)
	}

	return nil
}

// Seed applies settings only when none are configured yet, persisting them.
// It is used to carry environment configuration into the store on first run.
func (m *Manager) Seed(ctx context.Context, settings Settings) error {
	m.mu.RLock()
	alreadyConfigured := m.settings.BaseURL != "" || m.settings.Model != ""
	m.mu.RUnlock()

	if alreadyConfigured || !settings.Configured() {
		return nil
	}

	return m.Update(ctx, settings)
}

// Update applies and persists new settings, rebuilding the provider.
func (m *Manager) Update(ctx context.Context, settings Settings) error {
	settings.Provider = strings.TrimSpace(settings.Provider)
	settings.BaseURL = strings.TrimSpace(settings.BaseURL)
	settings.Model = strings.TrimSpace(settings.Model)

	if m.repo != nil {
		if err := m.repo.SaveLLMSettings(ctx, settings); err != nil {
			return err
		}
	}

	m.set(settings)

	return nil
}

func (m *Manager) set(settings Settings) {
	var provider Provider

	if settings.Configured() {
		name := settings.Provider
		if name == "" {
			name = "llm"
		}

		provider = NewOpenAICompatible(name, OpenAIConfig{
			BaseURL: settings.BaseURL,
			APIKey:  settings.APIKey,
			Model:   settings.Model,
		})
	}

	m.mu.Lock()
	m.settings = settings
	m.provider = provider
	m.mu.Unlock()
}

// Provider returns the current provider, or nil when the assistant is not
// configured.
func (m *Manager) Provider() Provider {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.provider
}

// Settings returns the current settings.
func (m *Manager) Settings() Settings {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.settings
}
