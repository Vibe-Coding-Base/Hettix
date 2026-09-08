package sqlite_test

import (
	"context"
	"testing"
	"time"

	"github.com/Vibe-Coding-Base/Hettix/pkg/finding"
)

func TestFindingRoundTrip(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	projectID := newULID(t)
	reqLogID := newULID(t)

	f := finding.Finding{
		ID: newULID(t), ProjectID: projectID, Title: "Reflected XSS", Description: "q param reflected",
		Severity: finding.SeverityHigh, RequestLogID: &reqLogID, CreatedAt: time.UnixMilli(time.Now().UnixMilli()),
	}
	if err := db.StoreFinding(ctx, f); err != nil {
		t.Fatalf("store: %v", err)
	}

	got, err := db.FindFindingByID(ctx, projectID, f.ID)
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if got.Title != "Reflected XSS" || got.Severity != finding.SeverityHigh {
		t.Fatalf("unexpected finding: %+v", got)
	}
	if got.RequestLogID == nil || *got.RequestLogID != reqLogID {
		t.Fatalf("expected linked request log id, got %v", got.RequestLogID)
	}

	list, err := db.FindFindings(ctx, projectID)
	if err != nil || len(list) != 1 {
		t.Fatalf("expected 1 finding, got %d (err %v)", len(list), err)
	}

	// Project isolation.
	other := newULID(t)
	if _, err := db.FindFindingByID(ctx, other, f.ID); err != finding.ErrFindingNotFound {
		t.Fatalf("expected ErrFindingNotFound for other project, got %v", err)
	}
	if err := db.DeleteFinding(ctx, other, f.ID); err != finding.ErrFindingNotFound {
		t.Fatalf("expected ErrFindingNotFound deleting from other project, got %v", err)
	}

	if err := db.DeleteFinding(ctx, projectID, f.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if list, _ := db.FindFindings(ctx, projectID); len(list) != 0 {
		t.Fatalf("expected 0 findings after delete, got %d", len(list))
	}
}
