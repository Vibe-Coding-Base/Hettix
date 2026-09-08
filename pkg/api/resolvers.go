package api

//go:generate go run github.com/99designs/gqlgen

import (
	"bytes"
	"context"
	crand "crypto/rand"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/oklog/ulid"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"github.com/Vibe-Coding-Base/Hettix/pkg/agent"
	"github.com/Vibe-Coding-Base/Hettix/pkg/aitools"
	"github.com/Vibe-Coding-Base/Hettix/pkg/httpql"
	"github.com/Vibe-Coding-Base/Hettix/pkg/llm"
	"github.com/Vibe-Coding-Base/Hettix/pkg/matchreplace"
	"github.com/Vibe-Coding-Base/Hettix/pkg/proj"
	"github.com/Vibe-Coding-Base/Hettix/pkg/proxy"
	"github.com/Vibe-Coding-Base/Hettix/pkg/proxy/intercept"
	"github.com/Vibe-Coding-Base/Hettix/pkg/reqlog"
	"github.com/Vibe-Coding-Base/Hettix/pkg/scope"
	"github.com/Vibe-Coding-Base/Hettix/pkg/sender"
	"github.com/Vibe-Coding-Base/Hettix/pkg/wslog"
	"github.com/Vibe-Coding-Base/Hettix/pkg/wsproxy"
)

var httpProtocolMap = map[string]HTTPProtocol{
	sender.HTTPProto10: HTTPProtocolHTTP10,
	sender.HTTPProto11: HTTPProtocolHTTP11,
	sender.HTTPProto20: HTTPProtocolHTTP20,
}

var revHTTPProtocolMap = map[HTTPProtocol]string{
	HTTPProtocolHTTP10: sender.HTTPProto10,
	HTTPProtocolHTTP11: sender.HTTPProto11,
	HTTPProtocolHTTP20: sender.HTTPProto20,
}

type Resolver struct {
	ProjectService    *proj.Service
	RequestLogService *reqlog.Service
	InterceptService  *intercept.Service
	SenderService     *sender.Service
	WebSocketService  *wslog.Service
	LLMProvider       llm.Provider
}

type (
	queryResolver    struct{ *Resolver }
	mutationResolver struct{ *Resolver }
)

func (r *Resolver) Query() QueryResolver       { return &queryResolver{r} }
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

func (r *queryResolver) HTTPRequestLogs(ctx context.Context, offset *int, limit *int) ([]HTTPRequestLog, error) {
	reqs, err := r.RequestLogService.FindRequests(ctx, offset, limit)
	if errors.Is(err, proj.ErrNoProject) {
		return nil, noActiveProjectErr(ctx)
	} else if err != nil {
		return nil, fmt.Errorf("could not query repository for requests: %w", err)
	}

	logs := make([]HTTPRequestLog, len(reqs))

	for i, req := range reqs {
		req, err := parseRequestLog(req)
		if err != nil {
			return nil, err
		}

		logs[i] = req
	}

	return logs, nil
}

func (r *queryResolver) HTTPRequestLog(ctx context.Context, id ulid.ULID) (*HTTPRequestLog, error) {
	log, err := r.RequestLogService.FindRequestLogByID(ctx, id)
	if errors.Is(err, reqlog.ErrRequestNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("could not get request by ID: %w", err)
	}

	req, err := parseRequestLog(log)
	if err != nil {
		return nil, err
	}

	return &req, nil
}

func parseRequestLog(reqLog reqlog.RequestLog) (HTTPRequestLog, error) {
	method := HTTPMethod(reqLog.Method)
	if method != "" && !method.IsValid() {
		return HTTPRequestLog{}, fmt.Errorf("request has invalid method: %v", method)
	}

	log := HTTPRequestLog{
		ID:        reqLog.ID,
		Proto:     reqLog.Proto,
		Method:    method,
		Timestamp: ulid.Time(reqLog.ID.Time()),
	}

	if reqLog.URL != nil {
		log.URL = reqLog.URL.String()
	}

	if len(reqLog.Body) > 0 {
		bodyStr := string(reqLog.Body)
		log.Body = &bodyStr
	}

	if reqLog.Header != nil {
		log.Headers = make([]HTTPHeader, 0)

		for key, values := range reqLog.Header {
			for _, value := range values {
				log.Headers = append(log.Headers, HTTPHeader{
					Key:   key,
					Value: value,
				})
			}
		}

		sort.Sort(HTTPHeaders(log.Headers))
	}

	if reqLog.Response != nil {
		resLog, err := parseResponseLog(*reqLog.Response)
		if err != nil {
			return HTTPRequestLog{}, err
		}

		resLog.ID = reqLog.ID

		log.Response = &resLog
	}

	return log, nil
}

