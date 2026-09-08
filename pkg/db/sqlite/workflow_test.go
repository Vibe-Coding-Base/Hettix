package sqlite_test

import (
	"context"
	"testing"
	"time"

	"github.com/Vibe-Coding-Base/Hettix/pkg/workflow"
)

func TestWorkflowRoundTrip(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	projectID := newULID(t)

	wf := workflow.Workflow{
		ID: newULID(t), ProjectID: projectID, Name: "login audit",
		Steps: []workflow.Step{
			{Type: workflow.StepSearch, Query: `req.path cont "login"`},
			{Type: workflow.StepFuzz, URL: "https://t/login?u=§", Payloads: []string{"a", "b"}},
		},
		CreatedAt: time.UnixMilli(time.Now().UnixMilli()),
	}
	if err := db.StoreWorkflow(ctx, wf); err != nil {
		t.Fatalf("store: %v", err)
	}

	got, err := db.FindWorkflowByID(ctx, projectID, wf.ID)
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if got.Name != "login audit" || len(got.Steps) != 2 || got.Steps[1].Type != workflow.StepFuzz {
		t.Fatalf("unexpected workflow: %+v", got)
	}
	if len(got.Steps[1].Payloads) != 2 {
		t.Fatalf("expected 2 payloads preserved, got %v", got.Steps[1].Payloads)
	}

	// Update (upsert) changes name/steps but keeps the id.
	wf.Name = "login audit v2"
	wf.Steps = wf.Steps[:1]
	if err := db.StoreWorkflow(ctx, wf); err != nil {
		t.Fatalf("update: %v", err)
	}
	got, _ = db.FindWorkflowByID(ctx, projectID, wf.ID)
	if got.Name != "login audit v2" || len(got.Steps) != 1 {
		t.Fatalf("update not applied: %+v", got)
	}

	list, _ := db.FindWorkflows(ctx, projectID)
	if len(list) != 1 {
		t.Fatalf("expected 1 workflow, got %d", len(list))
	}

	// Project isolation.
	other := newULID(t)
	if _, err := db.FindWorkflowByID(ctx, other, wf.ID); err != workflow.ErrWorkflowNotFound {
		t.Fatalf("expected ErrWorkflowNotFound for other project, got %v", err)
	}

	if err := db.DeleteWorkflow(ctx, projectID, wf.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if list, _ := db.FindWorkflows(ctx, projectID); len(list) != 0 {
		t.Fatalf("expected 0 after delete, got %d", len(list))
	}
}
