package sqlite_test

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/Vibe-Coding-Base/Hettix/pkg/reqlog"
)

func TestSitemap(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	projectID := newULID(t)

	type row struct {
		method, rawURL string
		code           int
	}
	rows := []row{
		{"GET", "https://api.example.com/users", 200},
		{"POST", "https://api.example.com/users", 201},
		{"GET", "https://api.example.com/users", 200}, // duplicate endpoint
		{"GET", "https://api.example.com/health", 204},
	}

	for _, r := range rows {
		u, _ := url.Parse(r.rawURL)
		id := newULID(t)
		if err := db.StoreRequestLog(ctx, reqlog.RequestLog{
			ID: id, ProjectID: projectID, Method: r.method, URL: u, Proto: "HTTP/1.1",
			Header: http.Header{"X": {"y"}},
		}); err != nil {
			t.Fatal(err)
		}
		if err := db.StoreResponseLog(ctx, projectID, id, reqlog.ResponseLog{
			Proto: "HTTP/1.1", StatusCode: r.code,
		}); err != nil {
			t.Fatal(err)
		}
	}

	entries, err := db.Sitemap(ctx, projectID)
	if err != nil {
		t.Fatalf("sitemap: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 endpoints, got %d: %+v", len(entries), entries)
	}

	var users reqlog.SitemapEntry
	for _, e := range entries {
		if e.Path == "/users" {
			users = e
		}
	}
	if users.Host != "api.example.com" || users.Count != 3 {
		t.Fatalf("unexpected /users entry: %+v", users)
	}
	if len(users.Methods) != 2 || users.Methods[0] != "GET" || users.Methods[1] != "POST" {
		t.Fatalf("expected sorted methods [GET POST], got %v", users.Methods)
	}
	if len(users.StatusCodes) != 2 || users.StatusCodes[0] != 200 || users.StatusCodes[1] != 201 {
		t.Fatalf("expected status codes [200 201], got %v", users.StatusCodes)
	}
}
