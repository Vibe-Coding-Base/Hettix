import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
export type Maybe<T> = T | null;
export type InputMaybe<T> = Maybe<T>;
export type Exact<T extends { [key: string]: unknown }> = { [K in keyof T]: T[K] };
export type MakeOptional<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]?: Maybe<T[SubKey]> };
export type MakeMaybe<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]: Maybe<T[SubKey]> };
const defaultOptions = {} as const;
/** All built-in and custom scalars, mapped to their actual values */
export type Scalars = {
  ID: string;
  String: string;
  Boolean: boolean;
  Int: number;
  Float: number;
  Regexp: any;
  Time: any;
  URL: any;
};

export type AgentAction = {
  __typename?: 'AgentAction';
  denied: Scalars['Boolean'];
  input: Scalars['String'];
  output: Scalars['String'];
  tool: Scalars['String'];
};

export enum AgentMode {
  Ask = 'ASK',
  Assist = 'ASSIST',
  Auto = 'AUTO'
}

export type AgentReply = {
  __typename?: 'AgentReply';
  actions: Array<AgentAction>;
  reply: Scalars['String'];
};

export type CancelRequestResult = {
  __typename?: 'CancelRequestResult';
  success: Scalars['Boolean'];
};

export type CancelResponseResult = {
  __typename?: 'CancelResponseResult';
  success: Scalars['Boolean'];
};

export type ClearHttpRequestLogResult = {
  __typename?: 'ClearHTTPRequestLogResult';
  success: Scalars['Boolean'];
};

export type CloseProjectResult = {
  __typename?: 'CloseProjectResult';
  success: Scalars['Boolean'];
};

export type DeleteProjectResult = {
  __typename?: 'DeleteProjectResult';
  success: Scalars['Boolean'];
};

export type DeleteSenderRequestsResult = {
  __typename?: 'DeleteSenderRequestsResult';
  success: Scalars['Boolean'];
};

export type DropWebSocketMessageResult = {
  __typename?: 'DropWebSocketMessageResult';
  success: Scalars['Boolean'];
};

export type HttpHeader = {
  __typename?: 'HttpHeader';
  key: Scalars['String'];
  value: Scalars['String'];
};

export type HttpHeaderInput = {
  key: Scalars['String'];
  value: Scalars['String'];
};

export enum HttpMethod {
  Connect = 'CONNECT',
  Delete = 'DELETE',
  Get = 'GET',
  Head = 'HEAD',
  Options = 'OPTIONS',
  Patch = 'PATCH',
  Post = 'POST',
  Put = 'PUT',
  Trace = 'TRACE'
}

export enum HttpProtocol {
  Http10 = 'HTTP10',
  Http11 = 'HTTP11',
  Http20 = 'HTTP20'
}

export type HttpRequest = {
  __typename?: 'HttpRequest';
  body?: Maybe<Scalars['String']>;
  headers: Array<HttpHeader>;
  id: Scalars['ID'];
  method: HttpMethod;
  proto: HttpProtocol;
  response?: Maybe<HttpResponse>;
  url: Scalars['URL'];
};

export type HttpRequestLog = {
  __typename?: 'HttpRequestLog';
  body?: Maybe<Scalars['String']>;
  headers: Array<HttpHeader>;
  id: Scalars['ID'];
  method: HttpMethod;
  proto: Scalars['String'];
  response?: Maybe<HttpResponseLog>;
  timestamp: Scalars['Time'];
  url: Scalars['String'];
};

export type HttpRequestLogFilter = {
  __typename?: 'HttpRequestLogFilter';
  onlyInScope: Scalars['Boolean'];
  searchExpression?: Maybe<Scalars['String']>;
};

export type HttpRequestLogFilterInput = {
  onlyInScope?: InputMaybe<Scalars['Boolean']>;
  searchExpression?: InputMaybe<Scalars['String']>;
};

export type HttpResponse = {
  __typename?: 'HttpResponse';
  body?: Maybe<Scalars['String']>;
  headers: Array<HttpHeader>;
  /** Will be the same ID as its related request ID. */
  id: Scalars['ID'];
  proto: HttpProtocol;
  statusCode: Scalars['Int'];
  statusReason: Scalars['String'];
};

export type HttpResponseLog = {
  __typename?: 'HttpResponseLog';
  body?: Maybe<Scalars['String']>;
  headers: Array<HttpHeader>;
  /** Will be the same ID as its related request ID. */
  id: Scalars['ID'];
  proto: HttpProtocol;
  statusCode: Scalars['Int'];
  statusReason: Scalars['String'];
};

export type InterceptSettings = {
  __typename?: 'InterceptSettings';
  requestFilter?: Maybe<Scalars['String']>;
  requestsEnabled: Scalars['Boolean'];
  responseFilter?: Maybe<Scalars['String']>;
  responsesEnabled: Scalars['Boolean'];
};

export type InterceptedWebSocketMessage = {
  __typename?: 'InterceptedWebSocketMessage';
  connectionId: Scalars['ID'];
  direction: WebSocketDirection;
  id: Scalars['ID'];
  opcode: Scalars['Int'];
  payload: Scalars['String'];
};

export type IntruderAttack = {
  __typename?: 'IntruderAttack';
  completed: Scalars['Int'];
  id: Scalars['ID'];
  name: Scalars['String'];
  status: IntruderAttackStatus;
  timestamp: Scalars['Time'];
  total: Scalars['Int'];
};

export enum IntruderAttackStatus {
  Completed = 'COMPLETED',
  Running = 'RUNNING'
}

export type IntruderHeaderInput = {
  key: Scalars['String'];
  value: Scalars['String'];
};

export type IntruderResult = {
  __typename?: 'IntruderResult';
  durationMs: Scalars['Int'];
  error?: Maybe<Scalars['String']>;
  index: Scalars['Int'];
  length: Scalars['Int'];
  payload: Scalars['String'];
  statusCode: Scalars['Int'];
};

export enum MatchReplacePhase {
  Request = 'REQUEST',
  Response = 'RESPONSE'
}

export type MatchReplaceRule = {
  __typename?: 'MatchReplaceRule';
  bodyMatcher?: Maybe<Scalars['String']>;
  bodyReplacement?: Maybe<Scalars['String']>;
  condition?: Maybe<Scalars['String']>;
  enabled: Scalars['Boolean'];
  headerName?: Maybe<Scalars['String']>;
  headerValue?: Maybe<Scalars['String']>;
  id: Scalars['ID'];
  name: Scalars['String'];
  phase: MatchReplacePhase;
  removeHeader: Scalars['Boolean'];
};

export type MatchReplaceRuleInput = {
  bodyMatcher?: InputMaybe<Scalars['String']>;
  bodyReplacement?: InputMaybe<Scalars['String']>;
  condition?: InputMaybe<Scalars['String']>;
  enabled: Scalars['Boolean'];
  headerName?: InputMaybe<Scalars['String']>;
  headerValue?: InputMaybe<Scalars['String']>;
  name: Scalars['String'];
  phase: MatchReplacePhase;
  removeHeader?: InputMaybe<Scalars['Boolean']>;
};

export type ModifyRequestInput = {
  body?: InputMaybe<Scalars['String']>;
  headers?: InputMaybe<Array<HttpHeaderInput>>;
  id: Scalars['ID'];
  method: HttpMethod;
  modifyResponse?: InputMaybe<Scalars['Boolean']>;
  proto: HttpProtocol;
  url: Scalars['URL'];
};

export type ModifyRequestResult = {
  __typename?: 'ModifyRequestResult';
  success: Scalars['Boolean'];
};

export type ModifyResponseInput = {
  body?: InputMaybe<Scalars['String']>;
  headers?: InputMaybe<Array<HttpHeaderInput>>;
  proto: HttpProtocol;
  requestID: Scalars['ID'];
  statusCode: Scalars['Int'];
  statusReason: Scalars['String'];
};

export type ModifyResponseResult = {
  __typename?: 'ModifyResponseResult';
  success: Scalars['Boolean'];
};

export type ModifyWebSocketMessageInput = {
  id: Scalars['ID'];
  payload: Scalars['String'];
};

export type ModifyWebSocketMessageResult = {
  __typename?: 'ModifyWebSocketMessageResult';
  success: Scalars['Boolean'];
};

export type Mutation = {
  __typename?: 'Mutation';
  cancelRequest: CancelRequestResult;
  cancelResponse: CancelResponseResult;
  clearHTTPRequestLog: ClearHttpRequestLogResult;
  closeProject: CloseProjectResult;
  createOrUpdateSenderRequest: SenderRequest;
  createProject?: Maybe<Project>;
  createSenderRequestFromHttpRequestLog: SenderRequest;
  deleteProject: DeleteProjectResult;
  deleteSenderRequests: DeleteSenderRequestsResult;
  dropWebSocketMessage: DropWebSocketMessageResult;
  forwardWebSocketMessage: ModifyWebSocketMessageResult;
  modifyRequest: ModifyRequestResult;
  modifyResponse: ModifyResponseResult;
  modifyWebSocketMessage: ModifyWebSocketMessageResult;
  openProject?: Maybe<Project>;
  runAgent: AgentReply;
  sendRequest: SenderRequest;
  setHttpRequestLogFilter?: Maybe<HttpRequestLogFilter>;
  setMatchReplaceRules: Array<MatchReplaceRule>;
  setScope: Array<ScopeRule>;
  setSenderRequestFilter?: Maybe<SenderRequestFilter>;
  startIntruderAttack: IntruderAttack;
  updateInterceptSettings: InterceptSettings;
  updateWebSocketInterceptSettings: WebSocketInterceptSettings;
};


export type MutationCancelRequestArgs = {
  id: Scalars['ID'];
};


export type MutationCancelResponseArgs = {
  requestID: Scalars['ID'];
};


export type MutationCreateOrUpdateSenderRequestArgs = {
  request: SenderRequestInput;
};


export type MutationCreateProjectArgs = {
  name: Scalars['String'];
};


export type MutationCreateSenderRequestFromHttpRequestLogArgs = {
  id: Scalars['ID'];
};


export type MutationDeleteProjectArgs = {
  id: Scalars['ID'];
};


export type MutationDropWebSocketMessageArgs = {
  id: Scalars['ID'];
};


export type MutationForwardWebSocketMessageArgs = {
  id: Scalars['ID'];
};


export type MutationModifyRequestArgs = {
  request: ModifyRequestInput;
};


export type MutationModifyResponseArgs = {
  response: ModifyResponseInput;
};


export type MutationModifyWebSocketMessageArgs = {
  input: ModifyWebSocketMessageInput;
};


export type MutationOpenProjectArgs = {
  id: Scalars['ID'];
};


export type MutationRunAgentArgs = {
  input: RunAgentInput;
};


