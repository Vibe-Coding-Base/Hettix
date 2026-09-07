// Package agent orchestrates an LLM over a set of tools: it runs the
// call/observe/iterate loop, enforces the operating mode and guardrails, and
// records every action. The concrete tools (traffic search, replay, scope, ...)
// are registered by the caller so this package stays free of service
// dependencies.
package agent

import (
	"context"
	"encoding/json"

	"github.com/Vibe-Coding-Base/Hettix/pkg/llm"
)

// Tool is a capability the agent can invoke on the model's behalf.
type Tool struct {
	Name        string
	Description string
	// Schema is the JSON Schema for the tool's arguments.
	Schema json.RawMessage
	// Mutating marks tools that change state or send traffic; they are gated by
	// the operating mode.
	Mutating bool
	// Handler runs the tool and returns a textual result for the model.
	Handler func(ctx context.Context, args json.RawMessage) (string, error)
}

// Registry holds the tools available to an agent.
type Registry struct {
	tools map[string]Tool
	order []string
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{tools: map[string]Tool{}}
}

// Register adds a tool, replacing any existing tool with the same name.
func (r *Registry) Register(t Tool) {
	if _, exists := r.tools[t.Name]; !exists {
		r.order = append(r.order, t.Name)
	}

	r.tools[t.Name] = t
}

func (r *Registry) get(name string) (Tool, bool) {
	t, ok := r.tools[name]
	return t, ok
}

// llmTools returns the tools advertised to the model for a mode. Read-only mode
// hides mutating tools so the model never proposes an action it cannot take.
func (r *Registry) llmTools(mode Mode) []llm.Tool {
	var tools []llm.Tool

	for _, name := range r.order {
		t := r.tools[name]
		if mode == ModeAsk && t.Mutating {
			continue
		}

		tools = append(tools, llm.Tool{
			Name:        t.Name,
			Description: t.Description,
			Parameters:  t.Schema,
		})
	}

	return tools
}
