package intercept

import (
	"net/http"

	"github.com/Vibe-Coding-Base/Hettix/pkg/httpql"
	"github.com/Vibe-Coding-Base/Hettix/pkg/scope"
)

// MatchRequestFilter reports whether an in-flight request satisfies the HTTPQL
// request filter.
func MatchRequestFilter(req *http.Request, expr httpql.Expression) (bool, error) {
	rec, err := httpql.RecordFromRequest(req)
	if err != nil {
		return false, err
	}

	return httpql.Eval(expr, rec)
}

// MatchResponseFilter reports whether an in-flight response satisfies the HTTPQL
// response filter.
func MatchResponseFilter(res *http.Response, expr httpql.Expression) (bool, error) {
	rec, err := httpql.RecordFromResponse(res)
	if err != nil {
		return false, err
	}

	return httpql.Eval(expr, rec)
}

// MatchRequestScope reports whether an in-flight request matches any scope rule.
func MatchRequestScope(req *http.Request, s *scope.Scope) (bool, error) {
	for _, rule := range s.Rules() {
		if rule.URL != nil && req.URL != nil {
			if matches := rule.URL.MatchString(req.URL.String()); matches {
				return true, nil
			}
		}

		for key, values := range req.Header {
			var keyMatches, valueMatches bool

			if rule.Header.Key != nil {
				if matches := rule.Header.Key.MatchString(key); matches {
					keyMatches = true
				}
			}

			if rule.Header.Value != nil {
				for _, value := range values {
					if matches := rule.Header.Value.MatchString(value); matches {
						valueMatches = true
						break
					}
				}
			}

			switch {
			case rule.Header.Key != nil && rule.Header.Value == nil && keyMatches:
				return true, nil
			case rule.Header.Key == nil && rule.Header.Value != nil && valueMatches:
				return true, nil
			case rule.Header.Key != nil && rule.Header.Value != nil && keyMatches && valueMatches:
				return true, nil
			}
		}

		if rule.Body != nil {
			body, err := httpql.DrainBody(&req.Body)
			if err != nil {
				return false, err
			}

			if matches := rule.Body.Match(body); matches {
				return true, nil
			}
		}
	}

	return false, nil
}
