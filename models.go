package meshapi

import (
	"context"
	"net/url"
)

// ModelsResource provides access to the /v1/models endpoints.
type ModelsResource struct {
	http *httpClient
}

// List returns all available models. Pass a non-nil Free pointer to filter.
func (r *ModelsResource) List(ctx context.Context, params ListModelsParams) ([]ModelInfo, error) {
	qs := url.Values{}
	if params.Free != nil {
		if *params.Free {
			qs.Set("free", "true")
		} else {
			qs.Set("free", "false")
		}
	}
	var models []ModelInfo
	if err := r.http.get(ctx, "/v1/models", qs, &models); err != nil {
		return nil, err
	}
	return models, nil
}

// Free returns only free-tier models.
func (r *ModelsResource) Free(ctx context.Context) ([]ModelInfo, error) {
	var models []ModelInfo
	if err := r.http.get(ctx, "/v1/models/free", nil, &models); err != nil {
		return nil, err
	}
	return models, nil
}

// Paid returns only paid-tier models.
func (r *ModelsResource) Paid(ctx context.Context) ([]ModelInfo, error) {
	var models []ModelInfo
	if err := r.http.get(ctx, "/v1/models/paid", nil, &models); err != nil {
		return nil, err
	}
	return models, nil
}
