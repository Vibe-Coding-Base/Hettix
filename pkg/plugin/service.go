package plugin

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/dop251/goja"
	"github.com/oklog/ulid"

	"github.com/Vibe-Coding-Base/Hettix/pkg/finding"
	"github.com/Vibe-Coding-Base/Hettix/pkg/log"
	"github.com/Vibe-Coding-Base/Hettix/pkg/proxy"
)

//go:embed prelude.js
var preludeJS string

//go:embed builtin/js-miner.js
var jsMinerSource string

// builtins are shipped plugins seeded into the plugins directory on first run so
// the operator can read and edit them.
var builtins = []struct {
	filename string
	source   string
}{
	{"js-miner.js", jsMinerSource},
}

const (
	dispatchQueueSize = 256
	pluginTimeout     = 3 * time.Second
	maxBodyBytes      = 4 << 20 // skip plugin scanning for bodies larger than this
)

// Repository persists plugin enabled state and discovered endpoints.
type Repository interface {
	LoadPluginEnabled(ctx context.Context) (map[string]bool, error)
	SavePluginEnabled(ctx context.Context, id string, enabled bool) error
	DeletePluginEnabled(ctx context.Context, id string) error
	StoreDiscoveredEndpoint(ctx context.Context, projectID ulid.ULID, host, path, source string) error
}

// FindingCreator records findings; satisfied by *finding.Service.
type FindingCreator interface {
	CreateFinding(
		ctx context.Context, title, description string, severity finding.Severity, requestLogID *ulid.ULID,
	) (finding.Finding, error)
}

// Status is a plugin's metadata plus its current enabled state, for listing.
type Status struct {
	Meta
	Enabled  bool
	Builtin  bool
	Filename string
}

// loaded is a plugin compiled into its own JavaScript runtime. A goja runtime is
// not safe for concurrent use, so all runtimes are driven by the single dispatch
// worker (and rebuilt under the service lock).
type loaded struct {
	meta       Meta
	filename   string
	builtin    bool
	source     string
	rt         *goja.Runtime
	dispatch   goja.Callable
	hasHandler bool
}

// Service loads plugins from a directory, dispatches proxied responses to the
// enabled ones, and lets the operator install, edit and remove them. It also
// implements Sink.
type Service struct {
	dir      string
	repo     Repository
	findings FindingCreator
	logger   log.Logger

	mu              sync.RWMutex
	plugins         []*loaded
	enabled         map[string]bool
	activeProjectID ulid.ULID

	queue chan Response
	once  sync.Once
}

// Config configures the plugin service.
type Config struct {
	Dir      string
	Repo     Repository
	Findings FindingCreator
	Logger   log.Logger
}

// NewService returns a plugin service. Call Load to seed and compile plugins.
func NewService(cfg Config) *Service {
	logger := cfg.Logger
	if logger == nil {
		logger = log.NewNopLogger()
	}

	return &Service{
		dir:      cfg.Dir,
		repo:     cfg.Repo,
		findings: cfg.Findings,
		logger:   logger,
		enabled:  make(map[string]bool),
		queue:    make(chan Response, dispatchQueueSize),
	}
}

// Load seeds built-in plugins on first run, compiles every plugin in the
// directory, and starts the dispatch worker.
func (s *Service) Load(ctx context.Context) error {
	if err := s.seedBuiltins(); err != nil {
		return err
	}
	if err := s.reload(ctx); err != nil {
		return err
	}

	s.once.Do(func() { go s.worker() })

	return nil
}

// seedBuiltins writes the shipped plugins into an empty/absent plugins directory
// so they exist as editable files. It never overwrites files the user has.
func (s *Service) seedBuiltins() error {
	if _, err := os.Stat(s.dir); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("plugin: stat plugins dir: %w", err)
	}

	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return fmt.Errorf("plugin: create plugins dir: %w", err)
	}
	for _, b := range builtins {
		if err := os.WriteFile(filepath.Join(s.dir, b.filename), []byte(b.source), 0o644); err != nil {
			return fmt.Errorf("plugin: seed %s: %w", b.filename, err)
		}
	}

	return nil
}

