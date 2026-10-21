package meshapi

// Unit tests for resilience: configurable transport retry, the chat
// client-side model-fallback chain, and observability events (retry /
// fallback / gateway-routing) via Logger. Mirrors the Node SDK's
// tests/resilience.test.ts behavioural contract.

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
)

// ── Helpers ───────────────────────────────────────────────────────────────────

const okChatBody = `{
	"id": "chatcmpl-1",
	"object": "chat.completion",
	"created": 0,
	"model": "openai/gpt-4o-mini",
	"choices": [
		{"index": 0, "message": {"role": "assistant", "content": "hi"}, "finish_reason": "stop"}
	],
	"usage": {"prompt_tokens": 1, "completion_tokens": 1, "total_tokens": 2}
}`

func jsonResponse(status int, body string, headers map[string]string) *http.Response {
	h := http.Header{}
	h.Set("Content-Type", "application/json")
	for k, v := range headers {
		h.Set(k, v)
	}
	return &http.Response{
		StatusCode: status,
		Header:     h,
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func testErrorResponse(status int, code, requestID string) *http.Response {
	body, _ := json.Marshal(map[string]interface{}{
		"error":      map[string]string{"code": code, "message": "boom"},
		"request_id": requestID,
	})
	return jsonResponse(status, string(body), nil)
}

type recordedCall struct {
	url  string
	body map[string]interface{}
}

// queueTransport is a RoundTripper fed by a queue of responses / errors;
// it records every request URL and JSON body.
type queueTransport struct {
	mu    sync.Mutex
	queue []interface{} // *http.Response or error
	calls []recordedCall
}

func (q *queueTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	call := recordedCall{url: req.URL.String()}
	if req.Body != nil && req.Body != http.NoBody {
		raw, _ := io.ReadAll(req.Body)
		req.Body.Close()
		_ = json.Unmarshal(raw, &call.body)
	}
	q.calls = append(q.calls, call)
	if len(q.queue) == 0 {
		return nil, errors.New("queue exhausted")
	}
	next := q.queue[0]
	q.queue = q.queue[1:]
	if err, ok := next.(error); ok {
		return nil, err
	}
	return next.(*http.Response), nil
}

func (q *queueTransport) callCount() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.calls)
}

// eventLog collects resilience events from Config.Logger.
type eventLog struct {
	mu     sync.Mutex
	events []ResilienceEvent
}

func (l *eventLog) log(e ResilienceEvent) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.events = append(l.events, e)
}

func (l *eventLog) ofKind(kind string) []ResilienceEvent {
	l.mu.Lock()
	defer l.mu.Unlock()
	var out []ResilienceEvent
	for _, e := range l.events {
		if e.Kind() == kind {
			out = append(out, e)
		}
	}
	return out
}

func zeroBackoffRetry(mutate func(*RetryPolicy)) *RetryPolicy {
	zero := 0
	two := 2
	p := &RetryPolicy{MaxRetries: &two, BackoffBaseMs: &zero, BackoffMaxMs: &zero}
	if mutate != nil {
		mutate(p)
	}
	return p
}

func makeTestClient(queue []interface{}, mutate func(*Config)) (*Client, *queueTransport, *eventLog) {
	transport := &queueTransport{queue: queue}
	events := &eventLog{}
	cfg := Config{
		BaseURL:    "https://gw.test",
		Token:      "rsk_test",
		Logger:     events.log,
		Retry:      zeroBackoffRetry(nil), // zero backoff so tests don't sleep
		HTTPClient: &http.Client{Transport: transport},
	}
	if mutate != nil {
		mutate(&cfg)
	}
	return New(cfg), transport, events
}

func chatParams() ChatCompletionParams {
	model := "openai/gpt-4o-mini"
	return ChatCompletionParams{
		Model:    &model,
		Messages: []ChatMessage{{Role: "user", Content: "hello"}},
	}
}

func ip(v int) *int   { return &v }
func bp(v bool) *bool { return &v }