func parseResponseLog(resLog reqlog.ResponseLog) (HTTPResponseLog, error) {
	proto := httpProtocolMap[resLog.Proto]
	if !proto.IsValid() {
		return HTTPResponseLog{}, fmt.Errorf("sender response has invalid protocol: %v", resLog.Proto)
	}

	httpResLog := HTTPResponseLog{
		Proto:      proto,
		StatusCode: resLog.StatusCode,
	}
	statusReasonSubs := strings.SplitN(resLog.Status, " ", 2)

	if len(statusReasonSubs) == 2 {
		httpResLog.StatusReason = statusReasonSubs[1]
	}

	if len(resLog.Body) > 0 {
		bodyStr := string(resLog.Body)
		httpResLog.Body = &bodyStr
	}

	if resLog.Header != nil {
		httpResLog.Headers = make([]HTTPHeader, 0)

		for key, values := range resLog.Header {
			for _, value := range values {
				httpResLog.Headers = append(httpResLog.Headers, HTTPHeader{
					Key:   key,
					Value: value,
				})
			}
		}

		sort.Sort(HTTPHeaders(httpResLog.Headers))
	}

	return httpResLog, nil
}

func (r *mutationResolver) CreateProject(ctx context.Context, name string) (*Project, error) {
	p, err := r.ProjectService.CreateProject(ctx, name)
	if errors.Is(err, proj.ErrInvalidName) {
		return nil, gqlerror.Errorf("Project name must only contain alphanumeric or space chars.")
	} else if err != nil {
		return nil, fmt.Errorf("could not open project: %w", err)
	}

	project := parseProject(r.ProjectService, p)

	return &project, nil
}

func (r *mutationResolver) OpenProject(ctx context.Context, id ulid.ULID) (*Project, error) {
	p, err := r.ProjectService.OpenProject(ctx, id)
	if errors.Is(err, proj.ErrInvalidName) {
		return nil, gqlerror.Errorf("Project name must only contain alphanumeric or space chars.")
	} else if err != nil {
		return nil, fmt.Errorf("could not open project: %w", err)
	}

	project := parseProject(r.ProjectService, p)

	return &project, nil
}

func (r *queryResolver) ActiveProject(ctx context.Context) (*Project, error) {
	p, err := r.ProjectService.ActiveProject(ctx)
	if errors.Is(err, proj.ErrNoProject) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("could not open project: %w", err)
	}

	project := parseProject(r.ProjectService, p)

	return &project, nil
}

func (r *queryResolver) Projects(ctx context.Context) ([]Project, error) {
	p, err := r.ProjectService.Projects(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get projects: %w", err)
	}

	projects := make([]Project, len(p))
	for i, proj := range p {
		projects[i] = parseProject(r.ProjectService, proj)
	}

	return projects, nil
}

func (r *queryResolver) Scope(ctx context.Context) ([]ScopeRule, error) {
	rules := r.ProjectService.Scope().Rules()
	return scopeToScopeRules(rules), nil
}

func regexpToStringPtr(r *regexp.Regexp) *string {
	if r == nil {
		return nil
	}

	s := r.String()

	return &s
}

func (r *mutationResolver) CloseProject(ctx context.Context) (*CloseProjectResult, error) {
	if err := r.ProjectService.CloseProject(); err != nil {
		return nil, fmt.Errorf("could not close project: %w", err)
	}

	return &CloseProjectResult{true}, nil
}

func (r *mutationResolver) DeleteProject(ctx context.Context, id ulid.ULID) (*DeleteProjectResult, error) {
	if err := r.ProjectService.DeleteProject(ctx, id); err != nil {
		return nil, fmt.Errorf("could not delete project: %w", err)
	}

	return &DeleteProjectResult{
		Success: true,
	}, nil
}

func (r *mutationResolver) ClearHTTPRequestLog(ctx context.Context) (*ClearHTTPRequestLogResult, error) {
	project, err := r.ProjectService.ActiveProject(ctx)
	if errors.Is(err, proj.ErrNoProject) {
		return nil, noActiveProjectErr(ctx)
	} else if err != nil {
		return nil, fmt.Errorf("could not get active project: %w", err)
	}

	if err := r.RequestLogService.ClearRequests(ctx, project.ID); err != nil {
		return nil, fmt.Errorf("could not clear request log: %w", err)
	}

	return &ClearHTTPRequestLogResult{true}, nil
}

