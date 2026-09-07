package httpql_test

import (
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Vibe-Coding-Base/Hettix/pkg/httpql"
)

func TestParseString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  string
	}{
		{`req.method eq "GET"`, `(req.method eq "GET")`},
		{`req.method = "GET"`, `(req.method eq "GET")`},
		{`resp.code gt 200`, `(resp.code gt 200)`},
		{`resp.code > 200`, `(resp.code gt 200)`},
		{`res.statusCode eq 404`, `(resp.code eq 404)`},
		{`req.tls eq true`, `(req.tls eq true)`},
		{`req.header["Authorization"] regex "^Bearer "`, `(req.header["Authorization"] regex "^Bearer ")`},
		{`a AND b`, `("a" AND "b")`},
		{`a b`, `("a" AND "b")`},
		{`a OR b AND c`, `("a" OR ("b" AND "c"))`},
		{`NOT req.tls eq true`, `(NOT (req.tls eq true))`},
		{`(a OR b) AND c`, `(("a" OR "b") AND "c")`},
		{`req.host cont "api" // comment`, `(req.host cont "api")`},
	}

	for _, tt := range tests {
		expr, err := httpql.Parse(tt.input)
		if err != nil {
			t.Errorf("Parse(%q) error: %v", tt.input, err)
			continue
		}
		if got := expr.String(); got != tt.want {
			t.Errorf("Parse(%q) = %s, want %s", tt.input, got, tt.want)
		}
	}
}

func TestParseErrors(t *testing.T) {
	t.Parallel()

	for _, input := range []string{
		`req.method eq`,
		`req. eq "x"`,
		`req.method "x"`,
		`(a`,
		`"unterminated`,
	} {
		if _, err := httpql.Parse(input); err == nil {
			t.Errorf("Parse(%q) expected error, got nil", input)
		}
	}
}

func TestParseDepthLimit(t *testing.T) {
	t.Parallel()

	_, err := httpql.Parse(strings.Repeat("(", 1_000_000))
	if !errors.Is(err, httpql.ErrMaxDepthExceeded) {
		t.Fatalf("expected ErrMaxDepthExceeded, got %v", err)
	}
}

func TestEmptyQueryMatchesEverything(t *testing.T) {
	t.Parallel()

	expr, err := httpql.Parse("   ")
	if err != nil {
		t.Fatal(err)
	}

	ok, err := httpql.Eval(expr, httpql.Record{})
	if err != nil || !ok {
		t.Fatalf("empty query should match, got ok=%v err=%v", ok, err)
	}
}

func sampleRecord() httpql.Record {
	return httpql.Record{
		Request: httpql.RequestData{
			ID: "01ABC", Method: "POST", Host: "api.example.com", Path: "/v1/login",
			Ext: "", URL: "https://api.example.com/v1/login?debug=1", Proto: "HTTP/1.1",
			Port: 443, Len: 120, TLS: true, CreatedAt: time.Date(2026, 9, 7, 12, 0, 0, 0, time.UTC),
			Header: http.Header{"Authorization": {"Bearer abc123"}, "Content-Type": {"application/json"}},
			Body:   `{"user":"admin","password":"secret"}`,
		},
		Response: &httpql.ResponseData{
			Proto: "HTTP/1.1", Reason: "200 OK", Code: 200, Len: 15, RoundTrip: 4200,
			Header: http.Header{"Content-Type": {"application/json"}},
			Body:   `{"token":"xyz"}`,
		},
	}
}

func TestEval(t *testing.T) {
	t.Parallel()

	rec := sampleRecord()

	tests := []struct {
		query string
		want  bool
	}{
		{`req.method eq "POST"`, true},
		{`req.method eq "GET"`, false},
		{`req.method ne "GET"`, true},
		{`req.host cont "example"`, true},
		{`req.host cont "EXAMPLE"`, true},
		{`req.host ncont "google"`, true},
		{`resp.code eq 200`, true},
		{`resp.code gt 200`, false},
		{`resp.code gte 200`, true},
		{`resp.code lt 500`, true},
		{`req.tls eq true`, true},
		{`req.port eq 443`, true},
		{`resp.roundtrip gt 3000`, true},
		{`req.header["Authorization"] regex "^Bearer "`, true},
		{`req.header["Authorization"] cont "abc123"`, true},
		{`req.header["X-Missing"] eq "y"`, false},
		{`req.body cont "password"`, true},
		{`req.body like "%admin%"`, true},
		{`req.path eq "/v1/login" AND resp.code eq 200`, true},
		{`req.path eq "/nope" OR resp.code eq 200`, true},
		{`NOT req.tls eq true`, false},
		{`req.created_at gt "2026-01-01T00:00:00Z"`, true},
		{`req.created_at lt "2026-01-01T00:00:00Z"`, false},
		{`"secret"`, true},
		{`"notpresent"`, false},
		{`req.method eq "POST" AND NOT req.tls eq false`, true},
	}

	for _, tt := range tests {
		expr, err := httpql.Parse(tt.query)
		if err != nil {
			t.Errorf("Parse(%q) error: %v", tt.query, err)
			continue
		}

		got, err := httpql.Eval(expr, rec)
		if err != nil {
			t.Errorf("Eval(%q) error: %v", tt.query, err)
			continue
		}

		if got != tt.want {
			t.Errorf("Eval(%q) = %v, want %v", tt.query, got, tt.want)
		}
	}
}

func TestEvalResponseAbsent(t *testing.T) {
	t.Parallel()

	rec := sampleRecord()
	rec.Response = nil

	expr, _ := httpql.Parse(`resp.code eq 200`)
	got, err := httpql.Eval(expr, rec)
	if err != nil {
		t.Fatal(err)
	}
	if got {
		t.Fatal("resp.code eq 200 should be false when there is no response")
	}
}

func TestEvalTypeErrors(t *testing.T) {
	t.Parallel()

	rec := sampleRecord()

	// Integer field compared against a string value.
	expr, err := httpql.Parse(`resp.code eq "abc"`)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := httpql.Eval(expr, rec); err == nil {
		t.Fatal("expected a type error comparing an int field to a string")
	}
}
