package main

import (
	"context"
	"crypto/tls"
	"embed"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/peterbourgon/ff/v3/ffcli"
	"go.uber.org/zap"

	"github.com/Vibe-Coding-Base/Hettix/pkg/api"
	"github.com/Vibe-Coding-Base/Hettix/pkg/chrome"
	"github.com/Vibe-Coding-Base/Hettix/pkg/db/sqlite"
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

var version = "0.0.0"

// The `all:` prefix embeds every file, including any whose name starts with
// "." or "_" that a plain embed would skip.
//
//go:embed all:admin
var adminContent embed.FS

var hettixUsage = `
Usage:
    hettix [flags] [subcommand] [flags]

Runs an HTTP server with (MITM) proxy, GraphQL service, and a web based admin interface.

Options:
    --cert         Path to root CA certificate. Creates file if it doesn't exist. (Default: "~/.hettix/hettix_cert.pem")
    --key          Path to root CA private key. Creates file if it doesn't exist. (Default: "~/.hettix/hettix_key.pem")
    --db           Database file path. Creates file if it doesn't exist. (Default: "~/.hettix/hettix.db")
    --addr         TCP address for HTTP server to listen on, in the form \"host:port\". (Default: ":8080")
    --chrome       Launch Chrome with proxy settings applied and certificate errors ignored. (Default: false)
    --verbose      Enable verbose logging.
    --json         Encode logs as JSON, instead of pretty/human readable output.
    --version, -v  Output version.
    --help, -h     Output this usage text.

Subcommands:
    - cert  Certificate management

Run ` + "`hettix <subcommand> --help`" + ` for subcommand specific usage instructions.

Hettix - an HTTP toolkit for security research.
`

type HettixCommand struct {
	config *Config

	cert    string
	key     string
	db      string
	addr    string
	chrome  bool
	version bool
}

func NewHettixCommand() (*ffcli.Command, *Config) {
	cmd := HettixCommand{
		config: &Config{},
	}

	fs := flag.NewFlagSet("hettix", flag.ExitOnError)

	fs.StringVar(&cmd.cert, "cert", "~/.hettix/hettix_cert.pem",
		"Path to root CA certificate. Creates a new certificate if file doesn't exist.")
	fs.StringVar(&cmd.key, "key", "~/.hettix/hettix_key.pem",
		"Path to root CA private key. Creates a new private key if file doesn't exist.")
	fs.StringVar(&cmd.db, "db", "~/.hettix/hettix.db", "Database file path. Creates file if it doesn't exist.")
	fs.StringVar(&cmd.addr, "addr", ":8080", "TCP address to listen on, in the form \"host:port\".")
	fs.BoolVar(&cmd.chrome, "chrome", false, "Launch Chrome with proxy settings applied and certificate errors ignored.")
	fs.BoolVar(&cmd.version, "version", false, "Output version.")
	fs.BoolVar(&cmd.version, "v", false, "Output version.")

	cmd.config.RegisterFlags(fs)

	return &ffcli.Command{
		Name:    "hettix",
		FlagSet: fs,
		Subcommands: []*ffcli.Command{
			NewCertCommand(cmd.config),
		},
		Exec: cmd.Exec,
		UsageFunc: func(*ffcli.Command) string {
			return hettixUsage
		},
	}, cmd.config
}

func (cmd *HettixCommand) Exec(ctx context.Context, _ []string) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	if cmd.version {
		fmt.Fprint(os.Stdout, version+"\n")
		return nil
	}

	mainLogger := cmd.config.logger.Named("main")

	listenHost, listenPort, err := net.SplitHostPort(cmd.addr)
	if err != nil {
		mainLogger.Fatal("Failed to parse listening address.", zap.Error(err))
	}

	url := fmt.Sprintf("http://%v:%v", listenHost, listenPort)
	if listenHost == "" || listenHost == "0.0.0.0" || listenHost == "127.0.0.1" || listenHost == "::1" {
		url = fmt.Sprintf("http://localhost:%v", listenPort)
	}

	// Expand `~` in filepaths.
	caCertFile, err := expandHome(cmd.cert)
	if err != nil {
		cmd.config.logger.Fatal("Failed to parse CA certificate filepath.", zap.Error(err))
	}

	caKeyFile, err := expandHome(cmd.key)
	if err != nil {
		cmd.config.logger.Fatal("Failed to parse CA private key filepath.", zap.Error(err))
	}

	dbPath, err := expandHome(cmd.db)
	if err != nil {
		cmd.config.logger.Fatal("Failed to parse database path.", zap.Error(err))
	}

	// Load existing CA certificate and key from disk, or generate and write
	// to disk if no files exist yet.
	caCert, caKey, err := proxy.LoadOrCreateCA(caKeyFile, caCertFile)
	if err != nil {
		cmd.config.logger.Fatal("Failed to load or create CA key pair.", zap.Error(err))
	}

	db, err := sqlite.OpenDatabase(dbPath)
	if err != nil {
		cmd.config.logger.Fatal("Failed to open database.", zap.Error(err))
	}
	defer func() { _ = db.Close() }()

	scope := &scope.Scope{}

	reqLogService := reqlog.NewService(reqlog.Config{
		Scope:      scope,
		Repository: db,
		Logger:     cmd.config.logger.Named("reqlog").Sugar(),
	})

	interceptService := intercept.NewService(intercept.Config{
		Logger: cmd.config.logger.Named("intercept").Sugar(),
	})

	senderService := sender.NewService(sender.Config{
		Repository:    db,
		ReqLogService: reqLogService,
	})

	matchReplaceEngine := matchreplace.NewEngine()

	wsLogService := wslog.NewService(wslog.Config{
		Repository: db,
		Logger:     cmd.config.logger.Named("websocket").Sugar(),
	})

	wsInterceptService := wsintercept.NewService(wsintercept.Config{
		Logger: cmd.config.logger.Named("wsintercept").Sugar(),
	})

	intruderClient := &http.Client{Transport: &sender.HTTPTransport{}, Timeout: 30 * time.Second}
	intruderService := intruder.NewService(intruder.Config{
		Repository: db,
		Runner:     intruder.NewRunner(intruderClient, 15),
		Logger:     cmd.config.logger.Named("intruder").Sugar(),
	})

	findingService := finding.NewService(finding.Config{
		Repository: db,
		Logger:     cmd.config.logger.Named("finding").Sugar(),
	})

	llmManager := llm.NewManager(db)
	if err := llmManager.Load(ctx); err != nil {
		cmd.config.logger.Fatal("Failed to load LLM settings.", zap.Error(err))
	}
	if err := llmManager.Seed(ctx, llmSettingsFromEnv()); err != nil {
		mainLogger.Warn("Failed to seed LLM settings from environment.", zap.Error(err))
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
		cmd.config.logger.Fatal("Failed to create new projects service.", zap.Error(err))
	}

	wsHandlers := wsLogService.Handlers()
	wsHandlers.Intercept = wsInterceptService.Intercept

	proxy, err := proxy.NewProxy(proxy.Config{
		CACert:            caCert,
		CAKey:             caKey,
		Logger:            cmd.config.logger.Named("proxy").Sugar(),
		WebSocketHandlers: wsHandlers,
	})
	if err != nil {
		cmd.config.logger.Fatal("Failed to create new proxy.", zap.Error(err))
	}

	proxy.UseRequestModifier(matchReplaceEngine.RequestModifier)
	proxy.UseResponseModifier(matchReplaceEngine.ResponseModifier)
	proxy.UseRequestModifier(reqLogService.RequestModifier)
	proxy.UseResponseModifier(reqLogService.ResponseModifier)
	proxy.UseRequestModifier(interceptService.RequestModifier)
	proxy.UseResponseModifier(interceptService.ResponseModifier)

	fsSub, err := fs.Sub(adminContent, "admin")
	if err != nil {
		cmd.config.logger.Fatal("Failed to construct file system subtree from admin dir.", zap.Error(err))
	}

	adminHandler := spaFileServer(fsSub)

	gqlEndpoint := "/api/graphql/"
	adminMux := http.NewServeMux()
	adminMux.Handle(gqlEndpoint, api.HTTPHandler(&api.Resolver{
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
		ProxyURL:                  url,
	}, gqlEndpoint))
	adminMux.Handle("/api/export/har", exportHandler(reqLogService, "har"))
	adminMux.Handle("/api/export/csv", exportHandler(reqLogService, "csv"))
	adminMux.Handle("/", adminHandler)

	hostname, _ := os.Hostname()

	// isAdminRequest reports whether a request targets the local admin interface
	// rather than the MITM proxy. It is an admin request when the Host is
	// well-known (the hostname, `hettix.proxy`, `localhost:[port]` or the listen
	// address), or when it is a plain request that is neither a CONNECT tunnel
	// nor a proxied absolute URL.
	isAdminRequest := func(req *http.Request) bool {
		host, _, _ := net.SplitHostPort(req.Host)

		return strings.EqualFold(host, hostname) ||
			req.Host == "hettix.proxy" ||
			req.Host == fmt.Sprintf("%v:%v", "localhost", listenPort) ||
			req.Host == fmt.Sprintf("%v:%v", listenHost, listenPort) ||
			req.Method != http.MethodConnect && !strings.HasPrefix(req.RequestURI, "http://")
	}

	// Proxy requests must reach the proxy handler with their path untouched, so
	// they bypass the admin ServeMux (which would clean and redirect paths).
	router := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if isAdminRequest(req) {
			adminMux.ServeHTTP(w, req)
			return
		}

		proxy.ServeHTTP(w, req)
	})

	httpServer := &http.Server{
		Addr:         cmd.addr,
		Handler:      router,
		TLSNextProto: map[string]func(*http.Server, *tls.Conn, http.Handler){}, // Disable HTTP/2
		ErrorLog:     zap.NewStdLog(cmd.config.logger.Named("http")),
	}

	go func() {
		mainLogger.Info(fmt.Sprintf("Hettix (v%v) is running on %v ...", version, cmd.addr))
		mainLogger.Info(fmt.Sprintf("\x1b[%dm%s\x1b[0m", uint8(32), "Get started at "+url))

		err := httpServer.ListenAndServe()
		if err != http.ErrServerClosed {
			mainLogger.Fatal("HTTP server closed unexpected.", zap.Error(err))
		}
	}()

	if cmd.chrome {
		ctx, cancel := chrome.NewExecAllocator(ctx, chrome.Config{
			ProxyServer:      url,
			ProxyBypassHosts: []string{url},
		})
		defer cancel()

		taskCtx, cancel := chromedp.NewContext(ctx)
		defer cancel()

		err = chromedp.Run(taskCtx, chromedp.Navigate(url))

		switch {
		case errors.Is(err, exec.ErrNotFound):
			mainLogger.Info("Chrome executable not found.")
		case err != nil:
			mainLogger.Error(fmt.Sprintf("Failed to navigate to %v.", url), zap.Error(err))
		default:
			mainLogger.Info("Launched Chrome.")
		}
	}

	// Wait for interrupt signal.
	<-ctx.Done()
	// Restore signal, allowing "force quit".
	stop()

	mainLogger.Info("Shutting down HTTP server. Press Ctrl+C to force quit.")

	// Note: We expect httpServer.Handler to handle timeouts, thus, we don't
	// need a context value with deadline here.
	//nolint:contextcheck
	err = httpServer.Shutdown(context.Background())
	if err != nil {
		return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	}

	return nil
}
