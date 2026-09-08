package matchreplace_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Vibe-Coding-Base/Hettix/pkg/httpql"
	"github.com/Vibe-Coding-Base/Hettix/pkg/matchreplace"
)

func newRequest(method, url, body string) *http.Request {
	req := httptest.NewRequest(method, url, strings.NewReader(body))
	req.Header.Set("Content-Type", "text/plain")
	return req
}

func runRequest(e *matchreplace.Engine, req *http.Request) {
	e.RequestModifier(func(*http.Request) {})(req)
}

func mustParse(t *testing.T, q string) httpql.Expression {
	t.Helper()
	expr, err := httpql.Parse(q)
	if err != nil {
		t.Fatalf("parse %q: %v", q, err)
	}
	return expr
}

func TestHeaderSetWithCondition(t *testing.T) {
	e := matchreplace.NewEngine()
	if err := e.SetRules([]matchreplace.Rule{{
		Name: "add header on POST", Enabled: true, Phase: matchreplace.PhaseRequest,
		Condition:  mustParse(t, `req.method eq "POST"`),
		HeaderName: "X-Injected", HeaderValue: "yes",
	}}); err != nil {
		t.Fatal(err)
	}

	post := newRequest("POST", "https://api.example.com/", "")
	runRequest(e, post)
	if post.Header.Get("X-Injected") != "yes" {
		t.Fatal("expected header to be set on POST")
	}

	get := newRequest("GET", "https://api.example.com/", "")
	runRequest(e, get)
	if get.Header.Get("X-Injected") != "" {
		t.Fatal("header must not be set on GET (condition mismatch)")
	}
}

func TestBodyReplace(t *testing.T) {
	e := matchreplace.NewEngine()
	if err := e.SetRules([]matchreplace.Rule{{
		Name: "redact", Enabled: true, Phase: matchreplace.PhaseRequest,
		BodyMatcher: `password=[^&]+`, BodyReplacement: "password=REDACTED",
	}}); err != nil {
		t.Fatal(err)
	}

	req := newRequest("POST", "https://api.example.com/login", "user=admin&password=secret")
	runRequest(e, req)

	body, _ := io.ReadAll(req.Body)
	if string(body) != "user=admin&password=REDACTED" {
		t.Fatalf("unexpected body: %q", body)
	}
	if req.Header.Get("Content-Length") != "28" {
		t.Fatalf("Content-Length not updated: %q", req.Header.Get("Content-Length"))
	}
}

func TestRemoveHeaderResponse(t *testing.T) {
	e := matchreplace.NewEngine()
	if err := e.SetRules([]matchreplace.Rule{{
		Name: "strip CSP", Enabled: true, Phase: matchreplace.PhaseResponse,
		HeaderName: "Content-Security-Policy", RemoveHeader: true,
	}}); err != nil {
		t.Fatal(err)
	}

	res := &http.Response{Header: http.Header{"Content-Security-Policy": {"default-src 'self'"}}, Body: http.NoBody}
	if err := e.ResponseModifier(func(*http.Response) error { return nil })(res); err != nil {
		t.Fatal(err)
	}

	if res.Header.Get("Content-Security-Policy") != "" {
		t.Fatal("expected CSP header to be removed")
	}
}

func TestDisabledRuleSkipped(t *testing.T) {
	e := matchreplace.NewEngine()
	if err := e.SetRules([]matchreplace.Rule{{
		Name: "off", Enabled: false, Phase: matchreplace.PhaseRequest,
		HeaderName: "X-Should-Not-Appear", HeaderValue: "1",
	}}); err != nil {
		t.Fatal(err)
	}

	req := newRequest("GET", "https://x/", "")
	runRequest(e, req)
	if req.Header.Get("X-Should-Not-Appear") != "" {
		t.Fatal("disabled rule must not apply")
	}
}

func TestInvalidMatcherRejected(t *testing.T) {
	e := matchreplace.NewEngine()
	if err := e.SetRules([]matchreplace.Rule{{Name: "bad", Enabled: true, BodyMatcher: "("}}); err == nil {
		t.Fatal("expected an error for an invalid regexp matcher")
	}
}
