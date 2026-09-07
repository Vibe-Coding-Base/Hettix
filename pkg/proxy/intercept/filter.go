package intercept

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
	"time"

	"github.com/Vibe-Coding-Base/Hettix/pkg/httpql"
	"github.com/Vibe-Coding-Base/Hettix/pkg/scope"
)

// MatchRequestFilter reports whether an in-flight request satisfies the HTTPQL
// request filter.
func MatchRequestFilter(req *http.Request, expr httpql.Expression) (bool, error) {
	rec, err := recordFromRequest(req)
	if err != nil {
		return false, err
	}

	return httpql.Eval(expr, rec)
}

// MatchResponseFilter reports whether an in-flight response satisfies the HTTPQL
// response filter. The originating request is included so that req.* clauses
// remain usable in response filters.
func MatchResponseFilter(res *http.Response, expr httpql.Expression) (bool, error) {
	rec := httpql.Record{}

	if res.Request != nil {
		reqRec, err := recordFromRequest(res.Request)
		if err != nil {
			return false, err
		}
		rec.Request = reqRec.Request
	}

	resData, err := responseData(res)
	if err != nil {
		return false, err
	}
	rec.Response = &resData

	return httpql.Eval(expr, rec)
}

func recordFromRequest(req *http.Request) (httpql.Record, error) {
	body, err := drainBody(&req.Body)
	if err != nil {
		return httpql.Record{}, err
	}

	reqData := httpql.RequestData{
		Method:    req.Method,
		Proto:     req.Proto,
		Header:    req.Header,
		Body:      string(body),
		Len:       len(body),
		CreatedAt: time.Now(),
	}

	if req.URL != nil {
		reqData.URL = req.URL.String()
		reqData.Host = req.URL.Hostname()
		reqData.Path = req.URL.Path
		reqData.Query = req.URL.RawQuery
		reqData.Ext = path.Ext(req.URL.Path)
		reqData.TLS = req.URL.Scheme == "https"
		reqData.Port = port(req.URL.Port(), reqData.TLS)
	}

	return httpql.Record{Request: reqData}, nil
}

func responseData(res *http.Response) (httpql.ResponseData, error) {
	body, err := drainBody(&res.Body)
	if err != nil {
		return httpql.ResponseData{}, err
	}

	reason := res.Status
	if _, after, ok := splitStatus(res.Status); ok {
		reason = after
	}

	return httpql.ResponseData{
		Proto:  res.Proto,
		Reason: reason,
		Code:   res.StatusCode,
		Header: res.Header,
		Body:   string(body),
		Len:    len(body),
	}, nil
}

// drainBody reads a body fully and replaces it with a fresh reader so it can be
// read again downstream.
func drainBody(body *io.ReadCloser) ([]byte, error) {
	if *body == nil {
		return nil, nil
	}

	buf, err := io.ReadAll(*body)
	if err != nil {
		return nil, fmt.Errorf("intercept: failed to read body: %w", err)
	}

	*body = io.NopCloser(bytes.NewBuffer(buf))

	return buf, nil
}

func splitStatus(status string) (code, reason string, ok bool) {
	for i := 0; i < len(status); i++ {
		if status[i] == ' ' {
			return status[:i], status[i+1:], true
		}
	}

	return status, "", false
}

func port(p string, tls bool) int {
	if p != "" {
		if n, err := strconv.Atoi(p); err == nil {
			return n
		}
	}

	if tls {
		return 443
	}

	return 80
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
			body, err := drainBody(&req.Body)
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
