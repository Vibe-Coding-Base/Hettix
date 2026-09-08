// Package intruder runs a base HTTP request repeatedly, substituting a marker
// in the request with each payload in a set, and collects the responses. It is
// the engine behind the fuzzer/intruder feature: given a request template and a
// list of payloads it sends the rendered requests concurrently and reports each
// result.
package intruder

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Marker delimits the region of the request template that each payload replaces.
const Marker = "§"

// maxBodyBytes bounds how much of each response body is read to measure its
// length, so a large response can't exhaust memory during an attack.
const maxBodyBytes = 8 << 20

// Request is a base request template. Any of URL, header values or Body may
// contain Marker occurrences, which are replaced with the payload.
type Request struct {
	Method string
	URL    string
	Proto  string
	Header http.Header
	Body   string
}

// Result is the outcome of a single rendered request.
type Result struct {
	Index      int
	Payload    string
	StatusCode int
	Length     int
	DurationMs int64
	Error      string
}

// Runner sends rendered requests with bounded concurrency.
type Runner struct {
	client      *http.Client
	concurrency int
}

// NewRunner returns a Runner. A concurrency of zero or less defaults to 10.
func NewRunner(client *http.Client, concurrency int) *Runner {
	if concurrency <= 0 {
		concurrency = 10
	}

	return &Runner{client: client, concurrency: concurrency}
}

// Run sends one request per payload and returns the results in payload order.
// onResult, if non-nil, is called for each result as it completes (from
// multiple goroutines, so it must be safe for concurrent use). Run stops early
// if ctx is cancelled; results for unsent payloads carry a cancellation error.
func (r *Runner) Run(ctx context.Context, tmpl Request, payloads []string, onResult func(Result)) []Result {
	results := make([]Result, len(payloads))
	sem := make(chan struct{}, r.concurrency)
	done := make(chan int, len(payloads))

	for i, payload := range payloads {
		select {
		case <-ctx.Done():
			results[i] = Result{Index: i, Payload: payload, Error: ctx.Err().Error()}
			done <- i

			continue
		case sem <- struct{}{}:
		}

		go func(i int, payload string) {
			defer func() { <-sem }()

			res := r.send(ctx, tmpl, i, payload)
			results[i] = res

			if onResult != nil {
				onResult(res)
			}

			done <- i
		}(i, payload)
	}

	for range payloads {
		<-done
	}

	return results
}

func (r *Runner) send(ctx context.Context, tmpl Request, index int, payload string) Result {
	result := Result{Index: index, Payload: payload}

	httpReq, err := buildRequest(ctx, render(tmpl, payload))
	if err != nil {
		result.Error = err.Error()
		return result
	}

	start := time.Now()

	res, err := r.client.Do(httpReq)
	result.DurationMs = time.Since(start).Milliseconds()

	if err != nil {
		result.Error = err.Error()
		return result
	}
	defer res.Body.Close()

	body, err := io.ReadAll(io.LimitReader(res.Body, maxBodyBytes))
	if err != nil {
		result.Error = err.Error()
		return result
	}

	result.StatusCode = res.StatusCode
	result.Length = len(body)

	return result
}

// render substitutes every Marker occurrence in the template with the payload.
func render(tmpl Request, payload string) Request {
	out := Request{
		Method: tmpl.Method,
		URL:    strings.ReplaceAll(tmpl.URL, Marker, payload),
		Proto:  tmpl.Proto,
		Body:   strings.ReplaceAll(tmpl.Body, Marker, payload),
	}

	if tmpl.Header != nil {
		out.Header = make(http.Header, len(tmpl.Header))
		for key, values := range tmpl.Header {
			replaced := make([]string, len(values))
			for i, v := range values {
				replaced[i] = strings.ReplaceAll(v, Marker, payload)
			}

			out.Header[key] = replaced
		}
	}

	return out
}

func buildRequest(ctx context.Context, req Request) (*http.Request, error) {
	method := req.Method
	if method == "" {
		method = http.MethodGet
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, req.URL, strings.NewReader(req.Body))
	if err != nil {
		return nil, fmt.Errorf("intruder: failed to build request: %w", err)
	}

	if req.Header != nil {
		httpReq.Header = req.Header.Clone()
	}

	return httpReq, nil
}

// HasMarker reports whether the template contains at least one insertion marker.
func (req Request) HasMarker() bool {
	if strings.Contains(req.URL, Marker) || strings.Contains(req.Body, Marker) {
		return true
	}

	for _, values := range req.Header {
		for _, v := range values {
			if strings.Contains(v, Marker) {
				return true
			}
		}
	}

	return false
}
