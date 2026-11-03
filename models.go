package meshapi

import (
	"context"
)

// ModelsResource provides access to the model catalog at GET /v1/models.
//
// Each model in the catalog carries a ModelPricing block. As of spec
// v1.0.378 the pricing block may also publish
// long_context_cache_write_1h_input_usd_per_1m: the per-million-token
// price of a 1-hour cache write on a prompt past the model's long-context
// threshold. The field is optional and is null or absent for models that
// do not price a long-context 1-hour cache write.
type ModelsResource struct {
	http *httpClient
}

// List returns every model currently available in the catalog.
//
// Use only the model ids returned here in requests; models retired from
// the catalog are not listed.
func (r *ModelsResource) List(ctx context.Context) (*ModelList, error) {
	var out ModelList
	if err := r.http.get(ctx, "/v1/models", &out); err != nil {
		return nil, err
	}
	return &out, nil
}
