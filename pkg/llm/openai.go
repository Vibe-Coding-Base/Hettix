package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// OpenAIConfig configures an OpenAI-compatible provider. BaseURL points at the
// API root (e.g. https://api.openai.com/v1, https://api.deepseek.com/v1, or
// http://localhost:11434/v1 for Ollama).
type OpenAIConfig struct {
	BaseURL    string
	APIKey     string
	Model      string
	HTTPClient *http.Client
}

type openAIProvider struct {
	name   string
	cfg    OpenAIConfig
	client *http.Client
}

// NewOpenAICompatible returns a Provider backed by an OpenAI-compatible
// /chat/completions endpoint. name is a human label (e.g. "deepseek").
func NewOpenAICompatible(name string, cfg OpenAIConfig) Provider {
	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 120 * time.Second}
	}

	return &openAIProvider{name: name, cfg: cfg, client: client}
}

func (p *openAIProvider) Name() string { return p.name }

func (p *openAIProvider) Chat(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	body, err := json.Marshal(newOpenAIRequest(p.cfg.Model, req))
	if err != nil {
		return ChatResponse{}, fmt.Errorf("llm: failed to encode request: %w", err)
	}

	url := p.cfg.BaseURL + "/chat/completions"

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return ChatResponse{}, fmt.Errorf("llm: failed to build request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if p.cfg.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+p.cfg.APIKey)
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("llm: request to %s failed: %w", p.name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ChatResponse{}, decodeError(p.name, resp)
	}

	var oaiResp openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&oaiResp); err != nil {
		return ChatResponse{}, fmt.Errorf("llm: failed to decode response: %w", err)
	}

	if len(oaiResp.Choices) == 0 {
		return ChatResponse{}, fmt.Errorf("llm: %s returned no choices", p.name)
	}

	return ChatResponse{Message: oaiResp.Choices[0].Message.toMessage()}, nil
}

// ---- wire types ----

type openAIRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	Tools       []openAITool    `json:"tools,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
}

type openAIMessage struct {
	Role       string           `json:"role"`
	Content    string           `json:"content,omitempty"`
	ToolCalls  []openAIToolCall `json:"tool_calls,omitempty"`
	ToolCallID string           `json:"tool_call_id,omitempty"`
	Name       string           `json:"name,omitempty"`
}

type openAIToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type openAITool struct {
	Type     string             `json:"type"`
	Function openAIToolFunction `json:"function"`
}

type openAIToolFunction struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters,omitempty"`
}

type openAIResponse struct {
	Choices []struct {
		Message openAIMessage `json:"message"`
	} `json:"choices"`
}

func newOpenAIRequest(model string, req ChatRequest) openAIRequest {
	out := openAIRequest{Model: model, Temperature: req.Temperature}

	for _, m := range req.Messages {
		msg := openAIMessage{
			Role:       string(m.Role),
			Content:    m.Content,
			ToolCallID: m.ToolCallID,
			Name:       m.Name,
		}

		for _, tc := range m.ToolCalls {
			oc := openAIToolCall{ID: tc.ID, Type: "function"}
			oc.Function.Name = tc.Name
			oc.Function.Arguments = tc.Arguments
			msg.ToolCalls = append(msg.ToolCalls, oc)
		}

		out.Messages = append(out.Messages, msg)
	}

	for _, t := range req.Tools {
		out.Tools = append(out.Tools, openAITool{
			Type:     "function",
			Function: openAIToolFunction(t),
		})
	}

	return out
}

func (m openAIMessage) toMessage() Message {
	msg := Message{
		Role:       Role(m.Role),
		Content:    m.Content,
		ToolCallID: m.ToolCallID,
		Name:       m.Name,
	}

	for _, tc := range m.ToolCalls {
		msg.ToolCalls = append(msg.ToolCalls, ToolCall{
			ID:        tc.ID,
			Name:      tc.Function.Name,
			Arguments: tc.Function.Arguments,
		})
	}

	return msg
}

func decodeError(name string, resp *http.Response) error {
	var body struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&body); err == nil && body.Error.Message != "" {
		return fmt.Errorf("llm: %s returned %d: %s", name, resp.StatusCode, body.Error.Message)
	}

	return fmt.Errorf("llm: %s returned status %d", name, resp.StatusCode)
}
