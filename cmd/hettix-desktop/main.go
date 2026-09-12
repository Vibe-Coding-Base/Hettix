// Command hettix-desktop runs Hettix as a native desktop application. It reuses
// the same backend and admin UI as the headless server, but presents them in an
// OS webview window and runs the MITM proxy on its own TCP port.
//
// On Windows the window is frameless and the admin UI paints its own title bar
// and window controls; on macOS and Linux the native window frame is used.
package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	llog "log"
	"net"
	"net/http"
	"os"
	"runtime"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"go.uber.org/zap"

	"github.com/Vibe-Coding-Base/Hettix/pkg/adminui"
	"github.com/Vibe-Coding-Base/Hettix/pkg/app"
	"github.com/Vibe-Coding-Base/Hettix/pkg/llm"
	"github.com/Vibe-Coding-Base/Hettix/pkg/log"
)

var version = "0.3.0"

func main() {
	var (
		proxyAddr string
		dbPath    string
		certPath  string
		keyPath   string
		verbose   bool
		jsonLogs  bool
	)

	fs := flag.NewFlagSet("hettix-desktop", flag.ExitOnError)
	fs.StringVar(&proxyAddr, "proxy-addr", ":8080", "TCP address for the MITM proxy to listen on, in the form \"host:port\".")
	fs.StringVar(&dbPath, "db", "~/.hettix/hettix.db", "Database file path. Creates file if it doesn't exist.")
	fs.StringVar(&certPath, "cert", "~/.hettix/hettix_cert.pem", "Path to root CA certificate. Creates it if absent.")
	fs.StringVar(&keyPath, "key", "~/.hettix/hettix_key.pem", "Path to root CA private key. Creates it if absent.")
	fs.BoolVar(&verbose, "verbose", false, "Enable verbose logging.")
	fs.BoolVar(&jsonLogs, "json", false, "Encode logs as JSON, instead of human readable output.")
	if err := fs.Parse(os.Args[1:]); err != nil {
		llog.Fatalf("Failed to parse command line arguments: %v", err)
	}

	logger, err := log.NewZapLogger(verbose, jsonLogs)
	if err != nil {
		llog.Fatal(err)
	}
	//nolint:errcheck
	defer logger.Sync()

	application, err := app.Build(context.Background(), app.Config{
		Logger:     logger,
		Version:    version,
		DBPath:     dbPath,
		CACertFile: certPath,
		CAKeyFile:  keyPath,
		ProxyURL:   proxyDisplayURL(proxyAddr),
		LLMEnv:     llmSettingsFromEnv(),
	})
	if err != nil {
		logger.Fatal("Failed to build application.", zap.Error(err))
	}
	defer func() { _ = application.Close() }()

	adminHandler, err := adminui.Handler()
	if err != nil {
		logger.Fatal("Failed to construct admin handler.", zap.Error(err))
	}

	// The webview serves the admin UI and GraphQL/export API from a single
	// handler: static assets and SPA routes fall through to the admin handler,
	// while /api/* is handled by the API mux.
	apiMux := application.APIMux
	apiMux.Handle("/", adminHandler)

	desk := &desktop{
		logger:    logger,
		proxyAddr: proxyAddr,
		proxy:     application.Proxy,
	}

	err = wails.Run(&options.App{
		Title:     "Hettix",
		Width:     1400,
		Height:    900,
		MinWidth:  960,
		MinHeight: 600,
		// On Windows the admin UI paints its own title bar and window controls, so
		// the window is frameless. macOS and Linux keep their native frame (and
		// window controls) instead.
		Frameless:        runtime.GOOS == "windows",
		BackgroundColour: &options.RGBA{R: 33, G: 33, B: 33, A: 255},
		AssetServer: &assetserver.Options{
			Handler: apiMux,
		},
		OnStartup:  desk.startup,
		OnShutdown: desk.shutdown,
		// Keep the browser's default context menu off so the app's own
		// Burp-style context menus are the only ones users see.
		EnableDefaultContextMenu: false,
		Windows: &windows.Options{
			Theme: windows.Dark,
		},
	})
	if err != nil {
		logger.Fatal("Failed to run desktop application.", zap.Error(err))
	}
}

// desktop holds the runtime state the Wails lifecycle callbacks need.
type desktop struct {
	logger    *zap.Logger
	proxyAddr string
	proxy     http.Handler

	proxyServer *http.Server
}

// startup starts the MITM proxy listener once the webview is ready.
func (d *desktop) startup(_ context.Context) {
	d.proxyServer = &http.Server{
		Addr:         d.proxyAddr,
		Handler:      d.proxy,
		TLSNextProto: map[string]func(*http.Server, *tls.Conn, http.Handler){}, // Disable HTTP/2
		ErrorLog:     zap.NewStdLog(d.logger.Named("proxy-http")),
	}

	go func() {
		d.logger.Info(fmt.Sprintf("MITM proxy listening on %v", d.proxyAddr))
		if err := d.proxyServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			d.logger.Error("Proxy server closed unexpectedly.", zap.Error(err))
		}
	}()
}

// shutdown stops the proxy listener when the window closes.
func (d *desktop) shutdown(_ context.Context) {
	if d.proxyServer != nil {
		_ = d.proxyServer.Shutdown(context.Background())
	}
}

// proxyDisplayURL renders a user-facing proxy URL, normalising wildcard and
// loopback hosts to localhost.
func proxyDisplayURL(addr string) string {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return "http://localhost:8080"
	}
	if host == "" || host == "0.0.0.0" || host == "127.0.0.1" || host == "::1" {
		host = "localhost"
	}
	return fmt.Sprintf("http://%v:%v", host, port)
}

// llmSettingsFromEnv reads LLM settings from HETTIX_LLM_* environment variables
// to seed stored settings on first run; afterwards the Settings page is the
// source of truth.
func llmSettingsFromEnv() llm.Settings {
	name := os.Getenv("HETTIX_LLM_PROVIDER")
	if name == "" {
		name = "llm"
	}

	return llm.Settings{
		Provider: name,
		BaseURL:  os.Getenv("HETTIX_LLM_BASE_URL"),
		APIKey:   os.Getenv("HETTIX_LLM_API_KEY"),
		Model:    os.Getenv("HETTIX_LLM_MODEL"),
		Enabled:  true,
	}
}
