package meshapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	sdkVersionHeader = "X-MeshAPI-SDK"
	sdkVersionValue  = "go/0.1.12"

	defaultTimeoutMs     = 60_000
	defaultMaxRetries    = 3
	defaultBackoffBaseMs = 500
	defaultBackoffMaxMs  = 30_000
)

// Gateway routing-outcome headers (FT-244) — present when the API key's
// routing_policy is active. See resilience.go (GatewayRoutingEvent).
const (
	routingAttemptsHeader = "X-Mesh-Routing-Attempts"
	routingFallbackHeader = "X-Mesh-Routing-Fallback"
	servedProviderHeader  = "X-Mesh-Served-Provider"
	requestIDHeader       = "X-Request-Id"
)

// httpClient wraps net/http.Client with retry, auth, and JSON helpers.
type httpClient struct {
	cfg    Config
	client *http.Client
	retry  resolvedRetryPolicy
}

func newHTTPClient(cfg Config) *httpClient {
	c := cfg.HTTPClient
	if c == nil {
		c = &http.Client{Timeout: time.Duration(cfg.timeoutMs()) * time.Millisecond}
	}
	return &httpClient{
		cfg:    cfg,
		client: c,
		retry:  resolveRetryPolicy(cfg.Retry, cfg.MaxRetries),
	}
}

// emit publishes a resilience event to the configured Config.Logger and, with
// Config.Debug, as a readable stderr line. Gateway-routing lines are only
// printed when a server-side retry/fallback actually happened; the logger
// receives every event. Also used by CompletionsResource for fallback hops.
func (h *httpClient) emit(event ResilienceEvent) {
	if h.cfg.Logger != nil {
		h.cfg.Logger(event)
	}
	if !h.cfg.Debug {
		return
	}
	if gw, ok := event.(GatewayRoutingEvent); ok && gw.Attempts <= 1 && !gw.Fallback {
		return
	}
	fmt.Fprintf(os.Stderr, "[meshapi] %s\n", FormatResilienceEvent(event))
}

func (h *httpClient) buildURL(path string, params url.Values) string {
	u := h.cfg.BaseURL + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}
	return u
}

func (h *httpClient) baseHeaders() map[string]string {
	return map[string]string{
		"Authorization":  "Bearer " + h.cfg.Token,
		"Content-Type":   "application/json",
		"Accept":         "application/json",
		sdkVersionHeader: sdkVersionValue,
	}
}

// do is the single transport retry loop shared by every non-streaming request
// (JSON, raw-bytes, and multipart). Re-sends on the policy's status set (and,
// opt-in, on pre-response network errors), with exponential backoff, jitter,
// and Retry-After support. Emits a RetryEvent per re-send and a
// GatewayRoutingEvent when the final response carries X-Mesh-Routing-*
// headers. Returns the final response — callers handle non-2xx statuses.
func (h *httpClient) do(ctx context.Context, req *http.Request) (*http.Response, error) {
	for k, v := range h.baseHeaders() {
		if req.Header.Get(k) == "" {
			req.Header.Set(k, v)
		}
	}

	// Buffer the body once so it can be re-sent on retry.
	var bodyBytes []byte
	if req.Body != nil && req.Body != http.NoBody {
		bodyBytes, _ = io.ReadAll(req.Body)
	}

	pol := h.retry
	attempt := 0
	for {
		if bodyBytes != nil {
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}

		resp, err := h.client.Do(req.WithContext(ctx))
		if err != nil {
			// Cancellations and timeouts always propagate. Other pre-response
			// failures (DNS, connection refused/reset) retry only when opted
			// in — they are ambiguous for non-idempotent POSTs.
			if !pol.retryOnNetworkError || attempt >= pol.maxRetries || isCancelOrTimeout(err) {
				return nil, err
			}
			delayMs := h.computeDelayMs(attempt, nil)
			h.emit(RetryEvent{
				Method:     req.Method,
				Path:       req.URL.Path,
				Attempt:    attempt + 1,
				MaxRetries: pol.maxRetries,
				DelayMs:    delayMs,
				Reason:     RetryReasonNetworkError,
			})
			if err := sleepCtx(ctx, delayMs); err != nil {
				return nil, err
			}
			attempt++
			continue
		}

		if pol.retryOnStatus[resp.StatusCode] && attempt < pol.maxRetries {
			delayMs := h.computeDelayMs(attempt, retryAfterFromResponse(resp, pol.respectRetryAfter))
			h.emit(RetryEvent{
				Method:     req.Method,
				Path:       req.URL.Path,
				Attempt:    attempt + 1,
				MaxRetries: pol.maxRetries,
				Status:     resp.StatusCode,
				RequestID:  resp.Header.Get(requestIDHeader),
				DelayMs:    delayMs,
				Reason:     RetryReasonStatus,
			})
			resp.Body.Close()
			if err := sleepCtx(ctx, delayMs); err != nil {
				return nil, err
			}
			attempt++
			continue
		}

		h.emitGatewayRouting(req.URL.Path, resp)
		return resp, nil
	}
}

// emitGatewayRouting surfaces the gateway's own routing outcome (server-side
// retry / provider fallback, FT-244) when the response reports it.
// Header-absence means the key has no active routing policy — nothing is
// emitted.
func (h *httpClient) emitGatewayRouting(path string, resp *http.Response) {
	attempts := resp.Header.Get(routingAttemptsHeader)
	if attempts == "" {
		return
	}
	n, err := strconv.Atoi(attempts)
	if err != nil || n == 0 {
		n = 1
	}
	h.emit(GatewayRoutingEvent{
		Path:           path,
		Attempts:       n,
		Fallback:       resp.Header.Get(routingFallbackHeader) == "true",
		ServedProvider: resp.Header.Get(servedProviderHeader),
		RequestID:      resp.Header.Get(requestIDHeader),
	})
}

