package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Vibe-Coding-Base/Hettix/pkg/llm"
)

// LoadLLMSettings returns the stored LLM settings. The second return value is
// false when no settings have been saved yet.
func (d *Database) LoadLLMSettings(ctx context.Context) (llm.Settings, bool, error) {
	row := d.db.QueryRowContext(ctx,
		`SELECT provider, base_url, api_key, model, enabled FROM llm_settings WHERE id = 1`)

	var (
		settings llm.Settings
		enabled  int
	)

	err := row.Scan(&settings.Provider, &settings.BaseURL, &settings.APIKey, &settings.Model, &enabled)
	if errors.Is(err, sql.ErrNoRows) {
		return llm.Settings{}, false, nil
	}
	if err != nil {
		return llm.Settings{}, false, fmt.Errorf("sqlite: failed to load LLM settings: %w", err)
	}

	settings.Enabled = enabled != 0

	return settings, true, nil
}

// SaveLLMSettings stores the LLM settings as the single settings row.
func (d *Database) SaveLLMSettings(ctx context.Context, settings llm.Settings) error {
	enabled := 0
	if settings.Enabled {
		enabled = 1
	}

	_, err := d.db.ExecContext(ctx,
		`INSERT INTO llm_settings (id, provider, base_url, api_key, model, enabled)
		 VALUES (1, ?, ?, ?, ?, ?)
		 ON CONFLICT (id) DO UPDATE SET
		   provider = excluded.provider, base_url = excluded.base_url, api_key = excluded.api_key,
		   model = excluded.model, enabled = excluded.enabled`,
		settings.Provider, settings.BaseURL, settings.APIKey, settings.Model, enabled,
	)
	if err != nil {
		return fmt.Errorf("sqlite: failed to save LLM settings: %w", err)
	}

	return nil
}
