package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/oklog/ulid"

	"github.com/Vibe-Coding-Base/Hettix/pkg/workflow"
)

func (d *Database) StoreWorkflow(ctx context.Context, wf workflow.Workflow) error {
	steps, err := json.Marshal(wf.Steps)
	if err != nil {
		return fmt.Errorf("sqlite: failed to marshal workflow steps: %w", err)
	}

	_, err = d.db.ExecContext(ctx,
		`INSERT INTO workflows (id, project_id, name, steps, created_at)
		 VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT (id) DO UPDATE SET name = excluded.name, steps = excluded.steps`,
		wf.ID.String(), wf.ProjectID.String(), wf.Name, string(steps), wf.CreatedAt.UnixMilli(),
	)
	if err != nil {
		return fmt.Errorf("sqlite: failed to store workflow: %w", err)
	}

	return nil
}

const workflowSelect = `SELECT id, project_id, name, steps, created_at FROM workflows`

func (d *Database) FindWorkflows(ctx context.Context, projectID ulid.ULID) ([]workflow.Workflow, error) {
	rows, err := d.db.QueryContext(ctx,
		workflowSelect+` WHERE project_id = ? ORDER BY id DESC`, projectID.String())
	if err != nil {
		return nil, fmt.Errorf("sqlite: failed to query workflows: %w", err)
	}
	defer rows.Close()

	var workflows []workflow.Workflow

	for rows.Next() {
		wf, err := scanWorkflow(rows)
		if err != nil {
			return nil, err
		}

		workflows = append(workflows, wf)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("sqlite: failed to iterate workflows: %w", err)
	}

	return workflows, nil
}

func (d *Database) FindWorkflowByID(ctx context.Context, projectID, id ulid.ULID) (workflow.Workflow, error) {
	row := d.db.QueryRowContext(ctx,
		workflowSelect+` WHERE id = ? AND project_id = ?`, id.String(), projectID.String())

	wf, err := scanWorkflow(row)
	if errors.Is(err, sql.ErrNoRows) {
		return workflow.Workflow{}, workflow.ErrWorkflowNotFound
	}
	if err != nil {
		return workflow.Workflow{}, err
	}

	return wf, nil
}

func (d *Database) DeleteWorkflow(ctx context.Context, projectID, id ulid.ULID) error {
	res, err := d.db.ExecContext(ctx,
		`DELETE FROM workflows WHERE id = ? AND project_id = ?`, id.String(), projectID.String())
	if err != nil {
		return fmt.Errorf("sqlite: failed to delete workflow: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("sqlite: failed to read delete result: %w", err)
	}

	if affected == 0 {
		return workflow.ErrWorkflowNotFound
	}

	return nil
}

func scanWorkflow(row scanner) (workflow.Workflow, error) {
	var (
		idStr, projectIDStr, name, steps string
		createdAt                        int64
	)

	if err := row.Scan(&idStr, &projectIDStr, &name, &steps, &createdAt); err != nil {
		return workflow.Workflow{}, err
	}

	id, err := ulid.Parse(idStr)
	if err != nil {
		return workflow.Workflow{}, fmt.Errorf("sqlite: invalid workflow id %q: %w", idStr, err)
	}

	projectID, err := ulid.Parse(projectIDStr)
	if err != nil {
		return workflow.Workflow{}, fmt.Errorf("sqlite: invalid project id %q: %w", projectIDStr, err)
	}

	var parsedSteps []workflow.Step
	if err := json.Unmarshal([]byte(steps), &parsedSteps); err != nil {
		return workflow.Workflow{}, fmt.Errorf("sqlite: failed to unmarshal workflow steps: %w", err)
	}

	return workflow.Workflow{
		ID:        id,
		ProjectID: projectID,
		Name:      name,
		Steps:     parsedSteps,
		CreatedAt: time.UnixMilli(createdAt),
	}, nil
}
