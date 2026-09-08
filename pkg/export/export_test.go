package export_test

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/oklog/ulid"

	"github.com/Vibe-Coding-Base/Hettix/pkg/export"
	"github.com/Vibe-Coding-Base/Hettix/pkg/reqlog"
)

var entropy = ulid.Monotonic(rand.New(rand.NewSource(1)), 0)

func sampleLogs() []reqlog.RequestLog {
	u, _ := url.Parse("https://api.example.com/v1/login?next=/home")
	return []reqlog.RequestLog{{
		ID:     ulid.MustNew(ulid.Timestamp(time.Now()), entropy),
		Method: "POST", URL: u, Proto: "HTTP/1.1",
		Header: http.Header{"Content-Type": {"application/json"}},
		Body:   []byte(`{"u":"a"}`),
		Response: &reqlog.ResponseLog{
			Proto: "HTTP/1.1", StatusCode: 200, Status: "200 OK",
			Header: http.Header{"Content-Type": {"application/json"}}, Body: []byte("{}"), RoundTripMillis: 42,
		},
	}}
}

func TestWriteHAR(t *testing.T) {
	var buf bytes.Buffer
	if err := export.WriteHAR(&buf, "1.0.0", sampleLogs()); err != nil {
		t.Fatalf("WriteHAR: %v", err)
	}

	var root map[string]any
	if err := json.Unmarshal(buf.Bytes(), &root); err != nil {
		t.Fatalf("HAR is not valid JSON: %v", err)
	}

	log := root["log"].(map[string]any)
	if log["version"] != "1.2" {
		t.Fatalf("expected HAR version 1.2, got %v", log["version"])
	}
	entries := log["entries"].([]any)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	entry := entries[0].(map[string]any)
	req := entry["request"].(map[string]any)
	if req["method"] != "POST" || req["url"] != "https://api.example.com/v1/login?next=/home" {
		t.Fatalf("unexpected request: %+v", req)
	}
	res := entry["response"].(map[string]any)
	if res["status"].(float64) != 200 {
		t.Fatalf("expected status 200, got %v", res["status"])
	}
}

func TestWriteCSV(t *testing.T) {
	var buf bytes.Buffer
	if err := export.WriteCSV(&buf, sampleLogs()); err != nil {
		t.Fatalf("WriteCSV: %v", err)
	}

	records, err := csv.NewReader(strings.NewReader(buf.String())).ReadAll()
	if err != nil {
		t.Fatalf("CSV is not valid: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected header + 1 row, got %d records", len(records))
	}
	if records[0][2] != "method" || records[1][2] != "POST" {
		t.Fatalf("unexpected CSV content: %v", records)
	}
	if records[1][4] != "200" || records[1][6] != "42" {
		t.Fatalf("expected status 200 and roundtrip 42, got %v", records[1])
	}
}
