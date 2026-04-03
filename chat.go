package meshapi

import "context"

// CompletionsResource handles POST /v1/chat/completions.
type CompletionsResource struct {
	http *httpClient
}

// Create sends a non-streaming chat completion request and returns the full response.
func (r *CompletionsResource) Create(ctx context.Context, params ChatCompletionParams) (*ChatCompletionResponse, error) {
	f := false
	params.Stream = &f
	var out ChatCompletionResponse
	if err := r.http.post(ctx, "/v1/chat/completions", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Stream opens a streaming chat completion. It returns two channels:
//   - chunkCh: receives parsed ChatCompletionChunks until [DONE] or error
//   - errCh:   receives at most one error, then is closed
//
// Both channels are always closed when the stream finishes. Callers must
// drain chunkCh before reading errCh, or use a select loop.
//
// Streams are NEVER retried. On failure, catch the error from errCh and
// restart a new Stream call if reconnection is needed.
func (r *CompletionsResource) Stream(ctx context.Context, params ChatCompletionParams) (<-chan ChatCompletionChunk, <-chan error) {
	t := true
	params.Stream = &t

	chunkCh := make(chan ChatCompletionChunk)
	errCh := make(chan error, 1)

	go func() {
		resp, err := r.http.stream(ctx, "/v1/chat/completions", params)
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

// ChatResource groups chat-related sub-resources.
type ChatResource struct {
	Completions *CompletionsResource
}
