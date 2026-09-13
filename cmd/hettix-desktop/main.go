// Command hettix-desktop runs Hettix as a native desktop application: it hosts
// the backend and admin UI in an OS webview window and runs the MITM proxy on
// its own TCP port.
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
	"strconv"
	"sync"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"go.uber.org/zap"

	"github.com/Vibe-Coding-Base/Hettix/pkg/adminui"
	"github.com/Vibe-Coding-Base/Hettix/pkg/api"
	"github.com/Vibe-Coding-Base/Hettix/pkg/app"
	"github.com/Vibe-Coding-Base/Hettix/pkg/log"
)

var version = "0.4.0"

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

	// Apply a previously-configured proxy port over the flag default.
	if port, ok, err := application.DB.LoadProxyPort(context.Background()); err == nil && ok {
		host, _, _ := net.SplitHostPort(proxyAddr)
		proxyAddr = net.JoinHostPort(host, strconv.Itoa(port))
		application.Resolver.ProxyURL = proxyDisplayURL(proxyAddr)
	}

	desk := &desktop{
		logger:    logger,
		db:        application.DB,
		resolver:  application.Resolver,
		proxyAddr: proxyAddr,
		proxy:     application.Proxy,
	}
	application.Resolver.ProxyController = desk

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

// desktop holds the runtime state the Wails lifecycle callbacks need, and
// implements api.ProxyController so the GUI can change the proxy port live.
type desktop struct {
	logger   *zap.Logger
	db       proxyPortStore
	resolver *api.Resolver
	proxy    http.Handler

	mu          sync.Mutex
	proxyAddr   string
	proxyServer *http.Server
}

// proxyPortStore persists the chosen proxy port.
type proxyPortStore interface {
	SaveProxyPort(ctx context.Context, port int) error
}

// startup starts the MITM proxy listener once the webview is ready.
func (d *desktop) startup(_ context.Context) {
	ln, err := net.Listen("tcp", d.proxyAddr)
	if err != nil {
		d.logger.Fatal("Failed to bind proxy port.", zap.Error(err))
	}
	d.serve(ln)
}

// serve runs the proxy on ln, replacing any previous server. The caller holds
// no lock; serve takes it.
func (d *desktop) serve(ln net.Listener) {
	server := &http.Server{
		Handler:      d.proxy,
		TLSNextProto: map[string]func(*http.Server, *tls.Conn, http.Handler){}, // Disable HTTP/2
		ErrorLog:     zap.NewStdLog(d.logger.Named("proxy-http")),
	}

	d.mu.Lock()
	d.proxyServer = server
	d.proxyAddr = ln.Addr().String()
	addr := d.proxyAddr
	d.mu.Unlock()

	go func() {
		d.logger.Info(fmt.Sprintf("MITM proxy listening on %v", addr))
		if err := server.Serve(ln); err != nil && err != http.ErrServerClosed {
			d.logger.Error("Proxy server closed unexpectedly.", zap.Error(err))
		}
	}()
}

// shutdown stops the proxy listener when the window closes.
func (d *desktop) shutdown(_ context.Context) {
	d.mu.Lock()
	server := d.proxyServer
	d.mu.Unlock()
	if server != nil {
		_ = server.Shutdown(context.Background())
	}
}

// Port reports the port the proxy currently listens on.
func (d *desktop) Port() int {
	d.mu.Lock()
	addr := d.proxyAddr
	d.mu.Unlock()
	_, portStr, _ := net.SplitHostPort(addr)
	port, _ := strconv.Atoi(portStr)
	return port
}

// SetPort rebinds the proxy to a new port, persists it, and refreshes the
// advertised proxy URL. It binds the new port before dropping the old listener,
// so a failure leaves the current one running.
func (d *desktop) SetPort(ctx context.Context, port int) error {
	d.mu.Lock()
	host, _, _ := net.SplitHostPort(d.proxyAddr)
	old := d.proxyServer
	d.mu.Unlock()

	newAddr := net.JoinHostPort(host, strconv.Itoa(port))
	ln, err := net.Listen("tcp", newAddr)
	if err != nil {
		return fmt.Errorf("cannot listen on port %d: %w", port, err)
	}

	d.serve(ln)

	if old != nil {
		_ = old.Shutdown(context.Background())
	}

	if err := d.db.SaveProxyPort(ctx, port); err != nil {
		d.logger.Error("Failed to persist proxy port.", zap.Error(err))
	}
	d.resolver.ProxyURL = proxyDisplayURL(newAddr)

	return nil
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
