// Package wsproxy man-in-the-middles WebSocket connections: it completes the
// upgrade with the client, dials the upstream, and relays messages in both
// directions, invoking lifecycle hooks so connections and messages can be
// stored. It is self-contained so it can be tested against a real WebSocket
// server before being wired into the live proxy.
package wsproxy

import (
	"context"
	"crypto/sha1" //nolint:gosec // required by RFC 6455 for the handshake accept key
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/oklog/ulid"
)

// Direction is the travel direction of a WebSocket message.
type Direction int

const (
	ClientToServer Direction = iota
	ServerToClient
)

func (d Direction) String() string {
	if d == ServerToClient {
		return "server->client"
	}

	return "client->server"
}

// ConnInfo describes a proxied WebSocket connection. ID is a ULID that links
// the connection to its messages.
type ConnInfo struct {
	ID   string
	URL  string
	Host string
	Path string
	Time time.Time
}

// Message is a captured WebSocket message. ConnID matches the ConnInfo.ID of
// the connection it belongs to.
type Message struct {
	ConnID    string
	Direction Direction
	Opcode    ws.OpCode
	Payload   []byte
	Time      time.Time
}

// Handlers receive the lifecycle events of a proxied connection. Every field is
// optional.
type Handlers struct {
	// OnOpen is called once, after the upstream is dialed and the client
	// handshake completes, before any message is relayed.
	OnOpen func(ConnInfo)
	// OnMessage is called for each relayed data message.
	OnMessage func(Message)
	// OnClose is called once when the connection is torn down.
	OnClose func(connID string)
}

const handshakeMagic = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

//nolint:gosec // connection IDs don't need cryptographic randomness
var (
	connEntropy   = rand.New(rand.NewSource(time.Now().UnixNano()))
	connEntropyMu sync.Mutex
)

func newConnID() string {
	connEntropyMu.Lock()
	defer connEntropyMu.Unlock()

	return ulid.MustNew(ulid.Timestamp(time.Now()), connEntropy).String()
}

// IsUpgrade reports whether a request is a WebSocket upgrade.
func IsUpgrade(r *http.Request) bool {
	return tokenListContains(r.Header.Get("Connection"), "upgrade") &&
		strings.EqualFold(strings.TrimSpace(r.Header.Get("Upgrade")), "websocket")
}

// Handler proxies a WebSocket upgrade request: it dials the upstream, hijacks
// the client connection, completes the handshake and relays messages until
// either side closes, invoking the lifecycle hooks along the way.
func Handler(w http.ResponseWriter, r *http.Request, h Handlers) error {
	target := targetURL(r)

	upstream, err := dialUpstream(r, target)
	if err != nil {
		http.Error(w, "websocket upstream dial failed", http.StatusBadGateway)
		return err
	}
	defer upstream.Close()

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		return fmt.Errorf("wsproxy: response writer does not support hijacking")
	}

	client, _, err := hijacker.Hijack()
	if err != nil {
		return fmt.Errorf("wsproxy: failed to hijack client connection: %w", err)
	}
	defer client.Close()

	if err := writeClientHandshake(client, r); err != nil {
		return err
	}

	connID := newConnID()

	if h.OnOpen != nil {
		h.OnOpen(ConnInfo{
			ID:   connID,
			URL:  target.String(),
			Host: target.Host,
			Path: target.Path,
			Time: time.Now(),
		})
	}

	if h.OnClose != nil {
		defer h.OnClose(connID)
	}

	return relay(connID, client, upstream, h.OnMessage)
}

func targetURL(r *http.Request) url.URL {
	scheme := "ws"
	if r.TLS != nil {
		scheme = "wss"
	}

	host := r.URL.Host
	if host == "" {
		host = r.Host
	}

	return url.URL{Scheme: scheme, Host: host, Path: r.URL.Path, RawQuery: r.URL.RawQuery}
}

func dialUpstream(r *http.Request, target url.URL) (net.Conn, error) {
	dialer := ws.Dialer{
		Header:    ws.HandshakeHeaderHTTP(upstreamHeaders(r.Header)),
		TLSConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // MITM proxy dials many hosts
		Timeout:   30 * time.Second,
	}

	conn, _, _, err := dialer.Dial(context.Background(), target.String())
	if err != nil {
		return nil, fmt.Errorf("wsproxy: upstream dial failed: %w", err)
	}

	return conn, nil
}

func relay(connID string, client, upstream net.Conn, onMessage func(Message)) error {
	errc := make(chan error, 2)

	go pump(connID, client, upstream, true, ClientToServer, onMessage, errc)
	go pump(connID, upstream, client, false, ServerToClient, onMessage, errc)

	return <-errc
}

// pump reads data messages from src and writes them to dst. srcIsClient selects
// the framing (client frames are masked); the write to dst uses the same role so
// masking stays correct end to end.
func pump(
	connID string,
	src, dst net.Conn,
	srcIsClient bool,
	dir Direction,
	onMessage func(Message),
	errc chan<- error,
) {
	for {
		var (
			data []byte
			op   ws.OpCode
			err  error
		)

		if srcIsClient {
			data, op, err = wsutil.ReadClientData(src)
		} else {
			data, op, err = wsutil.ReadServerData(src)
		}
		if err != nil {
			errc <- err
			return
		}

		if onMessage != nil {
			onMessage(Message{ConnID: connID, Direction: dir, Opcode: op, Payload: data, Time: time.Now()})
		}

		if srcIsClient {
			err = wsutil.WriteClientMessage(dst, op, data)
		} else {
			err = wsutil.WriteServerMessage(dst, op, data)
		}
		if err != nil {
			errc <- err
			return
		}
	}
}

// upstreamHeaders forwards application headers to the upstream handshake,
// dropping hop-by-hop and WebSocket-negotiation headers that the dialer sets.
func upstreamHeaders(in http.Header) http.Header {
	out := http.Header{}

	for key, values := range in {
		switch http.CanonicalHeaderKey(key) {
		case "Upgrade", "Connection", "Sec-Websocket-Key", "Sec-Websocket-Version",
			"Sec-Websocket-Extensions", "Host", "Content-Length":
			continue
		default:
			for _, v := range values {
				out.Add(key, v)
			}
		}
	}

	return out
}

func writeClientHandshake(client net.Conn, r *http.Request) error {
	accept := acceptKey(r.Header.Get("Sec-WebSocket-Key"))

	resp := "HTTP/1.1 101 Switching Protocols\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Accept: " + accept + "\r\n"

	if proto := r.Header.Get("Sec-WebSocket-Protocol"); proto != "" {
		resp += "Sec-WebSocket-Protocol: " + proto + "\r\n"
	}

	resp += "\r\n"

	if _, err := client.Write([]byte(resp)); err != nil {
		return fmt.Errorf("wsproxy: failed to write client handshake: %w", err)
	}

	return nil
}

func acceptKey(key string) string {
	h := sha1.New() //nolint:gosec // RFC 6455 mandates SHA-1 here
	h.Write([]byte(key + handshakeMagic))

	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func tokenListContains(header, token string) bool {
	for _, part := range strings.Split(header, ",") {
		if strings.EqualFold(strings.TrimSpace(part), token) {
			return true
		}
	}

	return false
}
