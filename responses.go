package meshapi

import "context"

type ResponsesResource struct {
	http *httpClient
}

func (r *ResponsesResource) Create(ctx context.Context, params ResponsesParams) (*ResponsesResponse, error) {
	f := false
	params.Stream = &f
	var out ResponsesResponse
	if err := r.http.post(ctx, "/v1/responses", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *ResponsesResource) Stream(ctx context.Context, params ResponsesParams) (<-chan ResponsesStreamEvent, <-chan error) {
	t := true
	params.Stream = &t
	chunkCh := make(chan ResponsesStreamEvent)
	errCh := make(chan error, 1)
	go func() {
		resp, err := r.http.stream(ctx, "/v1/responses", params)
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
