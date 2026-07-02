package meshapi

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
)

type ImagesResource struct {
	http *httpClient
}

// Bytes returns the raw image bytes regardless of how the provider returned
// them. It handles both B64JSON and a data: URI in URL (some models, e.g.
// openai/gpt-image-1, inline the image as a data URL rather than populating
// B64JSON). It returns an error for a remote http(s) URL (fetch it yourself)
// or when no image data is present.
func (i *ImageItem) Bytes() ([]byte, error) {
	if i.B64JSON != nil && *i.B64JSON != "" {
		return base64.StdEncoding.DecodeString(*i.B64JSON)
	}
	if i.URL != nil && strings.HasPrefix(*i.URL, "data:") {
		parts := strings.SplitN(*i.URL, ",", 2)
		if len(parts) == 2 {
			return base64.StdEncoding.DecodeString(parts[1])
		}
	}
	if i.URL != nil && *i.URL != "" {
		return nil, fmt.Errorf("image is a remote URL; fetch it with an HTTP client: %s", *i.URL)
	}
	return nil, fmt.Errorf("image item has neither b64_json nor url")
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
