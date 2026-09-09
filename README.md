<h1 align="center">Hettix</h1>

<p align="center"><em>An HTTP toolkit for security research, evolving into an AI-driven pentest platform.</em></p>

<p align="center">
  <a href="https://github.com/Vibe-Coding-Base/Hettix/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-MIT-e53935" alt="License: MIT"></a>
  <img src="https://img.shields.io/badge/status-in%20active%20development-e53935" alt="Status: in active development">
  <img src="https://img.shields.io/badge/go-1.26-e53935" alt="Go 1.26">
</p>

**Hettix** is a machine-in-the-middle HTTP proxy for security research and bug
bounty work. It began as a revival of [Hetty](https://github.com/dstotijn/hetty),
an open-source proxy that had been dormant for years, and is being rebuilt into
a full security-assessment platform in the spirit of tools like
[Caido](https://caido.io) and Burp Suite — with one thing neither of them has at
its core: an **AI agent that can drive the assessment itself**.

## Vision

Hettix is more than a request logger. The goal is a tool where every capability
is exposed both to the operator and to a large language model, so you can hand a
goal to an assistant in plain language and have it plan, query traffic, replay
and mutate requests, and record findings — under scope enforcement and an
approval policy you control.

## Features

- **MITM proxy** — request/response logging with **HTTPQL**, a typed query
  language for filtering traffic that compiles to SQL
- **Interception** — pause, edit, forward or cancel requests and responses for
  manual review
- **Sender** — craft, edit and replay HTTP requests by hand
- **Intruder** — fuzz requests with wordlists loaded from file
- **Match & Replace** — rewrite traffic on the fly with rules
- **WebSocket** capture and interception
- **Sitemap**, **Scope** rules, and **Findings** tracking
- **Workflows** — chain search, fuzzing and findings into repeatable jobs
- **Decoder** — base64 / URL / hex / HTML, plus an in-editor right-click menu to
  decode selected text in place or send it to the Decoder
- **AI Assistant** — an LLM agent woven into every flow, backed by any
  OpenAI-compatible provider (DeepSeek, OpenAI, Ollama/local, and similar)
- **Native desktop app** (Windows) and a **headless server** (all platforms),
  sharing the same backend and admin UI
- **Bundled browser** — launch a pre-wired, pentest-optimised portable Chromium
  straight from the app, Burp-style
- Project-based storage backed by SQLite

## Getting started

Prebuilt binaries are attached to each [release](https://github.com/Vibe-Coding-Base/Hettix/releases):
the headless `hettix` server for Linux/macOS/Windows, the native desktop app
(Windows installer, macOS `.app`, Linux tarball). Or build from source below.

### Prerequisites

- [Go](https://go.dev/dl/) 1.26 or newer
- [Node.js](https://nodejs.org/) and [Yarn](https://classic.yarnpkg.com/) (to build the admin UI)

### Build

The admin frontend is compiled to a static bundle and embedded into the Go
binaries, so `make` builds the frontend first:

```sh
make build          # headless server -> ./hettix (all platforms)
make build-desktop  # native desktop app -> cmd/hettix-desktop/build/bin/
```

`make build` produces a `hettix` binary in the repository root. To install the
server into your `$PATH`:

```sh
make build-admin && go install ./cmd/hettix
```

The desktop app runs on Windows, macOS and Linux and is built with the
[Wails](https://wails.io) CLI (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`);
a plain `go build` of the desktop package will not produce a working window. On
Windows it is frameless with a custom title bar; macOS and Linux use the native
window frame. Linux needs the WebKit dev packages (`libgtk-3-dev`,
`libwebkit2gtk-4.1-dev`).

### Run

Headless server (all platforms):

```sh
./hettix
```

Open the printed URL (default `http://localhost:8080`) and point your browser's
HTTP proxy at it. On the desktop app, the admin UI runs in a native window and
the proxy listens on `:8080` (configurable with `--proxy-addr`).

### Configure the AI assistant

The assistant works with any OpenAI-compatible endpoint. Configure it from the
**Settings** page, or seed it on first run with environment variables:

```sh
export HETTIX_LLM_BASE_URL=https://api.deepseek.com/v1
export HETTIX_LLM_API_KEY=your-api-key   # omit for a local Ollama
export HETTIX_LLM_MODEL=deepseek-chat
```

### Windows installer

`scripts/build-installer.ps1` packages the desktop app together with the bundled
portable browser into an Inno Setup installer (written to `installer/out/`). It
needs the [Wails](https://wails.io) CLI and [Inno Setup](https://jrsoftware.org/isinfo.php).

### Usage

```
$ hettix --help

Usage:
    hettix [flags] [subcommand] [flags]

Runs an HTTP server with (MITM) proxy, GraphQL service, and a web based admin interface.

Options:
    --cert         Path to root CA certificate. Creates file if it doesn't exist. (Default: "~/.hettix/hettix_cert.pem")
    --key          Path to root CA private key. Creates file if it doesn't exist. (Default: "~/.hettix/hettix_key.pem")
    --db           Database file path. Creates file if it doesn't exist. (Default: "~/.hettix/hettix.db")
    --addr         TCP address for HTTP server to listen on, in the form "host:port". (Default: ":8080")
    --chrome       Launch Chrome with proxy settings applied and certificate errors ignored. (Default: false)
    --verbose      Enable verbose logging.
    --json         Encode logs as JSON, instead of pretty/human readable output.
    --version, -v  Output version.
    --help, -h     Output this usage text.

Subcommands:
    - cert  Certificate management

Run `hettix <subcommand> --help` for subcommand specific usage instructions.
```

## Contributing

Contributions are welcome. See the [contribution guidelines](CONTRIBUTING.md)
and the [code of conduct](CODE_OF_CONDUCT.md).

## Acknowledgements

- Hettix builds on [Hetty](https://github.com/dstotijn/hetty) by David Stotijn
  and its contributors. The original project is MIT-licensed, and Hettix
  continues under the same license.
- The font used in the admin interface is [JetBrains Mono](https://www.jetbrains.com/lp/mono/).

## License

[MIT](LICENSE)
