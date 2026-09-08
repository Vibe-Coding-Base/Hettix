// Package workflow runs saved, declarative automations over a project's
// capabilities: an ordered list of typed steps (search traffic, fuzz a request,
// record a finding) that pass simple values forward through ${var} substitution.
// It composes the existing services rather than exposing a plugin ABI, so the
// internal API stays free to evolve.
package workflow

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/oklog/ulid"

	"github.com/Vibe-Coding-Base/Hettix/pkg/finding"
	"github.com/Vibe-Coding-Base/Hettix/pkg/httpql"
	"github.com/Vibe-Coding-Base/Hettix/pkg/intruder"
	"github.com/Vibe-Coding-Base/Hettix/pkg/reqlog"
	"github.com/Vibe-Coding-Base/Hettix/pkg/scope"
)

// StepType identifies a workflow step's action.
type StepType string

const (
	StepSearch  StepType = "search"
	StepFuzz    StepType = "fuzz"
	StepFinding StepType = "finding"
)

var (
	ErrProjectIDMustBeSet = errors.New("workflow: project ID must be set")
	ErrWorkflowNotFound   = errors.New("workflow: workflow not found")
	ErrNameRequired       = errors.New("workflow: name is required")
	ErrUnknownStep        = errors.New("workflow: unknown step type")
)

// Step is one action in a workflow. Only the fields relevant to its Type are used.
type Step struct {
	Type StepType `json:"type"`

	// search
	Query string `json:"query,omitempty"`

	// fuzz
	Name     string   `json:"name,omitempty"`
	Method   string   `json:"method,omitempty"`
	URL      string   `json:"url,omitempty"`
	Body     string   `json:"body,omitempty"`
	Payloads []string `json:"payloads,omitempty"`

	// finding
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Severity    string `json:"severity,omitempty"`
}

// Workflow is a named, ordered list of steps.
type Workflow struct {
	ID        ulid.ULID
	ProjectID ulid.ULID
	Name      string
	Steps     []Step
	CreatedAt time.Time
}

// StepResult is the outcome of running one step.
type StepResult struct {
	Type   StepType
	Output string
	Error  string
}

// SearchService runs HTTPQL queries against captured traffic.
type SearchService interface {
	FindByQuery(ctx context.Context, expr httpql.Expression, limit int) ([]reqlog.RequestLog, error)
}

// FuzzService starts fuzzing attacks.
type FuzzService interface {
	StartAttack(ctx context.Context, name string, tmpl intruder.Request, payloads []string) (intruder.Attack, error)
}

// FindingService records findings.
type FindingService interface {
	CreateFinding(
		ctx context.Context,
		title, description string,
		severity finding.Severity,
		requestLogID *ulid.ULID,
	) (finding.Finding, error)
}

// Runner executes workflow steps against the services.
type Runner struct {
	search   SearchService
	fuzz     FuzzService
	findings FindingService
	scope    *scope.Scope
}

type RunnerConfig struct {
	Search   SearchService
	Fuzz     FuzzService
	Findings FindingService
	Scope    *scope.Scope
}

func NewRunner(cfg RunnerConfig) *Runner {
	return &Runner{search: cfg.Search, fuzz: cfg.Fuzz, findings: cfg.Findings, scope: cfg.Scope}
}

var varPattern = regexp.MustCompile(`\$\{(\w+)\}`)

// Run executes steps in order, threading ${var} substitutions forward, and
// returns a result per step. It stops at the first step that errors.
func (r *Runner) Run(ctx context.Context, steps []Step) []StepResult {
	vars := map[string]string{}

	results := make([]StepResult, 0, len(steps))

	for _, step := range steps {
		result := r.runStep(ctx, step, vars)
		results = append(results, result)

		if result.Error != "" {
			break
		}
	}

	return results
}

