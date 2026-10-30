package meshapi

import (
	"context"
	"net/url"
	"strconv"
)

type ResponsesResource struct {
	http *httpClient
}

// ResponsesListParams holds query parameters for Responses.List.
type ResponsesListParams struct {
	After *string
	Limit *int
}

// ResponsesInputTokensParams is the request body for POST /v1/responses/input_tokens.
//
// The body follows the /v1/responses request shape. Input, tool and reasoning
// values are passed through as-is.
type ResponsesInputTokensParams struct {
	// Model is the model id to count against, e.g. "openai/gpt-4o-mini".
	Model string `json:"model"`
	// Input is the input to count (string or list of input items).
	Input any `json:"input,omitempty"`
	// Instructions is an optional system-level instruction.
	Instructions *string `json:"instructions,omitempty"`
	// Tools is an optional list of tool definitions.
	Tools []any `json:"tools,omitempty"`
	// ToolChoice optionally controls tool selection.
	ToolChoice any `json:"tool_choice,omitempty"`
	// PreviousResponseID optionally counts on top of a prior response you own.
	PreviousResponseID *string `json:"previous_response_id,omitempty"`
	// Reasoning optionally configures reasoning.
	Reasoning any `json:"reasoning,omitempty"`
	// Text optionally configures text output.
	Text any `json:"text,omitempty"`
	// Provider optionally carries provider routing preferences.
	Provider any `json:"provider,omitempty"`
}

// ResponsesInputTokensResponse is the response from POST /v1/responses/input_tokens.
type ResponsesInputTokensResponse struct {
	// Object is the object type, e.g. "response.input_tokens".
	Object string `json:"object,omitempty"`
	// InputTokens is the number of input tokens the request would consume.
	InputTokens int `json:"input_tokens"`
}

func (r *ResponsesResource) Create(ctx context.Context, params ResponsesParams) (*ResponsesResponse, error) {
	f := false
	params.Stream = &f
	var out ResponsesResponse
	if err := r.http.post(ctx, "/v1/responses", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *ResponsesResource) Stream(ctx context.Context, params ResponsesParams) (<-chan ResponsesStreamEvent, <-chan error) {
	t := true
	params.Stream = &t
	chunkCh := make(chan ResponsesStreamEvent)
	errCh := make(chan error, 1)
	go func() {
		resp, err := r.http.stream(ctx, "/v1/responses", params)
		if err != nil {
			close(chunkCh)
			errCh <- err
			close(errCh)
			return
		}
		parseJSONSSEStream(resp, chunkCh, errCh)
	}()
	return chunkCh, errCh
}

// List returns the caller's background response jobs (OpenAI list envelope).
func (r *ResponsesResource) List(ctx context.Context, params ResponsesListParams) (*ResponsesListResponse, error) {
	qs := url.Values{}
	if params.After != nil {
		qs.Set("after", *params.After)
	}
	if params.Limit != nil {
		qs.Set("limit", strconv.Itoa(*params.Limit))
	}
	var out ResponsesListResponse
	if err := r.http.get(ctx, "/v1/responses", qs, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get fetches a background response job by id.
func (r *ResponsesResource) Get(ctx context.Context, responseID string) (*ResponsesResponse, error) {
	var out ResponsesResponse
	if err := r.http.get(ctx, "/v1/responses/"+responseID, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// InputTokens counts the input tokens of a /v1/responses request without
// running inference. The call is not billed and does not create a usage record.
//
// The count is provider-exact when the upstream offers a native counter and a
// local estimate otherwise; the server reports which in the
// X-Mesh-Token-Count-Source response header.
func (r *ResponsesResource) InputTokens(ctx context.Context, params ResponsesInputTokensParams) (*ResponsesInputTokensResponse, error) {
	var out ResponsesInputTokensResponse
	if err := r.http.post(ctx, "/v1/responses/input_tokens", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
