package main

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"

	"github.com/chromedp/chromedp"
	"github.com/peterbourgon/ff/v3/ffcli"
	"go.uber.org/zap"

	"github.com/Vibe-Coding-Base/Hettix/pkg/adminui"
	"github.com/Vibe-Coding-Base/Hettix/pkg/app"
	"github.com/Vibe-Coding-Base/Hettix/pkg/chrome"
)

var version = "0.2.0"

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

	application, err := app.Build(ctx, app.Config{
		Logger:     cmd.config.logger,
		Version:    version,
		DBPath:     cmd.db,
		CACertFile: cmd.cert,
		CAKeyFile:  cmd.key,
		ProxyURL:   url,
		LLMEnv:     llmSettingsFromEnv(),
	})
	if err != nil {
		cmd.config.logger.Fatal("Failed to build application.", zap.Error(err))
	}
	defer func() { _ = application.Close() }()

	proxyHandler := application.Proxy

	adminHandler, err := adminui.Handler()
	if err != nil {
		cmd.config.logger.Fatal("Failed to construct admin handler.", zap.Error(err))
	}

	adminMux := application.APIMux
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

		proxyHandler.ServeHTTP(w, req)
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
