// Package wsintercept holds proxied WebSocket messages so an operator can
// modify or drop them before they are forwarded. It adapts the proxy's
// interception hook to a queue of pending frames that the API drives, mirroring
// the HTTP intercept service.
package wsintercept

import (
	"context"
	"errors"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/oklog/ulid"

	"github.com/Vibe-Coding-Base/Hettix/pkg/httpql"
	"github.com/Vibe-Coding-Base/Hettix/pkg/log"
	"github.com/Vibe-Coding-Base/Hettix/pkg/wsproxy"
)

var (
	ErrFrameNotFound = errors.New("wsintercept: frame not found")
	ErrFrameDone     = errors.New("wsintercept: frame is done")
)

// Frame is a held WebSocket message awaiting an operator decision.
type Frame struct {
	ID        ulid.ULID
	ConnID    string
	Direction wsproxy.Direction
	Opcode    int
	Payload   []byte
}

type decision struct {
	payload []byte
	drop    bool
}

type pending struct {
	frame Frame
	ch    chan decision
	done  chan struct{}
}

// Settings configures message interception.
type Settings struct {
	Enabled bool
	Filter  httpql.Expression
}

type Service struct {
	mu     sync.RWMutex
	frames map[ulid.ULID]pending

	enabled bool
	filter  httpql.Expression

	logger    log.Logger
	entropy   *rand.Rand
	entropyMu sync.Mutex
}

type Config struct {
	Logger  log.Logger
	Enabled bool
	Filter  httpql.Expression
}

func NewService(cfg Config) *Service {
	s := &Service{
		frames:  make(map[ulid.ULID]pending),
		enabled: cfg.Enabled,
		filter:  cfg.Filter,
		logger:  cfg.Logger,
		//nolint:gosec // frame IDs don't need cryptographic randomness
		entropy: rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	if s.logger == nil {
		s.logger = log.NewNopLogger()
	}

	return s
}

func (s *Service) newID() ulid.ULID {
	s.entropyMu.Lock()
	defer s.entropyMu.Unlock()

	return ulid.MustNew(ulid.Timestamp(time.Now()), s.entropy)
}

// Intercept is the wsproxy.InterceptFunc: it holds a matching message until an
// operator forwards, modifies or drops it, or the connection closes.
func (s *Service) Intercept(ctx context.Context, m wsproxy.Message) ([]byte, bool) {
	s.mu.RLock()
	enabled := s.enabled
	filter := s.filter
	s.mu.RUnlock()

	if !enabled {
		return m.Payload, false
	}

	if filter != nil {
		match, err := httpql.Eval(filter, recordForMessage(m))
		if err != nil {
			s.logger.Errorw("Failed to match WebSocket intercept filter.", "error", err)
			return m.Payload, false
		}

		if !match {
			return m.Payload, false
		}
	}

	id := s.newID()
	ch := make(chan decision)
	done := make(chan struct{})

	s.mu.Lock()
	s.frames[id] = pending{
		frame: Frame{
			ID:        id,
			ConnID:    m.ConnID,
			Direction: m.Direction,
			Opcode:    int(m.Opcode),
			Payload:   m.Payload,
		},
		ch:   ch,
		done: done,
	}
	s.mu.Unlock()

	defer func() {
		close(done)
		s.mu.Lock()
		delete(s.frames, id)
		s.mu.Unlock()
	}()

	select {
	case d := <-ch:
		return d.payload, d.drop
	case <-ctx.Done():
		// The connection is closing; forward as-is rather than block.
		return m.Payload, false
	}
}

// ForwardFrame releases a held frame unchanged.
func (s *Service) ForwardFrame(id ulid.ULID) error {
	s.mu.RLock()
	p, ok := s.frames[id]
	s.mu.RUnlock()

	if !ok {
		return ErrFrameNotFound
	}

	return s.send(p, decision{payload: p.frame.Payload})
}

// ModifyFrame releases a held frame with a replaced payload.
func (s *Service) ModifyFrame(id ulid.ULID, payload []byte) error {
	s.mu.RLock()
	p, ok := s.frames[id]
	s.mu.RUnlock()

	if !ok {
		return ErrFrameNotFound
	}

	return s.send(p, decision{payload: payload})
}

// DropFrame drops a held frame without forwarding it.
func (s *Service) DropFrame(id ulid.ULID) error {
	s.mu.RLock()
	p, ok := s.frames[id]
	s.mu.RUnlock()

	if !ok {
		return ErrFrameNotFound
	}

	return s.send(p, decision{drop: true})
}

func (s *Service) send(p pending, d decision) error {
	select {
	case <-p.done:
		return ErrFrameDone
	case p.ch <- d:
		return nil
	}
}

// Frames returns the pending frames ordered by ID (capture order).
func (s *Service) Frames() []Frame {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := make([]ulid.ULID, 0, len(s.frames))
	for id := range s.frames {
		ids = append(ids, id)
	}

	sort.Slice(ids, func(i, j int) bool { return ids[i].Compare(ids[j]) == -1 })

	frames := make([]Frame, len(ids))
	for i, id := range ids {
		frames[i] = s.frames[id].frame
	}

	return frames
}

// FrameByID returns a pending frame by ID.
func (s *Service) FrameByID(id ulid.ULID) (Frame, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	p, ok := s.frames[id]
	if !ok {
		return Frame{}, ErrFrameNotFound
	}

	return p.frame, nil
}

func (s *Service) Settings() Settings {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return Settings{Enabled: s.enabled, Filter: s.filter}
}

// UpdateSettings applies new interception settings, forwarding any pending
// frames when interception is turned off.
func (s *Service) UpdateSettings(settings Settings) {
	s.mu.Lock()
	wasEnabled := s.enabled
	s.enabled = settings.Enabled
	s.filter = settings.Filter
	s.mu.Unlock()

	if wasEnabled && !settings.Enabled {
		s.forwardAll()
	}
}

func (s *Service) forwardAll() {
	s.mu.RLock()
	pendings := make([]pending, 0, len(s.frames))
	for _, p := range s.frames {
		pendings = append(pendings, p)
	}
	s.mu.RUnlock()

	for _, p := range pendings {
		select {
		case <-p.done:
		case p.ch <- decision{payload: p.frame.Payload}:
		}
	}
}

func recordForMessage(m wsproxy.Message) httpql.Record {
	return httpql.Record{WebSocket: &httpql.WebSocketData{
		Direction: m.Direction.String(),
		Opcode:    int(m.Opcode),
		Payload:   string(m.Payload),
	}}
}