// reload compiles every .js file in the directory and swaps them in, preserving
// the enabled state of plugins that are already known.
func (s *Service) reload(ctx context.Context) error {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return fmt.Errorf("plugin: read plugins dir: %w", err)
	}

	persisted := map[string]bool{}
	if s.repo != nil {
		if persisted, err = s.repo.LoadPluginEnabled(ctx); err != nil {
			return fmt.Errorf("plugin: load enabled state: %w", err)
		}
	}

	var compiled []*loaded
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".js") {
			continue
		}
		source, err := os.ReadFile(filepath.Join(s.dir, entry.Name()))
		if err != nil {
			s.logger.Errorw("Failed to read plugin file", "file", entry.Name(), "error", err)
			continue
		}
		lp, err := s.compile(entry.Name(), string(source))
		if err != nil {
			s.logger.Errorw("Failed to load plugin", "file", entry.Name(), "error", err)
			continue
		}
		compiled = append(compiled, lp)
	}

	s.mu.Lock()
	prev := s.enabled
	enabled := make(map[string]bool, len(compiled))
	for _, lp := range compiled {
		id := lp.meta.ID
		if on, ok := prev[id]; ok {
			enabled[id] = on
		} else if on, ok := persisted[id]; ok {
			enabled[id] = on
		} else {
			enabled[id] = lp.meta.DefaultEnabled
		}
	}
	s.plugins = compiled
	s.enabled = enabled
	s.mu.Unlock()

	return nil
}

// compile builds a runtime for one plugin: the prelude, then the plugin source.
// Running the source registers its metadata and hooks but performs no scanning.
func (s *Service) compile(filename, source string) (*loaded, error) {
	rt := goja.New()
	lp := &loaded{filename: filename, source: source, builtin: isBuiltin(filename)}

	host := rt.NewObject()
	_ = host.Set("register", func(m map[string]interface{}) { lp.meta = parseMeta(m) })
	_ = host.Set("addEndpoint", func(host, path string) {
		if err := s.AddEndpoint(context.Background(), host, path, lp.meta.ID); err != nil {
			s.logger.Errorw("Plugin failed to add endpoint", "plugin", lp.meta.ID, "error", err)
		}
	})
	_ = host.Set("addFinding", func(title, description, severity string, requestLogID goja.Value) {
		var id *ulid.ULID
		if requestLogID != nil && !goja.IsUndefined(requestLogID) && !goja.IsNull(requestLogID) {
			if parsed, err := ulid.Parse(requestLogID.String()); err == nil {
				id = &parsed
			}
		}
		if err := s.AddFinding(context.Background(), title, description, severity, id); err != nil {
			s.logger.Errorw("Plugin failed to add finding", "plugin", lp.meta.ID, "error", err)
		}
	})
	_ = host.Set("log", func(msg string) { s.logger.Infow("Plugin log", "plugin", lp.meta.ID, "message", msg) })
	if err := rt.Set("__host", host); err != nil {
		return nil, err
	}

	if _, err := rt.RunString(preludeJS); err != nil {
		return nil, fmt.Errorf("prelude: %w", err)
	}
	if _, err := rt.RunString(source); err != nil {
		return nil, fmt.Errorf("plugin code: %w", err)
	}

	if strings.TrimSpace(lp.meta.ID) == "" {
		return nil, errors.New("plugin did not declare an id via hettix.plugin({...})")
	}

	dispatch, ok := goja.AssertFunction(rt.Get("__dispatch"))
	if !ok {
		return nil, errors.New("plugin runtime is missing the dispatch entry point")
	}
	if has, ok := goja.AssertFunction(rt.Get("__hasResponseHandlers")); ok {
		if v, err := has(goja.Undefined()); err == nil {
			lp.hasHandler = v.ToBoolean()
		}
	}

	lp.rt = rt
	lp.dispatch = dispatch

	return lp, nil
}

// SetActiveProjectID sets the project discovered endpoints are recorded against.
func (s *Service) SetActiveProjectID(id ulid.ULID) {
	s.mu.Lock()
	s.activeProjectID = id
	s.mu.Unlock()
}

// List returns each loaded plugin with its enabled state, sorted by name.
func (s *Service) List() []Status {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]Status, 0, len(s.plugins))
	for _, lp := range s.plugins {
		out = append(out, Status{
			Meta:     lp.meta,
			Enabled:  s.enabled[lp.meta.ID],
			Builtin:  lp.builtin,
			Filename: lp.filename,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })

	return out
}

