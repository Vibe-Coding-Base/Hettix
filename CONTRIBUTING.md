# Contribution Guidelines

Thank you for your interest in Hettix. Please read the guidelines below before
contributing.

## Code of conduct

Please read the [code of conduct](CODE_OF_CONDUCT.md) and abide by it in all
project interactions.

## Issues

Use [issues](https://github.com/Vibe-Coding-Base/Hettix/issues) to report bugs
and request features. Prefer labels over category prefixes in titles.

## Pull requests

Hettix is in an early, fast-moving stage — major design decisions are still
being made. Before submitting a pull request that adds a feature or significantly
changes behavior, open an issue first to align on direction. Small fixes and
documentation improvements are welcome directly.

## Development

### Prerequisites

- Go 1.26 or newer
- Node.js and Yarn (for the admin UI)

### Build and run

Hettix is a native desktop app. The admin frontend compiles to a static bundle
embedded into it:

```sh
make build-desktop   # builds the admin UI, then the desktop app -> ./releases
```

Building the desktop app needs the [Wails](https://wails.io) CLI
(`go install github.com/wailsapp/wails/v2/cmd/wails@latest`). To iterate on the
backend only, rebuild the frontend once with `make build-admin` and then use
`go build ./pkg/...`.

### Tests and linting

```sh
go test ./pkg/...
golangci-lint run          # backend
cd admin && yarn lint      # frontend
```

### Conventions

- All identifiers, comments, and log/error messages are written in English.
- Comment sparingly — explain *why*, not *what*. Let the code speak for itself.
- Keep packages cohesive and dependencies explicit; favor small, testable units.