func (r *mutationResolver) SetScope(ctx context.Context, input []ScopeRuleInput) ([]ScopeRule, error) {
	rules := make([]scope.Rule, len(input))

	for i, rule := range input {
		u, err := stringPtrToRegexp(rule.URL)
		if err != nil {
			return nil, fmt.Errorf("invalid URL in scope rule: %w", err)
		}

		var headerKey, headerValue *regexp.Regexp

		if rule.Header != nil {
			headerKey, err = stringPtrToRegexp(rule.Header.Key)
			if err != nil {
				return nil, fmt.Errorf("invalid header key in scope rule: %w", err)
			}

			headerValue, err = stringPtrToRegexp(rule.Header.Value)
			if err != nil {
				return nil, fmt.Errorf("invalid header value in scope rule: %w", err)
			}
		}

		body, err := stringPtrToRegexp(rule.Body)
		if err != nil {
			return nil, fmt.Errorf("invalid body in scope rule: %w", err)
		}

		rules[i] = scope.Rule{
			URL: u,
			Header: scope.Header{
				Key:   headerKey,
				Value: headerValue,
			},
			Body: body,
		}
	}

	err := r.ProjectService.SetScopeRules(ctx, rules)
	if err != nil {
		return nil, fmt.Errorf("could not set scope rules: %w", err)
	}

	return scopeToScopeRules(rules), nil
}

func (r *queryResolver) HTTPRequestLogFilter(ctx context.Context) (*HTTPRequestLogFilter, error) {
	return findReqFilterToHTTPReqLogFilter(r.RequestLogService.FindReqsFilter()), nil
}

func (r *mutationResolver) SetHTTPRequestLogFilter(
	ctx context.Context,
	input *HTTPRequestLogFilterInput,
) (*HTTPRequestLogFilter, error) {
	filter, err := findRequestsFilterFromInput(input)
	if err != nil {
		return nil, fmt.Errorf("could not parse request log filter: %w", err)
	}

	err = r.ProjectService.SetRequestLogFindFilter(ctx, filter)
	if errors.Is(err, proj.ErrNoProject) {
		return nil, noActiveProjectErr(ctx)
	} else if err != nil {
		return nil, fmt.Errorf("could not set request log filter: %w", err)
	}

	return findReqFilterToHTTPReqLogFilter(filter), nil
}

func (r *queryResolver) SenderRequest(ctx context.Context, id ulid.ULID) (*SenderRequest, error) {
	senderReq, err := r.SenderService.FindRequestByID(ctx, id)
	if errors.Is(err, sender.ErrRequestNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("could not get request by ID: %w", err)
	}

	req, err := parseSenderRequest(senderReq)
	if err != nil {
		return nil, err
	}

	return &req, nil
}

func (r *queryResolver) SenderRequests(ctx context.Context, offset *int, limit *int) ([]SenderRequest, error) {
	reqs, err := r.SenderService.FindRequests(ctx, offset, limit)
	if errors.Is(err, proj.ErrNoProject) {
		return nil, noActiveProjectErr(ctx)
	} else if err != nil {
		return nil, fmt.Errorf("failed to find sender requests: %w", err)
	}

	senderReqs := make([]SenderRequest, len(reqs))

	for i, req := range reqs {
		req, err := parseSenderRequest(req)
		if err != nil {
			return nil, err
		}

		senderReqs[i] = req
	}

	return senderReqs, nil
}

func (r *mutationResolver) SetSenderRequestFilter(
	ctx context.Context,
	input *SenderRequestFilterInput,
) (*SenderRequestFilter, error) {
	filter, err := findSenderRequestsFilterFromInput(input)
	if err != nil {
		return nil, fmt.Errorf("could not parse request log filter: %w", err)
	}

	err = r.ProjectService.SetSenderRequestFindFilter(ctx, filter)
	if errors.Is(err, proj.ErrNoProject) {
		return nil, noActiveProjectErr(ctx)
	} else if err != nil {
		return nil, fmt.Errorf("could not set request log filter: %w", err)
	}

	return findReqFilterToSenderReqFilter(filter), nil
}

func (r *mutationResolver) CreateOrUpdateSenderRequest(
	ctx context.Context,
	input SenderRequestInput,
) (*SenderRequest, error) {
	req := sender.Request{
		URL:    input.URL,
		Header: make(http.Header),
	}

	if input.ID != nil {
		req.ID = *input.ID
	}

	if input.Method != nil {
		req.Method = input.Method.String()
	}

	if input.Proto != nil {
		req.Proto = revHTTPProtocolMap[*input.Proto]
	}

	for _, header := range input.Headers {
		req.Header.Add(header.Key, header.Value)
	}

	if input.Body != nil {
		req.Body = []byte(*input.Body)
	}

	req, err := r.SenderService.CreateOrUpdateRequest(ctx, req)
	if errors.Is(err, proj.ErrNoProject) {
		return nil, noActiveProjectErr(ctx)
	} else if err != nil {
		return nil, fmt.Errorf("could not create sender request: %w", err)
	}

	senderReq, err := parseSenderRequest(req)
	if err != nil {
		return nil, err
	}

	return &senderReq, nil
}

