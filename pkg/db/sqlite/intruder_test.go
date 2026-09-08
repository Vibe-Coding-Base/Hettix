package sqlite_test

import (
	"context"
	"testing"
	"time"

	"github.com/Vibe-Coding-Base/Hettix/pkg/intruder"
)

func TestIntruderRoundTrip(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	projectID := newULID(t)
	attackID := newULID(t)

	attack := intruder.Attack{
		ID:        attackID,
		ProjectID: projectID,
		Name:      "login fuzz",
		Status:    intruder.StatusRunning,
		Total:     2,
		CreatedAt: time.UnixMilli(time.Now().UnixMilli()),
	}
	if err := db.StoreAttack(ctx, attack); err != nil {
		t.Fatalf("store attack: %v", err)
	}

	results := []intruder.Result{
		{Index: 0, Payload: "admin", StatusCode: 200, Length: 120, DurationMs: 10},
		{Index: 1, Payload: "root", StatusCode: 500, Length: 40, DurationMs: 15, Error: ""},
	}
	for _, r := range results {
		if err := db.StoreResult(ctx, attackID, r); err != nil {
			t.Fatalf("store result: %v", err)
		}
	}

	if err := db.UpdateAttackStatus(ctx, projectID, attackID, intruder.StatusCompleted); err != nil {
		t.Fatalf("update status: %v", err)
	}

	got, err := db.FindAttackByID(ctx, projectID, attackID)
	if err != nil {
		t.Fatalf("find attack: %v", err)
	}
	if got.Status != intruder.StatusCompleted || got.Total != 2 || got.Completed != 2 {
		t.Fatalf("unexpected attack state: %+v", got)
	}

	gotResults, err := db.FindResults(ctx, projectID, attackID)
	if err != nil {
		t.Fatalf("find results: %v", err)
	}
	if len(gotResults) != 2 || gotResults[0].Payload != "admin" || gotResults[1].StatusCode != 500 {
		t.Fatalf("unexpected results: %+v", gotResults)
	}

	// Project isolation: another project sees nothing.
	other := newULID(t)
	if _, err := db.FindAttackByID(ctx, other, attackID); err != intruder.ErrAttackNotFound {
		t.Fatalf("expected ErrAttackNotFound for other project, got %v", err)
	}
	if _, err := db.FindResults(ctx, other, attackID); err != intruder.ErrAttackNotFound {
		t.Fatalf("expected ErrAttackNotFound querying other project's results, got %v", err)
	}
}
