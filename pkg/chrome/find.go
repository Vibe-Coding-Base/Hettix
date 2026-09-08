package chrome

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// FindBrowser locates a Chromium-based browser to drive. It prefers a Chromium
// bundled next to Hettix (see BundledPath), then a browser on PATH, then the
// well-known install locations for the current OS. It returns "" when none is
// found.
func FindBrowser() string {
	if p := BundledPath(); p != "" {
		return p
	}

	for _, name := range pathCandidates() {
		if p, err := exec.LookPath(name); err == nil {
			return p
		}
	}

	for _, p := range wellKnownPaths() {
		if fileExists(p) {
			return p
		}
	}

	return ""
}

// BundledPath returns the path to a Chromium bundled with Hettix: a "browser"
// folder next to the executable (the installed layout) or under the data
// directory (~/.hettix/browser). It returns "" when none is present.
func BundledPath() string {
	var dirs []string

	if exe, err := os.Executable(); err == nil {
		dirs = append(dirs, filepath.Join(filepath.Dir(exe), "browser"))
	}

	if home, err := os.UserHomeDir(); err == nil {
		dirs = append(dirs, filepath.Join(home, ".hettix", "browser"))
	}

	for _, dir := range dirs {
		for _, rel := range bundledRelPaths() {
			p := filepath.Join(dir, rel)
			if fileExists(p) {
				return p
			}
		}
	}

	return ""
}

func pathCandidates() []string {
	switch runtime.GOOS {
	case "windows":
		return []string{"chrome.exe", "chromium.exe", "msedge.exe"}
	case "darwin":
		return []string{"chromium", "google-chrome", "google-chrome-stable"}
	default:
		return []string{"chromium", "chromium-browser", "google-chrome", "google-chrome-stable", "microsoft-edge"}
	}
}

func wellKnownPaths() []string {
	switch runtime.GOOS {
	case "windows":
		var roots []string
		for _, env := range []string{"ProgramFiles", "ProgramFiles(x86)", "LocalAppData"} {
			if v := os.Getenv(env); v != "" {
				roots = append(roots, v)
			}
		}

		suffixes := []string{
			`Google\Chrome\Application\chrome.exe`,
			`Chromium\Application\chrome.exe`,
			`Microsoft\Edge\Application\msedge.exe`,
		}

		var paths []string
		for _, root := range roots {
			for _, s := range suffixes {
				paths = append(paths, filepath.Join(root, s))
			}
		}

		return paths
	case "darwin":
		return []string{
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
		}
	default:
		return []string{
			"/usr/bin/chromium",
			"/usr/bin/chromium-browser",
			"/usr/bin/google-chrome",
			"/usr/bin/google-chrome-stable",
			"/usr/bin/microsoft-edge",
			"/snap/bin/chromium",
		}
	}
}

// bundledRelPaths lists where the browser executable sits inside an extracted
// Chromium snapshot for the current OS.
func bundledRelPaths() []string {
	switch runtime.GOOS {
	case "windows":
		return []string{`chrome.exe`, `chrome-win\chrome.exe`}
	case "darwin":
		return []string{
			"Chromium.app/Contents/MacOS/Chromium",
			"chrome-mac/Chromium.app/Contents/MacOS/Chromium",
		}
	default:
		return []string{"chrome", "chrome-linux/chrome"}
	}
}

func fileExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}