// SetEnabled toggles a plugin and persists the choice.
func (s *Service) SetEnabled(ctx context.Context, id string, on bool) (Status, bool) {
	s.mu.Lock()
	lp := s.find(id)
	if lp == nil {
		s.mu.Unlock()
		return Status{}, false
	}
	s.enabled[id] = on
	status := Status{Meta: lp.meta, Enabled: on, Builtin: lp.builtin, Filename: lp.filename}
	s.mu.Unlock()

	if s.repo != nil {
		if err := s.repo.SavePluginEnabled(ctx, id, on); err != nil {
			s.logger.Errorw("Failed to persist plugin state", "plugin", id, "error", err)
		}
	}

	return status, true
}

// Install validates and stores a new plugin file, then reloads. The filename is
// taken from the plugin id to keep the directory predictable.
func (s *Service) Install(ctx context.Context, content string) (Status, error) {
	lp, err := s.compile("<new>", content)
	if err != nil {
		return Status{}, fmt.Errorf("invalid plugin: %w", err)
	}
	id := lp.meta.ID

	s.mu.RLock()
	existing := s.find(id)
	s.mu.RUnlock()
	if existing != nil {
		return Status{}, fmt.Errorf("a plugin with id %q is already installed", id)
	}

	filename := sanitizeFilename(id) + ".js"
	if err := os.WriteFile(filepath.Join(s.dir, filename), []byte(content), 0o644); err != nil {
		return Status{}, fmt.Errorf("write plugin: %w", err)
	}
	if err := s.reload(ctx); err != nil {
		return Status{}, err
	}

	return s.status(id)
}

// Update replaces an installed plugin's source, then reloads.
func (s *Service) Update(ctx context.Context, id, content string) (Status, error) {
	s.mu.RLock()
	lp := s.find(id)
	s.mu.RUnlock()
	if lp == nil {
		return Status{}, fmt.Errorf("unknown plugin: %s", id)
	}

	compiled, err := s.compile(lp.filename, content)
	if err != nil {
		return Status{}, fmt.Errorf("invalid plugin: %w", err)
	}
	if compiled.meta.ID != id {
		return Status{}, fmt.Errorf("plugin id must stay %q", id)
	}

	if err := os.WriteFile(filepath.Join(s.dir, lp.filename), []byte(content), 0o644); err != nil {
		return Status{}, fmt.Errorf("write plugin: %w", err)
	}
	if err := s.reload(ctx); err != nil {
		return Status{}, err
	}

	return s.status(id)
}

// Delete removes an installed plugin file and its persisted state, then reloads.
func (s *Service) Delete(ctx context.Context, id string) error {
	s.mu.RLock()
	lp := s.find(id)
	s.mu.RUnlock()
	if lp == nil {
		return fmt.Errorf("unknown plugin: %s", id)
	}

	if err := os.Remove(filepath.Join(s.dir, lp.filename)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove plugin: %w", err)
	}
	if s.repo != nil {
		if err := s.repo.DeletePluginEnabled(ctx, id); err != nil {
			s.logger.Errorw("Failed to remove plugin state", "plugin", id, "error", err)
		}
	}

	return s.reload(ctx)
}

// Source returns a plugin's editable source.
func (s *Service) Source(id string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if lp := s.find(id); lp != nil {
		return lp.source, true
	}

	return "", false
}

func (s *Service) status(id string) (Status, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	lp := s.find(id)
	if lp == nil {
		return Status{}, fmt.Errorf("unknown plugin: %s", id)
	}

	return Status{Meta: lp.meta, Enabled: s.enabled[id], Builtin: lp.builtin, Filename: lp.filename}, nil
}

// find returns the loaded plugin with the given id. Callers hold s.mu.
func (s *Service) find(id string) *loaded {
	for _, lp := range s.plugins {
		if lp.meta.ID == id {
			return lp
		}
	}

	return nil
}

// ResponseModifier is a proxy response middleware that queues each response for
// the plugin worker. It reads the (already-decompressed) body, restores it for
// downstream consumers, and never blocks the proxy.
func (s *Service) ResponseModifier(next proxy.ResponseModifyFunc) proxy.ResponseModifyFunc {
	return func(res *http.Response) error {
		if err := next(res); err != nil {
			return err
		}

		if res.Body == nil || res.Request == nil || !s.hasEnabled() {
			return nil
		}

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}
		_ = res.Body.Close()
		res.Body = io.NopCloser(bytes.NewBuffer(body))

		if len(body) > maxBodyBytes {
			return nil
		}

		resp := Response{
			Method:      res.Request.Method,
			URL:         res.Request.URL,
			Host:        res.Request.Host,
			StatusCode:  res.StatusCode,
			ContentType: res.Header.Get("Content-Type"),
			Body:        append([]byte(nil), body...),
		}

		select {
		case s.queue <- resp:
		default:
			s.logger.Debugw("Plugin queue full, dropping response")
		}

		return nil
	}
}

