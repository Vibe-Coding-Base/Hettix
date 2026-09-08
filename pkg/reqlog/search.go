package reqlog

import (
	"path"
	"strconv"

	"github.com/oklog/ulid"

	"github.com/Vibe-Coding-Base/Hettix/pkg/httpql"
	"github.com/Vibe-Coding-Base/Hettix/pkg/scope"
)

// Matches reports whether the request log satisfies the HTTPQL expression.
func (reqLog RequestLog) Matches(expr httpql.Expression) (bool, error) {
	return httpql.Eval(expr, reqLog.toRecord())
}

func (reqLog RequestLog) toRecord() httpql.Record {
	req := httpql.RequestData{
		ID:        reqLog.ID.String(),
		Method:    reqLog.Method,
		Proto:     reqLog.Proto,
		Len:       len(reqLog.Body),
		CreatedAt: ulid.Time(reqLog.ID.Time()),
		Header:    reqLog.Header,
		Body:      string(reqLog.Body),
	}

	if reqLog.URL != nil {
		req.URL = reqLog.URL.String()
		req.Host = reqLog.URL.Hostname()
		req.Path = reqLog.URL.Path
		req.Query = reqLog.URL.RawQuery
		req.Ext = path.Ext(reqLog.URL.Path)
		req.TLS = reqLog.URL.Scheme == "https"
		req.Port = urlPort(reqLog.URL.Port(), req.TLS)
	}

	rec := httpql.Record{Request: req}

	if reqLog.Response != nil {
		rec.Response = &httpql.ResponseData{
			Proto:     reqLog.Response.Proto,
			Reason:    reqLog.Response.Status,
			Code:      reqLog.Response.StatusCode,
			Len:       len(reqLog.Response.Body),
			RoundTrip: int(reqLog.Response.RoundTripMillis),
			Header:    reqLog.Response.Header,
			Body:      string(reqLog.Response.Body),
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

// MatchScope reports whether the request log is in scope.
func (reqLog RequestLog) MatchScope(s *scope.Scope) bool {
	var rawURL string
	if reqLog.URL != nil {
		rawURL = reqLog.URL.String()
	}

	return s.InScope(rawURL, reqLog.Header, reqLog.Body)
}
