package httpql

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
	"time"
)

// RecordFromRequest builds an evaluation Record from a live HTTP request. The
// body is read and replaced with a fresh reader so it can be read again
// downstream.
func RecordFromRequest(req *http.Request) (Record, error) {
	body, err := DrainBody(&req.Body)
	if err != nil {
		return Record{}, err
	}

	reqData := RequestData{
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
		reqData.Port = portFor(req.URL.Port(), reqData.TLS)
	}

	return Record{Request: reqData}, nil
}

// RecordFromResponse builds an evaluation Record from a live HTTP response,
// including its originating request when available so req.* clauses remain
// usable in response contexts.
func RecordFromResponse(res *http.Response) (Record, error) {
	rec := Record{}

	if res.Request != nil {
		reqRec, err := RecordFromRequest(res.Request)
		if err != nil {
			return Record{}, err
		}
		rec.Request = reqRec.Request
	}

	body, err := DrainBody(&res.Body)
	if err != nil {
		return Record{}, err
	}

	reason := res.Status
	if _, after, ok := splitStatus(res.Status); ok {
		reason = after
	}

	rec.Response = &ResponseData{
		Proto:  res.Proto,
		Reason: reason,
		Code:   res.StatusCode,
		Header: res.Header,
		Body:   string(body),
		Len:    len(body),
	}

	return rec, nil
}

// DrainBody reads a body fully and replaces it with a fresh reader so it can be
// read again downstream. A nil body yields no bytes.
func DrainBody(body *io.ReadCloser) ([]byte, error) {
	if *body == nil {
		return nil, nil
	}

	buf, err := io.ReadAll(*body)
	if err != nil {
		return nil, fmt.Errorf("httpql: failed to read body: %w", err)
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

func portFor(p string, tls bool) int {
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
