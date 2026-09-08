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

	// Seed a pentest-friendly profile (safe browsing, password manager and
	// autofill off) before first launch. Ignore errors: the flags below still
	// cover the important cases.
	_ = seedPreferences(profileDir)

	// A pentest-optimized flag set, modelled on Burp's embedded browser: route
	// everything through the proxy, trust the MITM certificate, keep the profile
	// isolated and quiet (no background networking, updates, sync, telemetry or
	// popup blocking that would interfere with testing).
	args := []string{
		"--proxy-server=" + cfg.ProxyServer,
		"--proxy-bypass-list=" + bypass,
		"--ignore-certificate-errors",
		"--test-type", // suppresses the ignore-certificate-errors warning bar
		"--user-data-dir=" + profileDir,
		"--no-first-run",
		"--no-default-browser-check",
		"--no-service-autorun",
		"--disable-background-networking",
		"--disable-component-update",
		"--disable-client-side-phishing-detection",
		"--disable-sync",
		"--disable-default-apps",
		"--disable-popup-blocking",
		"--disable-prompt-on-repost",
		"--disable-breakpad",
		"--disable-backgrounding-occluded-windows",
		"--disable-search-engine-choice-screen",
		"--metrics-recording-only",
		"--safebrowsing-disable-auto-update",
		"--password-store=basic",
		"--use-mock-keychain",
		"--disable-features=Translate,OptimizationHints,MediaRouter,AutofillServerCommunication",
		startURL,
	}

	cmd := exec.Command(path, args...) //nolint:gosec // path is resolved from known browsers
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("chrome: failed to launch browser: %w", err)
	}

	// Detach: release the process so it outlives this call and Hettix.
	return cmd.Process.Release()
}

// pentestPreferences disables browser features that get in the way of testing:
// safe browsing (blocks "malicious" test payloads), the password manager and
// autofill (noise and false state), and translate prompts.
const pentestPreferences = `{
  "safebrowsing": {"enabled": false},
  "credentials_enable_service": false,
  "autofill": {"enabled": false, "profile_enabled": false, "credit_card_enabled": false},
  "translate": {"enabled": false},
  "profile": {
    "password_manager_enabled": false,
    "default_content_setting_values": {"popups": 1}
  }
}`

// seedPreferences writes a pentest-friendly default profile the first time the
// browser is launched. It never overwrites an existing profile so operator
// changes are preserved.
func seedPreferences(profileDir string) error {
	defaultDir := filepath.Join(profileDir, "Default")
	prefs := filepath.Join(defaultDir, "Preferences")

	if _, err := os.Stat(prefs); err == nil {
		return nil // profile already exists; leave it untouched
	}

	if err := os.MkdirAll(defaultDir, 0o700); err != nil {
		return err
	}

	return os.WriteFile(prefs, []byte(pentestPreferences), 0o600)
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