// ── resolveRetryPolicy ────────────────────────────────────────────────────────

func TestResolveRetryPolicy_Defaults(t *testing.T) {
	p := resolveRetryPolicy(nil, nil)
	if p.maxRetries != 3 {
		t.Errorf("maxRetries: expected 3, got %d", p.maxRetries)
	}
	for _, code := range []int{429, 502, 503, 504} {
		if !p.retryOnStatus[code] {
			t.Errorf("expected %d in default retryOnStatus", code)
		}
	}
	if len(p.retryOnStatus) != 4 {
		t.Errorf("expected 4 default retry statuses, got %d", len(p.retryOnStatus))
	}
	if p.backoffBaseMs != 500 {
		t.Errorf("backoffBaseMs: expected 500, got %d", p.backoffBaseMs)
	}
	if p.backoffMaxMs != 30_000 {
		t.Errorf("backoffMaxMs: expected 30000, got %d", p.backoffMaxMs)
	}
	if !p.respectRetryAfter {
		t.Error("respectRetryAfter: expected true")
	}
	if p.retryOnNetworkError {
		t.Error("retryOnNetworkError: expected false")
	}
}

func TestResolveRetryPolicy_MaxRetriesAliasPrecedence(t *testing.T) {
	// Retry.MaxRetries wins over the deprecated top-level MaxRetries.
	if got := resolveRetryPolicy(&RetryPolicy{MaxRetries: ip(5)}, ip(1)).maxRetries; got != 5 {
		t.Errorf("expected Retry.MaxRetries=5 to win, got %d", got)
	}
	if got := resolveRetryPolicy(nil, ip(1)).maxRetries; got != 1 {
		t.Errorf("expected legacy MaxRetries=1, got %d", got)
	}
}

// ── Transport retry ───────────────────────────────────────────────────────────

func TestTransportRetry_503ThenSuccess_EmitsRetryEvent(t *testing.T) {
	client, transport, events := makeTestClient([]interface{}{
		testErrorResponse(503, "provider_not_available", "req_err"),
		jsonResponse(200, okChatBody, nil),
	}, nil)

	res, err := client.Chat.Completions.Create(context.Background(), chatParams())
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if content := res.Choices[0].Message.Content; content == nil || *content != "hi" {
		t.Errorf("unexpected content: %v", content)
	}
	if transport.callCount() != 2 {
		t.Errorf("expected 2 calls, got %d", transport.callCount())
	}

	retries := events.ofKind("retry")
	if len(retries) != 1 {
		t.Fatalf("expected 1 retry event, got %d", len(retries))
	}
	retry := retries[0].(RetryEvent)
	if retry.Status != 503 {
		t.Errorf("Status: expected 503, got %d", retry.Status)
	}
	if retry.Attempt != 1 {
		t.Errorf("Attempt: expected 1, got %d", retry.Attempt)
	}
	if retry.Reason != RetryReasonStatus {
		t.Errorf("Reason: expected %q, got %q", RetryReasonStatus, retry.Reason)
	}
	if retry.RequestID != "" { // no X-Request-Id header on the mock response
		t.Errorf("RequestID: expected empty, got %q", retry.RequestID)
	}
	if retry.Method != "POST" || retry.Path != "/v1/chat/completions" {
		t.Errorf("unexpected Method/Path: %q %q", retry.Method, retry.Path)
	}
}

func TestTransportRetry_CustomRetryOnStatus(t *testing.T) {
	// 500 is not retryable by default; opt in explicitly.
	client, transport, _ := makeTestClient([]interface{}{
		testErrorResponse(500, "upstream_error", "req_err"),
		jsonResponse(200, okChatBody, nil),
	}, func(cfg *Config) {
		cfg.Retry = zeroBackoffRetry(func(p *RetryPolicy) {
			p.RetryOnStatus = []int{500}
		})
	})

	if _, err := client.Chat.Completions.Create(context.Background(), chatParams()); err != nil {
		t.Fatalf("create: %v", err)
	}
	if transport.callCount() != 2 {
		t.Errorf("expected 2 calls, got %d", transport.callCount())
	}
}

