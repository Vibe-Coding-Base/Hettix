package plugin

import (
	"context"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oklog/ulid"

	"github.com/Vibe-Coding-Base/Hettix/pkg/finding"
)

type fakeRepo struct {
	enabled   map[string]bool
	endpoints []string
}

func (r *fakeRepo) LoadPluginEnabled(context.Context) (map[string]bool, error) { return r.enabled, nil }
func (r *fakeRepo) SavePluginEnabled(_ context.Context, id string, on bool) error {
	r.enabled[id] = on
	return nil
}
func (r *fakeRepo) DeletePluginEnabled(_ context.Context, id string) error {
	delete(r.enabled, id)
	return nil
}
func (r *fakeRepo) StoreDiscoveredEndpoint(_ context.Context, _ ulid.ULID, host, path, source string) error {
	r.endpoints = append(r.endpoints, source+" "+host+path)
	return nil
}

type fakeFindings struct{ titles []string }

func (f *fakeFindings) CreateFinding(
	_ context.Context, title, _ string, _ finding.Severity, _ *ulid.ULID,
) (finding.Finding, error) {
	f.titles = append(f.titles, title)
	return finding.Finding{}, nil
}

func TestJSMinerDetectsEndpointsAndSecrets(t *testing.T) {
	repo := &fakeRepo{enabled: map[string]bool{}}
	findings := &fakeFindings{}

	svc := NewService(Config{Dir: filepath.Join(t.TempDir(), "plugins"), Repo: repo, Findings: findings})
	if err := svc.Load(context.Background()); err != nil {
		t.Fatalf("load: %v", err)
	}
	svc.SetActiveProjectID(ulid.MustNew(1, nil))

	if got := svc.List(); len(got) != 1 || got[0].ID != "js-miner" {
		t.Fatalf("expected js-miner to be seeded, got %+v", got)
	}

	u, _ := url.Parse("https://target.example/app.js")
	body := `
		fetch("/api/v2/users/profile");
		const cdn = "https://cdn.acme-corp.net/static/config.json";
		const awsKey = "AKIAZ7Q2W9E4R6T1Y3U5";
		const googleKey = "AIza` + strings.Repeat("a", 35) + `";
	`
	svc.process(Response{
		Method: "GET", URL: u, Host: "target.example",
		ContentType: "application/javascript", Body: []byte(body),
	})

	if len(repo.endpoints) == 0 {
		t.Errorf("expected discovered endpoints, got none")
	}
	assertContains(t, repo.endpoints, "js-miner target.example/api/v2/users/profile")
	if len(findings.titles) < 2 {
		t.Errorf("expected at least 2 findings (AWS + Google), got %v", findings.titles)
	}
}

func assertContains(t *testing.T, haystack []string, needle string) {
	t.Helper()
	for _, s := range haystack {
		if s == needle {
			return
		}
	}
	t.Errorf("expected %q in %v", needle, haystack)
}

func TestInstallAndDelete(t *testing.T) {
	repo := &fakeRepo{enabled: map[string]bool{}}
	svc := NewService(Config{Dir: filepath.Join(t.TempDir(), "plugins"), Repo: repo, Findings: &fakeFindings{}})
	if err := svc.Load(context.Background()); err != nil {
		t.Fatalf("load: %v", err)
	}

	src := `hettix.plugin({ id: "demo", name: "Demo", version: "1.0.0" });
	        hettix.onResponse(function (resp) {});`
	status, err := svc.Install(context.Background(), src)
	if err != nil {
		t.Fatalf("install: %v", err)
	}
	if status.ID != "demo" {
		t.Fatalf("expected demo, got %q", status.ID)
	}

	if _, err := svc.Install(context.Background(), src); err == nil {
		t.Errorf("expected duplicate install to fail")
	}

	if err := svc.Delete(context.Background(), "demo"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, ok := svc.Source("demo"); ok {
		t.Errorf("expected demo to be gone after delete")
	}
}
