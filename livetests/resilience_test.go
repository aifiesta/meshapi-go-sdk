package livetest

// Live tests for resilience: retry / fallback / observability. Mirrors the
// Node SDK's livetests/test-resilience.js scenarios.

import (
	"context"
	"strings"
	"sync"
	"testing"

	meshapi "meshapi-go-sdk"
)

// liveEventLog collects resilience events from Config.Logger.
type liveEventLog struct {
	mu     sync.Mutex
	events []meshapi.ResilienceEvent
}

func (l *liveEventLog) log(e meshapi.ResilienceEvent) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.events = append(l.events, e)
}

func (l *liveEventLog) ofKind(kind string) []meshapi.ResilienceEvent {
	l.mu.Lock()
	defer l.mu.Unlock()
	var out []meshapi.ResilienceEvent
	for _, e := range l.events {
		if e.Kind() == kind {
			out = append(out, e)
		}
	}
	return out
}

func liveToken() string {
	return liveEnv("MESHAPI_TOKEN", defaultToken)
}

// TestLive_Resilience_HealthyCallWithLogger verifies that a successful call
// with a logger attached produces no spurious events — gateway-routing only
// from headers, and shaped sanely when present.
func TestLive_Resilience_HealthyCallWithLogger(t *testing.T) {
	skipIfNoBackend(t)
	events := &liveEventLog{}
	client := meshapi.New(meshapi.Config{
		BaseURL: liveBaseURL(),
		Token:   liveToken(),
		Logger:  events.log,
	})

	maxTokens := 10
	resp, err := client.Chat.Completions.Create(context.Background(), meshapi.ChatCompletionParams{
		Model:     strPtr(liveModel()),
		Messages:  []meshapi.ChatMessage{{Role: "user", Content: "Reply with the word: ok"}},
		MaxTokens: &maxTokens,
	})
	if err != nil {
		t.Fatalf("chat create: %v", err)
	}
	if len(resp.Choices) == 0 || resp.Choices[0].Message == nil {
		t.Fatal("expected a message in the response")
	}

	// No client-side retry/fallback should have happened on a healthy call.
	if got := len(events.ofKind("retry")); got != 0 {
		t.Errorf("expected 0 retry events on a healthy call, got %d", got)
	}
	if got := len(events.ofKind("fallback")); got != 0 {
		t.Errorf("expected 0 fallback events on a healthy call, got %d", got)
	}
	// gateway-routing appears IFF the key has an active routing_policy; when
	// it does, the shape must be sane.
	for _, e := range events.ofKind("gateway-routing") {
		gw := e.(meshapi.GatewayRoutingEvent)
		if gw.Attempts < 1 {
			t.Errorf("gateway-routing Attempts must be >= 1, got %d", gw.Attempts)
		}
	}
}

// TestLive_Resilience_PerCallFallbackModelsStripped verifies that the
// per-call FallbackModels field is client-side only — the request validates
// server-side (the field was stripped) and the primary model answers.
func TestLive_Resilience_PerCallFallbackModelsStripped(t *testing.T) {
	skipIfNoBackend(t)
	client := meshapi.New(meshapi.Config{
		BaseURL: liveBaseURL(),
		Token:   liveToken(),
	})

	maxTokens := 10
	resp, err := client.Chat.Completions.Create(context.Background(), meshapi.ChatCompletionParams{
		Model:          strPtr(liveModel()),
		Messages:       []meshapi.ChatMessage{{Role: "user", Content: "Reply with the word: ok"}},
		MaxTokens:      &maxTokens,
		FallbackModels: []string{liveSecondModel()},
	})
	if err != nil {
		t.Fatalf("chat create with FallbackModels: %v", err)
	}
	if resp.Model == "" {
		t.Error("expected a model on the response")
	}
	if len(resp.Choices) == 0 || resp.Choices[0].Message == nil {
		t.Fatal("expected a message in the response")
	}
}

// TestLive_Resilience_UnreachableGateway verifies that against an unreachable
// gateway retry events fire, the chain advances, and the last error
// propagates.
func TestLive_Resilience_UnreachableGateway(t *testing.T) {
	events := &liveEventLog{}
	timeoutMs := 2_000
	client := meshapi.New(meshapi.Config{
		// A privileged, never-bound localhost port — connect fails instantly with
		// ECONNREFUSED (a network error, NOT a timeout), which is what we want to
		// exercise: retryable + fallback-eligible. TEST-NET-1 (192.0.2.x) is unroutable
		// but on networks that silently drop its packets the connect would instead time
		// out, and timeouts are deliberately never retried — making this test flaky.
		BaseURL:   "http://127.0.0.1:1",
		Token:     liveToken(),
		TimeoutMs: &timeoutMs,
		Retry: &meshapi.RetryPolicy{
			MaxRetries:          intPtr(1),
			BackoffBaseMs:       intPtr(10),
			BackoffMaxMs:        intPtr(20),
			RetryOnNetworkError: boolPtr(true),
		},
		Fallback: &meshapi.FallbackConfig{Models: []string{liveSecondModel()}},
		Logger:   events.log,
	})

	_, err := client.Chat.Completions.Create(context.Background(), meshapi.ChatCompletionParams{
		Model:    strPtr(liveModel()),
		Messages: []meshapi.ChatMessage{{Role: "user", Content: "hello"}},
	})
	if err == nil {
		t.Fatal("expected an error from the unreachable gateway")
	}

	// Depending on the OS, connecting to TEST-NET-1 either fails fast
	// (EHOSTUNREACH — a retryable network error) or hangs until the client
	// timeout (a timeout — never retried, and terminal for the chain).
	if strings.Contains(strings.ToLower(err.Error()), "timeout") ||
		strings.Contains(strings.ToLower(err.Error()), "deadline") {
		t.Skipf("TEST-NET-1 connect timed out instead of failing fast; timeouts are never retried: %v", err)
	}

	// Each model attempt retried once on the network error…
	networkRetries := 0
	for _, e := range events.ofKind("retry") {
		if e.(meshapi.RetryEvent).Reason == meshapi.RetryReasonNetworkError {
			networkRetries++
		}
	}
	if networkRetries < 1 {
		t.Errorf("expected network-error retry events, got %d (events: %+v)", networkRetries, events.events)
	}

	// …and the chain advanced to the fallback model before giving up.
	fbs := events.ofKind("fallback")
	if len(fbs) == 0 {
		t.Fatal("expected a fallback event")
	}
	fb := fbs[0].(meshapi.FallbackEvent)
	if fb.FromModel != liveModel() {
		t.Errorf("FromModel: expected %q, got %q", liveModel(), fb.FromModel)
	}
	if fb.ToModel != liveSecondModel() {
		t.Errorf("ToModel: expected %q, got %q", liveSecondModel(), fb.ToModel)
	}
}
