package intruder_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"sync"
	"testing"

	"github.com/Vibe-Coding-Base/Hettix/pkg/intruder"
)

func TestRunSubstitutesPayloads(t *testing.T) {
	var (
		mu   sync.Mutex
		seen []string
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		seen = append(seen, r.URL.Query().Get("q"))
		mu.Unlock()

		if r.URL.Query().Get("q") == "boom" {
			w.WriteHeader(http.StatusInternalServerError)
		}
		fmt.Fprint(w, "response body")
	}))
	defer srv.Close()

	runner := intruder.NewRunner(srv.Client(), 4)

	tmpl := intruder.Request{
		Method: http.MethodGet,
		URL:    srv.URL + "/search?q=" + intruder.Marker,
		Header: http.Header{"X-Fuzz": {intruder.Marker}},
	}
	payloads := []string{"alpha", "boom", "gamma"}

	results := runner.Run(context.Background(), tmpl, payloads, nil)

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	// Results are ordered by payload.
	for i, want := range payloads {
		if results[i].Payload != want {
			t.Errorf("result %d: payload %q, want %q", i, results[i].Payload, want)
		}
	}

	if results[1].StatusCode != http.StatusInternalServerError {
		t.Errorf("expected payload %q to yield 500, got %d", "boom", results[1].StatusCode)
	}
	for _, r := range []intruder.Result{results[0], results[2]} {
		if r.StatusCode != http.StatusOK {
			t.Errorf("payload %q: expected 200, got %d", r.Payload, r.StatusCode)
		}
		if r.Length != len("response body") {
			t.Errorf("payload %q: expected length %d, got %d", r.Payload, len("response body"), r.Length)
		}
	}

	mu.Lock()
	sort.Strings(seen)
	mu.Unlock()
	if len(seen) != 3 || seen[0] != "alpha" || seen[1] != "boom" || seen[2] != "gamma" {
		t.Errorf("server did not receive all substituted payloads: %v", seen)
	}
}

func TestRunReportsSendErrors(t *testing.T) {
	runner := intruder.NewRunner(&http.Client{}, 2)

	tmpl := intruder.Request{Method: http.MethodGet, URL: "http://127.0.0.1:0/" + intruder.Marker}
	results := runner.Run(context.Background(), tmpl, []string{"x"}, nil)

	if len(results) != 1 || results[0].Error == "" {
		t.Fatalf("expected a send error, got %+v", results)
	}
}

func TestRunOnResultCallback(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "ok")
	}))
	defer srv.Close()

	runner := intruder.NewRunner(srv.Client(), 3)

	var (
		mu    sync.Mutex
		count int
	)

	tmpl := intruder.Request{Method: http.MethodGet, URL: srv.URL + "/" + intruder.Marker}
	runner.Run(context.Background(), tmpl, []string{"a", "b", "c"}, func(intruder.Result) {
		mu.Lock()
		count++
		mu.Unlock()
	})

	if count != 3 {
		t.Fatalf("expected 3 callback invocations, got %d", count)
	}
}

func TestHasMarker(t *testing.T) {
	if (intruder.Request{URL: "https://example.com/"}).HasMarker() {
		t.Error("expected no marker")
	}
	if !(intruder.Request{URL: "https://example.com/" + intruder.Marker}).HasMarker() {
		t.Error("expected a marker in the URL")
	}
	if !(intruder.Request{Header: http.Header{"X": {intruder.Marker}}}).HasMarker() {
		t.Error("expected a marker in a header")
	}
}
