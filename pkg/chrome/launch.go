package chrome

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// ErrBrowserNotFound is returned when no Chromium-based browser can be located.
var ErrBrowserNotFound = errors.New("chrome: no Chromium-based browser found")

// LaunchConfig configures the pre-wired browser launch.
type LaunchConfig struct {
	// ProxyServer is the Hettix proxy address, e.g. "http://localhost:8080".
	ProxyServer string
	// ProxyBypassHosts are additional hosts to reach directly (loopback is
	// always bypassed so the admin UI isn't proxied).
	ProxyBypassHosts []string
	// StartURL is opened on launch; defaults to the proxy server address.
	StartURL string
}

// Launch opens a bundled or installed Chromium-based browser wired to the
// Hettix proxy, in an isolated profile with certificate errors ignored — the
// same model as Burp's embedded browser. The browser runs detached from
// Hettix, so it stays open after this returns.
func Launch(cfg LaunchConfig) error {
	path := FindBrowser()
	if path == "" {
		return ErrBrowserNotFound
	}

	profileDir, err := profileDir()
	if err != nil {
		return err
	}

	bypass := "<-loopback>"
	for _, host := range cfg.ProxyBypassHosts {
		bypass += ";" + host
	}

	startURL := cfg.StartURL
	if startURL == "" {
		startURL = cfg.ProxyServer
	}

	args := []string{
		"--proxy-server=" + cfg.ProxyServer,
		"--proxy-bypass-list=" + bypass,
		"--ignore-certificate-errors",
		"--test-type", // suppresses the ignore-certificate-errors warning bar
		"--no-first-run",
		"--no-default-browser-check",
		"--user-data-dir=" + profileDir,
		startURL,
	}

	cmd := exec.Command(path, args...) //nolint:gosec // path is resolved from known browsers
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("chrome: failed to launch browser: %w", err)
	}

	// Detach: release the process so it outlives this call and Hettix.
	return cmd.Process.Release()
}

// profileDir returns a dedicated, isolated browser profile directory so the
// launched browser never touches the operator's real browser profile.
func profileDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("chrome: failed to resolve home directory: %w", err)
	}

	dir := filepath.Join(home, ".hettix", "browser-profile")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("chrome: failed to create browser profile directory: %w", err)
	}

	return dir, nil
}
