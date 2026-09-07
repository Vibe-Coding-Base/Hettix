// Package llm abstracts chat-completion providers behind a single interface so
// the agent can talk to DeepSeek, OpenAI, Ollama or any other OpenAI-compatible
// endpoint (and, later, native adapters for other APIs) without caring which.
package llm

import (
	"context"
	"encoding/json"
)

// Role identifies who produced a message in a conversation.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// Message is a single turn in a conversation.
type Message struct {
	Role    Role
	Content string

	// ToolCalls is set on an assistant message that asks the caller to run one
	// or more tools.
	ToolCalls []ToolCall

	// ToolCallID and Name are set on a tool-result message (Role == RoleTool),
	// linking the result back to the assistant's ToolCall.
	ToolCallID string
	Name       string
}

// ToolCall is a request from the model to invoke a named tool with JSON
// arguments.
type ToolCall struct {
	ID        string
	Name      string
	Arguments string
}

// Tool describes a function the model may call. Parameters is a JSON Schema
// object describing the arguments.
type Tool struct {
	Name        string
	Description string
	Parameters  json.RawMessage
}

// ChatRequest is a single completion request.
type ChatRequest struct {
	Messages    []Message
	Tools       []Tool
	Temperature float64
}

// ChatResponse is the model's reply. Message may carry content, tool calls, or
// both.
type ChatResponse struct {
	Message Message
}

// Provider is a chat-completion backend.
type Provider interface {
	// Name identifies the provider for logging and UI.
	Name() string
	// Chat performs one completion round.
	Chat(ctx context.Context, req ChatRequest) (ChatResponse, error)
}
