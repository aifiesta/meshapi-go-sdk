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

// APIVersion is the dated MeshAPI contract version this SDK release was built
// against, sent as X-Mesh-Version on every request.
//
// Distinct from [Version]: that identifies this SDK build, this identifies the API
// contract it parses. APIVersion changes only when the SDK is updated for a newer
// response shape, which is rarer than a release — bump it together with whatever
// type changes that entails, and note it in CHANGELOG.md.
//
// Not sent on the realtime WebSocket handshake: the gateway's versioning applies to
// HTTP requests only (its middleware returns early on non-HTTP scopes), so a pin
// there would be a header nobody reads. Realtime negotiates its version separately.
const APIVersion = "2026-08"

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
	// APIVersion pins the dated MeshAPI contract version sent as X-Mesh-Version.
	//
	// nil (unset) sends [APIVersion], the version this SDK was built against. A
	// pointer to a non-empty string pins that version instead — useful if you have
	// migrated ahead of this SDK release. A pointer to the EMPTY string sends no
	// header at all, taking the gateway's baseline whatever it may become.
	//
	// nil and &"" mean different things on purpose: unset takes the SDK's version,
	// empty is an explicit opt-out.
	//
	// The gateway rejects a version it does not serve with 400 invalid_api_version
	// rather than falling back, so a typo cannot leave you believing you are pinned
	// when you are not.
	APIVersion *string
}

// apiVersion resolves the version to send, or "" to send no header.
func (c Config) apiVersion() string {
	if c.APIVersion == nil {
		return APIVersion
	}
	return *c.APIVersion
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
