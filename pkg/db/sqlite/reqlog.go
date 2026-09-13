package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"sort"

	"github.com/oklog/ulid"

	"github.com/Vibe-Coding-Base/Hettix/pkg/httpql"
	"github.com/Vibe-Coding-Base/Hettix/pkg/reqlog"
	"github.com/Vibe-Coding-Base/Hettix/pkg/scope"
)

// requestLogColumns maps HTTPQL fields to the columns of the request/response
// join used by FindRequestLogs, enabling query push-down for exact-comparison
// clauses. Fields absent here are refined by the in-memory evaluator.
var requestLogColumns = map[string]httpql.SQLColumn{
	"req.id":         {Expr: "r.id", Kind: httpql.ColumnString},
	"req.method":     {Expr: "r.method", Kind: httpql.ColumnString},
	"req.host":       {Expr: "r.host", Kind: httpql.ColumnString},
	"req.path":       {Expr: "r.path", Kind: httpql.ColumnString},
	"req.url":        {Expr: "r.url", Kind: httpql.ColumnString},
	"req.proto":      {Expr: "r.proto", Kind: httpql.ColumnString},
	"req.created_at": {Expr: "r.created_at", Kind: httpql.ColumnTime},
	"resp.code":      {Expr: "resp.status_code", Kind: httpql.ColumnInt},
	"resp.proto":     {Expr: "resp.proto", Kind: httpql.ColumnString},
	"resp.reason":    {Expr: "resp.status_reason", Kind: httpql.ColumnString},
	"resp.roundtrip": {Expr: "resp.roundtrip_ms", Kind: httpql.ColumnInt},
}

