package scope

import (
	"bytes"
	"encoding/gob"
	"net/http"
	"regexp"
	"sync"
)

type Scope struct {
	rules []Rule
	mu    sync.RWMutex
}

type Rule struct {
	URL    *regexp.Regexp
	Header Header
	Body   *regexp.Regexp

	// Exclude inverts the rule: a request matching an exclude rule is out of
	// scope, taking precedence over any include rule it also matches.
	Exclude bool
}

type Header struct {
	Key   *regexp.Regexp
	Value *regexp.Regexp
}

func (s *Scope) Rules() []Rule {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.rules
}

func (s *Scope) SetRules(rules []Rule) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.rules = rules
}

func (s *Scope) Match(req *http.Request, body []byte) bool {
	var rawURL string
	if req.URL != nil {
		rawURL = req.URL.String()
	}

	return s.InScope(rawURL, req.Header, body)
}

// InScope reports whether a request identified by its URL, headers and body is
// in scope. A request is in scope when it matches no exclude rule and either
// there are no include rules matching against an exclude-only set or it matches
// an include rule. An empty rule set puts everything out of scope.
func (s *Scope) InScope(rawURL string, header http.Header, body []byte) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.rules) == 0 {
		return false
	}

	matchedInclude := false
	hasInclude := false

	for _, rule := range s.rules {
		matches := rule.matches(rawURL, header, body)

		if rule.Exclude {
			if matches {
				return false
			}

			continue
		}

		hasInclude = true
		if matches {
			matchedInclude = true
		}
	}

	// With only exclude rules, everything that isn't excluded is in scope.
	if !hasInclude {
		return true
	}

	return matchedInclude
}

func (r Rule) Match(req *http.Request, body []byte) bool {
	var rawURL string
	if req.URL != nil {
		rawURL = req.URL.String()
	}

	return r.matches(rawURL, req.Header, body)
}

// matches reports whether the rule matches the given request fields, ignoring
// its Exclude flag (the caller applies exclusion semantics).
func (r Rule) matches(rawURL string, header http.Header, body []byte) bool {
	if r.URL != nil && rawURL != "" {
		if r.URL.MatchString(rawURL) {
			return true
		}
	}

	for key, values := range header {
		var keyMatches, valueMatches bool

		if r.Header.Key != nil {
			if r.Header.Key.MatchString(key) {
				keyMatches = true
			}
		}

		if r.Header.Value != nil {
			for _, value := range values {
				if r.Header.Value.MatchString(value) {
					valueMatches = true
					break
				}
			}
		}
		// When only key or value is set, match on whatever is set.
		// When both are set, both must match.
		switch {
		case r.Header.Key != nil && r.Header.Value == nil && keyMatches:
			return true
		case r.Header.Key == nil && r.Header.Value != nil && valueMatches:
			return true
		case r.Header.Key != nil && r.Header.Value != nil && keyMatches && valueMatches:
			return true
		}
	}

	if r.Body != nil {
		if r.Body.Match(body) {
			return true
		}
	}

	return false
}

func regexpToString(r *regexp.Regexp) string {
	if r == nil {
		return ""
	}

	return r.String()
}

func stringToRegexp(s string) (*regexp.Regexp, error) {
	if s == "" {
		return nil, nil
	}

	return regexp.Compile(s)
}

type ruleDTO struct {
	URL    string
	Header struct {
		Key   string
		Value string
	}
	Body    string
	Exclude bool
}

func (r Rule) MarshalBinary() ([]byte, error) {
	dto := ruleDTO{
		URL:     regexpToString(r.URL),
		Body:    regexpToString(r.Body),
		Exclude: r.Exclude,
	}
	dto.Header.Key = regexpToString(r.Header.Key)
	dto.Header.Value = regexpToString(r.Header.Value)

	buf := bytes.Buffer{}

	err := gob.NewEncoder(&buf).Encode(dto)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (r *Rule) UnmarshalBinary(data []byte) error {
	dto := ruleDTO{}

	err := gob.NewDecoder(bytes.NewReader(data)).Decode(&dto)
	if err != nil {
		return err
	}

	url, err := stringToRegexp(dto.URL)
	if err != nil {
		return err
	}

	headerKey, err := stringToRegexp(dto.Header.Key)
	if err != nil {
		return err
	}

	headerValue, err := stringToRegexp(dto.Header.Value)
	if err != nil {
		return err
	}

	body, err := stringToRegexp(dto.Body)
	if err != nil {
		return err
	}

	*r = Rule{
		URL: url,
		Header: Header{
			Key:   headerKey,
			Value: headerValue,
		},
		Body:    body,
		Exclude: dto.Exclude,
	}

	return nil
}
