package sqlite_test

import (
	"context"
	"math/rand"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/oklog/ulid"

	"github.com/Vibe-Coding-Base/Hettix/pkg/db/sqlite"
	"github.com/Vibe-Coding-Base/Hettix/pkg/httpql"
	"github.com/Vibe-Coding-Base/Hettix/pkg/proj"
	"github.com/Vibe-Coding-Base/Hettix/pkg/reqlog"
	"github.com/Vibe-Coding-Base/Hettix/pkg/sender"
)

var entropy = rand.New(rand.NewSource(1))

func newULID(t *testing.T) ulid.ULID {
	t.Helper()

	id, err := ulid.New(ulid.Timestamp(time.Now()), entropy)
	if err != nil {
		t.Fatalf("failed to create ULID: %v", err)
	}

	return id
}

func openTestDB(t *testing.T) *sqlite.Database {
	t.Helper()

	db, err := sqlite.OpenDatabase(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	return db
}

func TestProjectRoundTrip(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	project := proj.Project{ID: newULID(t), Name: "acme"}
	if err := db.UpsertProject(ctx, project); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	got, err := db.FindProjectByID(ctx, project.ID)
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if got.Name != "acme" {
		t.Fatalf("expected name acme, got %q", got.Name)
	}

	projects, err := db.Projects(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}

	if err := db.DeleteProject(ctx, project.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := db.FindProjectByID(ctx, project.ID); err != reqlogProjectNotFound() {
		t.Fatalf("expected ErrProjectNotFound, got %v", err)
	}
}

func reqlogProjectNotFound() error { return proj.ErrProjectNotFound }

func TestRequestLogRoundTrip(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	projectID := newULID(t)
	u, _ := url.Parse("https://api.example.com/v1/users?q=1")

	older := reqlog.RequestLog{
		ID: newULID(t), ProjectID: projectID, Method: "GET", URL: u, Proto: "HTTP/1.1",
		Header: http.Header{"X-Test": {"a"}}, Body: []byte("older"),
	}
	newer := reqlog.RequestLog{
		ID: newULID(t), ProjectID: projectID, Method: "POST", URL: u, Proto: "HTTP/1.1",
		Header: http.Header{"X-Test": {"b"}}, Body: []byte("newer"),
	}

	for _, rl := range []reqlog.RequestLog{older, newer} {
		if err := db.StoreRequestLog(ctx, rl); err != nil {
			t.Fatalf("store request: %v", err)
		}
	}

	if err := db.StoreResponseLog(ctx, projectID, newer.ID, reqlog.ResponseLog{
		Proto: "HTTP/1.1", StatusCode: 201, Status: "201 Created",
		Header: http.Header{"Content-Type": {"application/json"}}, Body: []byte("{}"), RoundTripMillis: 42,
	}); err != nil {
		t.Fatalf("store response: %v", err)
	}

	got, err := db.FindRequestLogByID(ctx, projectID, newer.ID)
	if err != nil {
		t.Fatalf("find by id: %v", err)
	}
	if got.Method != "POST" || string(got.Body) != "newer" {
		t.Fatalf("unexpected request: %+v", got)
	}
	if got.Response == nil || got.Response.StatusCode != 201 || got.Response.RoundTripMillis != 42 {
		t.Fatalf("unexpected response: %+v", got.Response)
	}

	logs, err := db.FindRequestLogs(ctx, reqlog.FindRequestsFilter{ProjectID: projectID}, nil)
	if err != nil {
		t.Fatalf("find logs: %v", err)
	}
	if len(logs) != 2 {
		t.Fatalf("expected 2 logs, got %d", len(logs))
	}
	// Newest first.
	if logs[0].ID != newer.ID {
		t.Fatalf("expected newest first, got %v", logs[0].ID)
	}

	if err := db.ClearRequestLogs(ctx, projectID); err != nil {
		t.Fatalf("clear: %v", err)
	}
	logs, err = db.FindRequestLogs(ctx, reqlog.FindRequestsFilter{ProjectID: projectID}, nil)
	if err != nil {
		t.Fatalf("find after clear: %v", err)
	}
	if len(logs) != 0 {
		t.Fatalf("expected 0 logs after clear, got %d", len(logs))
	}
}

func TestSenderRequestRoundTrip(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	projectID := newULID(t)
	u, _ := url.Parse("https://api.example.com/login")

	req := sender.Request{
		ID: newULID(t), ProjectID: projectID, Method: "POST", URL: u, Proto: "HTTP/1.1",
		Header: http.Header{"Content-Type": {"application/json"}}, Body: []byte(`{"u":"a"}`),
	}
	if err := db.StoreSenderRequest(ctx, req); err != nil {
		t.Fatalf("store: %v", err)
	}

	// Update the same request with a response.
	req.Response = &reqlog.ResponseLog{Proto: "HTTP/1.1", StatusCode: 200, Status: "200 OK", RoundTripMillis: 7}
	if err := db.StoreSenderRequest(ctx, req); err != nil {
		t.Fatalf("update: %v", err)
	}

	got, err := db.FindSenderRequestByID(ctx, projectID, req.ID)
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if got.Response == nil || got.Response.StatusCode != 200 {
		t.Fatalf("expected response 200, got %+v", got.Response)
	}

	reqs, err := db.FindSenderRequests(ctx, sender.FindRequestsFilter{ProjectID: projectID}, nil)
	if err != nil {
		t.Fatalf("find requests: %v", err)
	}
	if len(reqs) != 1 {
		t.Fatalf("expected 1 sender request, got %d", len(reqs))
	}

	if err := db.DeleteSenderRequests(ctx, projectID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	reqs, _ = db.FindSenderRequests(ctx, sender.FindRequestsFilter{ProjectID: projectID}, nil)
	if len(reqs) != 0 {
		t.Fatalf("expected 0 after delete, got %d", len(reqs))
	}
}

func TestRequestLogPushdownMatchesInMemory(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	projectID := newULID(t)

	type row struct {
		method, rawURL, body string
		code                 int
	}
	rows := []row{
		{"GET", "https://api.example.com/v1/users", "list", 200},
		{"POST", "https://api.example.com/v1/login", "password=secret", 401},
		{"GET", "https://web.example.com/home", "home", 200},
		{"DELETE", "http://api.example.com/v1/users/1", "", 500},
		{"GET", "https://api.other.com/health", "ok", 204},
	}

	for _, r := range rows {
		u, _ := url.Parse(r.rawURL)
		id := newULID(t)
		if err := db.StoreRequestLog(ctx, reqlog.RequestLog{
			ID: id, ProjectID: projectID, Method: r.method, URL: u, Proto: "HTTP/1.1",
			Header: http.Header{"X-Test": {"v"}}, Body: []byte(r.body),
		}); err != nil {
			t.Fatal(err)
		}
		if err := db.StoreResponseLog(ctx, projectID, id, reqlog.ResponseLog{
			Proto: "HTTP/1.1", StatusCode: r.code, Status: strconv.Itoa(r.code),
		}); err != nil {
			t.Fatal(err)
		}
	}

	// Reference set: everything, filtered purely in memory.
	all, err := db.FindRequestLogs(ctx, reqlog.FindRequestsFilter{ProjectID: projectID}, nil)
	if err != nil {
		t.Fatal(err)
	}

	queries := []string{
		`req.method eq "GET"`,
		`req.method ne "GET"`,
		`resp.code gte 400`,
		`resp.code eq 200`,
		`req.host eq "api.example.com"`,
		`req.host cont "example"`, // not pushed
		`req.path cont "users"`,   // not pushed
		`req.body cont "secret"`,  // not pushed
		`req.method eq "GET" AND resp.code eq 200`,
		`req.method eq "POST" OR resp.code eq 500`,
		`NOT req.method eq "GET"`,
		`req.url regex "v1/.*"`, // not pushed
		`req.tls eq true`,       // not a column
	}

	for _, q := range queries {
		expr, err := httpql.Parse(q)
		if err != nil {
			t.Fatalf("parse %q: %v", q, err)
		}

		want := map[string]bool{}
		for _, rl := range all {
			ok, err := rl.Matches(expr)
			if err != nil {
				t.Fatalf("in-memory match %q: %v", q, err)
			}
			if ok {
				want[rl.ID.String()] = true
			}
		}

		got, err := db.FindRequestLogs(ctx, reqlog.FindRequestsFilter{ProjectID: projectID, SearchExpr: expr}, nil)
		if err != nil {
			t.Fatalf("find %q: %v", q, err)
		}

		if len(got) != len(want) {
			t.Fatalf("query %q: got %d rows, want %d", q, len(got), len(want))
		}
		for _, rl := range got {
			if !want[rl.ID.String()] {
				t.Fatalf("query %q: pushdown returned a row the in-memory filter rejects (%v)", q, rl.ID)
			}
		}
	}
}