func (r *mutationResolver) CreateSenderRequestFromHTTPRequestLog(
	ctx context.Context,
	id ulid.ULID,
) (*SenderRequest, error) {
	req, err := r.SenderService.CloneFromRequestLog(ctx, id)
	if errors.Is(err, proj.ErrNoProject) {
		return nil, noActiveProjectErr(ctx)
	} else if err != nil {
		return nil, fmt.Errorf("could not create sender request from http request log: %w", err)
	}

	senderReq, err := parseSenderRequest(req)
	if err != nil {
		return nil, err
	}

	return &senderReq, nil
}

func (r *mutationResolver) SendRequest(ctx context.Context, id ulid.ULID) (*SenderRequest, error) {
	// Use new context, because we don't want to risk interrupting sending the request
	// or the subsequent storing of the response, e.g. if ctx gets cancelled or
	// times out.
	ctx2 := context.Background()

	var sendErr *sender.SendError

	//nolint:contextcheck
	req, err := r.SenderService.SendRequest(ctx2, id)

	switch {
	case errors.Is(err, proj.ErrNoProject):
		return nil, noActiveProjectErr(ctx)
	case errors.As(err, &sendErr):
		return nil, &gqlerror.Error{
			Path:    graphql.GetPath(ctx),
			Message: fmt.Sprintf("Sending request failed: %v", sendErr.Unwrap()),
			Extensions: map[string]interface{}{
				"code": "send_request_failed",
			},
		}
	case err != nil:
		return nil, fmt.Errorf("could not send request: %w", err)
	}

	senderReq, err := parseSenderRequest(req)
	if err != nil {
		return nil, err
	}

	return &senderReq, nil
}

func (r *mutationResolver) DeleteSenderRequests(ctx context.Context) (*DeleteSenderRequestsResult, error) {
	project, err := r.ProjectService.ActiveProject(ctx)
	if errors.Is(err, proj.ErrNoProject) {
		return nil, noActiveProjectErr(ctx)
	} else if err != nil {
		return nil, fmt.Errorf("could not get active project: %w", err)
	}

	if err := r.SenderService.DeleteRequests(ctx, project.ID); err != nil {
		return nil, fmt.Errorf("could not clear request log: %w", err)
	}

	return &DeleteSenderRequestsResult{true}, nil
}

func (r *queryResolver) InterceptedRequests(ctx context.Context) (httpReqs []HTTPRequest, err error) {
	items := r.InterceptService.Items()

	for _, item := range items {
		req, err := parseInterceptItem(item)
		if err != nil {
			return nil, err
		}

		httpReqs = append(httpReqs, req)
	}

	return httpReqs, nil
}

func (r *queryResolver) InterceptedRequest(ctx context.Context, id ulid.ULID) (*HTTPRequest, error) {
	item, err := r.InterceptService.ItemByID(id)
	if errors.Is(err, intercept.ErrRequestNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("could not get request by ID: %w", err)
	}

	req, err := parseInterceptItem(item)
	if err != nil {
		return nil, err
	}

	return &req, nil
}

func (r *mutationResolver) ModifyRequest(ctx context.Context, input ModifyRequestInput) (*ModifyRequestResult, error) {
	body := ""
	if input.Body != nil {
		body = *input.Body
	}

	//nolint:noctx
	req, err := http.NewRequest(input.Method.String(), input.URL.String(), strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to construct HTTP request: %w", err)
	}

	for _, header := range input.Headers {
		req.Header.Add(header.Key, header.Value)
	}

	err = r.InterceptService.ModifyRequest(input.ID, req, input.ModifyResponse)
	if err != nil {
		return nil, fmt.Errorf("could not modify http request: %w", err)
	}

	return &ModifyRequestResult{Success: true}, nil
}

func (r *mutationResolver) CancelRequest(ctx context.Context, id ulid.ULID) (*CancelRequestResult, error) {
	err := r.InterceptService.CancelRequest(id)
	if err != nil {
		return nil, fmt.Errorf("could not cancel http request: %w", err)
	}

	return &CancelRequestResult{Success: true}, nil
}

func (r *mutationResolver) ModifyResponse(
	ctx context.Context,
	input ModifyResponseInput,
) (*ModifyResponseResult, error) {
	res := &http.Response{
		Header:     make(http.Header),
		Status:     fmt.Sprintf("%v %v", input.StatusCode, input.StatusReason),
		StatusCode: input.StatusCode,
		Proto:      revHTTPProtocolMap[input.Proto],
	}

	var ok bool
	if res.ProtoMajor, res.ProtoMinor, ok = http.ParseHTTPVersion(res.Proto); !ok {
		return nil, fmt.Errorf("malformed HTTP version: %q", res.Proto)
	}

	var body string
	if input.Body != nil {
		body = *input.Body
	}

	res.Body = io.NopCloser(strings.NewReader(body))

	for _, header := range input.Headers {
		res.Header.Add(header.Key, header.Value)
	}

	err := r.InterceptService.ModifyResponse(input.RequestID, res)
	if err != nil {
		return nil, fmt.Errorf("could not modify http request: %w", err)
	}

	return &ModifyResponseResult{Success: true}, nil
}

