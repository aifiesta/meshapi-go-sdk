package meshapi

import "context"

// WebResource provides access to POST /v1/web/search.
//
// Gated server-side by WEB_SEARCH_ENABLED; when disabled the endpoint returns
// an error rather than results. Failover between the native engine and Tavily
// is opaque — inspect WebSearchResponse.Provider to see which engine served
// the request.
type WebResource struct {
	http *httpClient
}

// Search runs a live web search.
func (r *WebResource) Search(ctx context.Context, params WebSearchParams) (*WebSearchResponse, error) {
	var out WebSearchResponse
	if err := r.http.post(ctx, "/v1/web/search", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
