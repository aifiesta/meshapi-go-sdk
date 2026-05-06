package meshapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	sdkVersionHeader = "X-MeshAPI-SDK"
	sdkVersionValue  = "go/0.1.0"

	defaultTimeoutMs   = 60_000
	defaultMaxRetries  = 3
	backoffBaseMs      = 500
	backoffMaxMs       = 30_000
)

var retryStatusCodes = map[int]bool{429: true, 502: true, 503: true, 504: true}

// httpClient wraps net/http.Client with retry, auth, and JSON helpers.
type httpClient struct {
	cfg    Config
	client *http.Client
}

func newHTTPClient(cfg Config) *httpClient {
	c := cfg.HTTPClient
	if c == nil {
		c = &http.Client{Timeout: time.Duration(cfg.timeoutMs()) * time.Millisecond}
	}
	return &httpClient{cfg: cfg, client: c}
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
		"Authorization": "Bearer " + h.cfg.Token,
		"Content-Type":  "application/json",
		"Accept":        "application/json",
		sdkVersionHeader: sdkVersionValue,
	}
}

func (h *httpClient) do(ctx context.Context, req *http.Request) (*http.Response, error) {
	for k, v := range h.baseHeaders() {
		req.Header.Set(k, v)
	}

	maxRetries := h.cfg.maxRetries()

	var lastResp *http.Response
	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Clone body for retry
		var bodyBytes []byte
		if req.Body != nil && req.Body != http.NoBody {
			b, _ := io.ReadAll(req.Body)
			bodyBytes = b
			req.Body = io.NopCloser(bytes.NewReader(b))
		}

		resp, err := h.client.Do(req.WithContext(ctx))
		if err != nil {
			return nil, err
		}

		if retryStatusCodes[resp.StatusCode] && attempt < maxRetries {
			delay := computeDelay(attempt, retryAfterFromResponse(resp))
			resp.Body.Close()
			// Restore body for next attempt
			if bodyBytes != nil {
				req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			}
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
			continue
		}

		lastResp = resp
		break
	}

	return lastResp, nil
}

func computeDelay(attempt int, retryAfterSec *int) time.Duration {
	var baseMs float64
	if retryAfterSec != nil {
		baseMs = float64(*retryAfterSec * 1000)
	} else {
		baseMs = backoffBaseMs * math.Pow(2, float64(attempt))
	}
	if baseMs > backoffMaxMs {
		baseMs = backoffMaxMs
	}
	// ±20% jitter
	jitter := baseMs * (0.8 + rand.Float64()*0.4)
	return time.Duration(jitter) * time.Millisecond
}

func retryAfterFromResponse(resp *http.Response) *int {
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