func (r *mutationResolver) CancelResponse(ctx context.Context, requestID ulid.ULID) (*CancelResponseResult, error) {
	err := r.InterceptService.CancelResponse(requestID)
	if err != nil {
		return nil, fmt.Errorf("could not cancel http response: %w", err)
	}

	return &CancelResponseResult{Success: true}, nil
}

func (r *mutationResolver) UpdateInterceptSettings(
	ctx context.Context,
	input UpdateInterceptSettingsInput,
) (*InterceptSettings, error) {
	settings := intercept.Settings{
		RequestsEnabled:  input.RequestsEnabled,
		ResponsesEnabled: input.ResponsesEnabled,
	}

	if input.RequestFilter != nil && *input.RequestFilter != "" {
		expr, err := httpql.Parse(*input.RequestFilter)
		if err != nil {
			return nil, fmt.Errorf("could not parse request filter: %w", err)
		}

		settings.RequestFilter = expr
	}

	if input.ResponseFilter != nil && *input.ResponseFilter != "" {
		expr, err := httpql.Parse(*input.ResponseFilter)
		if err != nil {
			return nil, fmt.Errorf("could not parse response filter: %w", err)
		}

		settings.ResponseFilter = expr
	}

	err := r.ProjectService.UpdateInterceptSettings(ctx, settings)
	if errors.Is(err, proj.ErrNoProject) {
		return nil, noActiveProjectErr(ctx)
	} else if err != nil {
		return nil, fmt.Errorf("could not update intercept settings: %w", err)
	}

	updated := &InterceptSettings{
		RequestsEnabled:  settings.RequestsEnabled,
		ResponsesEnabled: settings.ResponsesEnabled,
	}

	if settings.RequestFilter != nil {
		reqFilter := settings.RequestFilter.String()
		updated.RequestFilter = &reqFilter
	}

	if settings.ResponseFilter != nil {
		resFilter := settings.ResponseFilter.String()
		updated.ResponseFilter = &resFilter
	}

	return updated, nil
}

func (r *mutationResolver) RunAgent(ctx context.Context, input RunAgentInput) (*AgentReply, error) {
	if r.LLMProvider == nil {
		return nil, errors.New("AI assistant is not configured; set the HETTIX_LLM_* environment variables")
	}

	mode := agent.ModeAsk
	if input.Mode != nil {
		switch *input.Mode {
		case AgentModeAssist:
			mode = agent.ModeAssist
		case AgentModeAuto:
			mode = agent.ModeAuto
		case AgentModeAsk:
			mode = agent.ModeAsk
		}
	}

	var actions []AgentAction

	ag := agent.New(agent.Config{
		Provider:     r.LLMProvider,
		Registry:     aitools.NewRegistry(r.RequestLogService, r.SenderService, r.ProjectService),
		Mode:         mode,
		SystemPrompt: aitools.SystemPrompt,
		OnEvent: func(e agent.Event) {
			output := e.Result
			switch {
			case e.Err != nil:
				output = e.Err.Error()
			case e.Denied:
				output = e.Message
			}

			actions = append(actions, AgentAction{
				Tool:   e.Tool,
				Input:  string(e.Args),
				Output: output,
				Denied: e.Denied,
			})
		},
	})

	msgs, err := ag.Run(ctx, []llm.Message{{Role: llm.RoleUser, Content: input.Message}})
	if err != nil {
		return nil, fmt.Errorf("agent run failed: %w", err)
	}

	var reply string
	if len(msgs) > 0 {
		reply = msgs[len(msgs)-1].Content
	}

	return &AgentReply{Reply: reply, Actions: actions}, nil
}

func (r *queryResolver) MatchReplaceRules(ctx context.Context) ([]MatchReplaceRule, error) {
	rules, err := r.ProjectService.MatchReplaceRules(ctx)
	if errors.Is(err, proj.ErrNoProject) {
		return nil, noActiveProjectErr(ctx)
	} else if err != nil {
		return nil, fmt.Errorf("could not get match & replace rules: %w", err)
	}

	return matchReplaceRulesToGraphQL(rules), nil
}

func (r *queryResolver) WebSocketConnections(ctx context.Context) ([]WebSocketConnection, error) {
	conns, err := r.WebSocketService.Connections(ctx)
	if errors.Is(err, wslog.ErrProjectIDMustBeSet) {
		return nil, noActiveProjectErr(ctx)
	} else if err != nil {
		return nil, fmt.Errorf("could not get websocket connections: %w", err)
	}

	out := make([]WebSocketConnection, len(conns))
	for i, conn := range conns {
		out[i] = webSocketConnectionToGraphQL(conn)
	}

	return out, nil
}

