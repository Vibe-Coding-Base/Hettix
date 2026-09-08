package scope_test

import (
	"net/http"
	"net/url"
	"regexp"
	"testing"

	"github.com/Vibe-Coding-Base/Hettix/pkg/scope"
)

func req(t *testing.T, rawURL string) *http.Request {
	t.Helper()

	u, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("parse url: %v", err)
	}

	return &http.Request{URL: u, Header: http.Header{}}
}

func urlRule(pattern string, exclude bool) scope.Rule {
	return scope.Rule{URL: regexp.MustCompile(pattern), Exclude: exclude}
}

func TestInScope(t *testing.T) {
	tests := []struct {
		name  string
		rules []scope.Rule
		url   string
		want  bool
	}{
		{"empty scope excludes everything", nil, "https://a.com/x", false},
		{"include match", []scope.Rule{urlRule(`a\.com`, false)}, "https://a.com/x", true},
		{"include no match", []scope.Rule{urlRule(`a\.com`, false)}, "https://b.com/x", false},
		{
			"exclude beats include",
			[]scope.Rule{urlRule(`example\.com`, false), urlRule(`/admin`, true)},
			"https://example.com/admin",
			false,
		},
		{
			"include without exclusion match",
			[]scope.Rule{urlRule(`example\.com`, false), urlRule(`/admin`, true)},
			"https://example.com/users",
			true,
		},
		{
			"exclude-only lets everything else through",
			[]scope.Rule{urlRule(`/logout`, true)},
			"https://example.com/home",
			true,
		},
		{
			"exclude-only blocks the excluded",
			[]scope.Rule{urlRule(`/logout`, true)},
			"https://example.com/logout",
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &scope.Scope{}
			s.SetRules(tt.rules)

			if got := s.Match(req(t, tt.url), nil); got != tt.want {
				t.Errorf("Match(%q) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}

func TestRuleMarshalRoundTripExclude(t *testing.T) {
	rule := scope.Rule{URL: regexp.MustCompile(`/admin`), Exclude: true}

	data, err := rule.MarshalBinary()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got scope.Rule
	if err := got.UnmarshalBinary(data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if !got.Exclude {
		t.Error("expected Exclude to survive the round trip")
	}
	if got.URL == nil || got.URL.String() != `/admin` {
		t.Errorf("expected URL /admin, got %v", got.URL)
	}
}