export type MutationSendRequestArgs = {
  id: Scalars['ID'];
};


export type MutationSetHttpRequestLogFilterArgs = {
  filter?: InputMaybe<HttpRequestLogFilterInput>;
};


export type MutationSetMatchReplaceRulesArgs = {
  rules: Array<MatchReplaceRuleInput>;
};


export type MutationSetScopeArgs = {
  scope: Array<ScopeRuleInput>;
};


export type MutationSetSenderRequestFilterArgs = {
  filter?: InputMaybe<SenderRequestFilterInput>;
};


export type MutationStartIntruderAttackArgs = {
  input: StartIntruderAttackInput;
};


export type MutationUpdateInterceptSettingsArgs = {
  input: UpdateInterceptSettingsInput;
};


export type MutationUpdateWebSocketInterceptSettingsArgs = {
  input: UpdateWebSocketInterceptSettingsInput;
};

export type Project = {
  __typename?: 'Project';
  id: Scalars['ID'];
  isActive: Scalars['Boolean'];
  name: Scalars['String'];
  settings: ProjectSettings;
};

export type ProjectSettings = {
  __typename?: 'ProjectSettings';
  intercept: InterceptSettings;
};

export type Query = {
  __typename?: 'Query';
  activeProject?: Maybe<Project>;
  httpRequestLog?: Maybe<HttpRequestLog>;
  httpRequestLogFilter?: Maybe<HttpRequestLogFilter>;
  httpRequestLogs: Array<HttpRequestLog>;
  interceptedRequest?: Maybe<HttpRequest>;
  interceptedRequests: Array<HttpRequest>;
  interceptedWebSocketMessages: Array<InterceptedWebSocketMessage>;
  intruderAttack?: Maybe<IntruderAttack>;
  intruderAttacks: Array<IntruderAttack>;
  intruderResults: Array<IntruderResult>;
  matchReplaceRules: Array<MatchReplaceRule>;
  projects: Array<Project>;
  scope: Array<ScopeRule>;
  senderRequest?: Maybe<SenderRequest>;
  senderRequests: Array<SenderRequest>;
  sitemap: Array<SitemapEntry>;
  webSocketConnection?: Maybe<WebSocketConnection>;
  webSocketConnections: Array<WebSocketConnection>;
  webSocketInterceptSettings: WebSocketInterceptSettings;
  webSocketMessages: Array<WebSocketMessage>;
};


export type QueryHttpRequestLogArgs = {
  id: Scalars['ID'];
};


export type QueryHttpRequestLogsArgs = {
  limit?: InputMaybe<Scalars['Int']>;
  offset?: InputMaybe<Scalars['Int']>;
};


export type QueryInterceptedRequestArgs = {
  id: Scalars['ID'];
};


export type QueryIntruderAttackArgs = {
  id: Scalars['ID'];
};


export type QueryIntruderResultsArgs = {
  attackId: Scalars['ID'];
};


export type QuerySenderRequestArgs = {
  id: Scalars['ID'];
};


export type QuerySenderRequestsArgs = {
  limit?: InputMaybe<Scalars['Int']>;
  offset?: InputMaybe<Scalars['Int']>;
};


export type QueryWebSocketConnectionArgs = {
  id: Scalars['ID'];
};


export type QueryWebSocketConnectionsArgs = {
  searchExpression?: InputMaybe<Scalars['String']>;
};


export type QueryWebSocketMessagesArgs = {
  connectionId: Scalars['ID'];
  searchExpression?: InputMaybe<Scalars['String']>;
};

export type RunAgentInput = {
  message: Scalars['String'];
  mode?: InputMaybe<AgentMode>;
};

export type ScopeHeader = {
  __typename?: 'ScopeHeader';
  key?: Maybe<Scalars['Regexp']>;
  value?: Maybe<Scalars['Regexp']>;
};

export type ScopeHeaderInput = {
  key?: InputMaybe<Scalars['Regexp']>;
  value?: InputMaybe<Scalars['Regexp']>;
};

export type ScopeRule = {
  __typename?: 'ScopeRule';
  body?: Maybe<Scalars['Regexp']>;
  exclude: Scalars['Boolean'];
  header?: Maybe<ScopeHeader>;
  url?: Maybe<Scalars['Regexp']>;
};

export type ScopeRuleInput = {
  body?: InputMaybe<Scalars['Regexp']>;
  exclude?: InputMaybe<Scalars['Boolean']>;
  header?: InputMaybe<ScopeHeaderInput>;
  url?: InputMaybe<Scalars['Regexp']>;
};

export type SenderRequest = {
  __typename?: 'SenderRequest';
  body?: Maybe<Scalars['String']>;
  headers?: Maybe<Array<HttpHeader>>;
  id: Scalars['ID'];
  method: HttpMethod;
  proto: HttpProtocol;
  response?: Maybe<HttpResponseLog>;
  sourceRequestLogID?: Maybe<Scalars['ID']>;
  timestamp: Scalars['Time'];
  url: Scalars['URL'];
};

export type SenderRequestFilter = {
  __typename?: 'SenderRequestFilter';
  onlyInScope: Scalars['Boolean'];
  searchExpression?: Maybe<Scalars['String']>;
};

export type SenderRequestFilterInput = {
  onlyInScope?: InputMaybe<Scalars['Boolean']>;
  searchExpression?: InputMaybe<Scalars['String']>;
};

export type SenderRequestInput = {
  body?: InputMaybe<Scalars['String']>;
  headers?: InputMaybe<Array<HttpHeaderInput>>;
  id?: InputMaybe<Scalars['ID']>;
  method?: InputMaybe<HttpMethod>;
  proto?: InputMaybe<HttpProtocol>;
  url: Scalars['URL'];
};

export type SitemapEntry = {
  __typename?: 'SitemapEntry';
  count: Scalars['Int'];
  host: Scalars['String'];
  methods: Array<Scalars['String']>;
  path: Scalars['String'];
  statusCodes: Array<Scalars['Int']>;
};

export type StartIntruderAttackInput = {
  body?: InputMaybe<Scalars['String']>;
  headers?: InputMaybe<Array<IntruderHeaderInput>>;
  method: HttpMethod;
  name: Scalars['String'];
  payloads: Array<Scalars['String']>;
  url: Scalars['String'];
};

export type UpdateInterceptSettingsInput = {
  requestFilter?: InputMaybe<Scalars['String']>;
  requestsEnabled: Scalars['Boolean'];
  responseFilter?: InputMaybe<Scalars['String']>;
  responsesEnabled: Scalars['Boolean'];
};

export type UpdateWebSocketInterceptSettingsInput = {
  enabled: Scalars['Boolean'];
  filter?: InputMaybe<Scalars['String']>;
};

export type WebSocketConnection = {
  __typename?: 'WebSocketConnection';
  closedAt?: Maybe<Scalars['Time']>;
  host: Scalars['String'];
  id: Scalars['ID'];
  messageCount: Scalars['Int'];
  path: Scalars['String'];
  timestamp: Scalars['Time'];
  url: Scalars['String'];
};

export enum WebSocketDirection {
  ClientToServer = 'CLIENT_TO_SERVER',
  ServerToClient = 'SERVER_TO_CLIENT'
}

export type WebSocketInterceptSettings = {
  __typename?: 'WebSocketInterceptSettings';
  enabled: Scalars['Boolean'];
  filter?: Maybe<Scalars['String']>;
};

export type WebSocketMessage = {
  __typename?: 'WebSocketMessage';
  direction: WebSocketDirection;
  opcode: Scalars['Int'];
  payload: Scalars['String'];
  timestamp: Scalars['Time'];
};

export type RunAgentMutationVariables = Exact<{
  input: RunAgentInput;
}>;


export type RunAgentMutation = { __typename?: 'Mutation', runAgent: { __typename?: 'AgentReply', reply: string, actions: Array<{ __typename?: 'AgentAction', tool: string, input: string, output: string, denied: boolean }> } };

export type CancelRequestMutationVariables = Exact<{
  id: Scalars['ID'];
}>;


export type CancelRequestMutation = { __typename?: 'Mutation', cancelRequest: { __typename?: 'CancelRequestResult', success: boolean } };

export type CancelResponseMutationVariables = Exact<{
  requestID: Scalars['ID'];
}>;


export type CancelResponseMutation = { __typename?: 'Mutation', cancelResponse: { __typename?: 'CancelResponseResult', success: boolean } };

export type GetInterceptedRequestQueryVariables = Exact<{
  id: Scalars['ID'];
}>;


export type GetInterceptedRequestQuery = { __typename?: 'Query', interceptedRequest?: { __typename?: 'HttpRequest', id: string, url: any, method: HttpMethod, proto: HttpProtocol, body?: string | null, headers: Array<{ __typename?: 'HttpHeader', key: string, value: string }>, response?: { __typename?: 'HttpResponse', id: string, proto: HttpProtocol, statusCode: number, statusReason: string, body?: string | null, headers: Array<{ __typename?: 'HttpHeader', key: string, value: string }> } | null } | null };

export type ModifyRequestMutationVariables = Exact<{
  request: ModifyRequestInput;
}>;


export type ModifyRequestMutation = { __typename?: 'Mutation', modifyRequest: { __typename?: 'ModifyRequestResult', success: boolean } };

export type ModifyResponseMutationVariables = Exact<{
  response: ModifyResponseInput;
}>;


export type ModifyResponseMutation = { __typename?: 'Mutation', modifyResponse: { __typename?: 'ModifyResponseResult', success: boolean } };

export type IntruderAttacksQueryVariables = Exact<{ [key: string]: never; }>;


export type IntruderAttacksQuery = { __typename?: 'Query', intruderAttacks: Array<{ __typename?: 'IntruderAttack', id: string, name: string, status: IntruderAttackStatus, total: number, completed: number, timestamp: any }> };

export type IntruderResultsQueryVariables = Exact<{
  attackId: Scalars['ID'];
}>;


export type IntruderResultsQuery = { __typename?: 'Query', intruderResults: Array<{ __typename?: 'IntruderResult', index: number, payload: string, statusCode: number, length: number, durationMs: number, error?: string | null }> };

export type StartIntruderAttackMutationVariables = Exact<{
  input: StartIntruderAttackInput;
}>;


export type StartIntruderAttackMutation = { __typename?: 'Mutation', startIntruderAttack: { __typename?: 'IntruderAttack', id: string, name: string, status: IntruderAttackStatus, total: number, completed: number, timestamp: any } };

export type MatchReplaceRulesQueryVariables = Exact<{ [key: string]: never; }>;


