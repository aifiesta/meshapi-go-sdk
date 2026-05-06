package meshapi

import "context"

type CompareResource struct {
	http *httpClient
}

func (r *CompareResource) Create(ctx context.Context, params CompareParams) (*CompareResponse, error) {
	f := false
	params.Stream = &f
	var out CompareResponse
	if err := r.http.post(ctx, "/v1/chat/compare", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *CompareResource) Stream(ctx context.Context, params CompareParams) (<-chan CompareStreamEvent, <-chan error) {
	t := true
	params.Stream = &t
	chunkCh := make(chan CompareStreamEvent)
	errCh := make(chan error, 1)
	go func() {
		resp, err := r.http.stream(ctx, "/v1/chat/compare", params)
		if err != nil {
			close(chunkCh)
			errCh <- err
			close(errCh)
			return
		}
		parseJSONSSEStream(resp, chunkCh, errCh)
	}()
	return chunkCh, errCh
}
