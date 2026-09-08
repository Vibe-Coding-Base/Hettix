// Package aitools wires Hettix's capabilities into agent tools. It is the bridge
// between the generic agent loop and the request-log, sender and scope services,
// kept separate so pkg/agent stays dependency-free.
package aitools

import (
	"context"
	crand "crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/oklog/ulid"

	"github.com/Vibe-Coding-Base/Hettix/pkg/agent"
	"github.com/Vibe-Coding-Base/Hettix/pkg/finding"
	"github.com/Vibe-Coding-Base/Hettix/pkg/httpql"
	"github.com/Vibe-Coding-Base/Hettix/pkg/intruder"
	"github.com/Vibe-Coding-Base/Hettix/pkg/matchreplace"
	"github.com/Vibe-Coding-Base/Hettix/pkg/reqlog"
	"github.com/Vibe-Coding-Base/Hettix/pkg/scope"
	"github.com/Vibe-Coding-Base/Hettix/pkg/sender"
)

const maxBodyChars = 2000

// SystemPrompt primes the assistant with its role, the query language and how to
// use the tools.
const SystemPrompt = `You are Hetty, an assistant embedded in the Hettix HTTP security testing proxy.
You help a security researcher investigate captured HTTP traffic.

Use the provided tools to answer questions about the traffic; do not invent
requests or responses. Query traffic with HTTPQL, for example:
  req.method eq "POST" and resp.code gte 500
  req.host cont "api.target.com"
  req.header["Authorization"] regex "^Bearer "
Fields: req.{method,host,path,url,body,header["Name"],tls,...} and
resp.{code,reason,body,header["Name"],roundtrip}. Operators: eq ne cont regex
gt gte lt lte, combined with and/or/not.

You can also act: replay_request resends a captured request and returns the new
response, set_scope replaces the project scope with URL patterns, and run_fuzzer
starts a fuzzing attack that resends a request once per payload (mark the
insertion point in the URL or body with the § character), whose outcomes you
read back with get_fuzz_results. These change state or send traffic, so they are
only available in assist or auto mode and are refused for out-of-scope targets.
Investigate first, then act deliberately. Record what you confirm with
create_finding, linking the request log id that evidences it.

Be concise. When you reference a request, include its id so the operator can
open it.`

// RequestLogService is the subset of the request-log service the tools use.
type RequestLogService interface {
	FindByQuery(ctx context.Context, expr httpql.Expression, limit int) ([]reqlog.RequestLog, error)
	FindRequestLogByID(ctx context.Context, id ulid.ULID) (reqlog.RequestLog, error)
}

// SenderService is the subset of the sender service the tools use.
type SenderService interface {
	CloneFromRequestLog(ctx context.Context, reqLogID ulid.ULID) (sender.Request, error)
	SendRequest(ctx context.Context, id ulid.ULID) (sender.Request, error)
}

// ProjectService is the subset of the project service the tools use.
type ProjectService interface {
	SetScopeRules(ctx context.Context, rules []scope.Rule) error
	Scope() *scope.Scope
	MatchReplaceRules(ctx context.Context) ([]matchreplace.Rule, error)
	SetMatchReplaceRules(ctx context.Context, rules []matchreplace.Rule) error
}

// IntruderService is the subset of the fuzzer service the tools use.
type IntruderService interface {
	StartAttack(ctx context.Context, name string, tmpl intruder.Request, payloads []string) (intruder.Attack, error)
	AttackByID(ctx context.Context, id ulid.ULID) (intruder.Attack, error)
	Results(ctx context.Context, attackID ulid.ULID) ([]intruder.Result, error)
}

// FindingService is the subset of the finding service the tools use.
type FindingService interface {
	CreateFinding(
		ctx context.Context,
		title, description string,
		severity finding.Severity,
		requestLogID *ulid.ULID,
	) (finding.Finding, error)
}