export type MatchReplaceRulesQuery = { __typename?: 'Query', matchReplaceRules: Array<{ __typename?: 'MatchReplaceRule', id: string, name: string, enabled: boolean, phase: MatchReplacePhase, condition?: string | null, headerName?: string | null, headerValue?: string | null, removeHeader: boolean, bodyMatcher?: string | null, bodyReplacement?: string | null }> };

export type SetMatchReplaceRulesMutationVariables = Exact<{
  rules: Array<MatchReplaceRuleInput> | MatchReplaceRuleInput;
}>;


export type SetMatchReplaceRulesMutation = { __typename?: 'Mutation', setMatchReplaceRules: Array<{ __typename?: 'MatchReplaceRule', id: string, name: string, enabled: boolean, phase: MatchReplacePhase, condition?: string | null, headerName?: string | null, headerValue?: string | null, removeHeader: boolean, bodyMatcher?: string | null, bodyReplacement?: string | null }> };

export type ActiveProjectQueryVariables = Exact<{ [key: string]: never; }>;


export type ActiveProjectQuery = { __typename?: 'Query', activeProject?: { __typename?: 'Project', id: string, name: string, isActive: boolean, settings: { __typename?: 'ProjectSettings', intercept: { __typename?: 'InterceptSettings', requestsEnabled: boolean, responsesEnabled: boolean, requestFilter?: string | null, responseFilter?: string | null } } } | null };

export type CloseProjectMutationVariables = Exact<{ [key: string]: never; }>;


export type CloseProjectMutation = { __typename?: 'Mutation', closeProject: { __typename?: 'CloseProjectResult', success: boolean } };

export type CreateProjectMutationVariables = Exact<{
  name: Scalars['String'];
}>;


export type CreateProjectMutation = { __typename?: 'Mutation', createProject?: { __typename?: 'Project', id: string, name: string } | null };

export type DeleteProjectMutationVariables = Exact<{
  id: Scalars['ID'];
}>;


export type DeleteProjectMutation = { __typename?: 'Mutation', deleteProject: { __typename?: 'DeleteProjectResult', success: boolean } };

export type OpenProjectMutationVariables = Exact<{
  id: Scalars['ID'];
}>;


export type OpenProjectMutation = { __typename?: 'Mutation', openProject?: { __typename?: 'Project', id: string, name: string, isActive: boolean } | null };

export type ProjectsQueryVariables = Exact<{ [key: string]: never; }>;


export type ProjectsQuery = { __typename?: 'Query', projects: Array<{ __typename?: 'Project', id: string, name: string, isActive: boolean }> };

export type ClearHttpRequestLogMutationVariables = Exact<{ [key: string]: never; }>;


export type ClearHttpRequestLogMutation = { __typename?: 'Mutation', clearHTTPRequestLog: { __typename?: 'ClearHTTPRequestLogResult', success: boolean } };

export type HttpRequestLogQueryVariables = Exact<{
  id: Scalars['ID'];
}>;


export type HttpRequestLogQuery = { __typename?: 'Query', httpRequestLog?: { __typename?: 'HttpRequestLog', id: string, method: HttpMethod, url: string, proto: string, body?: string | null, headers: Array<{ __typename?: 'HttpHeader', key: string, value: string }>, response?: { __typename?: 'HttpResponseLog', id: string, proto: HttpProtocol, statusCode: number, statusReason: string, body?: string | null, headers: Array<{ __typename?: 'HttpHeader', key: string, value: string }> } | null } | null };

export type HttpRequestLogFilterQueryVariables = Exact<{ [key: string]: never; }>;


export type HttpRequestLogFilterQuery = { __typename?: 'Query', httpRequestLogFilter?: { __typename?: 'HttpRequestLogFilter', onlyInScope: boolean, searchExpression?: string | null } | null };

export type HttpRequestLogsQueryVariables = Exact<{ [key: string]: never; }>;


export type HttpRequestLogsQuery = { __typename?: 'Query', httpRequestLogs: Array<{ __typename?: 'HttpRequestLog', id: string, method: HttpMethod, url: string, timestamp: any, response?: { __typename?: 'HttpResponseLog', statusCode: number, statusReason: string } | null }> };

export type SetHttpRequestLogFilterMutationVariables = Exact<{
  filter?: InputMaybe<HttpRequestLogFilterInput>;
}>;


export type SetHttpRequestLogFilterMutation = { __typename?: 'Mutation', setHttpRequestLogFilter?: { __typename?: 'HttpRequestLogFilter', onlyInScope: boolean, searchExpression?: string | null } | null };

export type ScopeQueryVariables = Exact<{ [key: string]: never; }>;


export type ScopeQuery = { __typename?: 'Query', scope: Array<{ __typename?: 'ScopeRule', url?: any | null, exclude: boolean }> };

export type SetScopeMutationVariables = Exact<{
  scope: Array<ScopeRuleInput> | ScopeRuleInput;
}>;


export type SetScopeMutation = { __typename?: 'Mutation', setScope: Array<{ __typename?: 'ScopeRule', url?: any | null, exclude: boolean }> };

export type CreateOrUpdateSenderRequestMutationVariables = Exact<{
  request: SenderRequestInput;
}>;


export type CreateOrUpdateSenderRequestMutation = { __typename?: 'Mutation', createOrUpdateSenderRequest: { __typename?: 'SenderRequest', id: string } };

export type CreateSenderRequestFromHttpRequestLogMutationVariables = Exact<{
  id: Scalars['ID'];
}>;


export type CreateSenderRequestFromHttpRequestLogMutation = { __typename?: 'Mutation', createSenderRequestFromHttpRequestLog: { __typename?: 'SenderRequest', id: string } };

export type SendRequestMutationVariables = Exact<{
  id: Scalars['ID'];
}>;


export type SendRequestMutation = { __typename?: 'Mutation', sendRequest: { __typename?: 'SenderRequest', id: string } };

export type GetSenderRequestQueryVariables = Exact<{
  id: Scalars['ID'];
}>;


export type GetSenderRequestQuery = { __typename?: 'Query', senderRequest?: { __typename?: 'SenderRequest', id: string, sourceRequestLogID?: string | null, url: any, method: HttpMethod, proto: HttpProtocol, body?: string | null, timestamp: any, headers?: Array<{ __typename?: 'HttpHeader', key: string, value: string }> | null, response?: { __typename?: 'HttpResponseLog', id: string, proto: HttpProtocol, statusCode: number, statusReason: string, body?: string | null, headers: Array<{ __typename?: 'HttpHeader', key: string, value: string }> } | null } | null };

export type GetSenderRequestsQueryVariables = Exact<{ [key: string]: never; }>;


export type GetSenderRequestsQuery = { __typename?: 'Query', senderRequests: Array<{ __typename?: 'SenderRequest', id: string, url: any, method: HttpMethod, response?: { __typename?: 'HttpResponseLog', id: string, statusCode: number, statusReason: string } | null }> };

export type UpdateInterceptSettingsMutationVariables = Exact<{
  input: UpdateInterceptSettingsInput;
}>;


export type UpdateInterceptSettingsMutation = { __typename?: 'Mutation', updateInterceptSettings: { __typename?: 'InterceptSettings', requestsEnabled: boolean, responsesEnabled: boolean, requestFilter?: string | null, responseFilter?: string | null } };

export type SitemapQueryVariables = Exact<{ [key: string]: never; }>;


export type SitemapQuery = { __typename?: 'Query', sitemap: Array<{ __typename?: 'SitemapEntry', host: string, path: string, methods: Array<string>, statusCodes: Array<number>, count: number }> };

export type WebSocketConnectionsQueryVariables = Exact<{
  searchExpression?: InputMaybe<Scalars['String']>;
}>;


export type WebSocketConnectionsQuery = { __typename?: 'Query', webSocketConnections: Array<{ __typename?: 'WebSocketConnection', id: string, url: string, host: string, path: string, timestamp: any, closedAt?: any | null, messageCount: number }> };

export type WebSocketMessagesQueryVariables = Exact<{
  connectionId: Scalars['ID'];
  searchExpression?: InputMaybe<Scalars['String']>;
}>;


export type WebSocketMessagesQuery = { __typename?: 'Query', webSocketMessages: Array<{ __typename?: 'WebSocketMessage', direction: WebSocketDirection, opcode: number, payload: string, timestamp: any }> };

export type WebSocketInterceptSettingsQueryVariables = Exact<{ [key: string]: never; }>;


export type WebSocketInterceptSettingsQuery = { __typename?: 'Query', webSocketInterceptSettings: { __typename?: 'WebSocketInterceptSettings', enabled: boolean, filter?: string | null } };

export type InterceptedWebSocketMessagesQueryVariables = Exact<{ [key: string]: never; }>;


export type InterceptedWebSocketMessagesQuery = { __typename?: 'Query', interceptedWebSocketMessages: Array<{ __typename?: 'InterceptedWebSocketMessage', id: string, connectionId: string, direction: WebSocketDirection, opcode: number, payload: string }> };

export type UpdateWebSocketInterceptSettingsMutationVariables = Exact<{
  input: UpdateWebSocketInterceptSettingsInput;
}>;


export type UpdateWebSocketInterceptSettingsMutation = { __typename?: 'Mutation', updateWebSocketInterceptSettings: { __typename?: 'WebSocketInterceptSettings', enabled: boolean, filter?: string | null } };

export type ModifyWebSocketMessageMutationVariables = Exact<{
  input: ModifyWebSocketMessageInput;
}>;


export type ModifyWebSocketMessageMutation = { __typename?: 'Mutation', modifyWebSocketMessage: { __typename?: 'ModifyWebSocketMessageResult', success: boolean } };

export type ForwardWebSocketMessageMutationVariables = Exact<{
  id: Scalars['ID'];
}>;


export type ForwardWebSocketMessageMutation = { __typename?: 'Mutation', forwardWebSocketMessage: { __typename?: 'ModifyWebSocketMessageResult', success: boolean } };

export type DropWebSocketMessageMutationVariables = Exact<{
  id: Scalars['ID'];
}>;


export type DropWebSocketMessageMutation = { __typename?: 'Mutation', dropWebSocketMessage: { __typename?: 'DropWebSocketMessageResult', success: boolean } };

export type GetInterceptedRequestsQueryVariables = Exact<{ [key: string]: never; }>;


export type GetInterceptedRequestsQuery = { __typename?: 'Query', interceptedRequests: Array<{ __typename?: 'HttpRequest', id: string, url: any, method: HttpMethod, response?: { __typename?: 'HttpResponse', statusCode: number, statusReason: string } | null }> };


export const RunAgentDocument = gql`
    mutation RunAgent($input: RunAgentInput!) {
  runAgent(input: $input) {
    reply
    actions {
      tool
      input
      output
      denied
    }
  }
}
    `;
export type RunAgentMutationFn = Apollo.MutationFunction<RunAgentMutation, RunAgentMutationVariables>;

