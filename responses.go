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

// ---------------------------------------------------------------------------
// Input token counting (POST /v1/responses/input_tokens)
// ---------------------------------------------------------------------------

// ResponsesInputTokensParams is the request body for Responses.InputTokens.
//
// It takes the same input shape as a Responses create request. The gateway
// counts the prompt tokens without running inference, so the call is not
// billed and records no usage.
type ResponsesInputTokensParams struct {
	// Model is the model id to count against, e.g. "openai/gpt-4o-mini".
	Model string `json:"model"`
	// Input is a string or a list of Responses input items.
	Input any `json:"input,omitempty"`
	// Instructions is an optional system-level instruction string.
	Instructions *string `json:"instructions,omitempty"`
	// Tools lists tool definitions that count toward the prompt.
	Tools []any `json:"tools,omitempty"`
	// ToolChoice is an optional tool choice (string or object).
	ToolChoice any `json:"tool_choice,omitempty"`
	// PreviousResponseID continues from a stored response you own.
	PreviousResponseID *string `json:"previous_response_id,omitempty"`
}

// ResponsesInputTokensResponse is the result of Responses.InputTokens.
type ResponsesInputTokensResponse struct {
	// Object is "response.input_tokens".
	Object string `json:"object"`
	// InputTokens is the number of input tokens the request would consume.
	InputTokens int `json:"input_tokens"`
}

// Create sends a non-streaming Responses request.
func (r *ResponsesResource) Create(ctx context.Context, params ResponsesParams) (*ResponsesResponse, error) {
	f := false
	params.Stream = &f
	var out ResponsesResponse
	if err := r.http.post(ctx, "/v1/responses", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Stream sends a streaming Responses request and returns event and error channels.
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

// InputTokens counts the input tokens a Responses request would use, without
// running inference. The call is not billed and writes no usage record.
//
// It calls POST /v1/responses/input_tokens.
func (r *ResponsesResource) InputTokens(ctx context.Context, params ResponsesInputTokensParams) (*ResponsesInputTokensResponse, error) {
	var out ResponsesInputTokensResponse
	if err := r.http.post(ctx, "/v1/responses/input_tokens", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
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
