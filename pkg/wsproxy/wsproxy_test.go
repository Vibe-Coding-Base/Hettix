package wsproxy_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/Vibe-Coding-Base/Hettix/pkg/wsproxy"
)

func TestProxyRelaysAndCaptures(t *testing.T) {
	// Upstream echo server.
	upgrader := websocket.Upgrader{}
	echo := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		for {
			mt, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if err := conn.WriteMessage(mt, msg); err != nil {
				return
			}
		}
	}))
	defer echo.Close()

	echoHost := strings.TrimPrefix(echo.URL, "http://")

	var (
		mu       sync.Mutex
		captured []wsproxy.Message
		openID   string
		closeID  string
	)

	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !wsproxy.IsUpgrade(r) {
			http.Error(w, "not a websocket upgrade", http.StatusBadRequest)
			return
		}
		r.URL.Host = echoHost
		r.URL.Scheme = "ws"
		_ = wsproxy.Handler(w, r, wsproxy.Handlers{
			OnOpen: func(c wsproxy.ConnInfo) {
				mu.Lock()
				openID = c.ID
				mu.Unlock()
			},
			OnMessage: func(m wsproxy.Message) {
				mu.Lock()
				captured = append(captured, wsproxy.Message{
					ConnID:    m.ConnID,
					Direction: m.Direction,
					Opcode:    m.Opcode,
					Payload:   append([]byte(nil), m.Payload...),
				})
				mu.Unlock()
			},
			OnClose: func(connID string) {
				mu.Lock()
				closeID = connID
				mu.Unlock()
			},
		})
	}))
	defer proxy.Close()

	wsURL := "ws" + strings.TrimPrefix(proxy.URL, "http") + "/chat"
	client, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("client dial through proxy failed: %v", err)
	}
	defer client.Close()

	if err := client.WriteMessage(websocket.TextMessage, []byte("hello hettix")); err != nil {
		t.Fatalf("write: %v", err)
	}

	_, msg, err := client.ReadMessage()
	if err != nil {
		t.Fatalf("read echo: %v", err)
	}
	if string(msg) != "hello hettix" {
		t.Fatalf("expected echo %q, got %q", "hello hettix", msg)
	}

	// Both directions should have been captured.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		mu.Lock()
		n := len(captured)
		mu.Unlock()
		if n >= 2 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	mu.Lock()

	if openID == "" {
		t.Error("OnOpen was not invoked")
	}

	var sawClient, sawServer bool
	for _, m := range captured {
		if m.ConnID != openID {
			t.Errorf("message ConnID %q does not match connection ID %q", m.ConnID, openID)
		}
		if string(m.Payload) != "hello hettix" {
			continue
		}
		if m.Direction == wsproxy.ClientToServer {
			sawClient = true
		}
		if m.Direction == wsproxy.ServerToClient {
			sawServer = true
		}
	}
	mu.Unlock()

	if !sawClient {
		t.Error("did not capture the client->server message")
	}
	if !sawServer {
		t.Error("did not capture the server->client message")
	}

	// Close the client so the relay tears down and OnClose fires.
	client.Close()

	deadline = time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		mu.Lock()
		id := closeID
		mu.Unlock()
		if id != "" {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	mu.Lock()
	gotClose := closeID
	mu.Unlock()
	if gotClose != openID {
		t.Errorf("OnClose ID %q does not match connection ID %q", gotClose, openID)
	}
}