/**
 * __useRunAgentMutation__
 *
 * To run a mutation, you first call `useRunAgentMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useRunAgentMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [runAgentMutation, { data, loading, error }] = useRunAgentMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useRunAgentMutation(baseOptions?: Apollo.MutationHookOptions<RunAgentMutation, RunAgentMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<RunAgentMutation, RunAgentMutationVariables>(RunAgentDocument, options);
      }
export type RunAgentMutationHookResult = ReturnType<typeof useRunAgentMutation>;
export type RunAgentMutationResult = Apollo.MutationResult<RunAgentMutation>;
export type RunAgentMutationOptions = Apollo.BaseMutationOptions<RunAgentMutation, RunAgentMutationVariables>;
export const CancelRequestDocument = gql`
    mutation CancelRequest($id: ID!) {
  cancelRequest(id: $id) {
    success
  }
}
    `;
export type CancelRequestMutationFn = Apollo.MutationFunction<CancelRequestMutation, CancelRequestMutationVariables>;

/**
 * __useCancelRequestMutation__
 *
 * To run a mutation, you first call `useCancelRequestMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCancelRequestMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [cancelRequestMutation, { data, loading, error }] = useCancelRequestMutation({
 *   variables: {
 *      id: // value for 'id'
 *   },
 * });
 */
export function useCancelRequestMutation(baseOptions?: Apollo.MutationHookOptions<CancelRequestMutation, CancelRequestMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CancelRequestMutation, CancelRequestMutationVariables>(CancelRequestDocument, options);
      }
export type CancelRequestMutationHookResult = ReturnType<typeof useCancelRequestMutation>;
export type CancelRequestMutationResult = Apollo.MutationResult<CancelRequestMutation>;
export type CancelRequestMutationOptions = Apollo.BaseMutationOptions<CancelRequestMutation, CancelRequestMutationVariables>;
export const CancelResponseDocument = gql`
    mutation CancelResponse($requestID: ID!) {
  cancelResponse(requestID: $requestID) {
    success
  }
}
    `;
export type CancelResponseMutationFn = Apollo.MutationFunction<CancelResponseMutation, CancelResponseMutationVariables>;

/**
 * __useCancelResponseMutation__
 *
 * To run a mutation, you first call `useCancelResponseMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCancelResponseMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [cancelResponseMutation, { data, loading, error }] = useCancelResponseMutation({
 *   variables: {
 *      requestID: // value for 'requestID'
 *   },
 * });
 */
export function useCancelResponseMutation(baseOptions?: Apollo.MutationHookOptions<CancelResponseMutation, CancelResponseMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CancelResponseMutation, CancelResponseMutationVariables>(CancelResponseDocument, options);
      }
export type CancelResponseMutationHookResult = ReturnType<typeof useCancelResponseMutation>;
export type CancelResponseMutationResult = Apollo.MutationResult<CancelResponseMutation>;
export type CancelResponseMutationOptions = Apollo.BaseMutationOptions<CancelResponseMutation, CancelResponseMutationVariables>;
export const GetInterceptedRequestDocument = gql`
    query GetInterceptedRequest($id: ID!) {
  interceptedRequest(id: $id) {
    id
    url
    method
    proto
    headers {
      key
      value
    }
    body
    response {
      id
      proto
      statusCode
      statusReason
      headers {
        key
        value
      }
      body
    }
  }
}
    `;

/**
 * __useGetInterceptedRequestQuery__
 *
 * To run a query within a React component, call `useGetInterceptedRequestQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetInterceptedRequestQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetInterceptedRequestQuery({
 *   variables: {
 *      id: // value for 'id'
 *   },
 * });
 */
export function useGetInterceptedRequestQuery(baseOptions: Apollo.QueryHookOptions<GetInterceptedRequestQuery, GetInterceptedRequestQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetInterceptedRequestQuery, GetInterceptedRequestQueryVariables>(GetInterceptedRequestDocument, options);
      }
export function useGetInterceptedRequestLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetInterceptedRequestQuery, GetInterceptedRequestQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetInterceptedRequestQuery, GetInterceptedRequestQueryVariables>(GetInterceptedRequestDocument, options);
        }
export type GetInterceptedRequestQueryHookResult = ReturnType<typeof useGetInterceptedRequestQuery>;
export type GetInterceptedRequestLazyQueryHookResult = ReturnType<typeof useGetInterceptedRequestLazyQuery>;
export type GetInterceptedRequestQueryResult = Apollo.QueryResult<GetInterceptedRequestQuery, GetInterceptedRequestQueryVariables>;
export const ModifyRequestDocument = gql`
    mutation ModifyRequest($request: ModifyRequestInput!) {
  modifyRequest(request: $request) {
    success
  }
}
    `;
export type ModifyRequestMutationFn = Apollo.MutationFunction<ModifyRequestMutation, ModifyRequestMutationVariables>;

/**
 * __useModifyRequestMutation__
 *
 * To run a mutation, you first call `useModifyRequestMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useModifyRequestMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [modifyRequestMutation, { data, loading, error }] = useModifyRequestMutation({
 *   variables: {
 *      request: // value for 'request'
 *   },
 * });
 */
export function useModifyRequestMutation(baseOptions?: Apollo.MutationHookOptions<ModifyRequestMutation, ModifyRequestMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<ModifyRequestMutation, ModifyRequestMutationVariables>(ModifyRequestDocument, options);
      }
export type ModifyRequestMutationHookResult = ReturnType<typeof useModifyRequestMutation>;
export type ModifyRequestMutationResult = Apollo.MutationResult<ModifyRequestMutation>;
export type ModifyRequestMutationOptions = Apollo.BaseMutationOptions<ModifyRequestMutation, ModifyRequestMutationVariables>;
export const ModifyResponseDocument = gql`
    mutation ModifyResponse($response: ModifyResponseInput!) {
  modifyResponse(response: $response) {
    success
  }
}
    `;
export type ModifyResponseMutationFn = Apollo.MutationFunction<ModifyResponseMutation, ModifyResponseMutationVariables>;

/**
 * __useModifyResponseMutation__
 *
 * To run a mutation, you first call `useModifyResponseMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useModifyResponseMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [modifyResponseMutation, { data, loading, error }] = useModifyResponseMutation({
 *   variables: {
 *      response: // value for 'response'
 *   },
 * });
 */
export function useModifyResponseMutation(baseOptions?: Apollo.MutationHookOptions<ModifyResponseMutation, ModifyResponseMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<ModifyResponseMutation, ModifyResponseMutationVariables>(ModifyResponseDocument, options);
      }
export type ModifyResponseMutationHookResult = ReturnType<typeof useModifyResponseMutation>;
export type ModifyResponseMutationResult = Apollo.MutationResult<ModifyResponseMutation>;
export type ModifyResponseMutationOptions = Apollo.BaseMutationOptions<ModifyResponseMutation, ModifyResponseMutationVariables>;
export const IntruderAttacksDocument = gql`
    query IntruderAttacks {
  intruderAttacks {
    id
    name
    status
    total
    completed
    timestamp
  }
}
    `;

/**
 * __useIntruderAttacksQuery__
 *
 * To run a query within a React component, call `useIntruderAttacksQuery` and pass it any options that fit your needs.
 * When your component renders, `useIntruderAttacksQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useIntruderAttacksQuery({
 *   variables: {
 *   },
 * });
 */
export function useIntruderAttacksQuery(baseOptions?: Apollo.QueryHookOptions<IntruderAttacksQuery, IntruderAttacksQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<IntruderAttacksQuery, IntruderAttacksQueryVariables>(IntruderAttacksDocument, options);
      }
export function useIntruderAttacksLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<IntruderAttacksQuery, IntruderAttacksQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<IntruderAttacksQuery, IntruderAttacksQueryVariables>(IntruderAttacksDocument, options);
        }
export type IntruderAttacksQueryHookResult = ReturnType<typeof useIntruderAttacksQuery>;
export type IntruderAttacksLazyQueryHookResult = ReturnType<typeof useIntruderAttacksLazyQuery>;
export type IntruderAttacksQueryResult = Apollo.QueryResult<IntruderAttacksQuery, IntruderAttacksQueryVariables>;
export const IntruderResultsDocument = gql`
    query IntruderResults($attackId: ID!) {
  intruderResults(attackId: $attackId) {
    index
    payload
    statusCode
    length
    durationMs
    error
  }
}
    `;

/**
 * __useIntruderResultsQuery__
 *
 * To run a query within a React component, call `useIntruderResultsQuery` and pass it any options that fit your needs.
 * When your component renders, `useIntruderResultsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useIntruderResultsQuery({
 *   variables: {
 *      attackId: // value for 'attackId'
 *   },
 * });
 */
export function useIntruderResultsQuery(baseOptions: Apollo.QueryHookOptions<IntruderResultsQuery, IntruderResultsQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<IntruderResultsQuery, IntruderResultsQueryVariables>(IntruderResultsDocument, options);
      }
export function useIntruderResultsLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<IntruderResultsQuery, IntruderResultsQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<IntruderResultsQuery, IntruderResultsQueryVariables>(IntruderResultsDocument, options);
        }
export type IntruderResultsQueryHookResult = ReturnType<typeof useIntruderResultsQuery>;
export type IntruderResultsLazyQueryHookResult = ReturnType<typeof useIntruderResultsLazyQuery>;
export type IntruderResultsQueryResult = Apollo.QueryResult<IntruderResultsQuery, IntruderResultsQueryVariables>;
export const StartIntruderAttackDocument = gql`
    mutation StartIntruderAttack($input: StartIntruderAttackInput!) {
  startIntruderAttack(input: $input) {
    id
    name
    status
    total
    completed
    timestamp
  }
}
    `;
export type StartIntruderAttackMutationFn = Apollo.MutationFunction<StartIntruderAttackMutation, StartIntruderAttackMutationVariables>;

/**
 * __useStartIntruderAttackMutation__
 *
 * To run a mutation, you first call `useStartIntruderAttackMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useStartIntruderAttackMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [startIntruderAttackMutation, { data, loading, error }] = useStartIntruderAttackMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useStartIntruderAttackMutation(baseOptions?: Apollo.MutationHookOptions<StartIntruderAttackMutation, StartIntruderAttackMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<StartIntruderAttackMutation, StartIntruderAttackMutationVariables>(StartIntruderAttackDocument, options);
      }
export type StartIntruderAttackMutationHookResult = ReturnType<typeof useStartIntruderAttackMutation>;
export type StartIntruderAttackMutationResult = Apollo.MutationResult<StartIntruderAttackMutation>;
export type StartIntruderAttackMutationOptions = Apollo.BaseMutationOptions<StartIntruderAttackMutation, StartIntruderAttackMutationVariables>;
export const MatchReplaceRulesDocument = gql`
    query MatchReplaceRules {
  matchReplaceRules {
    id
    name
    enabled
    phase
    condition
    headerName
    headerValue
    removeHeader
    bodyMatcher
    bodyReplacement
  }
}
    `;

/**
 * __useMatchReplaceRulesQuery__
 *
 * To run a query within a React component, call `useMatchReplaceRulesQuery` and pass it any options that fit your needs.
 * When your component renders, `useMatchReplaceRulesQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useMatchReplaceRulesQuery({
 *   variables: {
 *   },
 * });
 */
