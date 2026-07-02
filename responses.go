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
