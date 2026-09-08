// Package export serializes captured request logs to interchange formats: HAR
// (HTTP Archive) for import into other tools, and CSV for spreadsheets.
package export

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/oklog/ulid"

	"github.com/Vibe-Coding-Base/Hettix/pkg/reqlog"
)

// WriteHAR writes the request logs as an HAR 1.2 document.
func WriteHAR(w io.Writer, creatorVersion string, logs []reqlog.RequestLog) error {
	har := harLog{}
	har.Log.Version = "1.2"
	har.Log.Creator = harCreator{Name: "Hettix", Version: creatorVersion}
	har.Log.Entries = make([]harEntry, 0, len(logs))

	for _, rl := range logs {
		har.Log.Entries = append(har.Log.Entries, harEntryFromLog(rl))
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")

	if err := enc.Encode(har); err != nil {
		return fmt.Errorf("export: failed to encode HAR: %w", err)
	}

	return nil
}

// WriteCSV writes one row per request log with its key fields.
func WriteCSV(w io.Writer, logs []reqlog.RequestLog) error {
	cw := csv.NewWriter(w)

	header := []string{"id", "timestamp", "method", "url", "status_code", "response_length", "roundtrip_ms"}
	if err := cw.Write(header); err != nil {
		return fmt.Errorf("export: failed to write CSV header: %w", err)
	}

	for _, rl := range logs {
		var rawURL string
		if rl.URL != nil {
			rawURL = rl.URL.String()
		}

		statusCode, respLen, roundTrip := "", "", ""
		if rl.Response != nil {
			statusCode = strconv.Itoa(rl.Response.StatusCode)
			respLen = strconv.Itoa(len(rl.Response.Body))
			roundTrip = strconv.FormatInt(rl.Response.RoundTripMillis, 10)
		}

		row := []string{
			rl.ID.String(),
			ulid.Time(rl.ID.Time()).UTC().Format(time.RFC3339),
			rl.Method,
			rawURL,
			statusCode,
			respLen,
			roundTrip,
		}

		if err := cw.Write(row); err != nil {
			return fmt.Errorf("export: failed to write CSV row: %w", err)
		}
	}

	cw.Flush()

	if err := cw.Error(); err != nil {
		return fmt.Errorf("export: failed to flush CSV: %w", err)
	}

	return nil
}

type harLog struct {
	Log struct {
		Version string     `json:"version"`
		Creator harCreator `json:"creator"`
		Entries []harEntry `json:"entries"`
	} `json:"log"`
}

type harCreator struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type harEntry struct {
	StartedDateTime string      `json:"startedDateTime"`
	Time            int64       `json:"time"`
	Request         harRequest  `json:"request"`
	Response        harResponse `json:"response"`
	Cache           struct{}    `json:"cache"`
	Timings         harTimings  `json:"timings"`
}

type harTimings struct {
	Send    int   `json:"send"`
	Wait    int64 `json:"wait"`
	Receive int   `json:"receive"`
}

type harRequest struct {
	Method      string      `json:"method"`
	URL         string      `json:"url"`
	HTTPVersion string      `json:"httpVersion"`
	Headers     []harHeader `json:"headers"`
	QueryString []harQuery  `json:"queryString"`
	PostData    *harPost    `json:"postData,omitempty"`
	HeadersSize int         `json:"headersSize"`
	BodySize    int         `json:"bodySize"`
}

type harResponse struct {
	Status      int         `json:"status"`
	StatusText  string      `json:"statusText"`
	HTTPVersion string      `json:"httpVersion"`
	Headers     []harHeader `json:"headers"`
	Content     harContent  `json:"content"`
	RedirectURL string      `json:"redirectURL"`
	HeadersSize int         `json:"headersSize"`
	BodySize    int         `json:"bodySize"`
}

type harHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type harQuery struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type harPost struct {
	MimeType string `json:"mimeType"`
	Text     string `json:"text"`
}

type harContent struct {
	Size     int    `json:"size"`
	MimeType string `json:"mimeType"`
	Text     string `json:"text,omitempty"`
}

func harEntryFromLog(rl reqlog.RequestLog) harEntry {
	entry := harEntry{
		StartedDateTime: ulid.Time(rl.ID.Time()).UTC().Format(time.RFC3339Nano),
		Request: harRequest{
			Method:      rl.Method,
			HTTPVersion: rl.Proto,
			Headers:     harHeaders(rl.Header),
			QueryString: []harQuery{},
			HeadersSize: -1,
			BodySize:    len(rl.Body),
		},
	}

	if rl.URL != nil {
		entry.Request.URL = rl.URL.String()

		query := []harQuery{}
		for name, values := range rl.URL.Query() {
			for _, v := range values {
				query = append(query, harQuery{Name: name, Value: v})
			}
		}
		entry.Request.QueryString = query
	}

	if len(rl.Body) > 0 {
		entry.Request.PostData = &harPost{
			MimeType: rl.Header.Get("Content-Type"),
			Text:     string(rl.Body),
		}
	}

	if rl.Response != nil {
		entry.Time = rl.Response.RoundTripMillis
		entry.Timings = harTimings{Send: 0, Wait: rl.Response.RoundTripMillis, Receive: 0}
		entry.Response = harResponse{
			Status:      rl.Response.StatusCode,
			StatusText:  rl.Response.Status,
			HTTPVersion: rl.Response.Proto,
			Headers:     harHeaders(rl.Response.Header),
			Content: harContent{
				Size:     len(rl.Response.Body),
				MimeType: rl.Response.Header.Get("Content-Type"),
				Text:     string(rl.Response.Body),
			},
			HeadersSize: -1,
			BodySize:    len(rl.Response.Body),
		}
	} else {
		entry.Response = harResponse{Headers: []harHeader{}, Content: harContent{}}
	}

	return entry
}

func harHeaders(h map[string][]string) []harHeader {
	headers := []harHeader{}
	for name, values := range h {
		for _, v := range values {
			headers = append(headers, harHeader{Name: name, Value: v})
		}
	}

	return headers
}
