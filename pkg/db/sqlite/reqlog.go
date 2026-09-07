package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"

	"github.com/oklog/ulid"

	"github.com/Vibe-Coding-Base/Hettix/pkg/reqlog"
	"github.com/Vibe-Coding-Base/Hettix/pkg/scope"
)

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

	rows, err := d.db.QueryContext(ctx, requestLogSelect+` WHERE r.project_id = ? ORDER BY r.id DESC`,
		filter.ProjectID.String())
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

	return reqLogs, nil
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