func (r *queryResolver) WebSocketConnection(ctx context.Context, id ulid.ULID) (*WebSocketConnection, error) {
	conn, err := r.WebSocketService.ConnectionByID(ctx, id)
	switch {
	case errors.Is(err, wslog.ErrConnectionNotFound):
		return nil, nil
	case errors.Is(err, wslog.ErrProjectIDMustBeSet):
		return nil, noActiveProjectErr(ctx)
	case err != nil:
		return nil, fmt.Errorf("could not get websocket connection: %w", err)
	}

	out := webSocketConnectionToGraphQL(conn)

	return &out, nil
}

func (r *queryResolver) WebSocketMessages(ctx context.Context, connectionID ulid.ULID) ([]WebSocketMessage, error) {
	msgs, err := r.WebSocketService.Messages(ctx, connectionID)
	switch {
	case errors.Is(err, wslog.ErrConnectionNotFound):
		return nil, nil
	case errors.Is(err, wslog.ErrProjectIDMustBeSet):
		return nil, noActiveProjectErr(ctx)
	case err != nil:
		return nil, fmt.Errorf("could not get websocket messages: %w", err)
	}

	out := make([]WebSocketMessage, len(msgs))
	for i, msg := range msgs {
		out[i] = WebSocketMessage{
			Direction: webSocketDirectionToGraphQL(msg.Direction),
			Opcode:    msg.Opcode,
			Payload:   string(msg.Payload),
			Timestamp: msg.CreatedAt,
		}
	}

	return out, nil
}

func webSocketConnectionToGraphQL(conn wslog.Connection) WebSocketConnection {
	return WebSocketConnection{
		ID:           conn.ID,
		URL:          conn.URL,
		Host:         conn.Host,
		Path:         conn.Path,
		Timestamp:    conn.CreatedAt,
		ClosedAt:     conn.ClosedAt,
		MessageCount: conn.MessageCount,
	}
}

func webSocketDirectionToGraphQL(d wsproxy.Direction) WebSocketDirection {
	if d == wsproxy.ServerToClient {
		return WebSocketDirectionServerToClient
	}

	return WebSocketDirectionClientToServer
}

func (r *mutationResolver) SetMatchReplaceRules(ctx context.Context, input []MatchReplaceRuleInput) ([]MatchReplaceRule, error) {
	rules := make([]matchreplace.Rule, 0, len(input))

	for _, in := range input {
		id, err := ulid.New(ulid.Now(), crand.Reader)
		if err != nil {
			return nil, fmt.Errorf("could not generate rule id: %w", err)
		}

		rule := matchreplace.Rule{
			ID:      id,
			Name:    in.Name,
			Enabled: in.Enabled,
			Phase:   matchReplacePhaseFromGraphQL(in.Phase),
		}

		if in.Condition != nil && *in.Condition != "" {
			expr, err := httpql.Parse(*in.Condition)
			if err != nil {
				return nil, fmt.Errorf("invalid condition %q: %w", *in.Condition, err)
			}
			rule.Condition = expr
		}

		rule.HeaderName = strPtrValue(in.HeaderName)
		rule.HeaderValue = strPtrValue(in.HeaderValue)
		rule.BodyMatcher = strPtrValue(in.BodyMatcher)
		rule.BodyReplacement = strPtrValue(in.BodyReplacement)
		if in.RemoveHeader != nil {
			rule.RemoveHeader = *in.RemoveHeader
		}

		rules = append(rules, rule)
	}

	if err := r.ProjectService.SetMatchReplaceRules(ctx, rules); err != nil {
		if errors.Is(err, proj.ErrNoProject) {
			return nil, noActiveProjectErr(ctx)
		}
		return nil, fmt.Errorf("could not set match & replace rules: %w", err)
	}

	return matchReplaceRulesToGraphQL(rules), nil
}

func matchReplaceRulesToGraphQL(rules []matchreplace.Rule) []MatchReplaceRule {
	out := make([]MatchReplaceRule, len(rules))

	for i, rule := range rules {
		m := MatchReplaceRule{
			ID:           rule.ID,
			Name:         rule.Name,
			Enabled:      rule.Enabled,
			Phase:        matchReplacePhaseToGraphQL(rule.Phase),
			RemoveHeader: rule.RemoveHeader,
		}

		if rule.Condition != nil {
			s := rule.Condition.String()
			m.Condition = &s
		}
		m.HeaderName = nonEmptyPtr(rule.HeaderName)
		m.HeaderValue = nonEmptyPtr(rule.HeaderValue)
		m.BodyMatcher = nonEmptyPtr(rule.BodyMatcher)
		m.BodyReplacement = nonEmptyPtr(rule.BodyReplacement)

		out[i] = m
	}

	return out
}

func matchReplacePhaseFromGraphQL(p MatchReplacePhase) matchreplace.Phase {
	if p == MatchReplacePhaseResponse {
		return matchreplace.PhaseResponse
	}

	return matchreplace.PhaseRequest
}