export function useMatchReplaceRulesQuery(baseOptions?: Apollo.QueryHookOptions<MatchReplaceRulesQuery, MatchReplaceRulesQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<MatchReplaceRulesQuery, MatchReplaceRulesQueryVariables>(MatchReplaceRulesDocument, options);
      }
export function useMatchReplaceRulesLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<MatchReplaceRulesQuery, MatchReplaceRulesQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<MatchReplaceRulesQuery, MatchReplaceRulesQueryVariables>(MatchReplaceRulesDocument, options);
        }
export type MatchReplaceRulesQueryHookResult = ReturnType<typeof useMatchReplaceRulesQuery>;
export type MatchReplaceRulesLazyQueryHookResult = ReturnType<typeof useMatchReplaceRulesLazyQuery>;
export type MatchReplaceRulesQueryResult = Apollo.QueryResult<MatchReplaceRulesQuery, MatchReplaceRulesQueryVariables>;
export const SetMatchReplaceRulesDocument = gql`
    mutation SetMatchReplaceRules($rules: [MatchReplaceRuleInput!]!) {
  setMatchReplaceRules(rules: $rules) {
    id
    name
    enabled
    phase
    condition
    headerName
    headerValue
    removeHeader
    bodyMatcher
    bodyReplacement
  }
}
    `;
export type SetMatchReplaceRulesMutationFn = Apollo.MutationFunction<SetMatchReplaceRulesMutation, SetMatchReplaceRulesMutationVariables>;

/**
 * __useSetMatchReplaceRulesMutation__
 *
 * To run a mutation, you first call `useSetMatchReplaceRulesMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useSetMatchReplaceRulesMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [setMatchReplaceRulesMutation, { data, loading, error }] = useSetMatchReplaceRulesMutation({
 *   variables: {
 *      rules: // value for 'rules'
 *   },
 * });
 */
export function useSetMatchReplaceRulesMutation(baseOptions?: Apollo.MutationHookOptions<SetMatchReplaceRulesMutation, SetMatchReplaceRulesMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<SetMatchReplaceRulesMutation, SetMatchReplaceRulesMutationVariables>(SetMatchReplaceRulesDocument, options);
      }
export type SetMatchReplaceRulesMutationHookResult = ReturnType<typeof useSetMatchReplaceRulesMutation>;
export type SetMatchReplaceRulesMutationResult = Apollo.MutationResult<SetMatchReplaceRulesMutation>;
export type SetMatchReplaceRulesMutationOptions = Apollo.BaseMutationOptions<SetMatchReplaceRulesMutation, SetMatchReplaceRulesMutationVariables>;
export const ActiveProjectDocument = gql`
    query ActiveProject {
  activeProject {
    id
    name
    isActive
    settings {
      intercept {
        requestsEnabled
        responsesEnabled
        requestFilter
        responseFilter
      }
    }
  }
}
    `;

/**
 * __useActiveProjectQuery__
 *
 * To run a query within a React component, call `useActiveProjectQuery` and pass it any options that fit your needs.
 * When your component renders, `useActiveProjectQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useActiveProjectQuery({
 *   variables: {
 *   },
 * });
 */
export function useActiveProjectQuery(baseOptions?: Apollo.QueryHookOptions<ActiveProjectQuery, ActiveProjectQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<ActiveProjectQuery, ActiveProjectQueryVariables>(ActiveProjectDocument, options);
      }
export function useActiveProjectLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<ActiveProjectQuery, ActiveProjectQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<ActiveProjectQuery, ActiveProjectQueryVariables>(ActiveProjectDocument, options);
        }
export type ActiveProjectQueryHookResult = ReturnType<typeof useActiveProjectQuery>;
export type ActiveProjectLazyQueryHookResult = ReturnType<typeof useActiveProjectLazyQuery>;
export type ActiveProjectQueryResult = Apollo.QueryResult<ActiveProjectQuery, ActiveProjectQueryVariables>;
export const CloseProjectDocument = gql`
    mutation CloseProject {
  closeProject {
    success
  }
}
    `;
export type CloseProjectMutationFn = Apollo.MutationFunction<CloseProjectMutation, CloseProjectMutationVariables>;

/**
 * __useCloseProjectMutation__
 *
 * To run a mutation, you first call `useCloseProjectMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCloseProjectMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [closeProjectMutation, { data, loading, error }] = useCloseProjectMutation({
 *   variables: {
 *   },
 * });
 */
export function useCloseProjectMutation(baseOptions?: Apollo.MutationHookOptions<CloseProjectMutation, CloseProjectMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CloseProjectMutation, CloseProjectMutationVariables>(CloseProjectDocument, options);
      }
export type CloseProjectMutationHookResult = ReturnType<typeof useCloseProjectMutation>;
export type CloseProjectMutationResult = Apollo.MutationResult<CloseProjectMutation>;
export type CloseProjectMutationOptions = Apollo.BaseMutationOptions<CloseProjectMutation, CloseProjectMutationVariables>;
export const CreateProjectDocument = gql`
    mutation CreateProject($name: String!) {
  createProject(name: $name) {
    id
    name
  }
}
    `;
export type CreateProjectMutationFn = Apollo.MutationFunction<CreateProjectMutation, CreateProjectMutationVariables>;

/**
 * __useCreateProjectMutation__
 *
 * To run a mutation, you first call `useCreateProjectMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateProjectMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createProjectMutation, { data, loading, error }] = useCreateProjectMutation({
 *   variables: {
 *      name: // value for 'name'
 *   },
 * });
 */
export function useCreateProjectMutation(baseOptions?: Apollo.MutationHookOptions<CreateProjectMutation, CreateProjectMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CreateProjectMutation, CreateProjectMutationVariables>(CreateProjectDocument, options);
      }
export type CreateProjectMutationHookResult = ReturnType<typeof useCreateProjectMutation>;
export type CreateProjectMutationResult = Apollo.MutationResult<CreateProjectMutation>;
export type CreateProjectMutationOptions = Apollo.BaseMutationOptions<CreateProjectMutation, CreateProjectMutationVariables>;
export const DeleteProjectDocument = gql`
    mutation DeleteProject($id: ID!) {
  deleteProject(id: $id) {
    success
  }
}
    `;
export type DeleteProjectMutationFn = Apollo.MutationFunction<DeleteProjectMutation, DeleteProjectMutationVariables>;

/**
 * __useDeleteProjectMutation__
 *
 * To run a mutation, you first call `useDeleteProjectMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useDeleteProjectMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [deleteProjectMutation, { data, loading, error }] = useDeleteProjectMutation({
 *   variables: {
 *      id: // value for 'id'
 *   },
 * });
 */
export function useDeleteProjectMutation(baseOptions?: Apollo.MutationHookOptions<DeleteProjectMutation, DeleteProjectMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<DeleteProjectMutation, DeleteProjectMutationVariables>(DeleteProjectDocument, options);
      }
export type DeleteProjectMutationHookResult = ReturnType<typeof useDeleteProjectMutation>;
export type DeleteProjectMutationResult = Apollo.MutationResult<DeleteProjectMutation>;
export type DeleteProjectMutationOptions = Apollo.BaseMutationOptions<DeleteProjectMutation, DeleteProjectMutationVariables>;
export const OpenProjectDocument = gql`
    mutation OpenProject($id: ID!) {
  openProject(id: $id) {
    id
    name
    isActive
  }
}
    `;
export type OpenProjectMutationFn = Apollo.MutationFunction<OpenProjectMutation, OpenProjectMutationVariables>;

/**
 * __useOpenProjectMutation__
 *
 * To run a mutation, you first call `useOpenProjectMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useOpenProjectMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [openProjectMutation, { data, loading, error }] = useOpenProjectMutation({
 *   variables: {
 *      id: // value for 'id'
 *   },
 * });
 */
export function useOpenProjectMutation(baseOptions?: Apollo.MutationHookOptions<OpenProjectMutation, OpenProjectMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<OpenProjectMutation, OpenProjectMutationVariables>(OpenProjectDocument, options);
      }
export type OpenProjectMutationHookResult = ReturnType<typeof useOpenProjectMutation>;
export type OpenProjectMutationResult = Apollo.MutationResult<OpenProjectMutation>;
export type OpenProjectMutationOptions = Apollo.BaseMutationOptions<OpenProjectMutation, OpenProjectMutationVariables>;
export const ProjectsDocument = gql`
    query Projects {
  projects {
    id
    name
    isActive
  }
}
    `;

/**
 * __useProjectsQuery__
 *
 * To run a query within a React component, call `useProjectsQuery` and pass it any options that fit your needs.
 * When your component renders, `useProjectsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useProjectsQuery({
 *   variables: {
 *   },
 * });
 */
export function useProjectsQuery(baseOptions?: Apollo.QueryHookOptions<ProjectsQuery, ProjectsQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<ProjectsQuery, ProjectsQueryVariables>(ProjectsDocument, options);
      }
export function useProjectsLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<ProjectsQuery, ProjectsQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<ProjectsQuery, ProjectsQueryVariables>(ProjectsDocument, options);
        }
export type ProjectsQueryHookResult = ReturnType<typeof useProjectsQuery>;
export type ProjectsLazyQueryHookResult = ReturnType<typeof useProjectsLazyQuery>;
export type ProjectsQueryResult = Apollo.QueryResult<ProjectsQuery, ProjectsQueryVariables>;
export const ClearHttpRequestLogDocument = gql`
    mutation ClearHTTPRequestLog {
  clearHTTPRequestLog {
    success
  }
}
    `;
export type ClearHttpRequestLogMutationFn = Apollo.MutationFunction<ClearHttpRequestLogMutation, ClearHttpRequestLogMutationVariables>;

/**
 * __useClearHttpRequestLogMutation__
 *
 * To run a mutation, you first call `useClearHttpRequestLogMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useClearHttpRequestLogMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [clearHttpRequestLogMutation, { data, loading, error }] = useClearHttpRequestLogMutation({
 *   variables: {
 *   },
 * });
 */
