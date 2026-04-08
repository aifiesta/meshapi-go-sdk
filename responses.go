package meshapi

import "context"

// ResponsesResource handles POST /v1/responses.
type ResponsesResource struct {
	http *httpClient
}

// Create sends a non-streaming request and returns the full response.
func (r *ResponsesResource) Create(ctx context.Context, params ResponsesParams) (*ResponsesResponse, error) {
	f := false
	params.Stream = &f
	var out ResponsesResponse
	if err := r.http.post(ctx, "/v1/responses", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Stream opens a streaming request. It returns two channels:
//   - chunkCh: receives ChatCompletionChunks (identical SSE format to chat/completions)
//   - errCh:   receives at most one error, then is closed
//
// Both channels are always closed when the stream ends. Callers must
// drain chunkCh before reading errCh, or use a select loop.
//
// Streams are NEVER retried. On failure, catch the error from errCh and
// restart a new Stream call if reconnection is needed.
func (r *ResponsesResource) Stream(ctx context.Context, params ResponsesParams) (<-chan ChatCompletionChunk, <-chan error) {
	t := true
	params.Stream = &t

	chunkCh := make(chan ChatCompletionChunk)
	errCh := make(chan error, 1)

	go func() {
		resp, err := r.http.stream(ctx, "/v1/responses", params)
		if err != nil {
			close(chunkCh)
			errCh <- err
			close(errCh)
			return
		}
		parseSSEStream(resp, chunkCh, errCh)
	}()

	return chunkCh, errCh
}
