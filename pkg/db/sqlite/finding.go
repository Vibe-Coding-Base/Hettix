package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/oklog/ulid"

	"github.com/Vibe-Coding-Base/Hettix/pkg/finding"
)

func (d *Database) StoreFinding(ctx context.Context, f finding.Finding) error {
	var requestLogID sql.NullString
	if f.RequestLogID != nil {
		requestLogID = sql.NullString{String: f.RequestLogID.String(), Valid: true}
	}

	_, err := d.db.ExecContext(ctx,
		`INSERT INTO findings (id, project_id, title, description, severity, request_log_id, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		f.ID.String(), f.ProjectID.String(), f.Title, f.Description, string(f.Severity),
		requestLogID, f.CreatedAt.UnixMilli(),
	)
	if err != nil {
		return fmt.Errorf("sqlite: failed to store finding: %w", err)
	}

	return nil
}

const findingSelect = `SELECT id, project_id, title, description, severity, request_log_id, created_at FROM findings`

func (d *Database) FindFindings(ctx context.Context, projectID ulid.ULID) ([]finding.Finding, error) {
	rows, err := d.db.QueryContext(ctx,
		findingSelect+` WHERE project_id = ? ORDER BY id DESC`, projectID.String())
	if err != nil {
		return nil, fmt.Errorf("sqlite: failed to query findings: %w", err)
	}
	defer rows.Close()

	var findings []finding.Finding

	for rows.Next() {
		f, err := scanFinding(rows)
		if err != nil {
			return nil, err
		}

		findings = append(findings, f)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("sqlite: failed to iterate findings: %w", err)
	}

	return findings, nil
}

func (d *Database) FindFindingByID(ctx context.Context, projectID, id ulid.ULID) (finding.Finding, error) {
	row := d.db.QueryRowContext(ctx,
		findingSelect+` WHERE id = ? AND project_id = ?`, id.String(), projectID.String())

	f, err := scanFinding(row)
	if errors.Is(err, sql.ErrNoRows) {
		return finding.Finding{}, finding.ErrFindingNotFound
	}
	if err != nil {
		return finding.Finding{}, err
	}

	return f, nil
}

func (d *Database) DeleteFinding(ctx context.Context, projectID, id ulid.ULID) error {
	res, err := d.db.ExecContext(ctx,
		`DELETE FROM findings WHERE id = ? AND project_id = ?`, id.String(), projectID.String())
	if err != nil {
		return fmt.Errorf("sqlite: failed to delete finding: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("sqlite: failed to read delete result: %w", err)
	}

	if affected == 0 {
		return finding.ErrFindingNotFound
	}

	return nil
}

func scanFinding(row scanner) (finding.Finding, error) {
	var (
		idStr, projectIDStr, title, description, severity string
		requestLogID                                      sql.NullString
		createdAt                                         int64
	)

	if err := row.Scan(&idStr, &projectIDStr, &title, &description, &severity,
		&requestLogID, &createdAt); err != nil {
		return finding.Finding{}, err
	}

	id, err := ulid.Parse(idStr)
	if err != nil {
		return finding.Finding{}, fmt.Errorf("sqlite: invalid finding id %q: %w", idStr, err)
	}

	projectID, err := ulid.Parse(projectIDStr)
	if err != nil {
		return finding.Finding{}, fmt.Errorf("sqlite: invalid project id %q: %w", projectIDStr, err)
	}

	f := finding.Finding{
		ID:          id,
		ProjectID:   projectID,
		Title:       title,
		Description: description,
		Severity:    finding.Severity(severity),
		CreatedAt:   time.UnixMilli(createdAt),
	}

	if requestLogID.Valid {
		reqLogID, err := ulid.Parse(requestLogID.String)
		if err != nil {
			return finding.Finding{}, fmt.Errorf("sqlite: invalid request log id %q: %w", requestLogID.String, err)
		}

		f.RequestLogID = &reqLogID
	}

	return f, nil
}
