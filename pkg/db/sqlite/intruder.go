package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/oklog/ulid"

	"github.com/Vibe-Coding-Base/Hettix/pkg/intruder"
)

func (d *Database) StoreAttack(ctx context.Context, attack intruder.Attack) error {
	_, err := d.db.ExecContext(ctx,
		`INSERT INTO intruder_attacks (id, project_id, name, status, total, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		attack.ID.String(), attack.ProjectID.String(), attack.Name, attack.Status, attack.Total,
		attack.CreatedAt.UnixMilli(),
	)
	if err != nil {
		return fmt.Errorf("sqlite: failed to store attack: %w", err)
	}

	return nil
}

func (d *Database) UpdateAttackStatus(ctx context.Context, projectID, id ulid.ULID, status string) error {
	_, err := d.db.ExecContext(ctx,
		`UPDATE intruder_attacks SET status = ? WHERE id = ? AND project_id = ?`,
		status, id.String(), projectID.String(),
	)
	if err != nil {
		return fmt.Errorf("sqlite: failed to update attack status: %w", err)
	}

	return nil
}

func (d *Database) StoreResult(ctx context.Context, attackID ulid.ULID, result intruder.Result) error {
	_, err := d.db.ExecContext(ctx,
		`INSERT INTO intruder_results (attack_id, idx, payload, status_code, length, duration_ms, error)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		attackID.String(), result.Index, result.Payload, result.StatusCode, result.Length,
		result.DurationMs, result.Error,
	)
	if err != nil {
		return fmt.Errorf("sqlite: failed to store attack result: %w", err)
	}

	return nil
}

const attackSelect = `SELECT a.id, a.project_id, a.name, a.status, a.total, a.created_at,
	(SELECT COUNT(*) FROM intruder_results r WHERE r.attack_id = a.id) AS completed
	FROM intruder_attacks a`

func (d *Database) FindAttacks(ctx context.Context, projectID ulid.ULID) ([]intruder.Attack, error) {
	rows, err := d.db.QueryContext(ctx,
		attackSelect+` WHERE a.project_id = ? ORDER BY a.id DESC`, projectID.String())
	if err != nil {
		return nil, fmt.Errorf("sqlite: failed to query attacks: %w", err)
	}
	defer rows.Close()

	var attacks []intruder.Attack

	for rows.Next() {
		attack, err := scanAttack(rows)
		if err != nil {
			return nil, err
		}

		attacks = append(attacks, attack)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("sqlite: failed to iterate attacks: %w", err)
	}

	return attacks, nil
}

func (d *Database) FindAttackByID(ctx context.Context, projectID, id ulid.ULID) (intruder.Attack, error) {
	row := d.db.QueryRowContext(ctx,
		attackSelect+` WHERE a.id = ? AND a.project_id = ?`, id.String(), projectID.String())

	attack, err := scanAttack(row)
	if errors.Is(err, sql.ErrNoRows) {
		return intruder.Attack{}, intruder.ErrAttackNotFound
	}
	if err != nil {
		return intruder.Attack{}, err
	}

	return attack, nil
}

func (d *Database) FindResults(ctx context.Context, projectID, attackID ulid.ULID) ([]intruder.Result, error) {
	// The project scope is enforced through the parent attack.
	if _, err := d.FindAttackByID(ctx, projectID, attackID); err != nil {
		return nil, err
	}

	rows, err := d.db.QueryContext(ctx,
		`SELECT idx, payload, status_code, length, duration_ms, error
		 FROM intruder_results WHERE attack_id = ? ORDER BY idx ASC`, attackID.String())
	if err != nil {
		return nil, fmt.Errorf("sqlite: failed to query attack results: %w", err)
	}
	defer rows.Close()

	var results []intruder.Result

	for rows.Next() {
		var res intruder.Result
		if err := rows.Scan(&res.Index, &res.Payload, &res.StatusCode, &res.Length,
			&res.DurationMs, &res.Error); err != nil {
			return nil, fmt.Errorf("sqlite: failed to scan attack result: %w", err)
		}

		results = append(results, res)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("sqlite: failed to iterate attack results: %w", err)
	}

	return results, nil
}

func (d *Database) DeleteAttacks(ctx context.Context, projectID ulid.ULID) error {
	_, err := d.db.ExecContext(ctx, `DELETE FROM intruder_attacks WHERE project_id = ?`, projectID.String())
	if err != nil {
		return fmt.Errorf("sqlite: failed to delete attacks: %w", err)
	}

	return nil
}

func scanAttack(row scanner) (intruder.Attack, error) {
	var (
		idStr, projectIDStr, name, status string
		total, completed                  int
		createdAt                         int64
	)

	if err := row.Scan(&idStr, &projectIDStr, &name, &status, &total, &createdAt, &completed); err != nil {
		return intruder.Attack{}, err
	}

	id, err := ulid.Parse(idStr)
	if err != nil {
		return intruder.Attack{}, fmt.Errorf("sqlite: invalid attack id %q: %w", idStr, err)
	}

	projectID, err := ulid.Parse(projectIDStr)
	if err != nil {
		return intruder.Attack{}, fmt.Errorf("sqlite: invalid project id %q: %w", projectIDStr, err)
	}

	return intruder.Attack{
		ID:        id,
		ProjectID: projectID,
		Name:      name,
		Status:    status,
		Total:     total,
		Completed: completed,
		CreatedAt: time.UnixMilli(createdAt),
	}, nil
}