// isCancelOrTimeout reports whether a client.Do error is a context
// cancellation, a deadline expiry, or any network timeout — failure modes
// that are never retried (the caller gave up, or the request may already be
// executing server-side).
func isCancelOrTimeout(err error) bool {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}

// sleepCtx sleeps for delayMs milliseconds or until ctx is done.
func sleepCtx(ctx context.Context, delayMs float64) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(time.Duration(delayMs * float64(time.Millisecond))):
		return nil
	}
}

func (h *httpClient) computeDelayMs(attempt int, retryAfterSec *int) float64 {
	var baseMs float64
	if retryAfterSec != nil {
		baseMs = float64(*retryAfterSec * 1000)
	} else {
		baseMs = float64(h.retry.backoffBaseMs) * math.Pow(2, float64(attempt))
	}
	if maxMs := float64(h.retry.backoffMaxMs); baseMs > maxMs {
		baseMs = maxMs
	}
	// ±20% jitter
	return baseMs * (0.8 + rand.Float64()*0.4)
}

func retryAfterFromResponse(resp *http.Response, respectRetryAfter bool) *int {
	if !respectRetryAfter {
		return nil
	}
	val := resp.Header.Get("Retry-After")
	if val == "" {
		return nil
	}
	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return nil
	}
	n := int(math.Ceil(f))
	return &n
}

// get performs a GET request and decodes the JSON response into dst.
func (h *httpClient) get(ctx context.Context, path string, params url.Values, dst interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, h.buildURL(path, params), nil)
	if err != nil {
		return err
	}
	resp, err := h.do(ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return newErrorFromResponse(resp)
	}
	if resp.StatusCode == 204 {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(dst)
}

// post performs a POST request with a JSON body and decodes the response.
func (h *httpClient) post(ctx context.Context, path string, body interface{}, dst interface{}) error {
	return h.jsonRequest(ctx, http.MethodPost, path, body, dst)
}

// patch performs a PATCH request with a JSON body and decodes the response.
func (h *httpClient) patch(ctx context.Context, path string, body interface{}, dst interface{}) error {
	return h.jsonRequest(ctx, http.MethodPatch, path, body, dst)
}

// delete performs a DELETE request (expects 204).
func (h *httpClient) delete(ctx context.Context, path string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, h.buildURL(path, nil), nil)
	if err != nil {
		return err
	}
	resp, err := h.do(ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return newErrorFromResponse(resp)
	}
	return nil
}

func (h *httpClient) getBytes(ctx context.Context, path string, params url.Values) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, h.buildURL(path, params), nil)
	if err != nil {
		return nil, err
	}
	resp, err := h.do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, newErrorFromResponse(resp)
	}
	return io.ReadAll(resp.Body)
}

// postBytes performs a POST request with a JSON body and returns the raw response bytes.
func (h *httpClient) postBytes(ctx context.Context, path string, body interface{}) ([]byte, error) {
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.buildURL(path, nil), bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	resp, err := h.do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, newErrorFromResponse(resp)
	}
	return io.ReadAll(resp.Body)
}

// postMultipart sends a multipart/form-data POST.
// fields contains single-value form fields; multiValueFields contains repeated fields (e.g. keyterms).
// fileData (if non-nil) is the file content attached as the "file" field.
func (h *httpClient) postMultipart(ctx context.Context, path string, fields map[string]string, multiValueFields map[string][]string, fileData []byte, filename string, dst interface{}) error {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	for k, v := range fields {
		if err := mw.WriteField(k, v); err != nil {
			return fmt.Errorf("write field %s: %w", k, err)
		}
	}

	for k, vals := range multiValueFields {
		for _, v := range vals {
			if err := mw.WriteField(k, v); err != nil {
				return fmt.Errorf("write field %s: %w", k, err)
			}
		}
	}

	if fileData != nil {
		fw, err := mw.CreateFormFile("file", filename)
		if err != nil {
			return fmt.Errorf("create form file: %w", err)
		}
		if _, err := fw.Write(fileData); err != nil {
			return fmt.Errorf("write file data: %w", err)
		}
	}

	if err := mw.Close(); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.buildURL(path, nil), &buf)
	if err != nil {
		return err
	}
	// Set Content-Type before h.do() so the "set if absent" logic preserves the multipart boundary.
	req.Header.Set("Content-Type", mw.FormDataContentType())

	resp, err := h.do(ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return newErrorFromResponse(resp)
	}
	return json.NewDecoder(resp.Body).Decode(dst)
}

func (h *httpClient) jsonRequest(ctx context.Context, method, path string, body interface{}, dst interface{}) error {
	buf, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, method, h.buildURL(path, nil), bytes.NewReader(buf))
	if err != nil {
		return err
	}
	resp, err := h.do(ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return newErrorFromResponse(resp)
	}
	if resp.StatusCode == 204 || dst == nil {
		return nil
	}

	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 500))
		return &MeshAPIError{
			Status:  resp.StatusCode,
			Code:    "parse_error",
			Message: string(raw),
		}
	}
	return json.NewDecoder(resp.Body).Decode(dst)
}

// stream opens a streaming POST and returns the raw response for SSE parsing.
// The caller is responsible for closing resp.Body.
// Streaming requests are NEVER retried.
func (h *httpClient) stream(ctx context.Context, path string, body interface{}) (*http.Response, error) {
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal stream request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.buildURL(path, nil), bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	for k, v := range h.baseHeaders() {
		req.Header.Set(k, v)
	}
	req.Header.Set("Accept", "text/event-stream")

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		return nil, newErrorFromResponse(resp)
	}
	return resp, nil
}
