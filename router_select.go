package meshapi

import "context"

// RouterResource provides access to POST /v1/router/select.
//
// Select-only Auto Router: returns the model the Auto Router would pick for a
// prompt without running inference, so the caller can run inference on its own
// path. Gated server-side by AUTO_ROUTER_ENABLED. Fail-soft: on classification
// failure the router returns the configured default model with
// AutoRouterMeta.FallbackUsed = true rather than erroring.
type RouterResource struct {
	http *httpClient
}

// Select returns the model the Auto Router would pick for the given messages.
func (r *RouterResource) Select(ctx context.Context, params RouterSelectParams) (*RouterSelectResponse, error) {
	var out RouterSelectResponse
	if err := r.http.post(ctx, "/v1/router/select", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