func TestTransportRetry_ExhaustionThrowsAPIError(t *testing.T) {
	client, transport, events := makeTestClient([]interface{}{
		testErrorResponse(503, "provider_not_available", "req_err"),
		testErrorResponse(503, "provider_not_available", "req_err"),
		testErrorResponse(503, "provider_not_available", "req_err"),
	}, nil)

	_, err := client.Chat.Completions.Create(context.Background(), chatParams())
	var apiErr *MeshAPIError
	if !errors.As(err, &apiErr) || apiErr.Status != 503 {
		t.Fatalf("expected MeshAPIError with status 503, got %v", err)
	}
	if transport.callCount() != 3 { // 1 initial + 2 retries
		t.Errorf("expected 3 calls, got %d", transport.callCount())
	}
	if got := len(events.ofKind("retry")); got != 2 {
		t.Errorf("expected 2 retry events, got %d", got)
	}
}

func TestTransportRetry_NetworkErrorOffByDefault(t *testing.T) {
	client, transport, _ := makeTestClient([]interface{}{
		errors.New("connection refused"),
	}, nil)

	_, err := client.Models.List(context.Background(), ListModelsParams{})
	if err == nil || !strings.Contains(err.Error(), "connection refused") {
		t.Fatalf("expected connection refused error, got %v", err)
	}
	if transport.callCount() != 1 {
		t.Errorf("expected 1 call (no retry), got %d", transport.callCount())
	}
}

func TestTransportRetry_NetworkErrorOptIn(t *testing.T) {
	client, transport, events := makeTestClient([]interface{}{
		errors.New("connection refused"),
		jsonResponse(200, "[]", nil),
	}, func(cfg *Config) {
		cfg.Retry = zeroBackoffRetry(func(p *RetryPolicy) {
			p.RetryOnNetworkError = bp(true)
		})
	})

	if _, err := client.Models.List(context.Background(), ListModelsParams{}); err != nil {
		t.Fatalf("list: %v", err)
	}
	if transport.callCount() != 2 {
		t.Errorf("expected 2 calls, got %d", transport.callCount())
	}
	retries := events.ofKind("retry")
	if len(retries) != 1 {
		t.Fatalf("expected 1 retry event, got %d", len(retries))
	}
	retry := retries[0].(RetryEvent)
	if retry.Reason != RetryReasonNetworkError {
		t.Errorf("Reason: expected %q, got %q", RetryReasonNetworkError, retry.Reason)
	}
	if retry.Status != 0 {
		t.Errorf("Status: expected 0 for a network error, got %d", retry.Status)
	}
}

