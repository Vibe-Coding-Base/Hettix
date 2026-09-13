package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/oklog/ulid"
)

// LoadPluginEnabled returns the persisted enabled state for each plugin id.
func (d *Database) LoadPluginEnabled(ctx context.Context) (map[string]bool, error) {
	rows, err := d.db.QueryContext(ctx, `SELECT id, enabled FROM plugin_settings`)
	if err != nil {
		return nil, fmt.Errorf("sqlite: failed to query plugin settings: %w", err)
	}
	defer rows.Close()

	out := make(map[string]bool)
	for rows.Next() {
		var id string
		var enabled int
		if err := rows.Scan(&id, &enabled); err != nil {
			return nil, fmt.Errorf("sqlite: failed to scan plugin setting: %w", err)
		}
		out[id] = enabled != 0
	}

	return out, rows.Err()
}

// SavePluginEnabled upserts a plugin's enabled state.
func (d *Database) SavePluginEnabled(ctx context.Context, id string, enabled bool) error {
	_, err := d.db.ExecContext(ctx,
		`INSERT INTO plugin_settings (id, enabled) VALUES (?, ?)
		 ON CONFLICT (id) DO UPDATE SET enabled = excluded.enabled`,
		id, boolToInt(enabled),
	)
	if err != nil {
		return fmt.Errorf("sqlite: failed to save plugin setting: %w", err)
	}

	return nil
}

// DeletePluginEnabled removes a plugin's persisted enabled state.
func (d *Database) DeletePluginEnabled(ctx context.Context, id string) error {
	if _, err := d.db.ExecContext(ctx, `DELETE FROM plugin_settings WHERE id = ?`, id); err != nil {
		return fmt.Errorf("sqlite: failed to delete plugin setting: %w", err)
	}

	return nil
}

// StoreDiscoveredEndpoint records an endpoint a plugin discovered; duplicates
// (same project, host, path and source) are ignored.
func (d *Database) StoreDiscoveredEndpoint(ctx context.Context, projectID ulid.ULID, host, path, source string) error {
	_, err := d.db.ExecContext(ctx,
		`INSERT INTO discovered_endpoints (project_id, host, path, source, created_at)
		 VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT (project_id, host, path, source) DO NOTHING`,
		projectID.String(), host, path, source, time.Now().UnixMilli(),
	)
	if err != nil {
		return fmt.Errorf("sqlite: failed to store discovered endpoint: %w", err)
	}

	return nil
}

// LoadProxyPort returns the persisted proxy port, or ok=false when unset.
func (d *Database) LoadProxyPort(ctx context.Context) (int, bool, error) {
	var port int
	err := d.db.QueryRowContext(ctx, `SELECT port FROM proxy_settings WHERE id = 1`).Scan(&port)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("sqlite: failed to load proxy port: %w", err)
	}

	return port, true, nil
}

// SaveProxyPort persists the proxy port.
func (d *Database) SaveProxyPort(ctx context.Context, port int) error {
	_, err := d.db.ExecContext(ctx,
		`INSERT INTO proxy_settings (id, port) VALUES (1, ?)
		 ON CONFLICT (id) DO UPDATE SET port = excluded.port`,
		port,
	)
	if err != nil {
		return fmt.Errorf("sqlite: failed to save proxy port: %w", err)
	}

	return nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