export function useClearHttpRequestLogMutation(baseOptions?: Apollo.MutationHookOptions<ClearHttpRequestLogMutation, ClearHttpRequestLogMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<ClearHttpRequestLogMutation, ClearHttpRequestLogMutationVariables>(ClearHttpRequestLogDocument, options);
      }
export type ClearHttpRequestLogMutationHookResult = ReturnType<typeof useClearHttpRequestLogMutation>;
export type ClearHttpRequestLogMutationResult = Apollo.MutationResult<ClearHttpRequestLogMutation>;
export type ClearHttpRequestLogMutationOptions = Apollo.BaseMutationOptions<ClearHttpRequestLogMutation, ClearHttpRequestLogMutationVariables>;
export const HttpRequestLogDocument = gql`
    query HttpRequestLog($id: ID!) {
  httpRequestLog(id: $id) {
    id
    method
    url
    proto
    headers {
      key
      value
    }
    body
    response {
      id
      proto
      headers {
        key
        value
      }
      statusCode
      statusReason
      body
    }
  }
}
    `;

/**
 * __useHttpRequestLogQuery__
 *
 * To run a query within a React component, call `useHttpRequestLogQuery` and pass it any options that fit your needs.
 * When your component renders, `useHttpRequestLogQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useHttpRequestLogQuery({
 *   variables: {
 *      id: // value for 'id'
 *   },
 * });
 */
export function useHttpRequestLogQuery(baseOptions: Apollo.QueryHookOptions<HttpRequestLogQuery, HttpRequestLogQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<HttpRequestLogQuery, HttpRequestLogQueryVariables>(HttpRequestLogDocument, options);
      }
export function useHttpRequestLogLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<HttpRequestLogQuery, HttpRequestLogQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<HttpRequestLogQuery, HttpRequestLogQueryVariables>(HttpRequestLogDocument, options);
        }
export type HttpRequestLogQueryHookResult = ReturnType<typeof useHttpRequestLogQuery>;
export type HttpRequestLogLazyQueryHookResult = ReturnType<typeof useHttpRequestLogLazyQuery>;
export type HttpRequestLogQueryResult = Apollo.QueryResult<HttpRequestLogQuery, HttpRequestLogQueryVariables>;
export const HttpRequestLogFilterDocument = gql`
    query HttpRequestLogFilter {
  httpRequestLogFilter {
    onlyInScope
    searchExpression
  }
}
    `;

/**
 * __useHttpRequestLogFilterQuery__
 *
 * To run a query within a React component, call `useHttpRequestLogFilterQuery` and pass it any options that fit your needs.
 * When your component renders, `useHttpRequestLogFilterQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useHttpRequestLogFilterQuery({
 *   variables: {
 *   },
 * });
 */
export function useHttpRequestLogFilterQuery(baseOptions?: Apollo.QueryHookOptions<HttpRequestLogFilterQuery, HttpRequestLogFilterQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<HttpRequestLogFilterQuery, HttpRequestLogFilterQueryVariables>(HttpRequestLogFilterDocument, options);
      }
export function useHttpRequestLogFilterLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<HttpRequestLogFilterQuery, HttpRequestLogFilterQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<HttpRequestLogFilterQuery, HttpRequestLogFilterQueryVariables>(HttpRequestLogFilterDocument, options);
        }
export type HttpRequestLogFilterQueryHookResult = ReturnType<typeof useHttpRequestLogFilterQuery>;
export type HttpRequestLogFilterLazyQueryHookResult = ReturnType<typeof useHttpRequestLogFilterLazyQuery>;
export type HttpRequestLogFilterQueryResult = Apollo.QueryResult<HttpRequestLogFilterQuery, HttpRequestLogFilterQueryVariables>;
export const HttpRequestLogsDocument = gql`
    query HttpRequestLogs {
  httpRequestLogs {
    id
    method
    url
    timestamp
    response {
      statusCode
      statusReason
    }
  }
}
    `;

/**
 * __useHttpRequestLogsQuery__
 *
 * To run a query within a React component, call `useHttpRequestLogsQuery` and pass it any options that fit your needs.
 * When your component renders, `useHttpRequestLogsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useHttpRequestLogsQuery({
 *   variables: {
 *   },
 * });
 */
export function useHttpRequestLogsQuery(baseOptions?: Apollo.QueryHookOptions<HttpRequestLogsQuery, HttpRequestLogsQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<HttpRequestLogsQuery, HttpRequestLogsQueryVariables>(HttpRequestLogsDocument, options);
      }
export function useHttpRequestLogsLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<HttpRequestLogsQuery, HttpRequestLogsQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<HttpRequestLogsQuery, HttpRequestLogsQueryVariables>(HttpRequestLogsDocument, options);
        }
export type HttpRequestLogsQueryHookResult = ReturnType<typeof useHttpRequestLogsQuery>;
export type HttpRequestLogsLazyQueryHookResult = ReturnType<typeof useHttpRequestLogsLazyQuery>;
export type HttpRequestLogsQueryResult = Apollo.QueryResult<HttpRequestLogsQuery, HttpRequestLogsQueryVariables>;
export const SetHttpRequestLogFilterDocument = gql`
    mutation SetHttpRequestLogFilter($filter: HttpRequestLogFilterInput) {
  setHttpRequestLogFilter(filter: $filter) {
    onlyInScope
    searchExpression
  }
}
    `;
export type SetHttpRequestLogFilterMutationFn = Apollo.MutationFunction<SetHttpRequestLogFilterMutation, SetHttpRequestLogFilterMutationVariables>;

/**
 * __useSetHttpRequestLogFilterMutation__
 *
 * To run a mutation, you first call `useSetHttpRequestLogFilterMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useSetHttpRequestLogFilterMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [setHttpRequestLogFilterMutation, { data, loading, error }] = useSetHttpRequestLogFilterMutation({
 *   variables: {
 *      filter: // value for 'filter'
 *   },
 * });
 */
export function useSetHttpRequestLogFilterMutation(baseOptions?: Apollo.MutationHookOptions<SetHttpRequestLogFilterMutation, SetHttpRequestLogFilterMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<SetHttpRequestLogFilterMutation, SetHttpRequestLogFilterMutationVariables>(SetHttpRequestLogFilterDocument, options);
      }
export type SetHttpRequestLogFilterMutationHookResult = ReturnType<typeof useSetHttpRequestLogFilterMutation>;
export type SetHttpRequestLogFilterMutationResult = Apollo.MutationResult<SetHttpRequestLogFilterMutation>;
export type SetHttpRequestLogFilterMutationOptions = Apollo.BaseMutationOptions<SetHttpRequestLogFilterMutation, SetHttpRequestLogFilterMutationVariables>;
export const ScopeDocument = gql`
    query Scope {
  scope {
    url
    exclude
  }
}
    `;

/**
 * __useScopeQuery__
 *
 * To run a query within a React component, call `useScopeQuery` and pass it any options that fit your needs.
 * When your component renders, `useScopeQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useScopeQuery({
 *   variables: {
 *   },
 * });
 */
export function useScopeQuery(baseOptions?: Apollo.QueryHookOptions<ScopeQuery, ScopeQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<ScopeQuery, ScopeQueryVariables>(ScopeDocument, options);
      }
export function useScopeLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<ScopeQuery, ScopeQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<ScopeQuery, ScopeQueryVariables>(ScopeDocument, options);
        }
export type ScopeQueryHookResult = ReturnType<typeof useScopeQuery>;
export type ScopeLazyQueryHookResult = ReturnType<typeof useScopeLazyQuery>;
export type ScopeQueryResult = Apollo.QueryResult<ScopeQuery, ScopeQueryVariables>;
export const SetScopeDocument = gql`
    mutation SetScope($scope: [ScopeRuleInput!]!) {
  setScope(scope: $scope) {
    url
    exclude
  }
}
    `;
export type SetScopeMutationFn = Apollo.MutationFunction<SetScopeMutation, SetScopeMutationVariables>;

/**
 * __useSetScopeMutation__
 *
 * To run a mutation, you first call `useSetScopeMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useSetScopeMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [setScopeMutation, { data, loading, error }] = useSetScopeMutation({
 *   variables: {
 *      scope: // value for 'scope'
 *   },
 * });
 */
export function useSetScopeMutation(baseOptions?: Apollo.MutationHookOptions<SetScopeMutation, SetScopeMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<SetScopeMutation, SetScopeMutationVariables>(SetScopeDocument, options);
      }
export type SetScopeMutationHookResult = ReturnType<typeof useSetScopeMutation>;
export type SetScopeMutationResult = Apollo.MutationResult<SetScopeMutation>;
export type SetScopeMutationOptions = Apollo.BaseMutationOptions<SetScopeMutation, SetScopeMutationVariables>;
export const CreateOrUpdateSenderRequestDocument = gql`
    mutation CreateOrUpdateSenderRequest($request: SenderRequestInput!) {
  createOrUpdateSenderRequest(request: $request) {
    id
  }
}
    `;
export type CreateOrUpdateSenderRequestMutationFn = Apollo.MutationFunction<CreateOrUpdateSenderRequestMutation, CreateOrUpdateSenderRequestMutationVariables>;

/**
 * __useCreateOrUpdateSenderRequestMutation__
 *
 * To run a mutation, you first call `useCreateOrUpdateSenderRequestMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateOrUpdateSenderRequestMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createOrUpdateSenderRequestMutation, { data, loading, error }] = useCreateOrUpdateSenderRequestMutation({
 *   variables: {
 *      request: // value for 'request'
 *   },
 * });
 */
export function useCreateOrUpdateSenderRequestMutation(baseOptions?: Apollo.MutationHookOptions<CreateOrUpdateSenderRequestMutation, CreateOrUpdateSenderRequestMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CreateOrUpdateSenderRequestMutation, CreateOrUpdateSenderRequestMutationVariables>(CreateOrUpdateSenderRequestDocument, options);
      }
export type CreateOrUpdateSenderRequestMutationHookResult = ReturnType<typeof useCreateOrUpdateSenderRequestMutation>;
export type CreateOrUpdateSenderRequestMutationResult = Apollo.MutationResult<CreateOrUpdateSenderRequestMutation>;
export type CreateOrUpdateSenderRequestMutationOptions = Apollo.BaseMutationOptions<CreateOrUpdateSenderRequestMutation, CreateOrUpdateSenderRequestMutationVariables>;
export const CreateSenderRequestFromHttpRequestLogDocument = gql`
    mutation CreateSenderRequestFromHttpRequestLog($id: ID!) {
  createSenderRequestFromHttpRequestLog(id: $id) {
    id
  }
}
    `;
export type CreateSenderRequestFromHttpRequestLogMutationFn = Apollo.MutationFunction<CreateSenderRequestFromHttpRequestLogMutation, CreateSenderRequestFromHttpRequestLogMutationVariables>;

