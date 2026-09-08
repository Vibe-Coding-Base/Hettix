package wslog_test

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/oklog/ulid"

	"github.com/Vibe-Coding-Base/Hettix/pkg/httpql"
	"github.com/Vibe-Coding-Base/Hettix/pkg/wslog"
	"github.com/Vibe-Coding-Base/Hettix/pkg/wsproxy"
)

var entropy = rand.New(rand.NewSource(1))

type fakeRepo struct {
	conns    []wslog.Connection
	messages map[string][]wslog.Message
}

func (f *fakeRepo) StoreWebSocketConnection(context.Context, wslog.Connection) error { return nil }
func (f *fakeRepo) CloseWebSocketConnection(context.Context, ulid.ULID, ulid.ULID, time.Time) error {
	return nil
}
func (f *fakeRepo) StoreWebSocketMessage(context.Context, wslog.Message) error { return nil }

func (f *fakeRepo) FindWebSocketConnections(_ context.Context, _ ulid.ULID) ([]wslog.Connection, error) {
	return f.conns, nil
}

func (f *fakeRepo) FindWebSocketConnectionByID(
	_ context.Context,
	_, id ulid.ULID,
) (wslog.Connection, error) {
	for _, c := range f.conns {
		if c.ID == id {
			return c, nil
		}
	}

	return wslog.Connection{}, wslog.ErrConnectionNotFound
}

func (f *fakeRepo) FindWebSocketMessages(
	_ context.Context,
	_, connectionID ulid.ULID,
) ([]wslog.Message, error) {
	return f.messages[connectionID.String()], nil
}

func (f *fakeRepo) ClearWebSocketConnections(context.Context, ulid.ULID) error { return nil }

func mustParse(t *testing.T, q string) httpql.Expression {
	t.Helper()

	expr, err := httpql.Parse(q)
	if err != nil {
		t.Fatalf("parse %q: %v", q, err)
	}

	return expr
}

func TestConnectionsFilter(t *testing.T) {

	chatID := ulid.MustNew(ulid.Now(), entropy)
	feedID := ulid.MustNew(ulid.Now(), entropy)

	repo := &fakeRepo{
		conns: []wslog.Connection{
			{ID: chatID, Host: "chat.example.com", Path: "/ws", URL: "wss://chat.example.com/ws"},
			{ID: feedID, Host: "feed.example.com", Path: "/live", URL: "wss://feed.example.com/live"},
		},
		messages: map[string][]wslog.Message{
			chatID.String(): {
				{ConnectionID: chatID, Direction: wsproxy.ClientToServer, Opcode: 1, Payload: []byte(`{"token":"secret"}`)},
			},
			feedID.String(): {
				{ConnectionID: feedID, Direction: wsproxy.ServerToClient, Opcode: 1, Payload: []byte(`{"price":42}`)},
			},
		},
	}

	svc := wslog.NewService(wslog.Config{Repository: repo})
	svc.SetActiveProjectID(ulid.MustNew(ulid.Now(), entropy))

	ctx := context.Background()

	// Connection-level filter.
	got, err := svc.Connections(ctx, mustParse(t, `ws.host eq "chat.example.com"`))
	if err != nil {
		t.Fatalf("connections: %v", err)
	}
	if len(got) != 1 || got[0].ID != chatID {
		t.Fatalf("host filter: got %d connections, want chat only", len(got))
	}

	// Message-level filter matches the connection that carried the frame.
	got, err = svc.Connections(ctx, mustParse(t, `ws.payload cont "secret"`))
	if err != nil {
		t.Fatalf("connections: %v", err)
	}
	if len(got) != 1 || got[0].ID != chatID {
		t.Fatalf("payload filter: got %d connections, want chat only", len(got))
	}

	// Message-level filter that matches nothing.
	got, err = svc.Connections(ctx, mustParse(t, `ws.payload cont "nope"`))
	if err != nil {
		t.Fatalf("connections: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("payload filter: got %d connections, want none", len(got))
	}
}

func TestMessagesFilter(t *testing.T) {

	connID := ulid.MustNew(ulid.Now(), entropy)

	repo := &fakeRepo{
		conns: []wslog.Connection{{ID: connID, Host: "chat.example.com", Path: "/ws"}},
		messages: map[string][]wslog.Message{
			connID.String(): {
				{ConnectionID: connID, Direction: wsproxy.ClientToServer, Opcode: 1, Payload: []byte("hello")},
				{ConnectionID: connID, Direction: wsproxy.ServerToClient, Opcode: 1, Payload: []byte("world")},
			},
		},
	}

	svc := wslog.NewService(wslog.Config{Repository: repo})
	svc.SetActiveProjectID(ulid.MustNew(ulid.Now(), entropy))

	got, err := svc.Messages(context.Background(), connID, mustParse(t, `ws.direction eq "server->client"`))
	if err != nil {
		t.Fatalf("messages: %v", err)
	}
	if len(got) != 1 || string(got[0].Payload) != "world" {
		t.Fatalf("direction filter: got %+v, want the server->client message", got)
	}
}
