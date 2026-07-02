package meshapi

import "context"

type ImagesResource struct {
	http *httpClient
}

// Generate sends a non-streaming image generation request and returns the full response.
func (r *ImagesResource) Generate(ctx context.Context, params ImageGenerationParams) (*ImageGenerationResponse, error) {
	var out ImageGenerationResponse
	if err := r.http.post(ctx, "/v1/images/generations", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Edit edits a source image (JSON/base64 mode). Image must be a base64 or
// data:-URL string (or ImageRef) — remote http(s) URLs are rejected.
func (r *ImagesResource) Edit(ctx context.Context, params ImageEditParams) (*ImageGenerationResponse, error) {
	var out ImageGenerationResponse
	if err := r.http.post(ctx, "/v1/images/edits", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Stream opens a streaming image generation request. It returns two channels:
//   - chunkCh: receives parsed ImageGenerationChunks until [DONE] or error
//   - errCh:   receives at most one error, then is closed
//
// Both channels are always closed when the stream finishes. Callers must
// drain chunkCh before reading errCh, or use a select loop.
func (r *ImagesResource) Stream(ctx context.Context, params ImageGenerationParams) (<-chan ImageGenerationChunk, <-chan error) {
	t := true
	params.Stream = &t

	chunkCh := make(chan ImageGenerationChunk)
	errCh := make(chan error, 1)

	go func() {
		resp, err := r.http.stream(ctx, "/v1/images/generations", params)
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
