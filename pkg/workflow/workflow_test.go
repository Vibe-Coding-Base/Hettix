package workflow_test

import (
	"context"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/oklog/ulid"

	"github.com/Vibe-Coding-Base/Hettix/pkg/finding"
	"github.com/Vibe-Coding-Base/Hettix/pkg/httpql"
	"github.com/Vibe-Coding-Base/Hettix/pkg/intruder"
	"github.com/Vibe-Coding-Base/Hettix/pkg/reqlog"
	"github.com/Vibe-Coding-Base/Hettix/pkg/scope"
	"github.com/Vibe-Coding-Base/Hettix/pkg/workflow"
)

type fakeSearch struct{ logs []reqlog.RequestLog }

func (f *fakeSearch) FindByQuery(context.Context, httpql.Expression, int) ([]reqlog.RequestLog, error) {
	return f.logs, nil
}

type fakeFuzz struct {
	url      string
	payloads []string
}

func (f *fakeFuzz) StartAttack(
	_ context.Context,
	_ string,
	tmpl intruder.Request,
	payloads []string,
) (intruder.Attack, error) {
	f.url = tmpl.URL
	f.payloads = payloads

	return intruder.Attack{ID: ulid.MustNew(ulid.Now(), nil)}, nil
}

type fakeFinding struct {
	title    string
	severity finding.Severity
}

func (f *fakeFinding) CreateFinding(
	_ context.Context,
	title, _ string,
	severity finding.Severity,
	_ *ulid.ULID,
) (finding.Finding, error) {
	f.title = title
	f.severity = severity

	return finding.Finding{ID: ulid.MustNew(ulid.Now(), nil), Severity: severity}, nil
}

func TestRunnerChainsSteps(t *testing.T) {
	u, _ := url.Parse("https://api.example.com/login")
	search := &fakeSearch{logs: []reqlog.RequestLog{
		{ID: ulid.MustNew(ulid.Now(), nil), URL: u},
	}}
	fuzz := &fakeFuzz{}
	fnd := &fakeFinding{}

	sc := &scope.Scope{}
	sc.SetRules([]scope.Rule{{URL: regexp.MustCompile(`example\.com`)}})

	runner := workflow.NewRunner(workflow.RunnerConfig{Search: search, Fuzz: fuzz, Findings: fnd, Scope: sc})

	steps := []workflow.Step{
		{Type: workflow.StepSearch, Query: `req.host cont "example"`},
		{
			Type: workflow.StepFuzz, Name: "login fuzz", Method: "POST",
			URL: "https://api.example.com/login?u=§", Payloads: []string{"admin", "root"},
		},
		{
			Type: workflow.StepFinding, Title: "Fuzzed ${count} endpoint(s)",
			Severity: "medium", Description: "hosts: ${hosts}",
		},
	}

	results := runner.Run(context.Background(), steps)

	if len(results) != 3 {
		t.Fatalf("expected 3 step results, got %d", len(results))
	}
	for i, res := range results {
		if res.Error != "" {
			t.Fatalf("step %d errored: %s", i, res.Error)
		}
	}

	if len(fuzz.payloads) != 2 {
		t.Fatalf("expected 2 payloads sent to fuzzer, got %d", len(fuzz.payloads))
	}
	// The finding step substituted ${count} from the search step's output.
	if fnd.title != "Fuzzed 1 endpoint(s)" {
		t.Fatalf("expected substituted title, got %q", fnd.title)
	}
	if fnd.severity != finding.SeverityMedium {
		t.Fatalf("expected medium severity, got %q", fnd.severity)
	}
}

func TestRunnerStopsOnError(t *testing.T) {
	sc := &scope.Scope{}
	sc.SetRules([]scope.Rule{{URL: regexp.MustCompile(`allowed\.com`)}})

	runner := workflow.NewRunner(workflow.RunnerConfig{
		Search: &fakeSearch{}, Fuzz: &fakeFuzz{}, Findings: &fakeFinding{}, Scope: sc,
	})

	steps := []workflow.Step{
		// Out-of-scope fuzz must fail and stop the run.
		{Type: workflow.StepFuzz, URL: "https://evil.com/?q=§", Payloads: []string{"a"}},
		{Type: workflow.StepFinding, Title: "should not run", Severity: "low"},
	}

	results := runner.Run(context.Background(), steps)

	if len(results) != 1 {
		t.Fatalf("expected the run to stop after the failing step, got %d results", len(results))
	}
	if results[0].Error == "" || !strings.Contains(results[0].Error, "out of scope") {
		t.Fatalf("expected an out-of-scope error, got %q", results[0].Error)
	}
}
