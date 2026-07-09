package meshapi

// ── Resilience: retry policy, fallback chain, and observability events ───────
//
// Two independent layers, mirroring the gateway's design:
//
//  1. TRANSPORT RETRY (httpClient): re-send the same request on transient
//     failures (429/502/503/504, optionally network errors). Configured via
//     Config.Retry. Streaming requests are never retried.
//
//  2. MODEL FALLBACK (Chat.Completions.Create): after the transport gives up,
//     try the same request against the next model in a configured chain.
//     Configured via Config.Fallback or per-call FallbackModels.
//     Client-side only — the gateway additionally does its own server-side
//     retry + cross-provider fallback when the API key's routing_policy
//     enables it; that outcome is reported back via X-Mesh-Routing-*
//     response headers and surfaced as a GatewayRoutingEvent.
//
// Every retry, fallback hop, and gateway-routing outcome is observable through
// Config.Logger (structured events) and/or Config.Debug (readable stderr
// lines), so it is always clear which requests were retried and which were
// served by a fallback.

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// RetryPolicy is the transport-level retry policy. All fields are optional —
// nil fields keep the defaults.
type RetryPolicy struct {
	// MaxRetries is the maximum number of retries after the initial attempt
	// (default 3).
	MaxRetries *int
	// RetryOnStatus lists the HTTP status codes that trigger a retry of the
	// same request (default 429, 502, 503, 504).
	RetryOnStatus []int
	// BackoffBaseMs is the base delay for exponential backoff — doubles per
	// attempt, ±20% jitter (default 500).
	BackoffBaseMs *int
	// BackoffMaxMs is the upper bound on a single backoff delay (default 30_000).
	BackoffMaxMs *int
	// RespectRetryAfter honours the server's Retry-After response header when
	// present (default true).
	RespectRetryAfter *bool
	// RetryOnNetworkError also retries when the request fails before any
	// response arrives (DNS failure, connection refused/reset). Off by
	// default: a network error is ambiguous — the request may have reached
	// the server, and POST bodies are not idempotent. Context cancellations
	// and timeouts are never retried (default false).
	RetryOnNetworkError *bool
}

// resolvedRetryPolicy is a RetryPolicy with every field populated.
type resolvedRetryPolicy struct {
	maxRetries          int
	retryOnStatus       map[int]bool
	backoffBaseMs       int
	backoffMaxMs        int
	respectRetryAfter   bool
	retryOnNetworkError bool
}

// DefaultRetryStatusCodes are the statuses retried by the transport when
// RetryPolicy.RetryOnStatus is unset.
var DefaultRetryStatusCodes = []int{429, 502, 503, 504}

// DefaultFallbackStatusCodes are the statuses that advance the chat model
// fallback chain when FallbackConfig.OnStatus is unset.
var DefaultFallbackStatusCodes = []int{502, 503, 504}

func resolveRetryPolicy(policy *RetryPolicy, legacyMaxRetries *int) resolvedRetryPolicy {
	r := resolvedRetryPolicy{
		maxRetries:          defaultMaxRetries,
		retryOnStatus:       statusSet(DefaultRetryStatusCodes),
		backoffBaseMs:       defaultBackoffBaseMs,
		backoffMaxMs:        defaultBackoffMaxMs,
		respectRetryAfter:   true,
		retryOnNetworkError: false,
	}
	// Retry.MaxRetries wins over the deprecated top-level Config.MaxRetries.
	if legacyMaxRetries != nil {
		r.maxRetries = *legacyMaxRetries
	}
	if policy == nil {
		return r
	}
	if policy.MaxRetries != nil {
		r.maxRetries = *policy.MaxRetries
	}
	if policy.RetryOnStatus != nil {
		r.retryOnStatus = statusSet(policy.RetryOnStatus)
	}
	if policy.BackoffBaseMs != nil {
		r.backoffBaseMs = *policy.BackoffBaseMs
	}
	if policy.BackoffMaxMs != nil {
		r.backoffMaxMs = *policy.BackoffMaxMs
	}
	if policy.RespectRetryAfter != nil {
		r.respectRetryAfter = *policy.RespectRetryAfter
	}
	if policy.RetryOnNetworkError != nil {
		r.retryOnNetworkError = *policy.RetryOnNetworkError
	}
	return r
}

func statusSet(codes []int) map[int]bool {
	set := make(map[int]bool, len(codes))
	for _, c := range codes {
		set[c] = true
	}
	return set
}

