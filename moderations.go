package meshapi

import "context"

// ModerationsResource provides access to POST /v1/moderations.
type ModerationsResource struct {
	http *httpClient
}

// Create classifies the given input for policy violations.
func (r *ModerationsResource) Create(ctx context.Context, params ModerationParams) (*ModerationResponse, error) {
	var out ModerationResponse
	if err := r.http.post(ctx, "/v1/moderations", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
