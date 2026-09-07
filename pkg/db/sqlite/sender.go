package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"

	"github.com/oklog/ulid"

	"github.com/Vibe-Coding-Base/Hettix/pkg/httpql"
	"github.com/Vibe-Coding-Base/Hettix/pkg/reqlog"
	"github.com/Vibe-Coding-Base/Hettix/pkg/scope"
	"github.com/Vibe-Coding-Base/Hettix/pkg/sender"
)

// senderRequestColumns maps HTTPQL fields to sender_requests columns for query
// push-down. Fields absent here are refined by the in-memory evaluator.
var senderRequestColumns = map[string]httpql.SQLColumn{
	"req.id":         {Expr: "id", Kind: httpql.ColumnString},
	"req.method":     {Expr: "method", Kind: httpql.ColumnString},
	"req.host":       {Expr: "host", Kind: httpql.ColumnString},
	"req.path":       {Expr: "path", Kind: httpql.ColumnString},
	"req.url":        {Expr: "url", Kind: httpql.ColumnString},
	"req.proto":      {Expr: "proto", Kind: httpql.ColumnString},
	"req.created_at": {Expr: "created_at", Kind: httpql.ColumnTime},
	"resp.code":      {Expr: "res_status_code", Kind: httpql.ColumnInt},
	"resp.proto":     {Expr: "res_proto", Kind: httpql.ColumnString},
	"resp.reason":    {Expr: "res_status_reason", Kind: httpql.ColumnString},
	"resp.roundtrip": {Expr: "res_roundtrip_ms", Kind: httpql.ColumnInt},
}

func (d *Database) StoreSenderRequest(ctx context.Context, req sender.Request) error {
	headers, err := marshalHeader(req.Header)
	if err != nil {
		return err
	}

	var rawURL, host, path string
	if req.URL != nil {
		rawURL = req.URL.String()
		host = req.URL.Hostname()
		path = req.URL.Path
	}

	var (
		resProto, resStatusReason, resHeaders *string
		resStatusCode, resRoundTripMillis     *int64
		resBody                               []byte
	)

	if req.Response != nil {
		resProto = &req.Response.Proto
		resStatusReason = &req.Response.Status
		code := int64(req.Response.StatusCode)
		resStatusCode = &code
		rt := req.Response.RoundTripMillis
		resRoundTripMillis = &rt
		resBody = req.Response.Body

		h, err := marshalHeader(req.Response.Header)
		if err != nil {
			return err
		}
		resHeaders = &h
	}

	_, err = d.db.ExecContext(ctx,
		`INSERT INTO sender_requests
		   (id, project_id, source_request_log_id, method, url, host, path, proto, headers, body, created_at,
		    res_proto, res_status_code, res_status_reason, res_headers, res_body, res_roundtrip_ms)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT (id) DO UPDATE SET
		   method = excluded.method, url = excluded.url, host = excluded.host, path = excluded.path,
		   proto = excluded.proto, headers = excluded.headers, body = excluded.body,
		   res_proto = excluded.res_proto, res_status_code = excluded.res_status_code,
		   res_status_reason = excluded.res_status_reason, res_headers = excluded.res_headers,
		   res_body = excluded.res_body, res_roundtrip_ms = excluded.res_roundtrip_ms`,
		req.ID.String(), req.ProjectID.String(), req.SourceRequestLogID.String(), req.Method, rawURL, host, path,
		req.Proto, headers, req.Body, int64(req.ID.Time()),
		resProto, resStatusCode, resStatusReason, resHeaders, resBody, resRoundTripMillis,
	)
	if err != nil {
		return fmt.Errorf("sqlite: failed to store sender request: %w", err)
	}

	return nil
}

func (d *Database) FindSenderRequestByID(ctx context.Context, projectID, id ulid.ULID) (sender.Request, error) {
	row := d.db.QueryRowContext(ctx, senderSelect+` WHERE id = ? AND project_id = ?`, id.String(), projectID.String())

	req, err := scanSenderRequest(row)
	if errors.Is(err, sql.ErrNoRows) {
		return sender.Request{}, sender.ErrRequestNotFound
	}
	if err != nil {
		return sender.Request{}, err
	}

	return req, nil
}

