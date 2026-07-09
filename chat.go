package meshapi

import (
	"context"
	"errors"
	"net"
)

// CompletionsResource handles POST /v1/chat/completions.
type CompletionsResource struct {
	http *httpClient
}

// Create sends a non-streaming chat completion request and returns the full
// response.
//
// When a fallback chain is configured (Config.Fallback.Models, or the per-call
// params.FallbackModels override), a transient failure of the primary model
// (default 502/503/504, after the transport's own retries) re-issues the
// request against each chain model in order. Terminal errors (auth,
// validation, billing, rate limit) never advance the chain. The gateway's
// server-side routing (per-key routing_policy) runs within each attempt and is
// reported separately via GatewayRoutingEvents.
func (r *CompletionsResource) Create(ctx context.Context, params ChatCompletionParams) (*ChatCompletionResponse, error) {
	f := false
	params.Stream = &f

	// FallbackModels is a client-side directive — never sent to the server
	// (json:"-" keeps it off the wire).
	chainSrc := params.FallbackModels
	if chainSrc == nil && r.http.cfg.Fallback != nil {
		chainSrc = r.http.cfg.Fallback.Models
	}
	primary := ""
	if params.Model != nil {
		primary = *params.Model
	}
	chain := make([]string, 0, len(chainSrc))
	for _, m := range chainSrc {
		if m != primary {
			chain = append(chain, m)
		}
	}
	onStatusList := DefaultFallbackStatusCodes
	if r.http.cfg.Fallback != nil && r.http.cfg.Fallback.OnStatus != nil {
		onStatusList = r.http.cfg.Fallback.OnStatus
	}
	onStatus := statusSet(onStatusList)

	var lastErr error
	// Model may be unset (the key's default_model applies server-side) —
	// label it for fallback events; the chain always names explicit models.
	fromModel := primary
	if fromModel == "" {
		fromModel = "(key default)"
	}
	for index := 0; index <= len(chain); index++ {
		attemptParams := params
		if index > 0 {
			model := chain[index-1]
			var status int
			var errorCode, requestID string
			var apiErr *MeshAPIError
			if errors.As(lastErr, &apiErr) {
				status = apiErr.Status
				errorCode = apiErr.Code
				requestID = apiErr.RequestID
			}
			r.http.emit(FallbackEvent{
				FromModel:   fromModel,
				ToModel:     model,
				ChainIndex:  index - 1,
				ChainLength: len(chain),
				Status:      status,
				ErrorCode:   errorCode,
				RequestID:   requestID,
			})
			attemptParams.Model = &model
		}
		var out ChatCompletionResponse
		err := r.http.post(ctx, "/v1/chat/completions", attemptParams, &out)
		if err == nil {
			return &out, nil
		}
		lastErr = err
		if index > 0 {
			fromModel = chain[index-1]
		} else if primary != "" {
			fromModel = primary
		}
		if len(chain) == 0 || !isFallbackEligible(err, onStatus) {
			return nil, err
		}
	}
	return nil, lastErr
}

// isFallbackEligible reports whether a failure is worth trying on another
// model: a transient API error (default 502/503/504 — a provider/gateway path
// problem, not this request) or a pre-response network error. Cancellations
// and timeouts always propagate; terminal API errors (4xx auth/validation/
// billing) never advance the chain — they would fail identically on every
// model.
func isFallbackEligible(err error, onStatus map[int]bool) bool {
	var apiErr *MeshAPIError
	if errors.As(err, &apiErr) {
		return onStatus[apiErr.Status]
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return false
	}
	return true
}

// Stream opens a streaming chat completion. It returns two channels:
//   - chunkCh: receives parsed ChatCompletionChunks until [DONE] or error
//   - errCh:   receives at most one error, then is closed
//
// Both channels are always closed when the stream finishes. Callers must
// drain chunkCh before reading errCh, or use a select loop.
//
// Streams are NEVER retried. On failure, catch the error from errCh and
// restart a new Stream call if reconnection is needed.
func (r *CompletionsResource) Stream(ctx context.Context, params ChatCompletionParams) (<-chan ChatCompletionChunk, <-chan error) {
	t := true
	params.Stream = &t

	chunkCh := make(chan ChatCompletionChunk)
	errCh := make(chan error, 1)

	go func() {
		resp, err := r.http.stream(ctx, "/v1/chat/completions", params)
		if err != nil {
			close(chunkCh)
			errCh <- err
			close(errCh)
			return
		}
		parseSSEStream(resp, chunkCh, errCh)
	}()

	return chunkCh, errCh
}

// ChatResource groups chat-related sub-resources.
type ChatResource struct {
	Completions *CompletionsResource
}
