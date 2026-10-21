// Package meshapi is a Go client for the MeshAPI AI model gateway.
//
// # Quick start
//
//	client := meshapi.NewClient(meshapi.Config{
//	    BaseURL: "http://localhost:8000",
//	    Token:   "rsk_...",
//	})
//
//	model := "openai/gpt-4o-mini"
//	resp, err := client.Chat.Completions.Create(ctx, meshapi.ChatCompletionParams{
//	    Model:    &model,
//	    Messages: []meshapi.ChatMessage{{Role: "user", Content: "Hello!"}},
//	})
package meshapi

import "net/http"

// Version is the current SDK version.
const Version = "0.1.12"

// Config holds the client configuration.
type Config struct {
	// BaseURL is the MeshAPI gateway base URL (required).
	BaseURL string
	// Token is the Bearer token for this auth realm (required).
	Token string
	// TimeoutMs is the request timeout in milliseconds (default 60_000).
	// For streaming requests this applies to TTFB only.
	TimeoutMs *int
	// MaxRetries is the number of retry attempts on retryable errors (default 3).
	//
	// Deprecated: use Retry.MaxRetries — this alias maps onto it, and
	// Retry.MaxRetries wins when both are set.
	MaxRetries *int
	// Retry is the transport retry policy: which statuses to retry, backoff
	// shape, whether to honour Retry-After, and (opt-in) network-error retry.
	// Streaming requests are never retried.
	Retry *RetryPolicy
	// Fallback is the client-side model-fallback chain for
	// Chat.Completions.Create (non-streaming): when the primary model's
	// request exhausts its retries on a transient error, the SDK re-issues it
	// against each model in the chain until one succeeds. Each hop fires a
	// FallbackEvent.
	Fallback *FallbackConfig
	// Logger is a structured sink for resilience events — every transport
	// retry, every fallback hop, and every gateway-side routing outcome
	// (parsed from the X-Mesh-Routing-* response headers). Use this to pipe
	// into your own logging framework; use Debug for ready-made readable
	// lines instead.
	Logger func(ResilienceEvent)
	// Debug prints readable resilience lines to stderr ("[meshapi] retrying
	// POST …"). Gateway-routing lines are printed only when interesting (a
	// retry or a provider fallback actually happened). Independent of Logger
	// (default false).
	Debug bool
	// HTTPClient allows injecting a custom *http.Client (optional).
	HTTPClient *http.Client
}

func (c Config) timeoutMs() int {
	if c.TimeoutMs != nil {
		return *c.TimeoutMs
	}
	return defaultTimeoutMs
}

// Client is the MeshAPI SDK client.
//
// One instance = one auth realm. Use separate instances for different tokens:
//
//	inferenceClient := meshapi.New(meshapi.Config{Token: "rsk_..."})
//	mgmtClient      := meshapi.New(meshapi.Config{Token: "<jwt>"})
type Client struct {
	Chat        *ChatResource
	Responses   *ResponsesResource
	Embeddings  *EmbeddingsResource
	Compare     *CompareResource
	Batches     *BatchesResource
	Models      *ModelsResource
	Templates   *TemplatesResource
	Images      *ImagesResource
	RAG         *RagResource
	Realtime    *RealtimeResource
	Audio       *AudioResource
	Videos      *VideosResource
	Moderations *ModerationsResource
	Web         *WebResource
	Router      *RouterResource
}

// New creates a new MeshAPI client with the given configuration.
func New(cfg Config) *Client {
	http := newHTTPClient(cfg)
	return &Client{
		Chat: &ChatResource{
			Completions: &CompletionsResource{http: http},
		},
		Responses:   &ResponsesResource{http: http},
		Embeddings:  &EmbeddingsResource{http: http},
		Compare:     &CompareResource{http: http},
		Batches:     &BatchesResource{http: http},
		Models:      &ModelsResource{http: http},
		Templates:   &TemplatesResource{http: http},
		Images:      &ImagesResource{http: http},
		RAG:         &RagResource{http: http},
		Realtime:    &RealtimeResource{http: http},
		Audio:       &AudioResource{http: http},
		Videos:      &VideosResource{http: http},
		Moderations: &ModerationsResource{http: http},
		Web:         &WebResource{http: http},
		Router:      &RouterResource{http: http},
	}
}
