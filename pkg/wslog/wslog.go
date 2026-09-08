// Package wslog stores and retrieves the WebSocket connections and messages
// captured by the proxy, scoped to the active project. It adapts the proxy's
// wsproxy lifecycle hooks to a repository so captured traffic can be queried
// through the API.
package wslog

import (
	"context"
	"errors"
	"time"

	"github.com/oklog/ulid"

	"github.com/Vibe-Coding-Base/Hettix/pkg/log"
	"github.com/Vibe-Coding-Base/Hettix/pkg/wsproxy"
)

var (
	ErrConnectionNotFound = errors.New("wslog: connection not found")
	ErrProjectIDMustBeSet = errors.New("wslog: project ID must be set")
)

// Connection is a proxied WebSocket connection.
type Connection struct {
	ID           ulid.ULID
	ProjectID    ulid.ULID
	URL          string
	Host         string
	Path         string
	CreatedAt    time.Time
	ClosedAt     *time.Time
	MessageCount int
}

// Message is a single relayed WebSocket message.
type Message struct {
	ConnectionID ulid.ULID
	Direction    wsproxy.Direction
	Opcode       int
	Payload      []byte
	CreatedAt    time.Time
}

// Repository persists WebSocket connections and messages.
type Repository interface {
	StoreWebSocketConnection(ctx context.Context, conn Connection) error
	CloseWebSocketConnection(ctx context.Context, projectID, id ulid.ULID, at time.Time) error
	StoreWebSocketMessage(ctx context.Context, msg Message) error
	FindWebSocketConnections(ctx context.Context, projectID ulid.ULID) ([]Connection, error)
	FindWebSocketConnectionByID(ctx context.Context, projectID, id ulid.ULID) (Connection, error)
	FindWebSocketMessages(ctx context.Context, projectID, connectionID ulid.ULID) ([]Message, error)
	ClearWebSocketConnections(ctx context.Context, projectID ulid.ULID) error
}

// Service records captured WebSocket traffic for the active project and exposes
// it for querying.
type Service struct {
	activeProjectID ulid.ULID
	repo            Repository
	logger          log.Logger
}

type Config struct {
	Repository Repository
	Logger     log.Logger
}

func NewService(cfg Config) *Service {
	s := &Service{
		repo:   cfg.Repository,
		logger: cfg.Logger,
	}

	if s.logger == nil {
		s.logger = log.NewNopLogger()
	}

	return s
}

func (svc *Service) SetActiveProjectID(id ulid.ULID) {
	svc.activeProjectID = id
}

func (svc *Service) ActiveProjectID() ulid.ULID {
	return svc.activeProjectID
}

// Handlers returns the wsproxy lifecycle hooks that persist captured traffic.
// Capturing is bypassed while no project is active.
func (svc *Service) Handlers() wsproxy.Handlers {
	return wsproxy.Handlers{
		OnOpen:    svc.onOpen,
		OnMessage: svc.onMessage,
		OnClose:   svc.onClose,
	}
}

func (svc *Service) onOpen(info wsproxy.ConnInfo) {
	projectID := svc.activeProjectID
	if projectID.Compare(ulid.ULID{}) == 0 {
		return
	}

	id, err := ulid.Parse(info.ID)
	if err != nil {
		svc.logger.Errorw("Failed to parse WebSocket connection ID.", "error", err, "id", info.ID)
		return
	}

	conn := Connection{
		ID:        id,
		ProjectID: projectID,
		URL:       info.URL,
		Host:      info.Host,
		Path:      info.Path,
		CreatedAt: info.Time,
	}

	if err := svc.repo.StoreWebSocketConnection(context.Background(), conn); err != nil {
		svc.logger.Errorw("Failed to store WebSocket connection.", "error", err)
	}
}

func (svc *Service) onMessage(m wsproxy.Message) {
	projectID := svc.activeProjectID
	if projectID.Compare(ulid.ULID{}) == 0 {
		return
	}

	connID, err := ulid.Parse(m.ConnID)
	if err != nil {
		svc.logger.Errorw("Failed to parse WebSocket connection ID.", "error", err, "id", m.ConnID)
		return
	}

	msg := Message{
		ConnectionID: connID,
		Direction:    m.Direction,
		Opcode:       int(m.Opcode),
		Payload:      m.Payload,
		CreatedAt:    m.Time,
	}

	if err := svc.repo.StoreWebSocketMessage(context.Background(), msg); err != nil {
		svc.logger.Errorw("Failed to store WebSocket message.", "error", err)
	}
}

func (svc *Service) onClose(connID string) {
	projectID := svc.activeProjectID
	if projectID.Compare(ulid.ULID{}) == 0 {
		return
	}

	id, err := ulid.Parse(connID)
	if err != nil {
		svc.logger.Errorw("Failed to parse WebSocket connection ID.", "error", err, "id", connID)
		return
	}

	if err := svc.repo.CloseWebSocketConnection(context.Background(), projectID, id, time.Now()); err != nil {
		svc.logger.Errorw("Failed to mark WebSocket connection closed.", "error", err)
	}
}

func (svc *Service) Connections(ctx context.Context) ([]Connection, error) {
	if svc.activeProjectID.Compare(ulid.ULID{}) == 0 {
		return nil, ErrProjectIDMustBeSet
	}

	return svc.repo.FindWebSocketConnections(ctx, svc.activeProjectID)
}

func (svc *Service) ConnectionByID(ctx context.Context, id ulid.ULID) (Connection, error) {
	if svc.activeProjectID.Compare(ulid.ULID{}) == 0 {
		return Connection{}, ErrProjectIDMustBeSet
	}

	return svc.repo.FindWebSocketConnectionByID(ctx, svc.activeProjectID, id)
}

func (svc *Service) Messages(ctx context.Context, connectionID ulid.ULID) ([]Message, error) {
	if svc.activeProjectID.Compare(ulid.ULID{}) == 0 {
		return nil, ErrProjectIDMustBeSet
	}

	return svc.repo.FindWebSocketMessages(ctx, svc.activeProjectID, connectionID)
}
