package meshapi

import (
	"context"
	"net/url"
	"strconv"
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

// Search returns a paginated, filtered page of the model catalog (DB-only).
func (r *ModelsResource) Search(ctx context.Context, params ModelSearchParams) (*ModelsPage, error) {
	qs := url.Values{}
	if params.Q != nil {
		qs.Set("q", *params.Q)
	}
	if params.Free != nil {
		qs.Set("free", strconv.FormatBool(*params.Free))
	}
	if params.Discounted != nil {
		qs.Set("discounted", strconv.FormatBool(*params.Discounted))
	}
	for _, m := range params.InputModality {
		qs.Add("input_modality", m)
	}
	for _, m := range params.OutputModality {
		qs.Add("output_modality", m)
	}
	for _, b := range params.Brand {
		qs.Add("brand", b)
	}
	if params.Sort != nil {
		qs.Set("sort", *params.Sort)
	}
	if params.Order != nil {
		qs.Set("order", *params.Order)
	}
	if params.Limit != nil {
		qs.Set("limit", strconv.Itoa(*params.Limit))
	}
	if params.Offset != nil {
		qs.Set("offset", strconv.Itoa(*params.Offset))
	}
	var page ModelsPage
	if err := r.http.get(ctx, "/v1/models/search", qs, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Get returns a single model's detail by id (e.g. "openai/gpt-4o").
func (r *ModelsResource) Get(ctx context.Context, modelID string) (*ModelInfo, error) {
	var m ModelInfo
	if err := r.http.get(ctx, "/v1/models/"+modelID, nil, &m); err != nil {
		return nil, err
	}
	return &m, nil
}
