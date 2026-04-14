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
const Version = "0.1.0"

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
	MaxRetries *int
	// HTTPClient allows injecting a custom *http.Client (optional).
	HTTPClient *http.Client
}

func (c Config) timeoutMs() int {
	if c.TimeoutMs != nil {
		return *c.TimeoutMs
	}
	return defaultTimeoutMs
}

func (c Config) maxRetries() int {
	if c.MaxRetries != nil {
		return *c.MaxRetries
	}
	return defaultMaxRetries
}

// Client is the MeshAPI SDK client.
//
// One instance = one auth realm. Use separate instances for different tokens:
//
//	inferenceClient := meshapi.New(meshapi.Config{Token: "rsk_..."})
//	mgmtClient      := meshapi.New(meshapi.Config{Token: "<jwt>"})
type Client struct {
	Chat      *ChatResource
	Responses *ResponsesResource
	Models    *ModelsResource
	Templates *TemplatesResource
}

// New creates a new MeshAPI client with the given configuration.
func New(cfg Config) *Client {
	http := newHTTPClient(cfg)
	return &Client{
		Chat: &ChatResource{
			Completions: &CompletionsResource{http: http},
		},
		Responses: &ResponsesResource{http: http},
		Models:    &ModelsResource{http: http},
		Templates: &TemplatesResource{http: http},
	}
}
