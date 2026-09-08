package sender

import (
	"path"
	"strconv"

	"github.com/oklog/ulid"

	"github.com/Vibe-Coding-Base/Hettix/pkg/httpql"
	"github.com/Vibe-Coding-Base/Hettix/pkg/scope"
)

// Matches reports whether the sender request satisfies the HTTPQL expression.
func (req Request) Matches(expr httpql.Expression) (bool, error) {
	return httpql.Eval(expr, req.toRecord())
}

func (req Request) toRecord() httpql.Record {
	reqData := httpql.RequestData{
		ID:        req.ID.String(),
		Method:    req.Method,
		Proto:     req.Proto,
		Len:       len(req.Body),
		CreatedAt: ulid.Time(req.ID.Time()),
		Header:    req.Header,
		Body:      string(req.Body),
	}

	if req.URL != nil {
		reqData.URL = req.URL.String()
		reqData.Host = req.URL.Hostname()
		reqData.Path = req.URL.Path
		reqData.Query = req.URL.RawQuery
		reqData.Ext = path.Ext(req.URL.Path)
		reqData.TLS = req.URL.Scheme == "https"
		reqData.Port = urlPort(req.URL.Port(), reqData.TLS)
	}

	rec := httpql.Record{Request: reqData}

	if req.Response != nil {
		rec.Response = &httpql.ResponseData{
			Proto:     req.Response.Proto,
			Reason:    req.Response.Status,
			Code:      req.Response.StatusCode,
			Len:       len(req.Response.Body),
			RoundTrip: int(req.Response.RoundTripMillis),
			Header:    req.Response.Header,
			Body:      string(req.Response.Body),
		}
	}

	return rec
}

func urlPort(port string, tls bool) int {
	if port != "" {
		if n, err := strconv.Atoi(port); err == nil {
			return n
		}
	}

	if tls {
		return 443
	}

	return 80
}

// MatchScope reports whether the sender request is in scope.
func (req Request) MatchScope(s *scope.Scope) bool {
	var rawURL string
	if req.URL != nil {
		rawURL = req.URL.String()
	}

	return s.InScope(rawURL, req.Header, req.Body)
}