func matchReplacePhaseToGraphQL(p matchreplace.Phase) MatchReplacePhase {
	if p == matchreplace.PhaseResponse {
		return MatchReplacePhaseResponse
	}

	return MatchReplacePhaseRequest
}

func strPtrValue(s *string) string {
	if s == nil {
		return ""
	}

	return *s
}

func nonEmptyPtr(s string) *string {
	if s == "" {
		return nil
	}

	return &s
}

func parseSenderRequest(req sender.Request) (SenderRequest, error) {
	method := HTTPMethod(req.Method)
	if method != "" && !method.IsValid() {
		return SenderRequest{}, fmt.Errorf("sender request has invalid method: %v", method)
	}

	reqProto := httpProtocolMap[req.Proto]
	if !reqProto.IsValid() {
		return SenderRequest{}, fmt.Errorf("sender request has invalid protocol: %v", req.Proto)
	}

	senderReq := SenderRequest{
		ID:        req.ID,
		URL:       req.URL,
		Method:    method,
		Proto:     HTTPProtocol(req.Proto),
		Timestamp: ulid.Time(req.ID.Time()),
	}

	if req.SourceRequestLogID.Compare(ulid.ULID{}) != 0 {
		senderReq.SourceRequestLogID = &req.SourceRequestLogID
	}

	if req.Header != nil {
		senderReq.Headers = make([]HTTPHeader, 0)

		for key, values := range req.Header {
			for _, value := range values {
				senderReq.Headers = append(senderReq.Headers, HTTPHeader{
					Key:   key,
					Value: value,
				})
			}
		}

		sort.Sort(HTTPHeaders(senderReq.Headers))
	}

	if len(req.Body) > 0 {
		bodyStr := string(req.Body)
		senderReq.Body = &bodyStr
	}

	if req.Response != nil {
		resLog, err := parseResponseLog(*req.Response)
		if err != nil {
			return SenderRequest{}, err
		}

		resLog.ID = req.ID

		senderReq.Response = &resLog
	}

	return senderReq, nil
}

func parseHTTPRequest(req *http.Request) (HTTPRequest, error) {
	method := HTTPMethod(req.Method)
	if method != "" && !method.IsValid() {
		return HTTPRequest{}, fmt.Errorf("http request has invalid method: %v", method)
	}

	reqProto := httpProtocolMap[req.Proto]
	if !reqProto.IsValid() {
		return HTTPRequest{}, fmt.Errorf("http request has invalid protocol: %v", req.Proto)
	}

	id, ok := proxy.RequestIDFromContext(req.Context())
	if !ok {
		return HTTPRequest{}, errors.New("http request has missing ID")
	}

	httpReq := HTTPRequest{
		ID:     id,
		URL:    req.URL,
		Method: method,
		Proto:  HTTPProtocol(req.Proto),
	}

	if req.Header != nil {
		httpReq.Headers = make([]HTTPHeader, 0)

		for key, values := range req.Header {
			for _, value := range values {
				httpReq.Headers = append(httpReq.Headers, HTTPHeader{
					Key:   key,
					Value: value,
				})
			}
		}

		sort.Sort(HTTPHeaders(httpReq.Headers))
	}

	if req.Body != nil {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return HTTPRequest{}, fmt.Errorf("failed to read request body: %w", err)
		}

		req.Body = io.NopCloser(bytes.NewBuffer(body))
		bodyStr := string(body)
		httpReq.Body = &bodyStr
	}

	return httpReq, nil
}

func parseHTTPResponse(res *http.Response) (HTTPResponse, error) {
	resProto := httpProtocolMap[res.Proto]
	if !resProto.IsValid() {
		return HTTPResponse{}, fmt.Errorf("http response has invalid protocol: %v", res.Proto)
	}

	id, ok := proxy.RequestIDFromContext(res.Request.Context())
	if !ok {
		return HTTPResponse{}, errors.New("http response has missing ID")
	}

	httpRes := HTTPResponse{
		ID:         id,
		Proto:      resProto,
		StatusCode: res.StatusCode,
	}

	statusReasonSubs := strings.SplitN(res.Status, " ", 2)

	if len(statusReasonSubs) == 2 {
		httpRes.StatusReason = statusReasonSubs[1]
	}

	if res.Header != nil {
		httpRes.Headers = make([]HTTPHeader, 0)

		for key, values := range res.Header {
			for _, value := range values {
				httpRes.Headers = append(httpRes.Headers, HTTPHeader{
					Key:   key,
					Value: value,
				})
			}
		}

		sort.Sort(HTTPHeaders(httpRes.Headers))
	}

	if res.Body != nil {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return HTTPResponse{}, fmt.Errorf("failed to read response body: %w", err)
		}

		res.Body = io.NopCloser(bytes.NewBuffer(body))
		bodyStr := string(body)
		httpRes.Body = &bodyStr
	}

	return httpRes, nil
}

