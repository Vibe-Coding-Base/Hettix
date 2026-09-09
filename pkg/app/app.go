// Package app wires the Hettix backend services and exposes the shared HTTP API
// (GraphQL and exports). Both the headless HTTP-server entrypoint and the Wails
// desktop shell build their runtime through Build, so the two frontends stay in
// sync while differing only in how they present the admin UI and route traffic.
package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"

	"github.com/Vibe-Coding-Base/Hettix/pkg/api"
	"github.com/Vibe-Coding-Base/Hettix/pkg/db/sqlite"
	"github.com/Vibe-Coding-Base/Hettix/pkg/export"
	"github.com/Vibe-Coding-Base/Hettix/pkg/finding"
	"github.com/Vibe-Coding-Base/Hettix/pkg/intruder"
	"github.com/Vibe-Coding-Base/Hettix/pkg/llm"
	"github.com/Vibe-Coding-Base/Hettix/pkg/matchreplace"
	"github.com/Vibe-Coding-Base/Hettix/pkg/proj"
	"github.com/Vibe-Coding-Base/Hettix/pkg/proxy"
	"github.com/Vibe-Coding-Base/Hettix/pkg/proxy/intercept"
	"github.com/Vibe-Coding-Base/Hettix/pkg/reqlog"
	"github.com/Vibe-Coding-Base/Hettix/pkg/scope"
	"github.com/Vibe-Coding-Base/Hettix/pkg/sender"
	"github.com/Vibe-Coding-Base/Hettix/pkg/workflow"
	"github.com/Vibe-Coding-Base/Hettix/pkg/wsintercept"
	"github.com/Vibe-Coding-Base/Hettix/pkg/wslog"
)

const gqlEndpoint = "/api/graphql/"

// Config holds the inputs needed to build the backend. Paths are expected to be
// already expanded (no leading `~`).
type Config struct {
	Logger     *zap.Logger
	Version    string
	DBPath     string
	CACertFile string
	CAKeyFile  string
	// ProxyURL is the proxy address advertised to the admin UI so it can guide
	// browser configuration. It matches the address the proxy actually listens on.
	ProxyURL string
	// LLMEnv seeds stored LLM settings on first run; afterwards the Settings page
	// is the source of truth.
	LLMEnv llm.Settings
}

// App is the assembled backend: the MITM proxy, the resolver behind the GraphQL
// API, and an API mux serving GraphQL and exports. Frontends compose these with
// their own admin-UI delivery and request routing.
type App struct {
	DB       *sqlite.Database
	Proxy    *proxy.Proxy
	Resolver *api.Resolver
	// APIMux serves /api/graphql/ and /api/export/*. It does not serve the admin
	// SPA; each frontend delivers that itself.
	APIMux *http.ServeMux

	reqLogService *reqlog.Service
	logger        *zap.Logger
}

