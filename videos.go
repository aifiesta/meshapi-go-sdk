package meshapi

import "context"

// VideosResource provides access to the BytePlus Seedance video generation endpoints.
//
//   POST /v1/video/generations          — create an async task
//   GET  /v1/video/generations/{id}     — poll status / fetch result
type VideosResource struct {
	http *httpClient
}

// Create submits a video generation task and returns the task ID immediately.
//
// Video generation is asynchronous. Poll [Get] until Status is "succeeded",
// "failed", or "expired".
func (r *VideosResource) Create(ctx context.Context, params VideoGenerationParams) (*CreateVideoGenerationResponse, error) {
	var out CreateVideoGenerationResponse
	if err := r.http.post(ctx, "/v1/video/generations", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get retrieves the current status (and result) of a video generation task.
//
// When Status is "succeeded", Content.VideoURL is populated (valid for 24 h).
func (r *VideosResource) Get(ctx context.Context, taskID string) (*VideoTaskResponse, error) {
	var out VideoTaskResponse
	if err := r.http.get(ctx, "/v1/video/generations/"+taskID, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
