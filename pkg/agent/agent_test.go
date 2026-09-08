package agent_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/Vibe-Coding-Base/Hettix/pkg/agent"
	"github.com/Vibe-Coding-Base/Hettix/pkg/llm"
)

// scriptedProvider returns queued responses in order and captures the messages
// it was called with.
type scriptedProvider struct {
	responses []llm.ChatResponse
	calls     int
	lastMsgs  []llm.Message
}

func (p *scriptedProvider) Name() string { return "scripted" }

func (p *scriptedProvider) Chat(_ context.Context, req llm.ChatRequest) (llm.ChatResponse, error) {
	p.lastMsgs = req.Messages

	if p.calls >= len(p.responses) {
		// Default: a plain final answer, so runaway scripts still terminate.
		return llm.ChatResponse{Message: llm.Message{Role: llm.RoleAssistant, Content: "done"}}, nil
	}

	resp := p.responses[p.calls]
	p.calls++

	return resp, nil
}

func toolCallResp(id, name, args string) llm.ChatResponse {
	return llm.ChatResponse{Message: llm.Message{
		Role:      llm.RoleAssistant,
		ToolCalls: []llm.ToolCall{{ID: id, Name: name, Arguments: args}},
	}}
}

func finalResp(content string) llm.ChatResponse {
	return llm.ChatResponse{Message: llm.Message{Role: llm.RoleAssistant, Content: content}}
}

func newRegistry(executed *[]string) *agent.Registry {
	reg := agent.NewRegistry()
	reg.Register(agent.Tool{
		Name: "search", Description: "read-only search", Schema: json.RawMessage(`{"type":"object"}`),
		Handler: func(_ context.Context, _ json.RawMessage) (string, error) {
			*executed = append(*executed, "search")
			return "2 results", nil
		},
	})
	reg.Register(agent.Tool{
		Name: "replay", Description: "replay a request", Mutating: true, Schema: json.RawMessage(`{"type":"object"}`),
		Handler: func(_ context.Context, _ json.RawMessage) (string, error) {
			*executed = append(*executed, "replay")
			return "replayed", nil
		},
	})

	return reg
}

func TestRunToolThenFinal(t *testing.T) {
	var executed []string

	a := agent.New(agent.Config{
		Provider: &scriptedProvider{responses: []llm.ChatResponse{
			toolCallResp("c1", "search", `{"query":"req.method eq \"GET\""}`),
			finalResp("Found 2 GET requests."),
		}},
		Registry: newRegistry(&executed),
		Mode:     agent.ModeAsk,
	})

	msgs, err := a.Run(context.Background(), []llm.Message{{Role: llm.RoleUser, Content: "find GETs"}})
	if err != nil {
		t.Fatal(err)
	}

	if len(executed) != 1 || executed[0] != "search" {
		t.Fatalf("expected search to run once, got %v", executed)
	}

	last := msgs[len(msgs)-1]
	if last.Role != llm.RoleAssistant || last.Content != "Found 2 GET requests." {
		t.Fatalf("unexpected final message: %+v", last)
	}

	// The tool result must have been fed back before the final answer.
	var sawToolResult bool
	for _, m := range msgs {
		if m.Role == llm.RoleTool && m.Content == "2 results" {
			sawToolResult = true
		}
	}
	if !sawToolResult {
		t.Fatal("tool result was not appended to the history")
	}
}

func TestAskModeBlocksMutatingTool(t *testing.T) {
	var executed []string

	a := agent.New(agent.Config{
		Provider: &scriptedProvider{responses: []llm.ChatResponse{
			toolCallResp("c1", "replay", `{}`),
			finalResp("Cannot replay in ask mode."),
		}},
		Registry: newRegistry(&executed),
		Mode:     agent.ModeAsk,
	})

	_, err := a.Run(context.Background(), []llm.Message{{Role: llm.RoleUser, Content: "replay it"}})
	if err != nil {
		t.Fatal(err)
	}

	if len(executed) != 0 {
		t.Fatalf("mutating tool must not run in ask mode, got %v", executed)
	}
}

