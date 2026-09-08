package sqlite_test

import (
	"context"
	"testing"

	"github.com/Vibe-Coding-Base/Hettix/pkg/llm"
)

func TestLLMSettingsRoundTrip(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	if _, ok, err := db.LoadLLMSettings(ctx); err != nil || ok {
		t.Fatalf("expected no settings initially, got ok=%v err=%v", ok, err)
	}

	want := llm.Settings{Provider: "openai", BaseURL: "https://api.openai.com/v1", APIKey: "sk-x", Model: "gpt-4o", Enabled: true}
	if err := db.SaveLLMSettings(ctx, want); err != nil {
		t.Fatalf("save: %v", err)
	}

	got, ok, err := db.LoadLLMSettings(ctx)
	if err != nil || !ok {
		t.Fatalf("load: ok=%v err=%v", ok, err)
	}
	if got != want {
		t.Fatalf("round trip mismatch: got %+v want %+v", got, want)
	}

	// Upsert replaces the single row.
	want.Model = "gpt-4o-mini"
	want.Enabled = false
	if err := db.SaveLLMSettings(ctx, want); err != nil {
		t.Fatalf("update: %v", err)
	}
	got, _, _ = db.LoadLLMSettings(ctx)
	if got.Model != "gpt-4o-mini" || got.Enabled {
		t.Fatalf("upsert not applied: %+v", got)
	}
}
