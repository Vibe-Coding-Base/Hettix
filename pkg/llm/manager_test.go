package llm_test

import (
	"context"
	"testing"

	"github.com/Vibe-Coding-Base/Hettix/pkg/llm"
)

type memRepo struct {
	settings llm.Settings
	ok       bool
}

func (m *memRepo) LoadLLMSettings(context.Context) (llm.Settings, bool, error) {
	return m.settings, m.ok, nil
}

func (m *memRepo) SaveLLMSettings(_ context.Context, s llm.Settings) error {
	m.settings, m.ok = s, true
	return nil
}

func TestManagerUpdateBuildsProvider(t *testing.T) {
	repo := &memRepo{}
	m := llm.NewManager(repo)

	if m.Provider() != nil {
		t.Fatal("expected no provider before configuration")
	}

	err := m.Update(context.Background(), llm.Settings{
		Provider: "deepseek", BaseURL: "https://api.deepseek.com/v1", Model: "deepseek-chat", APIKey: "k", Enabled: true,
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if m.Provider() == nil {
		t.Fatal("expected a provider after configuration")
	}
	if !repo.ok || repo.settings.Model != "deepseek-chat" {
		t.Fatalf("settings not persisted: %+v", repo.settings)
	}

	// Disabling removes the provider.
	if err := m.Update(context.Background(), llm.Settings{BaseURL: "x", Model: "y", Enabled: false}); err != nil {
		t.Fatal(err)
	}
	if m.Provider() != nil {
		t.Fatal("expected no provider when disabled")
	}
}

func TestManagerSeedOnlyWhenEmpty(t *testing.T) {
	repo := &memRepo{}
	m := llm.NewManager(repo)

	// Seed applies when nothing is configured.
	if err := m.Seed(context.Background(), llm.Settings{BaseURL: "u", Model: "m", Enabled: true}); err != nil {
		t.Fatal(err)
	}
	if m.Settings().Model != "m" {
		t.Fatalf("seed should have applied, got %+v", m.Settings())
	}

	// A second seed must not overwrite existing settings.
	if err := m.Seed(context.Background(), llm.Settings{BaseURL: "other", Model: "other", Enabled: true}); err != nil {
		t.Fatal(err)
	}
	if m.Settings().Model != "m" {
		t.Fatalf("seed overwrote existing settings: %+v", m.Settings())
	}
}
