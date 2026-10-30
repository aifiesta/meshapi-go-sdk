package meshapi

import (
	"context"
)

// MessagesResource groups the Anthropic-compatible /v1/messages endpoints.
type MessagesResource struct {
	http *httpClient
}

// MessagesCountTokensParams is the request body for POST /v1/messages/count_tokens.
//
// The body follows the /v1/messages request shape. MaxTokens is optional on a
// count request. Message, tool and system entries are passed through as-is so
// any content block shape accepted by /v1/messages can be counted.
type MessagesCountTokensParams struct {
	// Model is the model id to count against, e.g. "anthropic/claude-haiku-4.5".
	Model string `json:"model"`
	// Messages is the conversation to count.
	Messages []any `json:"messages"`
	// System is an optional system prompt (string or list of content blocks).
	System any `json:"system,omitempty"`
	// Tools is an optional list of tool definitions.
	Tools []any `json:"tools,omitempty"`
	// ToolChoice optionally controls tool selection.
	ToolChoice any `json:"tool_choice,omitempty"`
	// Thinking optionally configures extended thinking.
	Thinking any `json:"thinking,omitempty"`
	// MaxTokens is optional on a count request.
	MaxTokens *int `json:"max_tokens,omitempty"`
	// MCPServers optionally lists MCP servers available to the request.
	MCPServers []any `json:"mcp_servers,omitempty"`
	// Container optionally names the container for the request.
	Container any `json:"container,omitempty"`
	// ServiceTier optionally selects the service tier.
	ServiceTier *string `json:"service_tier,omitempty"`
	// Provider optionally carries provider routing preferences.
	Provider any `json:"provider,omitempty"`
}

// MessagesCountTokensResponse is the response from POST /v1/messages/count_tokens.
type MessagesCountTokensResponse struct {
	// InputTokens is the number of input tokens the request would consume.
	InputTokens int `json:"input_tokens"`
}

// CountTokens counts the input tokens of a /v1/messages request without running
// inference. The call is not billed and does not create a usage record.
//
// The count is provider-exact when the upstream offers a native counter and a
// local estimate otherwise; the server reports which in the
// X-Mesh-Token-Count-Source response header.
func (r *MessagesResource) CountTokens(ctx context.Context, params MessagesCountTokensParams) (*MessagesCountTokensResponse, error) {
	var out MessagesCountTokensResponse
	if err := r.http.post(ctx, "/v1/messages/count_tokens", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