func (d *Database) FindSenderRequests(
	ctx context.Context,
	filter sender.FindRequestsFilter,
	scopeMatcher *scope.Scope,
) ([]sender.Request, error) {
	if filter.ProjectID.Compare(ulid.ULID{}) == 0 {
		return nil, sender.ErrProjectIDMustBeSet
	}

	where, whereArgs := httpql.CompileSQL(filter.SearchExpr, senderRequestColumns)
	args := append([]any{filter.ProjectID.String()}, whereArgs...)

	rows, err := d.db.QueryContext(ctx, senderSelect+` WHERE project_id = ? AND `+where+` ORDER BY id DESC`, args...)
	if err != nil {
		return nil, fmt.Errorf("sqlite: failed to query sender requests: %w", err)
	}
	defer rows.Close()

	var reqs []sender.Request

	for rows.Next() {
		req, err := scanSenderRequest(rows)
		if err != nil {
			return nil, err
		}

		if filter.OnlyInScope && !req.MatchScope(scopeMatcher) {
			continue
		}

		if filter.SearchExpr != nil {
			match, err := req.Matches(filter.SearchExpr)
			if err != nil {
				return nil, fmt.Errorf("sqlite: failed to match search expression (id: %v): %w", req.ID, err)
			}

			if !match {
				continue
			}
		}

		reqs = append(reqs, req)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("sqlite: failed to iterate sender requests: %w", err)
	}

	return reqs, nil
}

func (d *Database) DeleteSenderRequests(ctx context.Context, projectID ulid.ULID) error {
	_, err := d.db.ExecContext(ctx, `DELETE FROM sender_requests WHERE project_id = ?`, projectID.String())
	if err != nil {
		return fmt.Errorf("sqlite: failed to delete sender requests: %w", err)
	}

	return nil
}

const senderSelect = `SELECT id, project_id, source_request_log_id, method, url, proto, headers, body,
	res_proto, res_status_code, res_status_reason, res_headers, res_body, res_roundtrip_ms FROM sender_requests`

func scanSenderRequest(row scanner) (sender.Request, error) {
	var (
		idStr, projectIDStr, sourceIDStr, method, rawURL, proto, headers string
		body                                                             []byte
		resProto, resStatusReason, resHeaders                            sql.NullString
		resStatusCode, resRoundTripMillis                                sql.NullInt64
		resBody                                                          []byte
	)

	if err := row.Scan(&idStr, &projectIDStr, &sourceIDStr, &method, &rawURL, &proto, &headers, &body,
		&resProto, &resStatusCode, &resStatusReason, &resHeaders, &resBody, &resRoundTripMillis); err != nil {
		return sender.Request{}, err
	}

	id, err := ulid.Parse(idStr)
	if err != nil {
		return sender.Request{}, fmt.Errorf("sqlite: invalid sender request id %q: %w", idStr, err)
	}

	projectID, err := ulid.Parse(projectIDStr)
	if err != nil {
		return sender.Request{}, fmt.Errorf("sqlite: invalid project id %q: %w", projectIDStr, err)
	}

	var sourceID ulid.ULID
	if sourceIDStr != "" {
		if sourceID, err = ulid.Parse(sourceIDStr); err != nil {
			return sender.Request{}, fmt.Errorf("sqlite: invalid source request log id %q: %w", sourceIDStr, err)
		}
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return sender.Request{}, fmt.Errorf("sqlite: invalid sender request url %q: %w", rawURL, err)
	}

	header, err := unmarshalHeader(headers)
	if err != nil {
		return sender.Request{}, err
	}

	req := sender.Request{
		ID:                 id,
		ProjectID:          projectID,
		SourceRequestLogID: sourceID,
		Method:             method,
		URL:                parsedURL,
		Proto:              proto,
		Header:             header,
		Body:               body,
	}

	if resProto.Valid {
		resHeader, err := unmarshalHeader(resHeaders.String)
		if err != nil {
			return sender.Request{}, err
		}

		req.Response = &reqlog.ResponseLog{
			Proto:           resProto.String,
			StatusCode:      int(resStatusCode.Int64),
			Status:          resStatusReason.String,
			Header:          resHeader,
			Body:            resBody,
			RoundTripMillis: resRoundTripMillis.Int64,
		}
	}

	return req, nil
}
