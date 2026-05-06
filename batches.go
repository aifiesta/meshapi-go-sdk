package meshapi

import (
	"context"
	"net/url"
	"strconv"
)

type BatchesResource struct {
	http *httpClient
}

func (r *BatchesResource) Create(ctx context.Context, params CreateBatchParams) (*BatchObject, error) {
	var out BatchObject
	if err := r.http.post(ctx, "/v1/batches", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *BatchesResource) List(ctx context.Context, after *string, limit *int) (*BatchListResponse, error) {
	params := url.Values{}
	if after != nil {
		params.Set("after", *after)
	}
	if limit != nil {
		params.Set("limit", strconv.Itoa(*limit))
	}
	var out BatchListResponse
	if err := r.http.get(ctx, "/v1/batches", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *BatchesResource) Get(ctx context.Context, batchID string) (*BatchObject, error) {
	var out BatchObject
	if err := r.http.get(ctx, "/v1/batches/"+batchID, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *BatchesResource) Cancel(ctx context.Context, batchID string) (*BatchObject, error) {
	var out BatchObject
	if err := r.http.post(ctx, "/v1/batches/"+batchID+"/cancel", map[string]interface{}{}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