func TestAssistModeRequiresApproval(t *testing.T) {
	var executed []string

	cfg := agent.Config{
		Registry: newRegistry(&executed),
		Mode:     agent.ModeAssist,
		Approve:  func(string, json.RawMessage) bool { return false },
	}
	cfg.Provider = &scriptedProvider{responses: []llm.ChatResponse{
		toolCallResp("c1", "replay", `{}`),
		finalResp("ok"),
	}}

	if _, err := agent.New(cfg).Run(context.Background(), []llm.Message{{Role: llm.RoleUser}}); err != nil {
		t.Fatal(err)
	}
	if len(executed) != 0 {
		t.Fatalf("unapproved mutating tool must not run, got %v", executed)
	}

	// Now approve.
	executed = nil
	cfg.Approve = func(string, json.RawMessage) bool { return true }
	cfg.Provider = &scriptedProvider{responses: []llm.ChatResponse{
		toolCallResp("c1", "replay", `{}`),
		finalResp("ok"),
	}}
	if _, err := agent.New(cfg).Run(context.Background(), []llm.Message{{Role: llm.RoleUser}}); err != nil {
		t.Fatal(err)
	}
	if len(executed) != 1 || executed[0] != "replay" {
		t.Fatalf("approved mutating tool should run, got %v", executed)
	}
}

func TestGuardVetoes(t *testing.T) {
	var executed []string

	a := agent.New(agent.Config{
		Provider: &scriptedProvider{responses: []llm.ChatResponse{
			toolCallResp("c1", "search", `{}`),
			finalResp("ok"),
		}},
		Registry: newRegistry(&executed),
		Mode:     agent.ModeAuto,
		Guard:    func(string, json.RawMessage) error { return errors.New("host out of scope") },
	})

	if _, err := a.Run(context.Background(), []llm.Message{{Role: llm.RoleUser}}); err != nil {
		t.Fatal(err)
	}
	if len(executed) != 0 {
		t.Fatalf("guard should have vetoed the tool, got %v", executed)
	}
}

func TestMaxIterations(t *testing.T) {
	var executed []string

	a := agent.New(agent.Config{
		Provider:      &loopingProvider{},
		Registry:      newRegistry(&executed),
		Mode:          agent.ModeAuto,
		MaxIterations: 3,
	})

	_, err := a.Run(context.Background(), []llm.Message{{Role: llm.RoleUser}})
	if !errors.Is(err, agent.ErrMaxIterations) {
		t.Fatalf("expected ErrMaxIterations, got %v", err)
	}
}

// loopingProvider always asks for a tool call, never converging.
type loopingProvider struct{}

func (loopingProvider) Name() string { return "looping" }

func (loopingProvider) Chat(context.Context, llm.ChatRequest) (llm.ChatResponse, error) {
	return toolCallResp("c", "search", `{}`), nil
}

func TestSystemPromptPrepended(t *testing.T) {
	p := &scriptedProvider{responses: []llm.ChatResponse{finalResp("hi")}}
	a := agent.New(agent.Config{Provider: p, Registry: agent.NewRegistry(), SystemPrompt: "You are Hetty."})

	if _, err := a.Run(context.Background(), []llm.Message{{Role: llm.RoleUser, Content: "hello"}}); err != nil {
		t.Fatal(err)
	}

	if len(p.lastMsgs) == 0 || p.lastMsgs[0].Role != llm.RoleSystem || p.lastMsgs[0].Content != "You are Hetty." {
		t.Fatalf("system prompt not prepended: %+v", p.lastMsgs)
	}
}

// TestAutoPilotMultiStep exercises an autonomous run: a read tool, then a
// mutating tool, then a final summary — all executed without approval in auto
// mode, mirroring how auto-pilot drives a multi-step plan.
func TestAutoPilotMultiStep(t *testing.T) {
	var executed []string

	a := agent.New(agent.Config{
		Provider: &scriptedProvider{responses: []llm.ChatResponse{
			toolCallResp("c1", "search", `{"query":"req.method eq \"POST\""}`),
			toolCallResp("c2", "replay", `{"id":"x"}`),
			finalResp("Investigated and replayed the request."),
		}},
		Registry:      newRegistry(&executed),
		Mode:          agent.ModeAuto,
		MaxIterations: 30,
	})

	msgs, err := a.Run(context.Background(), []llm.Message{{Role: llm.RoleUser, Content: "audit POST requests"}})
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	if len(executed) != 2 || executed[0] != "search" || executed[1] != "replay" {
		t.Fatalf("expected search then replay to execute, got %v", executed)
	}

	if last := msgs[len(msgs)-1]; last.Content != "Investigated and replayed the request." {
		t.Fatalf("unexpected final message: %q", last.Content)
	}
}