func (d *Database) StoreRequestLog(ctx context.Context, reqLog reqlog.RequestLog) error {
	headers, err := marshalHeader(reqLog.Header)
	if err != nil {
		return err
	}

	var rawURL, host, path string
	if reqLog.URL != nil {
		rawURL = reqLog.URL.String()
		host = reqLog.URL.Hostname()
		path = reqLog.URL.Path
	}

	_, err = d.db.ExecContext(ctx,
		`INSERT INTO http_request_logs (id, project_id, method, url, host, path, proto, headers, body, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		reqLog.ID.String(), reqLog.ProjectID.String(), reqLog.Method, rawURL, host, path,
		reqLog.Proto, headers, reqLog.Body, int64(reqLog.ID.Time()),
	)
	if err != nil {
		return fmt.Errorf("sqlite: failed to store request log: %w", err)
	}

	return nil
}

func (d *Database) StoreResponseLog(ctx context.Context, projectID, reqLogID ulid.ULID, resLog reqlog.ResponseLog) error {
	headers, err := marshalHeader(resLog.Header)
	if err != nil {
		return err
	}

	_, err = d.db.ExecContext(ctx,
		`INSERT INTO http_response_logs (request_log_id, proto, status_code, status_reason, headers, body, roundtrip_ms)
		 VALUES (?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT (request_log_id) DO UPDATE SET
		   proto = excluded.proto, status_code = excluded.status_code, status_reason = excluded.status_reason,
		   headers = excluded.headers, body = excluded.body, roundtrip_ms = excluded.roundtrip_ms`,
		reqLogID.String(), resLog.Proto, resLog.StatusCode, resLog.Status, headers, resLog.Body, resLog.RoundTripMillis,
	)
	if err != nil {
		return fmt.Errorf("sqlite: failed to store response log: %w", err)
	}

	return nil
}

const requestLogSelect = `SELECT r.id, r.project_id, r.method, r.url, r.proto, r.headers, r.body,
	resp.proto, resp.status_code, resp.status_reason, resp.headers, resp.body, resp.roundtrip_ms
	FROM http_request_logs r
	LEFT JOIN http_response_logs resp ON resp.request_log_id = r.id`

func (d *Database) FindRequestLogByID(ctx context.Context, projectID, id ulid.ULID) (reqlog.RequestLog, error) {
	row := d.db.QueryRowContext(ctx, requestLogSelect+` WHERE r.id = ? AND r.project_id = ?`,
		id.String(), projectID.String())

	reqLog, err := scanRequestLog(row)
	if errors.Is(err, sql.ErrNoRows) {
		return reqlog.RequestLog{}, reqlog.ErrRequestNotFound
	}
	if err != nil {
		return reqlog.RequestLog{}, err
	}

	return reqLog, nil
}

func (d *Database) FindRequestLogs(
	ctx context.Context,
	filter reqlog.FindRequestsFilter,
	scopeMatcher *scope.Scope,
) ([]reqlog.RequestLog, error) {
	if filter.ProjectID.Compare(ulid.ULID{}) == 0 {
		return nil, reqlog.ErrProjectIDMustBeSet
	}

	where, whereArgs := httpql.CompileSQL(filter.SearchExpr, requestLogColumns)
	args := append([]any{filter.ProjectID.String()}, whereArgs...)

	rows, err := d.db.QueryContext(ctx,
		requestLogSelect+` WHERE r.project_id = ? AND `+where+` ORDER BY r.id DESC`, args...)
	if err != nil {
		return nil, fmt.Errorf("sqlite: failed to query request logs: %w", err)
	}
	defer rows.Close()

	var reqLogs []reqlog.RequestLog

	for rows.Next() {
		reqLog, err := scanRequestLog(rows)
		if err != nil {
			return nil, err
		}

		if filter.OnlyInScope && !reqLog.MatchScope(scopeMatcher) {
			continue
		}

		if filter.SearchExpr != nil {
			match, err := reqLog.Matches(filter.SearchExpr)
			if err != nil {
				return nil, fmt.Errorf("sqlite: failed to match search expression (id: %v): %w", reqLog.ID, err)
			}

			if !match {
				continue
			}
		}

		reqLogs = append(reqLogs, reqLog)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("sqlite: failed to iterate request logs: %w", err)
	}

	return paginate(reqLogs, filter.Offset, filter.Limit), nil
}

// Sitemap aggregates the project's request logs into one entry per host+path,
// collecting the distinct methods and response status codes seen and the total
// request count.
func (d *Database) Sitemap(ctx context.Context, projectID ulid.ULID) ([]reqlog.SitemapEntry, error) {
	rows, err := d.db.QueryContext(ctx,
		`SELECT r.host, r.path, r.method, resp.status_code
		 FROM http_request_logs r
		 LEFT JOIN http_response_logs resp ON resp.request_log_id = r.id
		 WHERE r.project_id = ?
		 ORDER BY r.host ASC, r.path ASC`, projectID.String())
	if err != nil {
		return nil, fmt.Errorf("sqlite: failed to query sitemap: %w", err)
	}
	defer rows.Close()

	type key struct{ host, path string }

	type agg struct {
		methods  map[string]struct{}
		statuses map[int]struct{}
		tags     map[string]struct{}
		count    int
	}

	order := make([]key, 0)
	byKey := make(map[key]*agg)

	ensure := func(k key) *agg {
		a, ok := byKey[k]
		if !ok {
			a = &agg{methods: map[string]struct{}{}, statuses: map[int]struct{}{}, tags: map[string]struct{}{}}
			byKey[k] = a
			order = append(order, k)
		}
		return a
	}

	for rows.Next() {
		var (
			host, path, method string
			statusCode         sql.NullInt64
		)

		if err := rows.Scan(&host, &path, &method, &statusCode); err != nil {
			return nil, fmt.Errorf("sqlite: failed to scan sitemap row: %w", err)
		}

		a := ensure(key{host: host, path: path})
		a.count++
		a.methods[method] = struct{}{}
		if statusCode.Valid {
			a.statuses[int(statusCode.Int64)] = struct{}{}
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("sqlite: failed to iterate sitemap rows: %w", err)
	}

	// Merge endpoints discovered by plugins, tagging them with their source.
	discRows, err := d.db.QueryContext(ctx,
		`SELECT host, path, source FROM discovered_endpoints WHERE project_id = ?`, projectID.String())
	if err != nil {
		return nil, fmt.Errorf("sqlite: failed to query discovered endpoints: %w", err)
	}
	defer discRows.Close()

	for discRows.Next() {
		var host, path, source string
		if err := discRows.Scan(&host, &path, &source); err != nil {
			return nil, fmt.Errorf("sqlite: failed to scan discovered endpoint: %w", err)
		}
		ensure(key{host: host, path: path}).tags[source] = struct{}{}
	}
	if err := discRows.Err(); err != nil {
		return nil, fmt.Errorf("sqlite: failed to iterate discovered endpoints: %w", err)
	}

	sort.Slice(order, func(i, j int) bool {
		if order[i].host != order[j].host {
			return order[i].host < order[j].host
		}
		return order[i].path < order[j].path
	})

	entries := make([]reqlog.SitemapEntry, 0, len(order))
	for _, k := range order {
		a := byKey[k]

		methods := keysSorted(a.methods)
		statuses := intKeysSorted(a.statuses)
		tags := keysSorted(a.tags)

		entries = append(entries, reqlog.SitemapEntry{
			Host:        k.host,
			Path:        k.path,
			Methods:     methods,
			StatusCodes: statuses,
			Count:       a.count,
			Tags:        tags,
		})
	}

	return entries, nil
}

func keysSorted(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func intKeysSorted(m map[int]struct{}) []int {
	out := make([]int, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Ints(out)
	return out
}

func (d *Database) ClearRequestLogs(ctx context.Context, projectID ulid.ULID) error {
	_, err := d.db.ExecContext(ctx, `DELETE FROM http_request_logs WHERE project_id = ?`, projectID.String())
	if err != nil {
		return fmt.Errorf("sqlite: failed to clear request logs: %w", err)
	}

	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanRequestLog(row scanner) (reqlog.RequestLog, error) {
	var (
		idStr, projectIDStr, method, rawURL, proto, headers string
		body                                                []byte
		resProto, resStatusReason, resHeaders               sql.NullString
		resStatusCode, resRoundTripMillis                   sql.NullInt64
		resBody                                             []byte
	)

	if err := row.Scan(&idStr, &projectIDStr, &method, &rawURL, &proto, &headers, &body,
		&resProto, &resStatusCode, &resStatusReason, &resHeaders, &resBody, &resRoundTripMillis); err != nil {
		return reqlog.RequestLog{}, err
	}

	id, err := ulid.Parse(idStr)
	if err != nil {
		return reqlog.RequestLog{}, fmt.Errorf("sqlite: invalid request log id %q: %w", idStr, err)
	}

	projectID, err := ulid.Parse(projectIDStr)
	if err != nil {
		return reqlog.RequestLog{}, fmt.Errorf("sqlite: invalid project id %q: %w", projectIDStr, err)
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return reqlog.RequestLog{}, fmt.Errorf("sqlite: invalid request url %q: %w", rawURL, err)
	}

	header, err := unmarshalHeader(headers)
	if err != nil {
		return reqlog.RequestLog{}, err
	}

	reqLog := reqlog.RequestLog{
		ID:        id,
		ProjectID: projectID,
		Method:    method,
		URL:       parsedURL,
		Proto:     proto,
		Header:    header,
		Body:      body,
	}

	if resProto.Valid {
		resHeader, err := unmarshalHeader(resHeaders.String)
		if err != nil {
			return reqlog.RequestLog{}, err
		}

		reqLog.Response = &reqlog.ResponseLog{
			Proto:           resProto.String,
			StatusCode:      int(resStatusCode.Int64),
			Status:          resStatusReason.String,
			Header:          resHeader,
			Body:            resBody,
			RoundTripMillis: resRoundTripMillis.Int64,
		}
	}

	return reqLog, nil
}
