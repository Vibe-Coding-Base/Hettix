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

See the [modernization roadmap](https://claude.ai/code/artifact/743a7fb6-3033-4e01-b511-c39fbf75bd8d)
for the architecture assessment, the HTTPQL query-language specification, and the
agentic-AI design.

## Features

Available today:

- Machine-in-the-middle (MITM) HTTP proxy with request/response logging and search
- HTTP client for manually crafting, editing, and replaying requests
- Request/response interception for manual review (edit, forward, cancel)
- Scope rules to keep traffic focused
- Web-based admin interface
- Project-based storage on disk

On the roadmap:

- **HTTPQL** — a typed query language for filtering traffic, compiled to SQL
- **AI Assistant** — a multi-provider LLM agent (DeepSeek, OpenAI, Anthropic,
  Gemini, Ollama/local) woven into every flow
- **WebSocket** capture and interception
- **Match & Replace**, **Automate** (fuzzing), **Sitemap**, and **Findings**

## Getting started

Hettix has no binary releases yet; build it from source.

### Prerequisites

- [Go](https://go.dev/dl/) 1.26 or newer
- [Node.js](https://nodejs.org/) and [Yarn](https://classic.yarnpkg.com/) (to build the admin UI)

### Build

The admin frontend is compiled to a static bundle and embedded into the Go
binary, so build the frontend first:

```sh
make build
```

This produces a `hettix` binary in the repository root. To build and install it
into your `$PATH` in one step (after `make build-admin`):

```sh
go install ./cmd/hettix
```

### Run

```sh
./hettix
```

Then open the printed URL (default `http://localhost:8080`) and configure your
browser to use Hettix as its HTTP proxy.

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
