package meshapi

import (
	"context"
	"net/url"
	"strconv"
)

// VideosResource provides access to /v1/video/generations endpoints.
type VideosResource struct {
	http *httpClient
}

// Generate submits a video generation task (POST /v1/video/generations).
func (r *VideosResource) Generate(ctx context.Context, params VideoGenerationParams) (*CreateVideoGenerationResponse, error) {
	var out CreateVideoGenerationResponse
	if err := r.http.post(ctx, "/v1/video/generations", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns video generation tasks (GET /v1/video/generations).
func (r *VideosResource) List(ctx context.Context, params *ListVideoGenerationsParams) (*VideoTaskListResponse, error) {
	qp := url.Values{}
	if params != nil {
		if params.Status != nil {
			qp.Set("status", *params.Status)
		}
		if params.Model != nil {
			qp.Set("model", *params.Model)
		}
		if params.CreatedAfter != nil {
			qp.Set("created_after", *params.CreatedAfter)
		}
		if params.CreatedBefore != nil {
			qp.Set("created_before", *params.CreatedBefore)
		}
		if params.Limit != nil {
			qp.Set("limit", strconv.Itoa(*params.Limit))
		}
		if params.Offset != nil {
			qp.Set("offset", strconv.Itoa(*params.Offset))
		}
	}
	var out VideoTaskListResponse
	if err := r.http.get(ctx, "/v1/video/generations", qp, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Retrieve fetches a single video generation task (GET /v1/video/generations/{task_id}).
func (r *VideosResource) Retrieve(ctx context.Context, taskID string) (*VideoTaskResponse, error) {
	var out VideoTaskResponse
	if err := r.http.get(ctx, "/v1/video/generations/"+taskID, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
