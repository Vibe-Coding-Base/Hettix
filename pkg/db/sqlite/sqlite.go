// Package sqlite provides a SQLite-backed implementation of the reqlog, sender
// and proj repositories. It replaces the previous bbolt storage: request and
// response logs are stored in normalized tables so that traffic queries can be
// pushed down to SQL, while project settings (which contain regexp and filter
// expression values that don't serialize cleanly) are kept as an opaque blob.
package sqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	_ "modernc.org/sqlite"
)

// Database is a SQLite-backed store implementing the repository interfaces of
// the reqlog, sender and proj packages.
type Database struct {
	db *sql.DB
}

const schema = `
CREATE TABLE IF NOT EXISTS projects (
	id         TEXT PRIMARY KEY,
	name       TEXT NOT NULL,
	data       BLOB NOT NULL,
	created_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS http_request_logs (
	id         TEXT PRIMARY KEY,
	project_id TEXT NOT NULL,
	method     TEXT NOT NULL,
	url        TEXT NOT NULL,
	host       TEXT NOT NULL,
	path       TEXT NOT NULL,
	proto      TEXT NOT NULL,
	headers    TEXT NOT NULL,
	body       BLOB,
	created_at INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_request_logs_project ON http_request_logs (project_id, id);

CREATE TABLE IF NOT EXISTS http_response_logs (
	request_log_id TEXT PRIMARY KEY,
	proto          TEXT NOT NULL,
	status_code    INTEGER NOT NULL,
	status_reason  TEXT NOT NULL,
	headers        TEXT NOT NULL,
	body           BLOB,
	roundtrip_ms   INTEGER NOT NULL,
	FOREIGN KEY (request_log_id) REFERENCES http_request_logs (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS sender_requests (
	id                    TEXT PRIMARY KEY,
	project_id            TEXT NOT NULL,
	source_request_log_id TEXT NOT NULL,
	method                TEXT NOT NULL,
	url                   TEXT NOT NULL,
	host                  TEXT NOT NULL,
	path                  TEXT NOT NULL,
	proto                 TEXT NOT NULL,
	headers               TEXT NOT NULL,
	body                  BLOB,
	created_at            INTEGER NOT NULL,
	res_proto             TEXT,
	res_status_code       INTEGER,
	res_status_reason     TEXT,
	res_headers           TEXT,
	res_body              BLOB,
	res_roundtrip_ms      INTEGER
);
CREATE INDEX IF NOT EXISTS idx_sender_requests_project ON sender_requests (project_id, id);

CREATE TABLE IF NOT EXISTS websocket_connections (
	id         TEXT PRIMARY KEY,
	project_id TEXT NOT NULL,
	url        TEXT NOT NULL,
	host       TEXT NOT NULL,
	path       TEXT NOT NULL,
	created_at INTEGER NOT NULL,
	closed_at  INTEGER
);
CREATE INDEX IF NOT EXISTS idx_ws_connections_project ON websocket_connections (project_id, id);

CREATE TABLE IF NOT EXISTS websocket_messages (
	id            INTEGER PRIMARY KEY AUTOINCREMENT,
	connection_id TEXT NOT NULL,
	direction     INTEGER NOT NULL,
	opcode        INTEGER NOT NULL,
	payload       BLOB,
	created_at    INTEGER NOT NULL,
	FOREIGN KEY (connection_id) REFERENCES websocket_connections (id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_ws_messages_connection ON websocket_messages (connection_id, id);
`

// OpenDatabase opens (creating if needed) a SQLite database at path and ensures
// the schema exists.
func OpenDatabase(path string) (*Database, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("sqlite: failed to open database: %w", err)
	}

	db.SetMaxOpenConns(1)

	if _, err := db.Exec("PRAGMA journal_mode = WAL; PRAGMA foreign_keys = ON;"); err != nil {
		db.Close()
		return nil, fmt.Errorf("sqlite: failed to set pragmas: %w", err)
	}

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("sqlite: failed to apply schema: %w", err)
	}

	return &Database{db: db}, nil
}

// Close closes the underlying database.
func (d *Database) Close() error {
	return d.db.Close()
}

// paginate applies an offset and limit to an already-filtered slice. A limit of
// zero means no limit. Slicing happens after in-memory refinement so the offset
// and limit count exact matches, not push-down candidates.
func paginate[T any](items []T, offset, limit int) []T {
	if offset < 0 {
		offset = 0
	}

	if offset >= len(items) {
		return nil
	}

	items = items[offset:]

	if limit > 0 && limit < len(items) {
		items = items[:limit]
	}

	return items
}

func marshalHeader(h http.Header) (string, error) {
	if h == nil {
		return "{}", nil
	}

	b, err := json.Marshal(h)
	if err != nil {
		return "", fmt.Errorf("sqlite: failed to marshal header: %w", err)
	}

	return string(b), nil
}

func unmarshalHeader(s string) (http.Header, error) {
	if s == "" {
		return nil, nil
	}

	var h http.Header
	if err := json.Unmarshal([]byte(s), &h); err != nil {
		return nil, fmt.Errorf("sqlite: failed to unmarshal header: %w", err)
	}

	return h, nil
}