// Build wires every backend service and returns the assembled App. The caller
// owns the lifecycle and must call Close when done.
func Build(ctx context.Context, cfg Config) (*App, error) {
	caKeyFile, err := expandHome(cfg.CAKeyFile)
	if err != nil {
		return nil, fmt.Errorf("expand CA key path: %w", err)
	}
	caCertFile, err := expandHome(cfg.CACertFile)
	if err != nil {
		return nil, fmt.Errorf("expand CA certificate path: %w", err)
	}
	dbPath, err := expandHome(cfg.DBPath)
	if err != nil {
		return nil, fmt.Errorf("expand database path: %w", err)
	}

	caCert, caKey, err := proxy.LoadOrCreateCA(caKeyFile, caCertFile)
	if err != nil {
		return nil, fmt.Errorf("load or create CA key pair: %w", err)
	}

	db, err := sqlite.OpenDatabase(dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	scope := &scope.Scope{}

	reqLogService := reqlog.NewService(reqlog.Config{
		Scope:      scope,
		Repository: db,
		Logger:     cfg.Logger.Named("reqlog").Sugar(),
	})

	interceptService := intercept.NewService(intercept.Config{
		Logger: cfg.Logger.Named("intercept").Sugar(),
	})

	senderService := sender.NewService(sender.Config{
		Repository:    db,
		ReqLogService: reqLogService,
	})

	matchReplaceEngine := matchreplace.NewEngine()

	wsLogService := wslog.NewService(wslog.Config{
		Repository: db,
		Logger:     cfg.Logger.Named("websocket").Sugar(),
	})

	wsInterceptService := wsintercept.NewService(wsintercept.Config{
		Logger: cfg.Logger.Named("wsintercept").Sugar(),
	})

	intruderClient := &http.Client{Transport: &sender.HTTPTransport{}, Timeout: 30 * time.Second}
	intruderService := intruder.NewService(intruder.Config{
		Repository: db,
		Runner:     intruder.NewRunner(intruderClient, 15),
		Logger:     cfg.Logger.Named("intruder").Sugar(),
	})

	findingService := finding.NewService(finding.Config{
		Repository: db,
		Logger:     cfg.Logger.Named("finding").Sugar(),
	})

	llmManager := llm.NewManager(db)
	if err := llmManager.Load(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("load LLM settings: %w", err)
	}
	if err := llmManager.Seed(ctx, cfg.LLMEnv); err != nil {
		cfg.Logger.Named("main").Warn("Failed to seed LLM settings from environment.", zap.Error(err))
	}

	workflowService := workflow.NewService(workflow.Config{
		Repository: db,
		Runner: workflow.NewRunner(workflow.RunnerConfig{
			Search:   reqLogService,
			Fuzz:     intruderService,
			Findings: findingService,
			Scope:    scope,
		}),
	})

	projService, err := proj.NewService(proj.Config{
		Repository:                db,
		InterceptService:          interceptService,
		ReqLogService:             reqLogService,
		SenderService:             senderService,
		WebSocketService:          wsLogService,
		WebSocketInterceptService: wsInterceptService,
		IntruderService:           intruderService,
		FindingService:            findingService,
		WorkflowService:           workflowService,
		Scope:                     scope,
		MatchReplaceEngine:        matchReplaceEngine,
	})
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("create projects service: %w", err)
	}

	wsHandlers := wsLogService.Handlers()
	wsHandlers.Intercept = wsInterceptService.Intercept

	prox, err := proxy.NewProxy(proxy.Config{
		CACert:            caCert,
		CAKey:             caKey,
		Logger:            cfg.Logger.Named("proxy").Sugar(),
		WebSocketHandlers: wsHandlers,
	})
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("create proxy: %w", err)
	}

	prox.UseRequestModifier(matchReplaceEngine.RequestModifier)
	prox.UseResponseModifier(matchReplaceEngine.ResponseModifier)
	prox.UseRequestModifier(reqLogService.RequestModifier)
	prox.UseResponseModifier(reqLogService.ResponseModifier)
	prox.UseRequestModifier(interceptService.RequestModifier)
	prox.UseResponseModifier(interceptService.ResponseModifier)

	resolver := &api.Resolver{
		ProjectService:            projService,
		RequestLogService:         reqLogService,
		InterceptService:          interceptService,
		SenderService:             senderService,
		WebSocketService:          wsLogService,
		WebSocketInterceptService: wsInterceptService,
		IntruderService:           intruderService,
		FindingService:            findingService,
		WorkflowService:           workflowService,
		LLMManager:                llmManager,
		ProxyURL:                  cfg.ProxyURL,
	}

	mux := http.NewServeMux()
	mux.Handle(gqlEndpoint, api.HTTPHandler(resolver, gqlEndpoint))
	mux.Handle("/api/export/har", exportHandler(reqLogService, cfg.Version, "har"))
	mux.Handle("/api/export/csv", exportHandler(reqLogService, cfg.Version, "csv"))

	return &App{
		DB:            db,
		Proxy:         prox,
		Resolver:      resolver,
		APIMux:        mux,
		reqLogService: reqLogService,
		logger:        cfg.Logger,
	}, nil
}

// Close releases resources held by the App.
func (a *App) Close() error {
	return a.DB.Close()
}

// expandHome expands a leading "~/" (or "~\") in path to the current user's
// home directory. The "~user" form is not supported.
func expandHome(path string) (string, error) {
	if path == "" || path[0] != '~' {
		return path, nil
	}

	if len(path) > 1 && path[1] != '/' && path[1] != filepath.Separator {
		return "", fmt.Errorf("cannot expand user-specific home path: %q", path)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}

	return filepath.Join(home, path[1:]), nil
}

// exportHandler streams the active project's request logs as HAR or CSV.
func exportHandler(reqLogSvc *reqlog.Service, version, format string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logs, err := reqLogSvc.FindByQuery(r.Context(), nil, 0)
		if err != nil {
			http.Error(w, "no active project or failed to load traffic", http.StatusConflict)
			return
		}

		switch format {
		case "har":
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Content-Disposition", `attachment; filename="hettix-export.har"`)

			if err := export.WriteHAR(w, version, logs); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		case "csv":
			w.Header().Set("Content-Type", "text/csv")
			w.Header().Set("Content-Disposition", `attachment; filename="hettix-export.csv"`)

			if err := export.WriteCSV(w, logs); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		default:
			http.Error(w, "unknown export format", http.StatusBadRequest)
		}
	}
}