/**
 * __useCreateSenderRequestFromHttpRequestLogMutation__
 *
 * To run a mutation, you first call `useCreateSenderRequestFromHttpRequestLogMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateSenderRequestFromHttpRequestLogMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createSenderRequestFromHttpRequestLogMutation, { data, loading, error }] = useCreateSenderRequestFromHttpRequestLogMutation({
 *   variables: {
 *      id: // value for 'id'
 *   },
 * });
 */
export function useCreateSenderRequestFromHttpRequestLogMutation(baseOptions?: Apollo.MutationHookOptions<CreateSenderRequestFromHttpRequestLogMutation, CreateSenderRequestFromHttpRequestLogMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CreateSenderRequestFromHttpRequestLogMutation, CreateSenderRequestFromHttpRequestLogMutationVariables>(CreateSenderRequestFromHttpRequestLogDocument, options);
      }
export type CreateSenderRequestFromHttpRequestLogMutationHookResult = ReturnType<typeof useCreateSenderRequestFromHttpRequestLogMutation>;
export type CreateSenderRequestFromHttpRequestLogMutationResult = Apollo.MutationResult<CreateSenderRequestFromHttpRequestLogMutation>;
export type CreateSenderRequestFromHttpRequestLogMutationOptions = Apollo.BaseMutationOptions<CreateSenderRequestFromHttpRequestLogMutation, CreateSenderRequestFromHttpRequestLogMutationVariables>;
export const SendRequestDocument = gql`
    mutation SendRequest($id: ID!) {
  sendRequest(id: $id) {
    id
  }
}
    `;
export type SendRequestMutationFn = Apollo.MutationFunction<SendRequestMutation, SendRequestMutationVariables>;

/**
 * __useSendRequestMutation__
 *
 * To run a mutation, you first call `useSendRequestMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useSendRequestMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [sendRequestMutation, { data, loading, error }] = useSendRequestMutation({
 *   variables: {
 *      id: // value for 'id'
 *   },
 * });
 */
export function useSendRequestMutation(baseOptions?: Apollo.MutationHookOptions<SendRequestMutation, SendRequestMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<SendRequestMutation, SendRequestMutationVariables>(SendRequestDocument, options);
      }
export type SendRequestMutationHookResult = ReturnType<typeof useSendRequestMutation>;
export type SendRequestMutationResult = Apollo.MutationResult<SendRequestMutation>;
export type SendRequestMutationOptions = Apollo.BaseMutationOptions<SendRequestMutation, SendRequestMutationVariables>;
export const GetSenderRequestDocument = gql`
    query GetSenderRequest($id: ID!) {
  senderRequest(id: $id) {
    id
    sourceRequestLogID
    url
    method
    proto
    headers {
      key
      value
    }
    body
    timestamp
    response {
      id
      proto
      statusCode
      statusReason
      body
      headers {
        key
        value
      }
    }
  }
}
    `;

/**
 * __useGetSenderRequestQuery__
 *
 * To run a query within a React component, call `useGetSenderRequestQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetSenderRequestQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetSenderRequestQuery({
 *   variables: {
 *      id: // value for 'id'
 *   },
 * });
 */
export function useGetSenderRequestQuery(baseOptions: Apollo.QueryHookOptions<GetSenderRequestQuery, GetSenderRequestQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetSenderRequestQuery, GetSenderRequestQueryVariables>(GetSenderRequestDocument, options);
      }
export function useGetSenderRequestLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetSenderRequestQuery, GetSenderRequestQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetSenderRequestQuery, GetSenderRequestQueryVariables>(GetSenderRequestDocument, options);
        }
export type GetSenderRequestQueryHookResult = ReturnType<typeof useGetSenderRequestQuery>;
export type GetSenderRequestLazyQueryHookResult = ReturnType<typeof useGetSenderRequestLazyQuery>;
export type GetSenderRequestQueryResult = Apollo.QueryResult<GetSenderRequestQuery, GetSenderRequestQueryVariables>;
export const GetSenderRequestsDocument = gql`
    query GetSenderRequests {
  senderRequests {
    id
    url
    method
    response {
      id
      statusCode
      statusReason
    }
  }
}
    `;

/**
 * __useGetSenderRequestsQuery__
 *
 * To run a query within a React component, call `useGetSenderRequestsQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetSenderRequestsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetSenderRequestsQuery({
 *   variables: {
 *   },
 * });
 */
export function useGetSenderRequestsQuery(baseOptions?: Apollo.QueryHookOptions<GetSenderRequestsQuery, GetSenderRequestsQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetSenderRequestsQuery, GetSenderRequestsQueryVariables>(GetSenderRequestsDocument, options);
      }
export function useGetSenderRequestsLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetSenderRequestsQuery, GetSenderRequestsQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetSenderRequestsQuery, GetSenderRequestsQueryVariables>(GetSenderRequestsDocument, options);
        }
export type GetSenderRequestsQueryHookResult = ReturnType<typeof useGetSenderRequestsQuery>;
export type GetSenderRequestsLazyQueryHookResult = ReturnType<typeof useGetSenderRequestsLazyQuery>;
export type GetSenderRequestsQueryResult = Apollo.QueryResult<GetSenderRequestsQuery, GetSenderRequestsQueryVariables>;
export const UpdateInterceptSettingsDocument = gql`
    mutation UpdateInterceptSettings($input: UpdateInterceptSettingsInput!) {
  updateInterceptSettings(input: $input) {
    requestsEnabled
    responsesEnabled
    requestFilter
    responseFilter
  }
}
    `;
export type UpdateInterceptSettingsMutationFn = Apollo.MutationFunction<UpdateInterceptSettingsMutation, UpdateInterceptSettingsMutationVariables>;

/**
 * __useUpdateInterceptSettingsMutation__
 *
 * To run a mutation, you first call `useUpdateInterceptSettingsMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useUpdateInterceptSettingsMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [updateInterceptSettingsMutation, { data, loading, error }] = useUpdateInterceptSettingsMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useUpdateInterceptSettingsMutation(baseOptions?: Apollo.MutationHookOptions<UpdateInterceptSettingsMutation, UpdateInterceptSettingsMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<UpdateInterceptSettingsMutation, UpdateInterceptSettingsMutationVariables>(UpdateInterceptSettingsDocument, options);
      }
export type UpdateInterceptSettingsMutationHookResult = ReturnType<typeof useUpdateInterceptSettingsMutation>;
export type UpdateInterceptSettingsMutationResult = Apollo.MutationResult<UpdateInterceptSettingsMutation>;
export type UpdateInterceptSettingsMutationOptions = Apollo.BaseMutationOptions<UpdateInterceptSettingsMutation, UpdateInterceptSettingsMutationVariables>;
export const SitemapDocument = gql`
    query Sitemap {
  sitemap {
    host
    path
    methods
    statusCodes
    count
  }
}
    `;

/**
 * __useSitemapQuery__
 *
 * To run a query within a React component, call `useSitemapQuery` and pass it any options that fit your needs.
 * When your component renders, `useSitemapQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useSitemapQuery({
 *   variables: {
 *   },
 * });
 */
export function useSitemapQuery(baseOptions?: Apollo.QueryHookOptions<SitemapQuery, SitemapQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<SitemapQuery, SitemapQueryVariables>(SitemapDocument, options);
      }
export function useSitemapLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<SitemapQuery, SitemapQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<SitemapQuery, SitemapQueryVariables>(SitemapDocument, options);
        }
export type SitemapQueryHookResult = ReturnType<typeof useSitemapQuery>;
export type SitemapLazyQueryHookResult = ReturnType<typeof useSitemapLazyQuery>;
export type SitemapQueryResult = Apollo.QueryResult<SitemapQuery, SitemapQueryVariables>;
export const WebSocketConnectionsDocument = gql`
    query WebSocketConnections($searchExpression: String) {
  webSocketConnections(searchExpression: $searchExpression) {
    id
    url
    host
    path
    timestamp
    closedAt
    messageCount
  }
}
    `;

/**
 * __useWebSocketConnectionsQuery__
 *
 * To run a query within a React component, call `useWebSocketConnectionsQuery` and pass it any options that fit your needs.
 * When your component renders, `useWebSocketConnectionsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useWebSocketConnectionsQuery({
 *   variables: {
 *      searchExpression: // value for 'searchExpression'
 *   },
 * });
 */
export function useWebSocketConnectionsQuery(baseOptions?: Apollo.QueryHookOptions<WebSocketConnectionsQuery, WebSocketConnectionsQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<WebSocketConnectionsQuery, WebSocketConnectionsQueryVariables>(WebSocketConnectionsDocument, options);
      }
export function useWebSocketConnectionsLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<WebSocketConnectionsQuery, WebSocketConnectionsQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<WebSocketConnectionsQuery, WebSocketConnectionsQueryVariables>(WebSocketConnectionsDocument, options);
        }
export type WebSocketConnectionsQueryHookResult = ReturnType<typeof useWebSocketConnectionsQuery>;
export type WebSocketConnectionsLazyQueryHookResult = ReturnType<typeof useWebSocketConnectionsLazyQuery>;
export type WebSocketConnectionsQueryResult = Apollo.QueryResult<WebSocketConnectionsQuery, WebSocketConnectionsQueryVariables>;
export const WebSocketMessagesDocument = gql`
    query WebSocketMessages($connectionId: ID!, $searchExpression: String) {
  webSocketMessages(
    connectionId: $connectionId
    searchExpression: $searchExpression
  ) {
    direction
    opcode
    payload
    timestamp
  }
}
    `;

/**
 * __useWebSocketMessagesQuery__
 *
 * To run a query within a React component, call `useWebSocketMessagesQuery` and pass it any options that fit your needs.
 * When your component renders, `useWebSocketMessagesQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useWebSocketMessagesQuery({
 *   variables: {
 *      connectionId: // value for 'connectionId'
 *      searchExpression: // value for 'searchExpression'
 *   },
 * });
 */
export function useWebSocketMessagesQuery(baseOptions: Apollo.QueryHookOptions<WebSocketMessagesQuery, WebSocketMessagesQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<WebSocketMessagesQuery, WebSocketMessagesQueryVariables>(WebSocketMessagesDocument, options);
      }
export function useWebSocketMessagesLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<WebSocketMessagesQuery, WebSocketMessagesQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<WebSocketMessagesQuery, WebSocketMessagesQueryVariables>(WebSocketMessagesDocument, options);
        }