func (s *Service) hasEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, on := range s.enabled {
		if on {
			return true
		}
	}

	return false
}

// worker runs plugin dispatch sequentially, so the single-threaded goja runtimes
// are never used concurrently.
func (s *Service) worker() {
	for resp := range s.queue {
		s.process(resp)
	}
}

func (s *Service) process(resp Response) {
	s.mu.RLock()
	active := make([]*loaded, 0, len(s.plugins))
	for _, lp := range s.plugins {
		if lp.hasHandler && s.enabled[lp.meta.ID] {
			active = append(active, lp)
		}
	}
	s.mu.RUnlock()

	if len(active) == 0 {
		return
	}

	arg := map[string]interface{}{
		"method":      resp.Method,
		"url":         resp.URL.String(),
		"host":        resp.Host,
		"statusCode":  resp.StatusCode,
		"contentType": resp.ContentType,
		"body":        string(resp.Body),
	}
	if resp.RequestLogID != nil {
		arg["requestLogId"] = resp.RequestLogID.String()
	} else {
		arg["requestLogId"] = nil
	}

	for _, lp := range active {
		s.run(lp, arg)
	}
}

func (s *Service) run(lp *loaded, arg map[string]interface{}) {
	timer := time.AfterFunc(pluginTimeout, func() { lp.rt.Interrupt("plugin execution timed out") })
	defer timer.Stop()
	defer lp.rt.ClearInterrupt()
	defer func() {
		if r := recover(); r != nil {
			s.logger.Errorw("Plugin panicked", "plugin", lp.meta.ID, "error", r)
		}
	}()

	if _, err := lp.dispatch(goja.Undefined(), lp.rt.ToValue(arg)); err != nil {
		s.logger.Errorw("Plugin failed on response", "plugin", lp.meta.ID, "error", err)
	}
}

// AddEndpoint records a discovered endpoint for the active project.
func (s *Service) AddEndpoint(ctx context.Context, host, path, source string) error {
	s.mu.RLock()
	projectID := s.activeProjectID
	s.mu.RUnlock()

	if projectID.Compare(ulid.ULID{}) == 0 || s.repo == nil {
		return nil
	}
	host = strings.TrimSpace(host)
	path = strings.TrimSpace(path)
	if host == "" || path == "" {
		return nil
	}

	return s.repo.StoreDiscoveredEndpoint(ctx, projectID, host, path, source)
}

// AddFinding records a finding for the active project.
func (s *Service) AddFinding(
	ctx context.Context, title, description, severity string, requestLogID *ulid.ULID,
) error {
	if s.findings == nil {
		return nil
	}
	_, err := s.findings.CreateFinding(ctx, title, description, finding.ParseSeverity(severity), requestLogID)

	return err
}

func parseMeta(m map[string]interface{}) Meta {
	meta := Meta{
		ID:          asString(m["id"]),
		Name:        asString(m["name"]),
		Description: asString(m["description"]),
		Version:     asString(m["version"]),
	}
	if enabled, ok := m["defaultEnabled"].(bool); ok {
		meta.DefaultEnabled = enabled
	}
	if caps, ok := m["capabilities"].([]interface{}); ok {
		for _, c := range caps {
			if str := asString(c); str != "" {
				meta.Capabilities = append(meta.Capabilities, str)
			}
		}
	}
	if meta.Name == "" {
		meta.Name = meta.ID
	}

	return meta
}

func asString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}

	return ""
}

func isBuiltin(filename string) bool {
	for _, b := range builtins {
		if b.filename == filename {
			return true
		}
	}

	return false
}

// sanitizeFilename reduces a plugin id to a safe file base name.
func sanitizeFilename(id string) string {
	safe := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			return r
		default:
			return '-'
		}
	}, id)
	safe = strings.Trim(safe, "-")
	if safe == "" {
		safe = "plugin"
	}

	return safe
}
