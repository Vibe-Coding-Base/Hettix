package aitools

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"
	"testing"

	"github.com/oklog/ulid"

	"github.com/Vibe-Coding-Base/Hettix/pkg/intruder"
	"github.com/Vibe-Coding-Base/Hettix/pkg/matchreplace"
	"github.com/Vibe-Coding-Base/Hettix/pkg/scope"
)

type fakeProjectService struct {
	scope *scope.Scope
}

func (f *fakeProjectService) SetScopeRules(context.Context, []scope.Rule) error { return nil }
func (f *fakeProjectService) Scope() *scope.Scope                               { return f.scope }
func (f *fakeProjectService) MatchReplaceRules(context.Context) ([]matchreplace.Rule, error) {
	return nil, nil
}
func (f *fakeProjectService) SetMatchReplaceRules(context.Context, []matchreplace.Rule) error {
	return nil
}

type fakeIntruderService struct {
	started  bool
	payloads []string
}

func (f *fakeIntruderService) StartAttack(
	_ context.Context,
	_ string,
	_ intruder.Request,
	payloads []string,
) (intruder.Attack, error) {
	f.started = true
	f.payloads = payloads

	return intruder.Attack{ID: ulid.MustNew(ulid.Now(), nil), Total: len(payloads)}, nil
}

func (f *fakeIntruderService) AttackByID(context.Context, ulid.ULID) (intruder.Attack, error) {
	return intruder.Attack{}, nil
}

func (f *fakeIntruderService) Results(context.Context, ulid.ULID) ([]intruder.Result, error) {
	return nil, nil
}

func TestRunFuzzerRequiresMarker(t *testing.T) {
	proj := &fakeProjectService{scope: &scope.Scope{}}
	intr := &fakeIntruderService{}

	handler := runFuzzer(proj, intr)
	args, _ := json.Marshal(map[string]any{"url": "https://target.com/", "payloads": []string{"a"}})

	if _, err := handler(context.Background(), args); err == nil {
		t.Fatal("expected an error when the request has no § marker")
	}
	if intr.started {
		t.Fatal("attack should not start without a marker")
	}
}

func TestRunFuzzerEnforcesScope(t *testing.T) {
	s := &scope.Scope{}
	s.SetRules([]scope.Rule{{URL: regexp.MustCompile(`^https://allowed\.com`)}})

	proj := &fakeProjectService{scope: s}
	intr := &fakeIntruderService{}
	handler := runFuzzer(proj, intr)

	// Out of scope: refused.
	args, _ := json.Marshal(map[string]any{"url": "https://evil.com/?q=§", "payloads": []string{"a"}})
	out, err := handler(context.Background(), args)
	if err == nil || !strings.Contains(err.Error(), "out of scope") {
		t.Fatalf("expected an out-of-scope refusal, got out=%q err=%v", out, err)
	}
	if intr.started {
		t.Fatal("out-of-scope attack must not start")
	}

	// In scope: allowed.
	args, _ = json.Marshal(map[string]any{"url": "https://allowed.com/?q=§", "payloads": []string{"a", "b"}})
	if _, err := handler(context.Background(), args); err != nil {
		t.Fatalf("expected in-scope attack to start, got %v", err)
	}
	if !intr.started || len(intr.payloads) != 2 {
		t.Fatalf("expected the attack to start with 2 payloads, got started=%v payloads=%v", intr.started, intr.payloads)
	}
}
