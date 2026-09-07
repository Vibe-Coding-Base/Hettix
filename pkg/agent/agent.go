package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Vibe-Coding-Base/Hettix/pkg/llm"
)

// Mode controls how much autonomy the agent has.
type Mode string

const (
	// ModeAsk is read-only: mutating tools are neither advertised nor executed.
	ModeAsk Mode = "ask"
	// ModeAssist executes read-only tools freely but requires approval for
	// mutating ones.
	ModeAssist Mode = "assist"
	// ModeAuto executes any tool within the guardrails.
	ModeAuto Mode = "auto"
)

// ErrMaxIterations is returned when the loop does not converge on a final answer
// within the configured budget.
var ErrMaxIterations = errors.New("agent: exceeded maximum iterations")

// Event records a single tool invocation for auditing.
type Event struct {
	Tool    string
	Args    json.RawMessage
	Result  string
	Err     error
	Denied  bool
	Message string
}

// Config configures an Agent.
type Config struct {
	Provider      llm.Provider
	Registry      *Registry
	Mode          Mode
	MaxIterations int
	SystemPrompt  string

	// Approve gates mutating tools in ModeAssist. If nil, mutating tools are
	// denied in ModeAssist.
	Approve func(tool string, args json.RawMessage) bool
	// Guard can veto any tool call regardless of mode (e.g. scope enforcement).
	Guard func(tool string, args json.RawMessage) error
	// OnEvent, if set, receives an Event for every tool call (the audit log).
	OnEvent func(Event)
}

// Agent runs an LLM against a tool registry.
type Agent struct {
	cfg Config
}

// New returns an Agent. Mode defaults to ModeAsk and MaxIterations to 10.
func New(cfg Config) *Agent {
	if cfg.Mode == "" {
		cfg.Mode = ModeAsk
	}

	if cfg.MaxIterations <= 0 {
		cfg.MaxIterations = 10
	}

	return &Agent{cfg: cfg}
}

// Run drives the conversation to a final assistant message, executing tool
// calls along the way. It returns the full message history including the tool
// exchanges and the final answer.
func (a *Agent) Run(ctx context.Context, history []llm.Message) ([]llm.Message, error) {
	msgs := a.withSystemPrompt(history)

	for i := 0; i < a.cfg.MaxIterations; i++ {
		if err := ctx.Err(); err != nil {
			return msgs, err
		}

		resp, err := a.cfg.Provider.Chat(ctx, llm.ChatRequest{
			Messages: msgs,
			Tools:    a.cfg.Registry.llmTools(a.cfg.Mode),
		})
		if err != nil {
			return msgs, fmt.Errorf("agent: provider chat failed: %w", err)
		}

		msgs = append(msgs, resp.Message)

		if len(resp.Message.ToolCalls) == 0 {
			return msgs, nil
		}

		for _, tc := range resp.Message.ToolCalls {
			result := a.executeToolCall(ctx, tc)
			msgs = append(msgs, llm.Message{
				Role:       llm.RoleTool,
				ToolCallID: tc.ID,
				Name:       tc.Name,
				Content:    result,
			})
		}
	}

	return msgs, ErrMaxIterations
}

func (a *Agent) withSystemPrompt(history []llm.Message) []llm.Message {
	if a.cfg.SystemPrompt == "" || (len(history) > 0 && history[0].Role == llm.RoleSystem) {
		return append([]llm.Message(nil), history...)
	}

	return append([]llm.Message{{Role: llm.RoleSystem, Content: a.cfg.SystemPrompt}}, history...)
}

// executeToolCall runs one tool call under the mode and guardrail policy,
// returning the text to feed back to the model. Denials and errors are returned
// as text so the model can adapt rather than aborting the run.
func (a *Agent) executeToolCall(ctx context.Context, tc llm.ToolCall) string {
	args := json.RawMessage(tc.Arguments)

	tool, ok := a.cfg.Registry.get(tc.Name)
	if !ok {
		return a.record(Event{Tool: tc.Name, Args: args, Denied: true, Message: "unknown tool"},
			"Error: unknown tool "+tc.Name)
	}

	if tool.Mutating {
		if a.cfg.Mode == ModeAsk {
			return a.record(Event{Tool: tc.Name, Args: args, Denied: true, Message: "blocked in ask mode"},
				"Denied: this tool changes state and is not permitted in read-only (ask) mode.")
		}

		if a.cfg.Mode == ModeAssist && (a.cfg.Approve == nil || !a.cfg.Approve(tc.Name, args)) {
			return a.record(Event{Tool: tc.Name, Args: args, Denied: true, Message: "not approved"},
				"Denied: this action was not approved by the operator.")
		}
	}

	if a.cfg.Guard != nil {
		if err := a.cfg.Guard(tc.Name, args); err != nil {
			return a.record(Event{Tool: tc.Name, Args: args, Denied: true, Message: err.Error()},
				"Denied by policy: "+err.Error())
		}
	}

	result, err := tool.Handler(ctx, args)
	if err != nil {
		return a.record(Event{Tool: tc.Name, Args: args, Err: err},
			"Error: "+err.Error())
	}

	return a.record(Event{Tool: tc.Name, Args: args, Result: result}, result)
}

func (a *Agent) record(e Event, result string) string {
	if a.cfg.OnEvent != nil {
		a.cfg.OnEvent(e)
	}

	return result
}
