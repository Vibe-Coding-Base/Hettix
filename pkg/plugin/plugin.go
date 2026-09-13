// Package plugin provides an extension mechanism backed by an embedded
// JavaScript runtime (goja). Each plugin is a single .js file that registers
// metadata and hooks (currently a response observer). Plugins can add
// discovered endpoints to the sitemap and record findings. They live as editable
// files under the plugins directory, can be installed and removed at runtime, and
// their enabled state is persisted per plugin.
package plugin

import (
	"context"
	"net/url"

	"github.com/oklog/ulid"
)

// Meta describes a plugin, as declared by its hettix.plugin({...}) call.
type Meta struct {
	ID             string
	Name           string
	Description    string
	Version        string
	Capabilities   []string
	DefaultEnabled bool
}

// Response is the data a plugin sees for each proxied response.
type Response struct {
	Method       string
	URL          *url.URL
	Host         string
	StatusCode   int
	ContentType  string
	Body         []byte
	RequestLogID *ulid.ULID
}

// Sink is how a plugin records what it finds. Implementations resolve the active
// project themselves.
type Sink interface {
	// AddEndpoint records a discovered endpoint, tagged with the plugin source.
	AddEndpoint(ctx context.Context, host, path, source string) error
	// AddFinding records a finding for the active project.
	AddFinding(ctx context.Context, title, description, severity string, requestLogID *ulid.ULID) error
}
