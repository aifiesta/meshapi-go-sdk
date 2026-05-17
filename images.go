package meshapi

import "context"

type ImagesResource struct {
	http *httpClient
}

func (r *ImagesResource) Generate(ctx context.Context, params ImageGenerationParams) (*ImageGenerationResponse, error) {
	var out ImageGenerationResponse
	if err := r.http.post(ctx, "/v1/images/generations", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