func TestTransportRetry_NeverRetriesContextCancel(t *testing.T) {
	client, transport, _ := makeTestClient([]interface{}{
		context.Canceled,
	}, func(cfg *Config) {
		cfg.Retry = zeroBackoffRetry(func(p *RetryPolicy) {
			p.MaxRetries = ip(3)
			p.RetryOnNetworkError = bp(true)
		})
	})

	_, err := client.Models.List(context.Background(), ListModelsParams{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if transport.callCount() != 1 {
		t.Errorf("expected 1 call (aborts never retry), got %d", transport.callCount())
	}
}

// timeoutError implements net.Error with Timeout() == true.
type timeoutError struct{}

func (timeoutError) Error() string   { return "i/o timeout" }
func (timeoutError) Timeout() bool   { return true }
func (timeoutError) Temporary() bool { return false }

func TestTransportRetry_NeverRetriesTimeouts(t *testing.T) {
	client, transport, _ := makeTestClient([]interface{}{
		timeoutError{},
	}, func(cfg *Config) {
		cfg.Retry = zeroBackoffRetry(func(p *RetryPolicy) {
			p.RetryOnNetworkError = bp(true)
		})
	})

	_, err := client.Models.List(context.Background(), ListModelsParams{})
	if err == nil {
		t.Fatal("expected a timeout error")
	}
	if transport.callCount() != 1 {
		t.Errorf("expected 1 call (timeouts never retry), got %d", transport.callCount())
	}
}

// ── Chat model-fallback chain ─────────────────────────────────────────────────

func TestFallback_AdvancesAfterRetriesExhaust_EmitsEvent(t *testing.T) {
	client, transport, events := makeTestClient([]interface{}{
		testErrorResponse(503, "provider_not_available", "req_err"), // primary attempt 1
		testErrorResponse(503, "provider_not_available", "req_err"), // primary retry 1
		jsonResponse(200, okChatBody, nil),                          // fallback model
	}, func(cfg *Config) {
		cfg.Retry = zeroBackoffRetry(func(p *RetryPolicy) { p.MaxRetries = ip(1) })
		cfg.Fallback = &FallbackConfig{Models: []string{"anthropic/claude-sonnet-5"}}
	})

	res, err := client.Chat.Completions.Create(context.Background(), chatParams())
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if content := res.Choices[0].Message.Content; content == nil || *content != "hi" {
		t.Errorf("unexpected content: %v", content)
	}
	if transport.callCount() != 3 {
		t.Fatalf("expected 3 calls, got %d", transport.callCount())
	}
	if got := transport.calls[2].body["model"]; got != "anthropic/claude-sonnet-5" {
		t.Errorf("expected fallback model on the wire, got %v", got)
	}

	fbs := events.ofKind("fallback")
	if len(fbs) != 1 {
		t.Fatalf("expected 1 fallback event, got %d", len(fbs))
	}
	fb := fbs[0].(FallbackEvent)
	if fb.FromModel != "openai/gpt-4o-mini" {
		t.Errorf("FromModel: got %q", fb.FromModel)
	}
	if fb.ToModel != "anthropic/claude-sonnet-5" {
		t.Errorf("ToModel: got %q", fb.ToModel)
	}
	if fb.Status != 503 {
		t.Errorf("Status: expected 503, got %d", fb.Status)
	}
	if fb.ErrorCode != "provider_not_available" {
		t.Errorf("ErrorCode: got %q", fb.ErrorCode)
	}
	if fb.RequestID != "req_err" {
		t.Errorf("RequestID: got %q", fb.RequestID)
	}
	if fb.ChainIndex != 0 || fb.ChainLength != 1 {
		t.Errorf("chain position: got %d/%d", fb.ChainIndex, fb.ChainLength)
	}
}

func TestFallback_PerCallOverride_NeverOnTheWire(t *testing.T) {
	client, transport, _ := makeTestClient([]interface{}{
		testErrorResponse(502, "provider_not_available", "req_err"),
		jsonResponse(200, okChatBody, nil),
	}, func(cfg *Config) {
		cfg.Retry = zeroBackoffRetry(func(p *RetryPolicy) { p.MaxRetries = ip(0) })
		cfg.Fallback = &FallbackConfig{Models: []string{"ignored/config-model"}}
	})

	params := chatParams()
	params.FallbackModels = []string{"mistral/mistral-large"}
	if _, err := client.Chat.Completions.Create(context.Background(), params); err != nil {
		t.Fatalf("create: %v", err)
	}

	if got := transport.calls[1].body["model"]; got != "mistral/mistral-large" {
		t.Errorf("expected per-call override model, got %v", got)
	}
	for i, call := range transport.calls {
		for key := range call.body {
			if key == "fallback_models" || key == "fallbackModels" || key == "FallbackModels" {
				t.Errorf("call %d: FallbackModels leaked to the wire as %q", i, key)
			}
		}
	}
}

func TestFallback_TerminalErrorNeverAdvances(t *testing.T) {
	client, transport, _ := makeTestClient([]interface{}{
		testErrorResponse(401, "unauthorized", "req_err"),
	}, func(cfg *Config) {
		cfg.Retry = zeroBackoffRetry(func(p *RetryPolicy) { p.MaxRetries = ip(0) })
		cfg.Fallback = &FallbackConfig{Models: []string{"anthropic/claude-sonnet-5"}}
	})

	_, err := client.Chat.Completions.Create(context.Background(), chatParams())
	var apiErr *MeshAPIError
	if !errors.As(err, &apiErr) || apiErr.Status != 401 {
		t.Fatalf("expected 401 MeshAPIError, got %v", err)
	}
	if transport.callCount() != 1 {
		t.Errorf("expected 1 call, got %d", transport.callCount())
	}
}

func TestFallback_ChainExhaustionReturnsLastError(t *testing.T) {
	client, transport, _ := makeTestClient([]interface{}{
		testErrorResponse(503, "provider_not_available", "req_1"),
		testErrorResponse(503, "provider_not_available", "req_2"),
		testErrorResponse(504, "gateway_timeout", "req_last"),
	}, func(cfg *Config) {
		cfg.Retry = zeroBackoffRetry(func(p *RetryPolicy) { p.MaxRetries = ip(0) })
		cfg.Fallback = &FallbackConfig{Models: []string{"m/a", "m/b"}}
	})

	_, err := client.Chat.Completions.Create(context.Background(), chatParams())
	var apiErr *MeshAPIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected MeshAPIError, got %v", err)
	}
	if apiErr.Status != 504 || apiErr.RequestID != "req_last" {
		t.Errorf("expected the LAST error (504 req_last), got %d %q", apiErr.Status, apiErr.RequestID)
	}
	if transport.callCount() != 3 {
		t.Errorf("expected 3 calls, got %d", transport.callCount())
	}
}

func TestFallback_SkipsPrimaryModelInChain(t *testing.T) {
	client, transport, _ := makeTestClient([]interface{}{
		testErrorResponse(503, "provider_not_available", "req_err"),
		jsonResponse(200, okChatBody, nil),
	}, func(cfg *Config) {
		cfg.Retry = zeroBackoffRetry(func(p *RetryPolicy) { p.MaxRetries = ip(0) })
		cfg.Fallback = &FallbackConfig{Models: []string{"openai/gpt-4o-mini", "m/b"}}
	})

	if _, err := client.Chat.Completions.Create(context.Background(), chatParams()); err != nil {
		t.Fatalf("create: %v", err)
	}
	if transport.callCount() != 2 {
		t.Fatalf("expected 2 calls, got %d", transport.callCount())
	}
	if got := transport.calls[1].body["model"]; got != "m/b" {
		t.Errorf("expected the primary to be skipped in the chain, got %v", got)
	}
}

func TestFallback_CustomOnStatus(t *testing.T) {
	// 429 not in the default fallback set — opt in.
	client, transport, _ := makeTestClient([]interface{}{
		testErrorResponse(429, "rate_limit_exceeded", "req_err"),
		jsonResponse(200, okChatBody, nil),
	}, func(cfg *Config) {
		cfg.Retry = zeroBackoffRetry(func(p *RetryPolicy) { p.MaxRetries = ip(0) })
		cfg.Fallback = &FallbackConfig{Models: []string{"m/b"}, OnStatus: []int{429}}
	})

	if _, err := client.Chat.Completions.Create(context.Background(), chatParams()); err != nil {
		t.Fatalf("create: %v", err)
	}
	if transport.callCount() != 2 {
		t.Errorf("expected 2 calls, got %d", transport.callCount())
	}
}

// ── Gateway routing observability ─────────────────────────────────────────────

func TestGatewayRouting_HeadersParsedIntoEvent(t *testing.T) {
	client, _, events := makeTestClient([]interface{}{
		jsonResponse(200, okChatBody, map[string]string{
			"X-Mesh-Routing-Attempts": "2",
			"X-Mesh-Routing-Fallback": "true",
			"X-Mesh-Served-Provider":  "bedrock",
			"X-Request-Id":            "req_routed",
		}),
	}, nil)

	if _, err := client.Chat.Completions.Create(context.Background(), chatParams()); err != nil {
		t.Fatalf("create: %v", err)
	}

	gws := events.ofKind("gateway-routing")
	if len(gws) != 1 {
		t.Fatalf("expected 1 gateway-routing event, got %d", len(gws))
	}
	gw := gws[0].(GatewayRoutingEvent)
	if gw.Attempts != 2 {
		t.Errorf("Attempts: expected 2, got %d", gw.Attempts)
	}
	if !gw.Fallback {
		t.Error("Fallback: expected true")
	}
	if gw.ServedProvider != "bedrock" {
		t.Errorf("ServedProvider: got %q", gw.ServedProvider)
	}
	if gw.RequestID != "req_routed" {
		t.Errorf("RequestID: got %q", gw.RequestID)
	}
	if gw.Path != "/v1/chat/completions" {
		t.Errorf("Path: got %q", gw.Path)
	}
}

func TestGatewayRouting_AbsentHeadersEmitNothing(t *testing.T) {
	client, _, events := makeTestClient([]interface{}{
		jsonResponse(200, okChatBody, nil),
	}, nil)

	if _, err := client.Chat.Completions.Create(context.Background(), chatParams()); err != nil {
		t.Fatalf("create: %v", err)
	}
	if got := len(events.ofKind("gateway-routing")); got != 0 {
		t.Errorf("expected no gateway-routing events, got %d", got)
	}
}

// ── Debug formatting ──────────────────────────────────────────────────────────

func TestFormatResilienceEvent_RetryLine(t *testing.T) {
	line := FormatResilienceEvent(RetryEvent{
		Method:     "POST",
		Path:       "/v1/chat/completions",
		Attempt:    1,
		MaxRetries: 3,
		Status:     503,
		RequestID:  "req_1",
		DelayMs:    512.4,
		Reason:     RetryReasonStatus,
	})
	want := "retrying POST /v1/chat/completions (attempt 1/4 failed: 503, next in 512ms) [req_1]"
	if line != want {
		t.Errorf("retry line:\n got %q\nwant %q", line, want)
	}
}

func TestFormatResilienceEvent_NetworkErrorRetryLine(t *testing.T) {
	line := FormatResilienceEvent(RetryEvent{
		Method:     "POST",
		Path:       "/v1/chat/completions",
		Attempt:    2,
		MaxRetries: 3,
		DelayMs:    1000,
		Reason:     RetryReasonNetworkError,
	})
	want := "retrying POST /v1/chat/completions (attempt 2/4 failed: network error, next in 1000ms)"
	if line != want {
		t.Errorf("retry line:\n got %q\nwant %q", line, want)
	}
}

func TestFormatResilienceEvent_FallbackLine(t *testing.T) {
	line := FormatResilienceEvent(FallbackEvent{
		FromModel:   "openai/gpt-4o",
		ToModel:     "anthropic/claude-sonnet-5",
		ChainIndex:  0,
		ChainLength: 2,
		Status:      503,
		ErrorCode:   "provider_not_available",
	})
	want := "falling back openai/gpt-4o → anthropic/claude-sonnet-5 (1/2: 503 provider_not_available)"
	if line != want {
		t.Errorf("fallback line:\n got %q\nwant %q", line, want)
	}
}

func TestFormatResilienceEvent_GatewayRoutingLine(t *testing.T) {
	line := FormatResilienceEvent(GatewayRoutingEvent{
		Path:           "/v1/chat/completions",
		Attempts:       2,
		Fallback:       true,
		ServedProvider: "bedrock",
		RequestID:      "req_2",
	})
	want := "gateway served /v1/chat/completions via bedrock (2 attempts, provider fallback) [req_2]"
	if line != want {
		t.Errorf("gateway-routing line:\n got %q\nwant %q", line, want)
	}
}
