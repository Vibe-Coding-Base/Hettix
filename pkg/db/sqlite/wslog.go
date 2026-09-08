package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/oklog/ulid"

	"github.com/Vibe-Coding-Base/Hettix/pkg/wslog"
	"github.com/Vibe-Coding-Base/Hettix/pkg/wsproxy"
)

func (d *Database) StoreWebSocketConnection(ctx context.Context, conn wslog.Connection) error {
	_, err := d.db.ExecContext(ctx,
		`INSERT INTO websocket_connections (id, project_id, url, host, path, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		conn.ID.String(), conn.ProjectID.String(), conn.URL, conn.Host, conn.Path, conn.CreatedAt.UnixMilli(),
	)
	if err != nil {
		return fmt.Errorf("sqlite: failed to store websocket connection: %w", err)
	}

	return nil
}

func (d *Database) CloseWebSocketConnection(ctx context.Context, projectID, id ulid.ULID, at time.Time) error {
	_, err := d.db.ExecContext(ctx,
		`UPDATE websocket_connections SET closed_at = ? WHERE id = ? AND project_id = ?`,
		at.UnixMilli(), id.String(), projectID.String(),
	)
	if err != nil {
		return fmt.Errorf("sqlite: failed to close websocket connection: %w", err)
	}

	return nil
}

func (d *Database) StoreWebSocketMessage(ctx context.Context, msg wslog.Message) error {
	_, err := d.db.ExecContext(ctx,
		`INSERT INTO websocket_messages (connection_id, direction, opcode, payload, created_at)
		 VALUES (?, ?, ?, ?, ?)`,
		msg.ConnectionID.String(), int(msg.Direction), msg.Opcode, msg.Payload, msg.CreatedAt.UnixMilli(),
	)
	if err != nil {
		return fmt.Errorf("sqlite: failed to store websocket message: %w", err)
	}

	return nil
}

const websocketConnectionSelect = `SELECT c.id, c.project_id, c.url, c.host, c.path, c.created_at, c.closed_at,
	(SELECT COUNT(*) FROM websocket_messages m WHERE m.connection_id = c.id) AS message_count
	FROM websocket_connections c`

func (d *Database) FindWebSocketConnections(ctx context.Context, projectID ulid.ULID) ([]wslog.Connection, error) {
	rows, err := d.db.QueryContext(ctx,
		websocketConnectionSelect+` WHERE c.project_id = ? ORDER BY c.id DESC`, projectID.String())
	if err != nil {
		return nil, fmt.Errorf("sqlite: failed to query websocket connections: %w", err)
	}
	defer rows.Close()

	var conns []wslog.Connection

	for rows.Next() {
		conn, err := scanWebSocketConnection(rows)
		if err != nil {
			return nil, err
		}

		conns = append(conns, conn)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("sqlite: failed to iterate websocket connections: %w", err)
	}

	return conns, nil
}

func (d *Database) FindWebSocketConnectionByID(
	ctx context.Context,
	projectID, id ulid.ULID,
) (wslog.Connection, error) {
	row := d.db.QueryRowContext(ctx,
		websocketConnectionSelect+` WHERE c.id = ? AND c.project_id = ?`, id.String(), projectID.String())

	conn, err := scanWebSocketConnection(row)
	if errors.Is(err, sql.ErrNoRows) {
		return wslog.Connection{}, wslog.ErrConnectionNotFound
	}
	if err != nil {
		return wslog.Connection{}, err
	}

	return conn, nil
}

func (d *Database) FindWebSocketMessages(
	ctx context.Context,
	projectID, connectionID ulid.ULID,
) ([]wslog.Message, error) {
	// The project scope is enforced through the parent connection.
	if _, err := d.FindWebSocketConnectionByID(ctx, projectID, connectionID); err != nil {
		return nil, err
	}

	rows, err := d.db.QueryContext(ctx,
		`SELECT connection_id, direction, opcode, payload, created_at
		 FROM websocket_messages WHERE connection_id = ? ORDER BY id ASC`, connectionID.String())
	if err != nil {
		return nil, fmt.Errorf("sqlite: failed to query websocket messages: %w", err)
	}
	defer rows.Close()

	var msgs []wslog.Message

	for rows.Next() {
		var (
			connIDStr string
			direction int
			opcode    int
			payload   []byte
			createdAt int64
		)

		if err := rows.Scan(&connIDStr, &direction, &opcode, &payload, &createdAt); err != nil {
			return nil, fmt.Errorf("sqlite: failed to scan websocket message: %w", err)
		}

		connID, err := ulid.Parse(connIDStr)
		if err != nil {
			return nil, fmt.Errorf("sqlite: invalid websocket connection id %q: %w", connIDStr, err)
		}

		msgs = append(msgs, wslog.Message{
			ConnectionID: connID,
			Direction:    wsproxy.Direction(direction),
			Opcode:       opcode,
			Payload:      payload,
			CreatedAt:    time.UnixMilli(createdAt),
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("sqlite: failed to iterate websocket messages: %w", err)
	}

	return msgs, nil
}

func (d *Database) ClearWebSocketConnections(ctx context.Context, projectID ulid.ULID) error {
	_, err := d.db.ExecContext(ctx,
		`DELETE FROM websocket_connections WHERE project_id = ?`, projectID.String())
	if err != nil {
		return fmt.Errorf("sqlite: failed to clear websocket connections: %w", err)
	}

	return nil
}

func scanWebSocketConnection(row scanner) (wslog.Connection, error) {
	var (
		idStr, projectIDStr, rawURL, host, path string
		createdAt                               int64
		closedAt                                sql.NullInt64
		messageCount                            int
	)

	if err := row.Scan(&idStr, &projectIDStr, &rawURL, &host, &path,
		&createdAt, &closedAt, &messageCount); err != nil {
		return wslog.Connection{}, err
	}

	id, err := ulid.Parse(idStr)
	if err != nil {
		return wslog.Connection{}, fmt.Errorf("sqlite: invalid websocket connection id %q: %w", idStr, err)
	}

	projectID, err := ulid.Parse(projectIDStr)
	if err != nil {
		return wslog.Connection{}, fmt.Errorf("sqlite: invalid project id %q: %w", projectIDStr, err)
	}

	conn := wslog.Connection{
		ID:           id,
		ProjectID:    projectID,
		URL:          rawURL,
		Host:         host,
		Path:         path,
		CreatedAt:    time.UnixMilli(createdAt),
		MessageCount: messageCount,
	}

	if closedAt.Valid {
		t := time.UnixMilli(closedAt.Int64)
		conn.ClosedAt = &t
	}

	return conn, nil
}
