// Package aitools wires Hettix's capabilities into agent tools. It is the bridge
// between the generic agent loop and the request-log, sender and scope services,
// kept separate so pkg/agent stays dependency-free.
package aitools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/oklog/ulid"

	"github.com/Vibe-Coding-Base/Hettix/pkg/agent"
	"github.com/Vibe-Coding-Base/Hettix/pkg/httpql"
	"github.com/Vibe-Coding-Base/Hettix/pkg/reqlog"
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

Be concise. When you reference a request, include its id so the operator can
open it.`

// RequestLogService is the subset of the request-log service the tools use.
type RequestLogService interface {
	FindByQuery(ctx context.Context, expr httpql.Expression, limit int) ([]reqlog.RequestLog, error)
	FindRequestLogByID(ctx context.Context, id ulid.ULID) (reqlog.RequestLog, error)
}

// NewRegistry builds the agent tool registry backed by the given services.
func NewRegistry(reqLog RequestLogService) *agent.Registry {
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

func urlString(rl reqlog.RequestLog) string {
	if rl.URL == nil {
		return ""
	}

	return rl.URL.String()
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