export type WebSocketMessagesQueryHookResult = ReturnType<typeof useWebSocketMessagesQuery>;
export type WebSocketMessagesLazyQueryHookResult = ReturnType<typeof useWebSocketMessagesLazyQuery>;
export type WebSocketMessagesQueryResult = Apollo.QueryResult<WebSocketMessagesQuery, WebSocketMessagesQueryVariables>;
export const WebSocketInterceptSettingsDocument = gql`
    query WebSocketInterceptSettings {
  webSocketInterceptSettings {
    enabled
    filter
  }
}
    `;

/**
 * __useWebSocketInterceptSettingsQuery__
 *
 * To run a query within a React component, call `useWebSocketInterceptSettingsQuery` and pass it any options that fit your needs.
 * When your component renders, `useWebSocketInterceptSettingsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useWebSocketInterceptSettingsQuery({
 *   variables: {
 *   },
 * });
 */
export function useWebSocketInterceptSettingsQuery(baseOptions?: Apollo.QueryHookOptions<WebSocketInterceptSettingsQuery, WebSocketInterceptSettingsQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<WebSocketInterceptSettingsQuery, WebSocketInterceptSettingsQueryVariables>(WebSocketInterceptSettingsDocument, options);
      }
export function useWebSocketInterceptSettingsLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<WebSocketInterceptSettingsQuery, WebSocketInterceptSettingsQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<WebSocketInterceptSettingsQuery, WebSocketInterceptSettingsQueryVariables>(WebSocketInterceptSettingsDocument, options);
        }
export type WebSocketInterceptSettingsQueryHookResult = ReturnType<typeof useWebSocketInterceptSettingsQuery>;
export type WebSocketInterceptSettingsLazyQueryHookResult = ReturnType<typeof useWebSocketInterceptSettingsLazyQuery>;
export type WebSocketInterceptSettingsQueryResult = Apollo.QueryResult<WebSocketInterceptSettingsQuery, WebSocketInterceptSettingsQueryVariables>;
export const InterceptedWebSocketMessagesDocument = gql`
    query InterceptedWebSocketMessages {
  interceptedWebSocketMessages {
    id
    connectionId
    direction
    opcode
    payload
  }
}
    `;

/**
 * __useInterceptedWebSocketMessagesQuery__
 *
 * To run a query within a React component, call `useInterceptedWebSocketMessagesQuery` and pass it any options that fit your needs.
 * When your component renders, `useInterceptedWebSocketMessagesQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useInterceptedWebSocketMessagesQuery({
 *   variables: {
 *   },
 * });
 */
export function useInterceptedWebSocketMessagesQuery(baseOptions?: Apollo.QueryHookOptions<InterceptedWebSocketMessagesQuery, InterceptedWebSocketMessagesQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<InterceptedWebSocketMessagesQuery, InterceptedWebSocketMessagesQueryVariables>(InterceptedWebSocketMessagesDocument, options);
      }
export function useInterceptedWebSocketMessagesLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<InterceptedWebSocketMessagesQuery, InterceptedWebSocketMessagesQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<InterceptedWebSocketMessagesQuery, InterceptedWebSocketMessagesQueryVariables>(InterceptedWebSocketMessagesDocument, options);
        }
export type InterceptedWebSocketMessagesQueryHookResult = ReturnType<typeof useInterceptedWebSocketMessagesQuery>;
export type InterceptedWebSocketMessagesLazyQueryHookResult = ReturnType<typeof useInterceptedWebSocketMessagesLazyQuery>;
export type InterceptedWebSocketMessagesQueryResult = Apollo.QueryResult<InterceptedWebSocketMessagesQuery, InterceptedWebSocketMessagesQueryVariables>;
export const UpdateWebSocketInterceptSettingsDocument = gql`
    mutation UpdateWebSocketInterceptSettings($input: UpdateWebSocketInterceptSettingsInput!) {
  updateWebSocketInterceptSettings(input: $input) {
    enabled
    filter
  }
}
    `;
export type UpdateWebSocketInterceptSettingsMutationFn = Apollo.MutationFunction<UpdateWebSocketInterceptSettingsMutation, UpdateWebSocketInterceptSettingsMutationVariables>;

/**
 * __useUpdateWebSocketInterceptSettingsMutation__
 *
 * To run a mutation, you first call `useUpdateWebSocketInterceptSettingsMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useUpdateWebSocketInterceptSettingsMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [updateWebSocketInterceptSettingsMutation, { data, loading, error }] = useUpdateWebSocketInterceptSettingsMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useUpdateWebSocketInterceptSettingsMutation(baseOptions?: Apollo.MutationHookOptions<UpdateWebSocketInterceptSettingsMutation, UpdateWebSocketInterceptSettingsMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<UpdateWebSocketInterceptSettingsMutation, UpdateWebSocketInterceptSettingsMutationVariables>(UpdateWebSocketInterceptSettingsDocument, options);
      }
export type UpdateWebSocketInterceptSettingsMutationHookResult = ReturnType<typeof useUpdateWebSocketInterceptSettingsMutation>;
export type UpdateWebSocketInterceptSettingsMutationResult = Apollo.MutationResult<UpdateWebSocketInterceptSettingsMutation>;
export type UpdateWebSocketInterceptSettingsMutationOptions = Apollo.BaseMutationOptions<UpdateWebSocketInterceptSettingsMutation, UpdateWebSocketInterceptSettingsMutationVariables>;
export const ModifyWebSocketMessageDocument = gql`
    mutation ModifyWebSocketMessage($input: ModifyWebSocketMessageInput!) {
  modifyWebSocketMessage(input: $input) {
    success
  }
}
    `;
export type ModifyWebSocketMessageMutationFn = Apollo.MutationFunction<ModifyWebSocketMessageMutation, ModifyWebSocketMessageMutationVariables>;

/**
 * __useModifyWebSocketMessageMutation__
 *
 * To run a mutation, you first call `useModifyWebSocketMessageMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useModifyWebSocketMessageMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [modifyWebSocketMessageMutation, { data, loading, error }] = useModifyWebSocketMessageMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useModifyWebSocketMessageMutation(baseOptions?: Apollo.MutationHookOptions<ModifyWebSocketMessageMutation, ModifyWebSocketMessageMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<ModifyWebSocketMessageMutation, ModifyWebSocketMessageMutationVariables>(ModifyWebSocketMessageDocument, options);
      }
export type ModifyWebSocketMessageMutationHookResult = ReturnType<typeof useModifyWebSocketMessageMutation>;
export type ModifyWebSocketMessageMutationResult = Apollo.MutationResult<ModifyWebSocketMessageMutation>;
export type ModifyWebSocketMessageMutationOptions = Apollo.BaseMutationOptions<ModifyWebSocketMessageMutation, ModifyWebSocketMessageMutationVariables>;
export const ForwardWebSocketMessageDocument = gql`
    mutation ForwardWebSocketMessage($id: ID!) {
  forwardWebSocketMessage(id: $id) {
    success
  }
}
    `;
export type ForwardWebSocketMessageMutationFn = Apollo.MutationFunction<ForwardWebSocketMessageMutation, ForwardWebSocketMessageMutationVariables>;

/**
 * __useForwardWebSocketMessageMutation__
 *
 * To run a mutation, you first call `useForwardWebSocketMessageMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useForwardWebSocketMessageMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [forwardWebSocketMessageMutation, { data, loading, error }] = useForwardWebSocketMessageMutation({
 *   variables: {
 *      id: // value for 'id'
 *   },
 * });
 */
export function useForwardWebSocketMessageMutation(baseOptions?: Apollo.MutationHookOptions<ForwardWebSocketMessageMutation, ForwardWebSocketMessageMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<ForwardWebSocketMessageMutation, ForwardWebSocketMessageMutationVariables>(ForwardWebSocketMessageDocument, options);
      }
export type ForwardWebSocketMessageMutationHookResult = ReturnType<typeof useForwardWebSocketMessageMutation>;
export type ForwardWebSocketMessageMutationResult = Apollo.MutationResult<ForwardWebSocketMessageMutation>;
export type ForwardWebSocketMessageMutationOptions = Apollo.BaseMutationOptions<ForwardWebSocketMessageMutation, ForwardWebSocketMessageMutationVariables>;
export const DropWebSocketMessageDocument = gql`
    mutation DropWebSocketMessage($id: ID!) {
  dropWebSocketMessage(id: $id) {
    success
  }
}
    `;
export type DropWebSocketMessageMutationFn = Apollo.MutationFunction<DropWebSocketMessageMutation, DropWebSocketMessageMutationVariables>;

/**
 * __useDropWebSocketMessageMutation__
 *
 * To run a mutation, you first call `useDropWebSocketMessageMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useDropWebSocketMessageMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [dropWebSocketMessageMutation, { data, loading, error }] = useDropWebSocketMessageMutation({
 *   variables: {
 *      id: // value for 'id'
 *   },
 * });
 */
export function useDropWebSocketMessageMutation(baseOptions?: Apollo.MutationHookOptions<DropWebSocketMessageMutation, DropWebSocketMessageMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<DropWebSocketMessageMutation, DropWebSocketMessageMutationVariables>(DropWebSocketMessageDocument, options);
      }
export type DropWebSocketMessageMutationHookResult = ReturnType<typeof useDropWebSocketMessageMutation>;
export type DropWebSocketMessageMutationResult = Apollo.MutationResult<DropWebSocketMessageMutation>;
export type DropWebSocketMessageMutationOptions = Apollo.BaseMutationOptions<DropWebSocketMessageMutation, DropWebSocketMessageMutationVariables>;
export const GetInterceptedRequestsDocument = gql`
    query GetInterceptedRequests {
  interceptedRequests {
    id
    url
    method
    response {
      statusCode
      statusReason
    }
  }
}
    `;

/**
 * __useGetInterceptedRequestsQuery__
 *
 * To run a query within a React component, call `useGetInterceptedRequestsQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetInterceptedRequestsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetInterceptedRequestsQuery({
 *   variables: {
 *   },
 * });
 */
export function useGetInterceptedRequestsQuery(baseOptions?: Apollo.QueryHookOptions<GetInterceptedRequestsQuery, GetInterceptedRequestsQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetInterceptedRequestsQuery, GetInterceptedRequestsQueryVariables>(GetInterceptedRequestsDocument, options);
      }
export function useGetInterceptedRequestsLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetInterceptedRequestsQuery, GetInterceptedRequestsQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetInterceptedRequestsQuery, GetInterceptedRequestsQueryVariables>(GetInterceptedRequestsDocument, options);
        }
export type GetInterceptedRequestsQueryHookResult = ReturnType<typeof useGetInterceptedRequestsQuery>;
export type GetInterceptedRequestsLazyQueryHookResult = ReturnType<typeof useGetInterceptedRequestsLazyQuery>;
export type GetInterceptedRequestsQueryResult = Apollo.QueryResult<GetInterceptedRequestsQuery, GetInterceptedRequestsQueryVariables>;