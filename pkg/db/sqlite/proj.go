package sqlite

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/gob"
	"errors"
	"fmt"

	"github.com/oklog/ulid"

	"github.com/Vibe-Coding-Base/Hettix/pkg/proj"
)

// Project settings hold regexp and filter-expression values that don't map onto
// SQL columns, so the whole project is persisted as a gob blob alongside its id
// and name. Request and response logs, which are queried, live in their own
// normalized tables.

func (d *Database) UpsertProject(ctx context.Context, project proj.Project) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(project); err != nil {
		return fmt.Errorf("sqlite: failed to encode project: %w", err)
	}

	_, err := d.db.ExecContext(ctx,
		`INSERT INTO projects (id, name, data, created_at) VALUES (?, ?, ?, ?)
		 ON CONFLICT (id) DO UPDATE SET name = excluded.name, data = excluded.data`,
		project.ID.String(), project.Name, buf.Bytes(), int64(project.ID.Time()),
	)
	if err != nil {
		return fmt.Errorf("sqlite: failed to upsert project: %w", err)
	}

	return nil
}

func (d *Database) FindProjectByID(ctx context.Context, id ulid.ULID) (proj.Project, error) {
	row := d.db.QueryRowContext(ctx, `SELECT data FROM projects WHERE id = ?`, id.String())

	var data []byte
	err := row.Scan(&data)
	if errors.Is(err, sql.ErrNoRows) {
		return proj.Project{}, proj.ErrProjectNotFound
	}
	if err != nil {
		return proj.Project{}, fmt.Errorf("sqlite: failed to find project: %w", err)
	}

	return decodeProject(data)
}

func (d *Database) DeleteProject(ctx context.Context, id ulid.ULID) error {
	_, err := d.db.ExecContext(ctx, `DELETE FROM projects WHERE id = ?`, id.String())
	if err != nil {
		return fmt.Errorf("sqlite: failed to delete project: %w", err)
	}

	return nil
}

func (d *Database) Projects(ctx context.Context) ([]proj.Project, error) {
	rows, err := d.db.QueryContext(ctx, `SELECT data FROM projects ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("sqlite: failed to query projects: %w", err)
	}
	defer rows.Close()

	var projects []proj.Project

	for rows.Next() {
		var data []byte
		if err := rows.Scan(&data); err != nil {
			return nil, fmt.Errorf("sqlite: failed to scan project: %w", err)
		}

		project, err := decodeProject(data)
		if err != nil {
			return nil, err
		}

		projects = append(projects, project)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("sqlite: failed to iterate projects: %w", err)
	}

	return projects, nil
}

func decodeProject(data []byte) (proj.Project, error) {
	var project proj.Project
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&project); err != nil {
		return proj.Project{}, fmt.Errorf("sqlite: failed to decode project: %w", err)
	}

	return project, nil
}
