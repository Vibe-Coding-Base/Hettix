// Package matchreplace applies operator-defined rules that rewrite proxied
// requests and responses. Each rule has an optional HTTPQL condition and either
// a header operation, a body regexp replacement, or both.
package matchreplace

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"sync"

	"github.com/oklog/ulid"

	"github.com/Vibe-Coding-Base/Hettix/pkg/httpql"
	"github.com/Vibe-Coding-Base/Hettix/pkg/proxy"
)

// Phase selects whether a rule runs on requests or responses.
type Phase int

const (
	PhaseRequest Phase = iota
	PhaseResponse
)

// Rule is a single match-and-replace rule. It is fully serializable (the body
// matcher is a pattern string, compiled by the engine) so it persists in a
// project's settings.
type Rule struct {
	ID      ulid.ULID
	Name    string
	Enabled bool
	Phase   Phase

	// Condition is an optional HTTPQL expression; when nil the rule always
	// applies.
	Condition httpql.Expression

	// Header operation: set HeaderName to HeaderValue, or remove it.
	HeaderName   string
	HeaderValue  string
	RemoveHeader bool

	// Body operation: replace matches of BodyMatcher (a regexp) with
	// BodyReplacement. Empty BodyMatcher means no body operation.
	BodyMatcher     string
	BodyReplacement string
}

type compiledRule struct {
	rule   Rule
	bodyRe *regexp.Regexp
}

// Engine holds the active rules and applies them as proxy modifiers.
type Engine struct {
	mu    sync.RWMutex
	rules []compiledRule
}

// NewEngine returns an empty engine.
func NewEngine() *Engine {
	return &Engine{}
}

// SetRules replaces the active rule set, compiling body matchers. It returns an
// error (and leaves the current rules unchanged) if a matcher is invalid.
func (e *Engine) SetRules(rules []Rule) error {
	compiled := make([]compiledRule, 0, len(rules))

	for _, r := range rules {
		cr := compiledRule{rule: r}

		if r.BodyMatcher != "" {
			re, err := regexp.Compile(r.BodyMatcher)
			if err != nil {
				return fmt.Errorf("matchreplace: invalid body matcher for rule %q: %w", r.Name, err)
			}
			cr.bodyRe = re
		}

		compiled = append(compiled, cr)
	}

	e.mu.Lock()
	e.rules = compiled
	e.mu.Unlock()

	return nil
}

// RequestModifier applies the request-phase rules and chains to next.
func (e *Engine) RequestModifier(next proxy.RequestModifyFunc) proxy.RequestModifyFunc {
	return func(req *http.Request) {
		e.applyRequest(req)
		next(req)
	}
}

// ResponseModifier applies the response-phase rules and chains to next.
func (e *Engine) ResponseModifier(next proxy.ResponseModifyFunc) proxy.ResponseModifyFunc {
	return func(res *http.Response) error {
		e.applyResponse(res)
		return next(res)
	}
}

func (e *Engine) applyRequest(req *http.Request) {
	e.mu.RLock()
	rules := e.rules
	e.mu.RUnlock()

	for _, cr := range rules {
		if !cr.rule.Enabled || cr.rule.Phase != PhaseRequest {
			continue
		}

		if !conditionMatches(cr.rule.Condition, func() (httpql.Record, error) { return httpql.RecordFromRequest(req) }) {
			continue
		}

		applyHeaderOp(cr.rule, req.Header)

		if cr.bodyRe != nil {
			replaceBody(cr, &req.Body, &req.ContentLength, req.Header)
		}
	}
}

func (e *Engine) applyResponse(res *http.Response) {
	e.mu.RLock()
	rules := e.rules
	e.mu.RUnlock()

	for _, cr := range rules {
		if !cr.rule.Enabled || cr.rule.Phase != PhaseResponse {
			continue
		}

		if !conditionMatches(cr.rule.Condition, func() (httpql.Record, error) { return httpql.RecordFromResponse(res) }) {
			continue
		}

		applyHeaderOp(cr.rule, res.Header)

		if cr.bodyRe != nil {
			replaceBody(cr, &res.Body, &res.ContentLength, res.Header)
		}
	}
}

func conditionMatches(cond httpql.Expression, record func() (httpql.Record, error)) bool {
	if cond == nil {
		return true
	}

	rec, err := record()
	if err != nil {
		return false
	}

	match, err := httpql.Eval(cond, rec)
	if err != nil {
		return false
	}

	return match
}

func applyHeaderOp(rule Rule, header http.Header) {
	if rule.HeaderName == "" {
		return
	}

	if rule.RemoveHeader {
		header.Del(rule.HeaderName)
		return
	}

	header.Set(rule.HeaderName, rule.HeaderValue)
}

func replaceBody(cr compiledRule, body *io.ReadCloser, contentLength *int64, header http.Header) {
	buf, err := httpql.DrainBody(body)
	if err != nil {
		return
	}

	replaced := cr.bodyRe.ReplaceAll(buf, []byte(cr.rule.BodyReplacement))

	*body = io.NopCloser(bytes.NewReader(replaced))
	*contentLength = int64(len(replaced))
	header.Set("Content-Length", strconv.Itoa(len(replaced)))
}