func (r *Runner) runStep(ctx context.Context, step Step, vars map[string]string) StepResult {
	result := StepResult{Type: step.Type}

	var (
		output string
		err    error
	)

	switch step.Type {
	case StepSearch:
		output, err = r.runSearch(ctx, subst(step.Query, vars), vars)
	case StepFuzz:
		output, err = r.runFuzz(ctx, step, vars)
	case StepFinding:
		output, err = r.runFinding(ctx, step, vars)
	default:
		err = fmt.Errorf("%w: %q", ErrUnknownStep, step.Type)
	}

	if err != nil {
		result.Error = err.Error()
		return result
	}

	result.Output = output

	return result
}

func (r *Runner) runSearch(ctx context.Context, query string, vars map[string]string) (string, error) {
	expr, err := httpql.Parse(query)
	if err != nil {
		return "", err
	}

	logs, err := r.search.FindByQuery(ctx, expr, 100)
	if err != nil {
		return "", err
	}

	hostSet := map[string]struct{}{}
	for _, rl := range logs {
		if rl.URL != nil {
			hostSet[rl.URL.Hostname()] = struct{}{}
		}
	}

	hosts := make([]string, 0, len(hostSet))
	for h := range hostSet {
		hosts = append(hosts, h)
	}
	sort.Strings(hosts)

	// Publish outputs for later steps.
	vars["count"] = strconv.Itoa(len(logs))
	vars["hosts"] = strings.Join(hosts, ", ")
	if len(logs) > 0 {
		vars["firstId"] = logs[0].ID.String()
	}

	return fmt.Sprintf("Matched %d request(s) across %d host(s).", len(logs), len(hosts)), nil
}

func (r *Runner) runFuzz(ctx context.Context, step Step, vars map[string]string) (string, error) {
	url := subst(step.URL, vars)
	body := subst(step.Body, vars)

	method := subst(step.Method, vars)
	if method == "" {
		method = "GET"
	}

	tmpl := intruder.Request{Method: method, URL: url, Body: body}
	if !tmpl.HasMarker() {
		return "", fmt.Errorf("fuzz step needs a %s marker in the url or body", intruder.Marker)
	}

	// Enforce scope: refuse to fuzz an out-of-scope host.
	if r.scope != nil {
		check := strings.ReplaceAll(url, intruder.Marker, "")
		if len(r.scope.Rules()) > 0 && !r.scope.InScope(check, nil, nil) {
			return "", fmt.Errorf("refusing to fuzz: %s is out of scope", check)
		}
	}

	payloads := make([]string, len(step.Payloads))
	for i, p := range step.Payloads {
		payloads[i] = subst(p, vars)
	}

	name := subst(step.Name, vars)
	if name == "" {
		name = "Workflow attack"
	}

	attack, err := r.fuzz.StartAttack(ctx, name, tmpl, payloads)
	if err != nil {
		return "", err
	}

	vars["attackId"] = attack.ID.String()

	return fmt.Sprintf("Started fuzzing attack %s with %d payloads.", attack.ID, len(payloads)), nil
}

func (r *Runner) runFinding(ctx context.Context, step Step, vars map[string]string) (string, error) {
	title := subst(step.Title, vars)
	description := subst(step.Description, vars)

	f, err := r.findings.CreateFinding(ctx, title, description, finding.ParseSeverity(subst(step.Severity, vars)), nil)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Recorded %s finding %s.", f.Severity, f.ID), nil
}

// subst replaces ${name} references in s with the matching variable's value.
func subst(s string, vars map[string]string) string {
	return varPattern.ReplaceAllStringFunc(s, func(match string) string {
		name := varPattern.FindStringSubmatch(match)[1]
		return vars[name]
	})
}

//nolint:gosec // workflow IDs don't need cryptographic randomness
var (
	entropy   = rand.New(rand.NewSource(time.Now().UnixNano()))
	entropyMu sync.Mutex
)

func newID() ulid.ULID {
	entropyMu.Lock()
	defer entropyMu.Unlock()

	return ulid.MustNew(ulid.Timestamp(time.Now()), entropy)
}

func (s StepType) valid() bool {
	switch s {
	case StepSearch, StepFuzz, StepFinding:
		return true
	default:
		return false
	}
}
