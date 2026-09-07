package llm_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Vibe-Coding-Base/Hettix/pkg/llm"
)

func TestOpenAIChatRequestAndResponse(t *testing.T) {
	var captured map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer sk-test" {
			t.Errorf("missing/incorrect auth header: %q", got)
		}

		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &captured); err != nil {
			t.Fatalf("request body not JSON: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{
			"choices": [{
				"message": {
					"role": "assistant",
					"content": "",
					"tool_calls": [{
						"id": "call_1",
						"type": "function",
						"function": {"name": "search_traffic", "arguments": "{\"query\":\"req.method eq \\\"GET\\\"\"}"}
					}]
				}
			}]
		}`)
	}))
	defer srv.Close()

	provider := llm.NewOpenAICompatible("test", llm.OpenAIConfig{
		BaseURL: srv.URL,
		APIKey:  "sk-test",
		Model:   "test-model",
	})

	resp, err := provider.Chat(context.Background(), llm.ChatRequest{
		Messages: []llm.Message{
			{Role: llm.RoleSystem, Content: "You are a pentest assistant."},
			{Role: llm.RoleUser, Content: "Find GET requests."},
		},
		Tools: []llm.Tool{
			{Name: "search_traffic", Description: "Run an HTTPQL query", Parameters: json.RawMessage(`{"type":"object"}`)},
		},
		Temperature: 0.2,
	})
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}

	if captured["model"] != "test-model" {
		t.Errorf("model = %v, want test-model", captured["model"])
	}
	if msgs, ok := captured["messages"].([]any); !ok || len(msgs) != 2 {
		t.Errorf("expected 2 messages, got %v", captured["messages"])
	}
	if tools, ok := captured["tools"].([]any); !ok || len(tools) != 1 {
		t.Errorf("expected 1 tool, got %v", captured["tools"])
	}

	if len(resp.Message.ToolCalls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(resp.Message.ToolCalls))
	}
	tc := resp.Message.ToolCalls[0]
	if tc.Name != "search_traffic" || tc.ID != "call_1" {
		t.Errorf("unexpected tool call: %+v", tc)
	}
	if tc.Arguments != `{"query":"req.method eq \"GET\""}` {
		t.Errorf("unexpected arguments: %s", tc.Arguments)
	}
}

func TestOpenAIChatError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, `{"error":{"message":"invalid api key"}}`)
	}))
	defer srv.Close()

	provider := llm.NewOpenAICompatible("test", llm.OpenAIConfig{BaseURL: srv.URL, Model: "m"})

	_, err := provider.Chat(context.Background(), llm.ChatRequest{
		Messages: []llm.Message{{Role: llm.RoleUser, Content: "hi"}},
	})
	if err == nil {
		t.Fatal("expected an error for a 401 response")
	}
}
