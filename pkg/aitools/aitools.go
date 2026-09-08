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
	"github.com/Vibe-Coding-Base/Hettix/pkg/httpql"
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
response, and set_scope replaces the project scope with URL patterns. These
change state or send traffic, so they are only available in assist or auto mode
and are refused for out-of-scope targets. Investigate first, then act
deliberately.

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

// NewRegistry builds the agent tool registry backed by the given services. The
// sender and project services power the mutating tools; pass nil to omit them.
func NewRegistry(reqLog RequestLogService, senderSvc SenderService, projSvc ProjectService) *agent.Registry {
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