func parseInterceptItem(item intercept.Item) (req HTTPRequest, err error) {
	if item.Response != nil {
		req, err = parseHTTPRequest(item.Response.Request)
		if err != nil {
			return HTTPRequest{}, err
		}

		res, err := parseHTTPResponse(item.Response)
		if err != nil {
			return HTTPRequest{}, err
		}

		req.Response = &res
	} else if item.Request != nil {
		req, err = parseHTTPRequest(item.Request)
		if err != nil {
			return HTTPRequest{}, err
		}
	}

	return req, nil
}

func parseProject(projSvc *proj.Service, p proj.Project) Project {
	project := Project{
		ID:       p.ID,
		Name:     p.Name,
		IsActive: projSvc.IsProjectActive(p.ID),
		Settings: &ProjectSettings{
			Intercept: &InterceptSettings{
				RequestsEnabled:  p.Settings.InterceptRequests,
				ResponsesEnabled: p.Settings.InterceptResponses,
			},
		},
	}

	if p.Settings.InterceptRequestFilter != nil {
		interceptReqFilter := p.Settings.InterceptRequestFilter.String()
		project.Settings.Intercept.RequestFilter = &interceptReqFilter
	}

	if p.Settings.InterceptResponseFilter != nil {
		interceptResFilter := p.Settings.InterceptResponseFilter.String()
		project.Settings.Intercept.ResponseFilter = &interceptResFilter
	}

	return project
}

func stringPtrToRegexp(s *string) (*regexp.Regexp, error) {
	if s == nil {
		return nil, nil
	}

	return regexp.Compile(*s)
}

func scopeToScopeRules(rules []scope.Rule) []ScopeRule {
	scopeRules := make([]ScopeRule, len(rules))
	for i, rule := range rules {
		scopeRules[i].URL = regexpToStringPtr(rule.URL)
		if rule.Header.Key != nil || rule.Header.Value != nil {
			scopeRules[i].Header = &ScopeHeader{
				Key:   regexpToStringPtr(rule.Header.Key),
				Value: regexpToStringPtr(rule.Header.Value),
			}
		}

		scopeRules[i].Body = regexpToStringPtr(rule.Body)
	}

	return scopeRules
}

func findRequestsFilterFromInput(input *HTTPRequestLogFilterInput) (findFilter reqlog.FindRequestsFilter, err error) {
	if input == nil {
		return
	}

	if input.OnlyInScope != nil {
		findFilter.OnlyInScope = *input.OnlyInScope
	}

	if input.SearchExpression != nil && *input.SearchExpression != "" {
		expr, err := httpql.Parse(*input.SearchExpression)
		if err != nil {
			return reqlog.FindRequestsFilter{}, fmt.Errorf("could not parse search query: %w", err)
		}

		findFilter.SearchExpr = expr
	}

	return
}

func findSenderRequestsFilterFromInput(input *SenderRequestFilterInput) (findFilter sender.FindRequestsFilter, err error) {
	if input == nil {
		return
	}

	if input.OnlyInScope != nil {
		findFilter.OnlyInScope = *input.OnlyInScope
	}

	if input.SearchExpression != nil && *input.SearchExpression != "" {
		expr, err := httpql.Parse(*input.SearchExpression)
		if err != nil {
			return sender.FindRequestsFilter{}, fmt.Errorf("could not parse search query: %w", err)
		}

		findFilter.SearchExpr = expr
	}

	return
}

func findReqFilterToHTTPReqLogFilter(findReqFilter reqlog.FindRequestsFilter) *HTTPRequestLogFilter {
	empty := reqlog.FindRequestsFilter{}
	if findReqFilter == empty {
		return nil
	}

	httpReqLogFilter := &HTTPRequestLogFilter{
		OnlyInScope: findReqFilter.OnlyInScope,
	}

	if findReqFilter.SearchExpr != nil {
		searchExpr := findReqFilter.SearchExpr.String()
		httpReqLogFilter.SearchExpression = &searchExpr
	}

	return httpReqLogFilter
}

func findReqFilterToSenderReqFilter(findReqFilter sender.FindRequestsFilter) *SenderRequestFilter {
	empty := sender.FindRequestsFilter{}
	if findReqFilter == empty {
		return nil
	}

	senderReqFilter := &SenderRequestFilter{
		OnlyInScope: findReqFilter.OnlyInScope,
	}

	if findReqFilter.SearchExpr != nil {
		searchExpr := findReqFilter.SearchExpr.String()
		senderReqFilter.SearchExpression = &searchExpr
	}

	return senderReqFilter
}

func noActiveProjectErr(ctx context.Context) error {
	return &gqlerror.Error{
		Path:    graphql.GetPath(ctx),
		Message: "No active project.",
		Extensions: map[string]interface{}{
			"code": "no_active_project",
		},
	}
}
