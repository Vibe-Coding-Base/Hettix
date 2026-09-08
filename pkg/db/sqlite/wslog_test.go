package sqlite_test

import (
	"context"
	"testing"
	"time"

	"github.com/Vibe-Coding-Base/Hettix/pkg/wslog"
	"github.com/Vibe-Coding-Base/Hettix/pkg/wsproxy"
)

func TestWebSocketRoundTrip(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	projectID := newULID(t)
	connID := newULID(t)

	conn := wslog.Connection{
		ID:        connID,
		ProjectID: projectID,
		URL:       "wss://chat.example.com/ws",
		Host:      "chat.example.com",
		Path:      "/ws",
		CreatedAt: time.UnixMilli(time.Now().UnixMilli()),
	}
	if err := db.StoreWebSocketConnection(ctx, conn); err != nil {
		t.Fatalf("store connection: %v", err)
	}

	messages := []wslog.Message{
		{ConnectionID: connID, Direction: wsproxy.ClientToServer, Opcode: 1, Payload: []byte("hello"), CreatedAt: time.UnixMilli(1)},
		{ConnectionID: connID, Direction: wsproxy.ServerToClient, Opcode: 1, Payload: []byte("hi"), CreatedAt: time.UnixMilli(2)},
	}
	for _, m := range messages {
		if err := db.StoreWebSocketMessage(ctx, m); err != nil {
			t.Fatalf("store message: %v", err)
		}
	}

	conns, err := db.FindWebSocketConnections(ctx, projectID)
	if err != nil {
		t.Fatalf("find connections: %v", err)
	}
	if len(conns) != 1 {
		t.Fatalf("expected 1 connection, got %d", len(conns))
	}
	if conns[0].MessageCount != 2 {
		t.Fatalf("expected message count 2, got %d", conns[0].MessageCount)
	}
	if conns[0].ClosedAt != nil {
		t.Fatalf("expected connection to be open, got closedAt %v", conns[0].ClosedAt)
	}

	got, err := db.FindWebSocketMessages(ctx, projectID, connID)
	if err != nil {
		t.Fatalf("find messages: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(got))
	}
	// Insertion order is preserved.
	if string(got[0].Payload) != "hello" || got[0].Direction != wsproxy.ClientToServer {
		t.Fatalf("unexpected first message: %+v", got[0])
	}
	if string(got[1].Payload) != "hi" || got[1].Direction != wsproxy.ServerToClient {
		t.Fatalf("unexpected second message: %+v", got[1])
	}

	closedAt := time.UnixMilli(time.Now().UnixMilli())
	if err := db.CloseWebSocketConnection(ctx, projectID, connID, closedAt); err != nil {
		t.Fatalf("close connection: %v", err)
	}
	reloaded, err := db.FindWebSocketConnectionByID(ctx, projectID, connID)
	if err != nil {
		t.Fatalf("find by id: %v", err)
	}
	if reloaded.ClosedAt == nil || !reloaded.ClosedAt.Equal(closedAt) {
		t.Fatalf("expected closedAt %v, got %v", closedAt, reloaded.ClosedAt)
	}
}

func TestWebSocketProjectIsolation(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	projectA := newULID(t)
	projectB := newULID(t)
	connID := newULID(t)

	if err := db.StoreWebSocketConnection(ctx, wslog.Connection{
		ID: connID, ProjectID: projectA, URL: "ws://a/ws", Host: "a", Path: "/ws",
		CreatedAt: time.UnixMilli(1),
	}); err != nil {
		t.Fatalf("store: %v", err)
	}
	if err := db.StoreWebSocketMessage(ctx, wslog.Message{
		ConnectionID: connID, Direction: wsproxy.ClientToServer, Opcode: 1, Payload: []byte("x"), CreatedAt: time.UnixMilli(1),
	}); err != nil {
		t.Fatalf("store message: %v", err)
	}

	// Project B must not see project A's connection or its messages.
	if conns, err := db.FindWebSocketConnections(ctx, projectB); err != nil || len(conns) != 0 {
		t.Fatalf("expected no connections for other project, got %d (err %v)", len(conns), err)
	}
	if _, err := db.FindWebSocketConnectionByID(ctx, projectB, connID); err != wslog.ErrConnectionNotFound {
		t.Fatalf("expected ErrConnectionNotFound for other project, got %v", err)
	}
	if _, err := db.FindWebSocketMessages(ctx, projectB, connID); err != wslog.ErrConnectionNotFound {
		t.Fatalf("expected ErrConnectionNotFound querying other project's messages, got %v", err)
	}
}
