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
	)

	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !wsproxy.IsUpgrade(r) {
			http.Error(w, "not a websocket upgrade", http.StatusBadRequest)
			return
		}
		r.URL.Host = echoHost
		r.URL.Scheme = "ws"
		_ = wsproxy.Handler(w, r, func(m wsproxy.Message) {
			mu.Lock()
			captured = append(captured, wsproxy.Message{Direction: m.Direction, Opcode: m.Opcode, Payload: append([]byte(nil), m.Payload...)})
			mu.Unlock()
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
	defer mu.Unlock()

	var sawClient, sawServer bool
	for _, m := range captured {
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

	if !sawClient {
		t.Error("did not capture the client->server message")
	}
	if !sawServer {
		t.Error("did not capture the server->client message")
	}
}
