package meshapi

import "context"

// WebResource provides access to POST /v1/web/search.
//
// Gated server-side by WEB_SEARCH_ENABLED; when disabled the endpoint returns
// an error rather than results. The native engine is tried first and failover
// to another engine is opaque — inspect WebSearchResponse.Provider to see which
// one served. Pinning WebSearchParams.Provider turns failover off.
//
// Set WebSearchParams.IncludePageContent to get each result's extracted page
// text alongside the snippet. Only the tinyfish engine returns it, so pin that
// engine when you need it.
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