// FallbackConfig is the client-side model-fallback chain for
// Chat.Completions.Create (non-streaming).
type FallbackConfig struct {
	// Models is the ordered list of models to try when the primary model's
	// request fails. Distinct from the Models request param (a server-side,
	// provider-handled fallback list): this chain is driven by the SDK, so it
	// works regardless of provider and is visible in your logs hop by hop.
	Models []string
	// OnStatus lists the error statuses eligible for advancing to the next
	// model (default 502, 503, 504). Terminal errors (auth, validation,
	// billing) never advance the chain.
	OnStatus []int
}

// ── Observability events ──────────────────────────────────────────────────────

// ResilienceEvent is a structured resilience event: a transport retry, a chat
// fallback hop, or a gateway-side routing outcome. Type-switch on the concrete
// types (RetryEvent, FallbackEvent, GatewayRoutingEvent) or dispatch on Kind().
type ResilienceEvent interface {
	// Kind returns "retry", "fallback", or "gateway-routing".
	Kind() string
}

// RetryReason values for RetryEvent.Reason.
const (
	RetryReasonStatus       = "status"
	RetryReasonNetworkError = "network-error"
)

// RetryEvent reports that the same request is being re-sent after a transient
// failure.
type RetryEvent struct {
	Method string
	Path   string
	// Attempt is the 1-based attempt that just failed; the next send is
	// Attempt + 1.
	Attempt    int
	MaxRetries int
	// Status is the HTTP status that triggered the retry; 0 for a network error.
	Status int
	// RequestID is the gateway request id of the failed attempt, when a
	// response was received.
	RequestID string
	DelayMs   float64
	// Reason is RetryReasonStatus or RetryReasonNetworkError.
	Reason string
}

// Kind implements ResilienceEvent.
func (RetryEvent) Kind() string { return "retry" }

// FallbackEvent reports that the chat fallback chain is advancing to the next
// model.
type FallbackEvent struct {
	FromModel string
	ToModel   string
	// ChainIndex is the 0-based index of ToModel within the configured chain.
	ChainIndex  int
	ChainLength int
	// Status is the HTTP status of the failed attempt; 0 for a network error.
	Status    int
	ErrorCode string
	RequestID string
}

// Kind implements ResilienceEvent.
func (FallbackEvent) Kind() string { return "fallback" }

// GatewayRoutingEvent reports that the GATEWAY retried or fell back
// server-side while serving this request — parsed from the X-Mesh-Routing-*
// response headers (present when the API key's routing_policy is active).
// Fallback true means a different provider than the primary served the request.
type GatewayRoutingEvent struct {
	Path           string
	Attempts       int
	Fallback       bool
	ServedProvider string
	RequestID      string
}

// Kind implements ResilienceEvent.
func (GatewayRoutingEvent) Kind() string { return "gateway-routing" }

// ── Built-in debug printer ────────────────────────────────────────────────────

// FormatResilienceEvent renders an event as a single readable line, e.g.
//
//	retrying POST /v1/chat/completions (attempt 1/3 failed: 503, next in 512ms) [req_abc]
//	falling back openai/gpt-4o → anthropic/claude-sonnet-5 (1/2: 503 provider_not_available)
//	gateway served /v1/chat/completions via bedrock (2 attempts, provider fallback) [req_abc]
func FormatResilienceEvent(event ResilienceEvent) string {
	switch e := event.(type) {
	case RetryEvent:
		why := strconv.Itoa(e.Status)
		if e.Reason == RetryReasonNetworkError {
			why = "network error"
		}
		return fmt.Sprintf("retrying %s %s (attempt %d/%d failed: %s, next in %dms)%s",
			e.Method, e.Path, e.Attempt, e.MaxRetries+1, why,
			int(math.Round(e.DelayMs)), ridSuffix(e.RequestID))
	case FallbackEvent:
		var parts []string
		if e.Status != 0 {
			parts = append(parts, strconv.Itoa(e.Status))
		}
		if e.ErrorCode != "" {
			parts = append(parts, e.ErrorCode)
		}
		why := strings.Join(parts, " ")
		if why == "" {
			why = "network error"
		}
		return fmt.Sprintf("falling back %s → %s (%d/%d: %s)%s",
			e.FromModel, e.ToModel, e.ChainIndex+1, e.ChainLength, why,
			ridSuffix(e.RequestID))
	case GatewayRoutingEvent:
		served := ""
		if e.ServedProvider != "" {
			served = " via " + e.ServedProvider
		}
		detail := fmt.Sprintf("%d attempts", e.Attempts)
		if e.Fallback {
			detail += ", provider fallback"
		}
		return fmt.Sprintf("gateway served %s%s (%s)%s", e.Path, served, detail, ridSuffix(e.RequestID))
	default:
		return fmt.Sprintf("%s event", event.Kind())
	}
}

func ridSuffix(requestID string) string {
	if requestID == "" {
		return ""
	}
	return " [" + requestID + "]"
}
