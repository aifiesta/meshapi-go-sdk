package meshapi

import "context"

type EmbeddingsResource struct {
	http *httpClient
}

func (r *EmbeddingsResource) Create(ctx context.Context, params EmbeddingsParams) (*EmbeddingsResponse, error) {
	var out EmbeddingsResponse
	if err := r.http.post(ctx, "/v1/embeddings", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