// NewRegistry builds the agent tool registry backed by the given services. The
// sender and project services power the mutating tools; pass nil to omit them.
// The intruder service adds the fuzzing tools; pass nil to omit them.
func NewRegistry(
	reqLog RequestLogService,
	senderSvc SenderService,
	projSvc ProjectService,
	intruderSvc IntruderService,
	findingSvc FindingService,
) *agent.Registry {
	reg := agent.NewRegistry()

	reg.Register(agent.Tool{
		Name:        "search_traffic",
		Description: "Search captured HTTP traffic with an HTTPQL query. Returns matching requests, newest first.",
		Schema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"query": {"type": "string", "description": "An HTTPQL query, e.g. req.method eq \"POST\""},
				"limit": {"type": "integer", "description": "Maximum results (default 20, max 50)"}
			},
			"required": ["query"]
		}`),
		Handler: searchTraffic(reqLog),
	})

	reg.Register(agent.Tool{
		Name:        "get_request",
		Description: "Fetch the full request and response for a request log id.",
		Schema: json.RawMessage(`{
			"type": "object",
			"properties": {"id": {"type": "string", "description": "The request log id (ULID)"}},
			"required": ["id"]
		}`),
		Handler: getRequest(reqLog),
	})

	if senderSvc != nil && projSvc != nil {
		reg.Register(agent.Tool{
			Name:        "replay_request",
			Description: "Resend a captured request (optionally already modified in Sender) and return the new response. The target must be in scope.",
			Mutating:    true,
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {"id": {"type": "string", "description": "The request log id (ULID) to replay"}},
				"required": ["id"]
			}`),
			Handler: replayRequest(reqLog, senderSvc, projSvc),
		})

		reg.Register(agent.Tool{
			Name:        "set_scope",
			Description: "Replace the project scope with a list of URL regular expressions. In-scope traffic is what the proxy focuses on.",
			Mutating:    true,
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"url_patterns": {
						"type": "array",
						"items": {"type": "string"},
						"description": "Regular expressions matched against request URLs"
					}
				},
				"required": ["url_patterns"]
			}`),
			Handler: setScope(projSvc),
		})

		reg.Register(agent.Tool{
			Name:        "add_match_replace_rule",
			Description: "Add a rule that rewrites proxied traffic: set or remove a header, or regexp-replace the body, of matching requests or responses.",
			Mutating:    true,
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"name": {"type": "string"},
					"phase": {"type": "string", "enum": ["request", "response"], "description": "Whether the rule runs on requests or responses"},
					"condition": {"type": "string", "description": "Optional HTTPQL condition; when omitted the rule always applies"},
					"header_name": {"type": "string", "description": "Header to set or remove"},
					"header_value": {"type": "string"},
					"remove_header": {"type": "boolean"},
					"body_matcher": {"type": "string", "description": "Regexp matched against the body"},
					"body_replacement": {"type": "string"}
				},
				"required": ["name", "phase"]
			}`),
			Handler: addMatchReplaceRule(projSvc),
		})
	}

	if intruderSvc != nil && projSvc != nil {
		reg.Register(agent.Tool{
			Name: "run_fuzzer",
			Description: "Start a fuzzing attack: send a request once per payload, substituting the § marker " +
				"in the URL or body with each payload. Returns the attack id. The target must be in scope.",
			Mutating: true,
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"name": {"type": "string", "description": "A short name for the attack"},
					"method": {"type": "string", "description": "HTTP method (default GET)"},
					"url": {"type": "string", "description": "Target URL; put § where each payload goes, e.g. https://t/search?q=§"},
					"body": {"type": "string", "description": "Optional request body; may also contain §"},
					"payloads": {"type": "array", "items": {"type": "string"}, "description": "The payloads to try"}
				},
				"required": ["url", "payloads"]
			}`),
			Handler: runFuzzer(projSvc, intruderSvc),
		})

		reg.Register(agent.Tool{
			Name:        "get_fuzz_results",
			Description: "Summarize the results of a fuzzing attack by its id: progress and notable responses.",
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {"id": {"type": "string", "description": "The attack id (ULID)"}},
				"required": ["id"]
			}`),
			Handler: getFuzzResults(intruderSvc),
		})
	}

	if findingSvc != nil {
		reg.Register(agent.Tool{
			Name: "create_finding",
			Description: "Record a security finding: a titled, severity-rated note, optionally linked to the " +
				"request log id that evidences it.",
			Mutating: true,
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"title": {"type": "string"},
					"description": {"type": "string"},
					"severity": {"type": "string", "enum": ["info", "low", "medium", "high", "critical"]},
					"request_log_id": {"type": "string", "description": "Optional request log id (ULID) that evidences the finding"}
				},
				"required": ["title", "severity"]
			}`),
			Handler: createFinding(findingSvc),
		})
	}

	return reg
}

func searchTraffic(reqLog RequestLogService) agent.Handler {
	return func(ctx context.Context, args json.RawMessage) (string, error) {
		var in struct {
			Query string `json:"query"`
			Limit int    `json:"limit"`
		}
		if err := json.Unmarshal(args, &in); err != nil {
			return "", fmt.Errorf("invalid arguments: %w", err)
		}

		expr, err := httpql.Parse(in.Query)
		if err != nil {
			return "", err
		}

		limit := in.Limit
		if limit <= 0 || limit > 50 {
			limit = 20
		}

		logs, err := reqLog.FindByQuery(ctx, expr, limit)
		if err != nil {
			return "", err
		}

		if len(logs) == 0 {
			return "No matching requests.", nil
		}

		var b strings.Builder
		fmt.Fprintf(&b, "%d matching requests:\n", len(logs))
		for _, rl := range logs {
			status := "(no response)"
			if rl.Response != nil {
				status = fmt.Sprintf("%d", rl.Response.StatusCode)
			}
			fmt.Fprintf(&b, "%s  %s %s  -> %s\n", rl.ID.String(), rl.Method, urlString(rl), status)
		}

		return b.String(), nil
	}
}

func getRequest(reqLog RequestLogService) agent.Handler {
	return func(ctx context.Context, args json.RawMessage) (string, error) {
		var in struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(args, &in); err != nil {
			return "", fmt.Errorf("invalid arguments: %w", err)
		}

		id, err := ulid.Parse(in.ID)
		if err != nil {
			return "", fmt.Errorf("invalid request id %q", in.ID)
		}

		rl, err := reqLog.FindRequestLogByID(ctx, id)
		if err != nil {
			return "", err
		}

		var b strings.Builder
		fmt.Fprintf(&b, "%s %s %s\n", rl.Method, urlString(rl), rl.Proto)
		writeHeaders(&b, rl.Header)
		writeBody(&b, rl.Body)

		if rl.Response != nil {
			fmt.Fprintf(&b, "\n%s %d %s\n", rl.Response.Proto, rl.Response.StatusCode, rl.Response.Status)
			writeHeaders(&b, rl.Response.Header)
			writeBody(&b, rl.Response.Body)
		}

		return b.String(), nil
	}
}

func replayRequest(reqLog RequestLogService, senderSvc SenderService, projSvc ProjectService) agent.Handler {
	return func(ctx context.Context, args json.RawMessage) (string, error) {
		var in struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(args, &in); err != nil {
			return "", fmt.Errorf("invalid arguments: %w", err)
		}

		id, err := ulid.Parse(in.ID)
		if err != nil {
			return "", fmt.Errorf("invalid request id %q", in.ID)
		}

		rl, err := reqLog.FindRequestLogByID(ctx, id)
		if err != nil {
			return "", err
		}

		// Scope enforcement at the tool layer: refuse to send traffic to a host
		// the operator has not put in scope. An empty scope means unrestricted.
		if s := projSvc.Scope(); len(s.Rules()) > 0 && !rl.MatchScope(s) {
			return "", fmt.Errorf("refusing to replay: %s is out of scope", urlString(rl))
		}

		cloned, err := senderSvc.CloneFromRequestLog(ctx, id)
		if err != nil {
			return "", err
		}

		sent, err := senderSvc.SendRequest(ctx, cloned.ID)
		if err != nil {
			return "", err
		}

		var b strings.Builder
		fmt.Fprintf(&b, "Replayed %s %s\n", sent.Method, senderURL(sent))
		if sent.Response != nil {
			fmt.Fprintf(&b, "-> %d %s (%dms)\n", sent.Response.StatusCode, sent.Response.Status, sent.Response.RoundTripMillis)
			writeBody(&b, sent.Response.Body)
		} else {
			b.WriteString("(no response)\n")
		}

		return b.String(), nil
	}
}

func runFuzzer(projSvc ProjectService, intruderSvc IntruderService) agent.Handler {
	return func(ctx context.Context, args json.RawMessage) (string, error) {
		var in struct {
			Name     string   `json:"name"`
			Method   string   `json:"method"`
			URL      string   `json:"url"`
			Body     string   `json:"body"`
			Payloads []string `json:"payloads"`
		}
		if err := json.Unmarshal(args, &in); err != nil {
			return "", fmt.Errorf("invalid arguments: %w", err)
		}

		method := in.Method
		if method == "" {
			method = "GET"
		}

		tmpl := intruder.Request{Method: method, URL: in.URL, Body: in.Body}
		if !tmpl.HasMarker() {
			return "", errors.New("the request needs a § marker in the url or body to mark the insertion point")
		}

		// Scope enforcement at the tool layer: refuse to fuzz a host the operator
		// has not put in scope. An empty scope means unrestricted.
		check := strings.ReplaceAll(in.URL, intruder.Marker, "")
		if s := projSvc.Scope(); len(s.Rules()) > 0 && !s.InScope(check, nil, nil) {
			return "", fmt.Errorf("refusing to fuzz: %s is out of scope", check)
		}

		name := in.Name
		if name == "" {
			name = "Agent attack"
		}

		attack, err := intruderSvc.StartAttack(ctx, name, tmpl, in.Payloads)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf(
			"Started fuzzing attack %s (%q) with %d payloads. Call get_fuzz_results with this id to see the outcomes.",
			attack.ID, name, len(in.Payloads),
		), nil
	}
}

func getFuzzResults(intruderSvc IntruderService) agent.Handler {
	return func(ctx context.Context, args json.RawMessage) (string, error) {
		var in struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(args, &in); err != nil {
			return "", fmt.Errorf("invalid arguments: %w", err)
		}

		id, err := ulid.Parse(in.ID)
		if err != nil {
			return "", fmt.Errorf("invalid attack id %q", in.ID)
		}

		attack, err := intruderSvc.AttackByID(ctx, id)
		if err != nil {
			return "", err
		}

		results, err := intruderSvc.Results(ctx, id)
		if err != nil {
			return "", err
		}

		var b strings.Builder
		fmt.Fprintf(&b, "Attack %q: %s, %d/%d done.\n", attack.Name, attack.Status, attack.Completed, attack.Total)

		counts := map[int]int{}
		var notable []intruder.Result

		for _, r := range results {
			counts[r.StatusCode]++
			if r.Error != "" || r.StatusCode == 0 || r.StatusCode >= 400 {
				notable = append(notable, r)
			}
		}

		if len(counts) > 0 {
			b.WriteString("Status codes:")
			for code, n := range counts {
				fmt.Fprintf(&b, " %d×%d", code, n)
			}
			b.WriteString("\n")
		}

		if len(notable) == 0 {
			b.WriteString("No error or 4xx/5xx responses.\n")
		} else {
			fmt.Fprintf(&b, "Notable (%d):\n", len(notable))
			for i, r := range notable {
				if i >= 15 {
					fmt.Fprintf(&b, "... and %d more\n", len(notable)-15)
					break
				}
				if r.Error != "" {
					fmt.Fprintf(&b, "  %q -> error: %s\n", r.Payload, r.Error)
				} else {
					fmt.Fprintf(&b, "  %q -> %d (%d bytes, %dms)\n", r.Payload, r.StatusCode, r.Length, r.DurationMs)
				}
			}
		}

		return b.String(), nil
	}
}

func createFinding(findingSvc FindingService) agent.Handler {
	return func(ctx context.Context, args json.RawMessage) (string, error) {
		var in struct {
			Title        string `json:"title"`
			Description  string `json:"description"`
			Severity     string `json:"severity"`
			RequestLogID string `json:"request_log_id"`
		}
		if err := json.Unmarshal(args, &in); err != nil {
			return "", fmt.Errorf("invalid arguments: %w", err)
		}

		var requestLogID *ulid.ULID
		if in.RequestLogID != "" {
			id, err := ulid.Parse(in.RequestLogID)
			if err != nil {
				return "", fmt.Errorf("invalid request log id %q", in.RequestLogID)
			}

			requestLogID = &id
		}

		f, err := findingSvc.CreateFinding(ctx, in.Title, in.Description, finding.ParseSeverity(in.Severity), requestLogID)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("Recorded %s finding %s: %q.", f.Severity, f.ID, f.Title), nil
	}
}

func setScope(projSvc ProjectService) agent.Handler {
	return func(ctx context.Context, args json.RawMessage) (string, error) {
		var in struct {
			URLPatterns []string `json:"url_patterns"`
		}
		if err := json.Unmarshal(args, &in); err != nil {
			return "", fmt.Errorf("invalid arguments: %w", err)
		}

		rules := make([]scope.Rule, 0, len(in.URLPatterns))
		for _, pattern := range in.URLPatterns {
			re, err := regexp.Compile(pattern)
			if err != nil {
				return "", fmt.Errorf("invalid url pattern %q: %w", pattern, err)
			}

			rules = append(rules, scope.Rule{URL: re})
		}

		if err := projSvc.SetScopeRules(ctx, rules); err != nil {
			return "", err
		}

		return fmt.Sprintf("Scope set to %d rule(s).", len(rules)), nil
	}
}

func addMatchReplaceRule(projSvc ProjectService) agent.Handler {
	return func(ctx context.Context, args json.RawMessage) (string, error) {
		var in struct {
			Name            string `json:"name"`
			Phase           string `json:"phase"`
			Condition       string `json:"condition"`
			HeaderName      string `json:"header_name"`
			HeaderValue     string `json:"header_value"`
			RemoveHeader    bool   `json:"remove_header"`
			BodyMatcher     string `json:"body_matcher"`
			BodyReplacement string `json:"body_replacement"`
		}
		if err := json.Unmarshal(args, &in); err != nil {
			return "", fmt.Errorf("invalid arguments: %w", err)
		}

		if in.HeaderName == "" && in.BodyMatcher == "" {
			return "", errors.New("a rule must set/remove a header or provide a body matcher")
		}

		id, err := ulid.New(ulid.Now(), crand.Reader)
		if err != nil {
			return "", err
		}

		rule := matchreplace.Rule{
			ID:              id,
			Name:            in.Name,
			Enabled:         true,
			Phase:           matchreplace.PhaseRequest,
			HeaderName:      in.HeaderName,
			HeaderValue:     in.HeaderValue,
			RemoveHeader:    in.RemoveHeader,
			BodyMatcher:     in.BodyMatcher,
			BodyReplacement: in.BodyReplacement,
		}

		if strings.EqualFold(in.Phase, "response") {
			rule.Phase = matchreplace.PhaseResponse
		}

		if in.Condition != "" {
			expr, err := httpql.Parse(in.Condition)
			if err != nil {
				return "", fmt.Errorf("invalid condition: %w", err)
			}
			rule.Condition = expr
		}

		existing, err := projSvc.MatchReplaceRules(ctx)
		if err != nil {
			return "", err
		}

		if err := projSvc.SetMatchReplaceRules(ctx, append(existing, rule)); err != nil {
			return "", err
		}

		return fmt.Sprintf("Added match & replace rule %q.", rule.Name), nil
	}
}

func urlString(rl reqlog.RequestLog) string {
	if rl.URL == nil {
		return ""
	}

	return rl.URL.String()
}

func senderURL(req sender.Request) string {
	if req.URL == nil {
		return ""
	}

	return req.URL.String()
}

func writeHeaders(b *strings.Builder, header map[string][]string) {
	for key, values := range header {
		for _, v := range values {
			fmt.Fprintf(b, "%s: %s\n", key, v)
		}
	}
}

func writeBody(b *strings.Builder, body []byte) {
	if len(body) == 0 {
		return
	}

	s := string(body)
	if len(s) > maxBodyChars {
		s = s[:maxBodyChars] + "\n…(truncated)"
	}

	b.WriteString("\n")
	b.WriteString(s)
	b.WriteString("\n")
}
