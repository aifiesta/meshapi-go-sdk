// Package meshapi provides a typed Go client for the MeshAPI AI model gateway.
package meshapi

import "encoding/json"

// ---------------------------------------------------------------------------
// Chat Completions
// ---------------------------------------------------------------------------

// ChatMessage represents a single message in the conversation.
type ChatMessage struct {
	Role             string                   `json:"role"`
	Content          interface{}              `json:"content,omitempty"` // string or []ContentPart
	Name             *string                  `json:"name,omitempty"`
	ToolCallID       *string                  `json:"tool_call_id,omitempty"`
	ToolCalls        []ToolCall               `json:"tool_calls,omitempty"`
	ReasoningDetails []map[string]interface{} `json:"reasoning_details,omitempty"`
}

// VideoURL holds the URL for a video content part.
type VideoURL struct {
	URL string `json:"url"`
}

// ContentPart is one element of a multimodal message content array.
type ContentPart struct {
	Type       string      `json:"type"`
	Text       *string     `json:"text,omitempty"`
	ImageURL   *ImageURL   `json:"image_url,omitempty"`
	InputAudio *InputAudio `json:"input_audio,omitempty"`
	VideoURL   *VideoURL   `json:"video_url,omitempty"`
	Fps        *string     `json:"fps,omitempty"`
}

// ImageURL holds the URL and rendering detail for an image content part.
type ImageURL struct {
	URL    string  `json:"url"`
	Detail *string `json:"detail,omitempty"`
}

// InputAudio holds an audio content part. One of Data, URI, or URL must be
// provided along with Format.
type InputAudio struct {
	Data   *string `json:"data,omitempty"`
	URI    *string `json:"uri,omitempty"`
	URL    *string `json:"url,omitempty"`
	Format string  `json:"format"`
}

// ToolCall represents a tool invocation in an assistant message.
type ToolCall struct {
	ID               string           `json:"id"`
	Type             string           `json:"type"`
	Function         ToolCallFunction `json:"function"`
	ThoughtSignature *string          `json:"thought_signature,omitempty"`
}

// ToolCallFunction holds the name and JSON-encoded arguments for a tool call.
type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// Tool defines a callable function available to the model.
type Tool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

// ToolFunction describes a callable function.
type ToolFunction struct {
	Name        string      `json:"name"`
	Description *string     `json:"description,omitempty"`
	Parameters  interface{} `json:"parameters,omitempty"`
}

// ToolChoice controls model tool use.
// Use "auto", "none", "required", or ToolChoiceObject.
type ToolChoice interface{}

// ToolChoiceObject selects a specific function.
type ToolChoiceObject struct {
	Type     string             `json:"type"`
	Function ToolChoiceFunction `json:"function"`
}

// ToolChoiceFunction names the function to call.
type ToolChoiceFunction struct {
	Name string `json:"name"`
}

type AudioOutputOptions struct {
	Voice  *string `json:"voice,omitempty"`
	Format *string `json:"format,omitempty"`
}

type ImageOptions struct {
	N              *int    `json:"n,omitempty"`
	Size           *string `json:"size,omitempty"`
	Quality        *string `json:"quality,omitempty"`
	ResponseFormat *string `json:"response_format,omitempty"`
}

// ChatCompletionParams is the request body for POST /v1/chat/completions.
type ChatCompletionParams struct {
	Messages         []ChatMessage          `json:"messages"`
	Model            *string                `json:"model,omitempty"`
	Stream           *bool                  `json:"stream,omitempty"`
	Template         *string                `json:"template,omitempty"`
	Variables        map[string]string      `json:"variables,omitempty"`
	SessionID        *string                `json:"session_id,omitempty"`
	Temperature      *float64               `json:"temperature,omitempty"`
	MaxTokens        *int                   `json:"max_tokens,omitempty"`
	TopP             *float64               `json:"top_p,omitempty"`
	FrequencyPenalty *float64               `json:"frequency_penalty,omitempty"`
	PresencePenalty  *float64               `json:"presence_penalty,omitempty"`
	Stop             interface{}            `json:"stop,omitempty"` // string or []string
	Seed             *int                   `json:"seed,omitempty"`
	Tools            []Tool                 `json:"tools,omitempty"`
	ToolChoice       interface{}            `json:"tool_choice,omitempty"`
	ResponseFormat   map[string]interface{} `json:"response_format,omitempty"`
	Transforms       []string               `json:"transforms,omitempty"`
	Models           []string               `json:"models,omitempty"`
	User             *string                `json:"user,omitempty"`
	Modality         *string                `json:"modality,omitempty"`
	Image            *ImageOptions          `json:"image,omitempty"`
	AsyncMode        *bool                  `json:"async_mode,omitempty"`
	Modalities       []string               `json:"modalities,omitempty"`
	Audio            *AudioOutputOptions    `json:"audio,omitempty"`
	// Cache enables prompt caching for this request (null = server default).
	Cache *bool `json:"cache,omitempty"`
	// Timeout overrides the server's upstream-provider timeout (default 300 s).
	// Set this for requests that may take longer than 5 minutes. This is
	// independent of the SDK-level TimeoutMs option on Config, which controls
	// the HTTP client timeout.
	Timeout *float64 `json:"timeout,omitempty"`
}

// UsageInfo holds token counts for a completion.
type UsageInfo struct {
	PromptTokens               int                    `json:"prompt_tokens"`
	CompletionTokens           int                    `json:"completion_tokens"`
	TotalTokens                int                    `json:"total_tokens"`
	PromptTokensDetails        map[string]interface{} `json:"prompt_tokens_details,omitempty"`
	CompletionTokensDetails    map[string]interface{} `json:"completion_tokens_details,omitempty"`
	ClassifierPromptTokens     *int                   `json:"classifier_prompt_tokens,omitempty"`
	ClassifierCompletionTokens *int                   `json:"classifier_completion_tokens,omitempty"`
	ClassifierTokens           *int                   `json:"classifier_tokens,omitempty"`
}

// ChatCompletionMessage is a completed message in a non-streaming response.
type ChatCompletionMessage struct {
	Role      string                 `json:"role"`
	Content   *string                `json:"content,omitempty"`
	ToolCalls []ToolCall             `json:"tool_calls,omitempty"`
	Audio     map[string]interface{} `json:"audio,omitempty"`
}

// ChatCompletionChoice is one result choice in a non-streaming response.
type ChatCompletionChoice struct {
	Index        int                    `json:"index"`
	Message      *ChatCompletionMessage `json:"message,omitempty"`
	FinishReason *string                `json:"finish_reason,omitempty"`
	Logprobs     interface{}            `json:"logprobs,omitempty"`
}

// ChatCompletionResponse is the full non-streaming response body.
type ChatCompletionResponse struct {
	ID                string                 `json:"id"`
	Object            string                 `json:"object"`
	Created           int64                  `json:"created"`
	Model             string                 `json:"model"`
	Choices           []ChatCompletionChoice `json:"choices"`
	Usage             *UsageInfo             `json:"usage,omitempty"`
	SystemFingerprint *string                `json:"system_fingerprint,omitempty"`
}

// ChatCompletionChunkDelta is the partial content in a streaming chunk.
type ChatCompletionChunkDelta struct {
	Role      *string                `json:"role,omitempty"`
	Content   *string                `json:"content,omitempty"`
	ToolCalls []ToolCall             `json:"tool_calls,omitempty"`
	Audio     map[string]interface{} `json:"audio,omitempty"`
}

// ChatCompletionChunkChoice is one choice in a streaming chunk.
type ChatCompletionChunkChoice struct {
	Index        int                       `json:"index"`
	Delta        *ChatCompletionChunkDelta `json:"delta,omitempty"`
	FinishReason *string                   `json:"finish_reason,omitempty"`
}

// ChatCompletionChunk is a single SSE chunk in a streaming completion.
type ChatCompletionChunk struct {
	ID      string                      `json:"id"`
	Object  string                      `json:"object"`
	Created int64                       `json:"created"`
	Model   string                      `json:"model"`
	Choices []ChatCompletionChunkChoice `json:"choices"`
	Usage   *UsageInfo                  `json:"usage,omitempty"`
	Cost    *string                     `json:"cost,omitempty"`
}

// ---------------------------------------------------------------------------
// Models
// ---------------------------------------------------------------------------

// ModelPricing holds per-token pricing for a model. All values are strings per
// the spec — do not coerce to float.
type ModelPricing struct {
	// Required
	PromptUSDPer1K     *string `json:"prompt_usd_per_1k,omitempty"`
	CompletionUSDPer1K *string `json:"completion_usd_per_1k,omitempty"`
	// Optional
	PricingUnit                          *string `json:"pricing_unit,omitempty"`
	PromptUSDPer1M                       *string `json:"prompt_usd_per_1m,omitempty"`
	CompletionUSDPer1M                   *string `json:"completion_usd_per_1m,omitempty"`
	ImageOutputUSDPerImage               *string `json:"image_output_usd_per_image,omitempty"`
	RequestUSD                           *string `json:"request_usd,omitempty"`
	LongContextInputUSDPer1M             *string `json:"long_context_input_usd_per_1m,omitempty"`
	LongContextOutputUSDPer1M            *string `json:"long_context_output_usd_per_1m,omitempty"`
	CacheReadInputUSDPer1M               *string `json:"cache_read_input_usd_per_1m,omitempty"`
	CacheWriteInputUSDPer1M              *string `json:"cache_write_input_usd_per_1m,omitempty"`
	CacheReadAudioInputUSDPer1M          *string `json:"cache_read_audio_input_usd_per_1m,omitempty"`
	LongContextCacheReadInputUSDPer1M    *string `json:"long_context_cache_read_input_usd_per_1m,omitempty"`
	LongContextCacheWriteInputUSDPer1M   *string `json:"long_context_cache_write_input_usd_per_1m,omitempty"`
	BatchInputUSDPer1M                   *string `json:"batch_input_usd_per_1m,omitempty"`
	BatchOutputUSDPer1M                  *string `json:"batch_output_usd_per_1m,omitempty"`
	TrainingUSDPer1M                     *string `json:"training_usd_per_1m,omitempty"`
	FineTunedInputUSDPer1M               *string `json:"fine_tuned_input_usd_per_1m,omitempty"`
	FineTunedOutputUSDPer1M              *string `json:"fine_tuned_output_usd_per_1m,omitempty"`
	AudioInputUSDPer1M                   *string `json:"audio_input_usd_per_1m,omitempty"`
	AudioOutputUSDPer1M                  *string `json:"audio_output_usd_per_1m,omitempty"`
	TranscriptionUSDPer1M                *string `json:"transcription_usd_per_1m,omitempty"`
	CachedAudioInputUSDPer1M             *string `json:"cached_audio_input_usd_per_1m,omitempty"`
	CachedTextInputUSDPer1M              *string `json:"cached_text_input_usd_per_1m,omitempty"`
	CacheHitUSDPer1M                     *string `json:"cache_hit_usd_per_1m,omitempty"`
	OutputWithAudioUSDPer1M              *string `json:"output_with_audio_usd_per_1m,omitempty"`
	OutputWithVideoUSDPer1M              *string `json:"output_with_video_usd_per_1m,omitempty"`
	ImageInputUSDPerImage                *string `json:"image_input_usd_per_image,omitempty"`
	ImageOutputSize                      *string `json:"image_output_size,omitempty"`
	EffectiveDate                        *string `json:"effective_date,omitempty"`
	DeprecatedDate                       *string `json:"deprecated_date,omitempty"`
	Notes                                *string `json:"notes,omitempty"`
	SourceURL                            *string `json:"source_url,omitempty"`
	DiscountPct                          *string `json:"discount_pct,omitempty"`
}

// ModelInfo describes an available model.
type ModelInfo struct {
	// Required fields
	ID                     string        `json:"id"`
	Name                   string        `json:"name"`
	ContextLength          *int          `json:"context_length,omitempty"`
	IsFree                 bool          `json:"is_free"`
	Pricing                *ModelPricing `json:"pricing,omitempty"`
	SupportsThinking       bool          `json:"supports_thinking"`
	SupportsCompletionsAPI bool          `json:"supports_completions_api"`
	SupportsResponsesAPI   bool          `json:"supports_responses_api"`
	ModelType              string        `json:"model_type"`
	InputModalities        []string      `json:"input_modalities,omitempty"`
	OutputModalities       []string      `json:"output_modalities,omitempty"`
	// Optional fields
	Brand                          *string  `json:"brand,omitempty"`
	Provider                       *string  `json:"provider,omitempty"`
	Description                    *string  `json:"description,omitempty"`
	SupportsRealtime               bool     `json:"supports_realtime"`
	SupportsEmbeddings             bool     `json:"supports_embeddings"`
	SupportsTools                  bool     `json:"supports_tools"`
	SupportsStructuredOutput       bool     `json:"supports_structured_output"`
	// SupportsSystemPrompt defaults to true in the spec (unlike the other
	// supports_* flags which default to false), so it is a pointer: nil means
	// the field was omitted and should be treated as true.
	SupportsSystemPrompt           *bool    `json:"supports_system_prompt,omitempty"`
	SupportsBatching               bool     `json:"supports_batching"`
	SupportsBackgroundResponse     bool     `json:"supports_background_response"`
	SupportsVideoGeneration        bool     `json:"supports_video_generation"`
	SupportsImageEdit              bool     `json:"supports_image_edit"`
	SupportsImageInpaint           bool     `json:"supports_image_inpaint"`
	SupportsImageOutpaint          bool     `json:"supports_image_outpaint"`
	SupportsImageMix               bool     `json:"supports_image_mix"`
	SupportsImageReframe           bool     `json:"supports_image_reframe"`
	SupportsImageUpscale           bool     `json:"supports_image_upscale"`
	SupportsImageRemoveBackground  bool     `json:"supports_image_remove_background"`
	SupportsImageReference         bool     `json:"supports_image_reference"`
	ContextWindow                  *int     `json:"context_window,omitempty"`
	StandardContextThreshold       *int     `json:"standard_context_threshold,omitempty"`
	RealtimeSessionMaxTokens       *int     `json:"realtime_session_max_tokens,omitempty"`
	RealtimeMaxConcurrentPerOwner  *int     `json:"realtime_max_concurrent_per_owner,omitempty"`
	IsComposite                    bool     `json:"is_composite"`
	CompositeModels                []string `json:"composite_models,omitempty"`
}

// ListModelsParams holds optional query parameters for listing models.
type ListModelsParams struct {
	Free     *bool   // nil = no filter
	Type     *string // "text" | "embedding" | "image" | "audio" | "video"
	Provider *string // e.g. "openai", "amazon-bedrock", "vertex"
}

// ---------------------------------------------------------------------------
// Templates
// ---------------------------------------------------------------------------

// CreateTemplateParams is the request body for POST /v1/templates.
type CreateTemplateParams struct {
	Name        string                   `json:"name"`
	Description *string                  `json:"description,omitempty"`
	System      *string                  `json:"system,omitempty"`
	Messages    []map[string]interface{} `json:"messages,omitempty"`
	Model       *string                  `json:"model,omitempty"`
	Params      map[string]interface{}   `json:"params,omitempty"`
	Variables   []string                 `json:"variables,omitempty"`
	// TeamID scopes the template to a specific team (optional).
	TeamID *string `json:"team_id,omitempty"`
}

// UpdateTemplateParams is the request body for PATCH /v1/templates/{id}.
type UpdateTemplateParams struct {
	Name        *string                  `json:"name,omitempty"`
	Description *string                  `json:"description,omitempty"`
	System      *string                  `json:"system,omitempty"`
	Messages    []map[string]interface{} `json:"messages,omitempty"`
	Model       *string                  `json:"model,omitempty"`
	Params      map[string]interface{}   `json:"params,omitempty"`
	Variables   []string                 `json:"variables,omitempty"`
}

// TemplateSummary is the response shape for all template operations.
type TemplateSummary struct {
	ID          string                   `json:"id"`
	Name        string                   `json:"name"`
	Owner       *string                  `json:"owner"`
	IsGlobal    bool                     `json:"is_global"`
	Description *string                  `json:"description,omitempty"`
	System      *string                  `json:"system,omitempty"`
	Messages    []map[string]interface{} `json:"messages,omitempty"`
	Model       *string                  `json:"model,omitempty"`
	Params      map[string]interface{}   `json:"params,omitempty"`
	Variables   []string                 `json:"variables,omitempty"`
	CreatedAt   string                   `json:"created_at"`
	UpdatedAt   string                   `json:"updated_at"`
}

type ProviderPreferences struct {
	Order             []string `json:"order,omitempty"`
	AllowFallbacks    *bool    `json:"allow_fallbacks,omitempty"`
	RequireParameters *bool    `json:"require_parameters,omitempty"`
	DataCollection    *string  `json:"data_collection,omitempty"`
}

// ImageEmbeddingUrl holds a URL for a multimodal image embedding input.
type ImageEmbeddingUrl struct {
	URL string `json:"url"`
}

// VideoEmbeddingUrl holds a URL for a multimodal video embedding input.
type VideoEmbeddingUrl struct {
	URL string `json:"url"`
}

// MultimodalEmbeddingInput is one element of a multimodal embeddings input
// array. Type is one of "text", "image_url", or "video_url".
type MultimodalEmbeddingInput struct {
	Type     string             `json:"type"`
	Text     *string            `json:"text,omitempty"`
	ImageURL *ImageEmbeddingUrl `json:"image_url,omitempty"`
	VideoURL *VideoEmbeddingUrl `json:"video_url,omitempty"`
}

type EmbeddingsParams struct {
	Model           *string     `json:"model,omitempty"`
	Input           interface{} `json:"input"` // string | []string | []int | [][]int | []MultimodalEmbeddingInput
	Dimensions      *int        `json:"dimensions,omitempty"`
	EncodingFormat  *string     `json:"encoding_format,omitempty"` // "float" | "base64"
	InputType       *string     `json:"input_type,omitempty"`
	Provider        interface{} `json:"provider,omitempty"`
	User            *string     `json:"user,omitempty"`
	Instructions    *string     `json:"instructions,omitempty"`
	SparseEmbedding map[string]interface{} `json:"sparse_embedding,omitempty"`
}

// EmbeddingVector holds an embedding value that can be either a float array
// (encoding_format=float, the default) or a base64 string
// (encoding_format=base64). Use Floats() or Base64() to access the value.
type EmbeddingVector struct {
	floats []float64
	base64 string
	isB64  bool
}

// UnmarshalJSON implements json.Unmarshaler. It accepts both a JSON array of
// floats and a JSON string (base64-encoded embedding).
func (e *EmbeddingVector) UnmarshalJSON(data []byte) error {
	// Try float array first (most common case).
	var floats []float64
	if err := json.Unmarshal(data, &floats); err == nil {
		e.floats = floats
		e.isB64 = false
		return nil
	}
	// Fall back to string (base64).
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	e.base64 = s
	e.isB64 = true
	return nil
}

// MarshalJSON implements json.Marshaler.
func (e EmbeddingVector) MarshalJSON() ([]byte, error) {
	if e.isB64 {
		return json.Marshal(e.base64)
	}
	return json.Marshal(e.floats)
}

// Floats returns the embedding as a float slice. Returns nil if the embedding
// is base64-encoded.
func (e *EmbeddingVector) Floats() []float64 {
	if e.isB64 {
		return nil
	}
	return e.floats
}

// Base64 returns the base64-encoded embedding string. Returns "" if the
// embedding is a float array.
func (e *EmbeddingVector) Base64() string {
	if !e.isB64 {
		return ""
	}
	return e.base64
}

// IsBase64 reports whether this embedding was returned as a base64 string.
func (e *EmbeddingVector) IsBase64() bool {
	return e.isB64
}

type EmbeddingItem struct {
	Object    string          `json:"object"`
	Index     int             `json:"index"`
	Embedding EmbeddingVector `json:"embedding"`
}

type EmbeddingsUsage struct {
	PromptTokens *int `json:"prompt_tokens,omitempty"`
	TotalTokens  *int `json:"total_tokens,omitempty"`
}

type EmbeddingsResponse struct {
	Object string           `json:"object"`
	Data   []EmbeddingItem  `json:"data"`
	Model  string           `json:"model"`
	Usage  *EmbeddingsUsage `json:"usage,omitempty"`
}

type ResponsesFunctionTool struct {
	Type        string      `json:"type"`
	Name        string      `json:"name"`
	Description *string     `json:"description,omitempty"`
	Parameters  interface{} `json:"parameters,omitempty"`
	Strict      *bool       `json:"strict,omitempty"`
}

type BuiltinTool struct {
	Type string `json:"type"`
}

type ResponsesParams struct {
	Model           *string                `json:"model,omitempty"`
	Input           interface{}            `json:"input"`
	Template        *string                `json:"template,omitempty"`
	Variables       map[string]string      `json:"variables,omitempty"`
	SessionID       *string                `json:"session_id,omitempty"`
	Stream          *bool                  `json:"stream,omitempty"`
	MaxOutputTokens *int                   `json:"max_output_tokens,omitempty"`
	Temperature     *float64               `json:"temperature,omitempty"`
	TopP            *float64               `json:"top_p,omitempty"`
	Seed            *int                   `json:"seed,omitempty"`
	Reasoning       map[string]interface{} `json:"reasoning,omitempty"`
	Tools           []interface{}          `json:"tools,omitempty"`
	ToolChoice      interface{}            `json:"tool_choice,omitempty"`
	ResponseFormat  map[string]interface{} `json:"response_format,omitempty"`
	Plugins         []interface{}          `json:"plugins,omitempty"`
	User            *string                `json:"user,omitempty"`
	// Fields added in pass-2 audit — all optional/nullable per spec.
	// PreviousResponseID chains this response to a prior one for multi-turn.
	PreviousResponseID *string `json:"previous_response_id,omitempty"`
	// Instructions overrides the system-level instructions for this request.
	Instructions *string `json:"instructions,omitempty"`
	// Thinking is a free-form object controlling chain-of-thought (model-specific).
	Thinking map[string]interface{} `json:"thinking,omitempty"`
	// Caching is a free-form object with prompt-caching settings.
	Caching map[string]interface{} `json:"caching,omitempty"`
	// Store controls whether the response is persisted for later retrieval.
	Store *bool `json:"store,omitempty"`
	// Include lists additional output fields to return (model-specific).
	Include []interface{} `json:"include,omitempty"`
	// ExpireAt is a Unix timestamp after which the stored response may be deleted.
	ExpireAt *int64 `json:"expire_at,omitempty"`
	// MaxToolCalls limits the number of tool calls permitted (1..10).
	MaxToolCalls *int `json:"max_tool_calls,omitempty"`
	// ContextManagement is a free-form object for context window management.
	ContextManagement map[string]interface{} `json:"context_management,omitempty"`
	// Timeout overrides the server's upstream-provider timeout (default 300 s).
	Timeout *float64 `json:"timeout,omitempty"`
}

type ResponsesUsage struct {
	InputTokens             *int                   `json:"input_tokens,omitempty"`
	OutputTokens            *int                   `json:"output_tokens,omitempty"`
	TotalTokens             *int                   `json:"total_tokens,omitempty"`
	PromptTokens            *int                   `json:"prompt_tokens,omitempty"`
	CompletionTokens        *int                   `json:"completion_tokens,omitempty"`
	PromptTokensDetails     map[string]interface{} `json:"prompt_tokens_details,omitempty"`
	CompletionTokensDetails map[string]interface{} `json:"completion_tokens_details,omitempty"`
	ClassifierTokens        *int                   `json:"classifier_tokens,omitempty"`
}

type ResponsesResponse struct {
	ID     *string                `json:"id,omitempty"`
	Object *string                `json:"object,omitempty"`
	Model  *string                `json:"model,omitempty"`
	Output []interface{}          `json:"output,omitempty"`
	Usage  *ResponsesUsage        `json:"usage,omitempty"`
	Status *string                `json:"status,omitempty"`
	Extra  map[string]interface{} `json:"-"`
}

type ResponsesStreamEvent map[string]interface{}

type ModelOverride struct {
	Model        string   `json:"model"`
	Temperature  *float64 `json:"temperature,omitempty"`
	MaxTokens    *int     `json:"max_tokens,omitempty"`
	SystemPrompt *string  `json:"system_prompt,omitempty"`
}

type CompareParams struct {
	Models                 []string          `json:"models"`
	Messages               []ChatMessage     `json:"messages"`
	ModelOverrides         []ModelOverride   `json:"model_overrides,omitempty"`
	ComparisonModel        *string           `json:"comparison_model,omitempty"`
	ComparisonInstructions *string           `json:"comparison_instructions,omitempty"`
	Temperature            *float64          `json:"temperature,omitempty"`
	MaxTokens              *int              `json:"max_tokens,omitempty"`
	Stream                 *bool             `json:"stream,omitempty"`
	Template               *string           `json:"template,omitempty"`
	Variables              map[string]string `json:"variables,omitempty"`
	SkipComparison         *bool             `json:"skip_comparison,omitempty"`
}

type TokenUsage struct {
	PromptTokens     *int `json:"prompt_tokens,omitempty"`
	CompletionTokens *int `json:"completion_tokens,omitempty"`
	TotalTokens      *int `json:"total_tokens,omitempty"`
}

type ModelCompareResult struct {
	Model        string                 `json:"model"`
	ResponseBody map[string]interface{} `json:"response_body,omitempty"`
	Content      *string                `json:"content,omitempty"`
	LatencyMs    int                    `json:"latency_ms"`
	Error        *string                `json:"error,omitempty"`
	ErrorCode    *string                `json:"error_code,omitempty"`
	Usage        *TokenUsage            `json:"usage,omitempty"`
	RequestID    string                 `json:"request_id"`
}

type CompareResponse struct {
	ComparisonID           string               `json:"comparison_id"`
	Object                 string               `json:"object"`
	Created                int64                `json:"created"`
	Models                 []string             `json:"models"`
	Results                []ModelCompareResult `json:"results"`
	Comparison             *string              `json:"comparison,omitempty"`
	ComparisonModel        *string              `json:"comparison_model,omitempty"`
	ComparisonUsage        *TokenUsage          `json:"comparison_usage,omitempty"`
	ComparisonFallbackUsed bool                 `json:"comparison_fallback_used"`
	TotalLatencyMs         int                  `json:"total_latency_ms"`
	Partial                bool                 `json:"partial"`
	SkipComparison         bool                 `json:"skip_comparison"`
}

type CompareStreamEvent map[string]interface{}

type BatchRequestItem struct {
	CustomID string                 `json:"custom_id"`
	Method   string                 `json:"method,omitempty"`
	URL      string                 `json:"url,omitempty"`
	Body     map[string]interface{} `json:"body"`
}

type CreateBatchParams struct {
	Requests         []BatchRequestItem     `json:"requests"`
	CompletionWindow *string                `json:"completion_window,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

type BatchObject struct {
	ID               string                   `json:"id"`
	Object           *string                  `json:"object,omitempty"`
	Endpoint         *string                  `json:"endpoint,omitempty"`
	InputFileID      *string                  `json:"input_file_id,omitempty"`
	OutputFileID     *string                  `json:"output_file_id,omitempty"`
	ErrorFileID      *string                  `json:"error_file_id,omitempty"`
	Status           string                   `json:"status"`
	Model            *string                  `json:"model,omitempty"`
	Provider         *string                  `json:"provider,omitempty"`
	CreatedAt        *int64                   `json:"created_at,omitempty"`
	CompletedAt      *int64                   `json:"completed_at,omitempty"`
	ExpiresAt        *int64                   `json:"expires_at,omitempty"`
	UsageSynced      *bool                    `json:"usage_synced,omitempty"`
	CompletionWindow *string                  `json:"completion_window,omitempty"`
	RequestCounts    map[string]interface{}   `json:"request_counts,omitempty"`
	Metadata         map[string]interface{}   `json:"metadata,omitempty"`
	Results          []map[string]interface{} `json:"results,omitempty"`
	ErrorsDetail     []map[string]interface{} `json:"errors_detail,omitempty"`
}

type BatchListResponse struct {
	Object  string        `json:"object"`
	Data    []BatchObject `json:"data"`
	HasMore bool          `json:"has_more"`
	FirstID *string       `json:"first_id,omitempty"`
	LastID  *string       `json:"last_id,omitempty"`
}

// ---------------------------------------------------------------------------
// Images
// ---------------------------------------------------------------------------

type ImageGenerationParams struct {
	Prompt         string  `json:"prompt"`
	Model          *string `json:"model,omitempty"`
	N              *int    `json:"n,omitempty"`
	Size           *string `json:"size,omitempty"`
	Quality        *string `json:"quality,omitempty"`
	ResponseFormat *string `json:"response_format,omitempty"` // "url" | "b64_json"
	OutputFormat   *string `json:"output_format,omitempty"`   // "png" | "jpeg" | "webp"
	Stream         *bool   `json:"stream,omitempty"`
	// Additional spec fields
	AspectRatio                        *string                `json:"aspect_ratio,omitempty"`
	Resolution                         *string                `json:"resolution,omitempty"`
	OutputCompression                  *int                   `json:"output_compression,omitempty"` // 0..100
	Background                         *string                `json:"background,omitempty"`         // "transparent"|"opaque"|"auto"
	Moderation                         *string                `json:"moderation,omitempty"`         // "low"|"auto"
	PartialImages                      *int                   `json:"partial_images,omitempty"`     // 0..3
	Image                              interface{}            `json:"image,omitempty"`              // string or []string
	Seed                               *int                   `json:"seed,omitempty"`               // -1..2147483647
	SequentialImageGeneration          *string                `json:"sequential_image_generation,omitempty"`         // "auto"|"disabled"
	SequentialImageGenerationOptions   map[string]interface{} `json:"sequential_image_generation_options,omitempty"`
	GuidanceScale                      *float64               `json:"guidance_scale,omitempty"` // 1..10
	Watermark                          *bool                  `json:"watermark,omitempty"`
	OptimizePromptOptions              map[string]interface{} `json:"optimize_prompt_options,omitempty"`
}

type ImageItem struct {
	URL           *string `json:"url,omitempty"`
	B64JSON       *string `json:"b64_json,omitempty"`
	RevisedPrompt *string `json:"revised_prompt,omitempty"`
}

type ImageUsage struct {
	PromptTokens        int                    `json:"prompt_tokens"`
	CompletionTokens    int                    `json:"completion_tokens"`
	TotalTokens         int                    `json:"total_tokens"`
	InputTokensDetails  map[string]interface{} `json:"input_tokens_details,omitempty"`
	OutputTokensDetails map[string]interface{} `json:"output_tokens_details,omitempty"`
}

type ImageGenerationResponse struct {
	Created      int64       `json:"created"`
	Data         []ImageItem `json:"data"`
	Background   *string     `json:"background,omitempty"`
	OutputFormat *string     `json:"output_format,omitempty"`
	Quality      *string     `json:"quality,omitempty"`
	Size         *string     `json:"size,omitempty"`
	Usage        *ImageUsage `json:"usage,omitempty"`
}

// ---------------------------------------------------------------------------
// RAG (Retrieval-Augmented Generation)
// ---------------------------------------------------------------------------

// InitUploadRequest initialises a RAG file upload and returns a signed URL.
type InitUploadRequest struct {
	FileName string                 `json:"file_name"`
	MimeType string                 `json:"mime_type"`
	Embed    *bool                  `json:"embed,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// UploadFileParams is used by RagResource.UploadFile — it combines the upload
// initialisation fields with the raw file Content to upload in one call.
type UploadFileParams struct {
	FileName string
	MimeType string
	Content  []byte
	Embed    *bool
	Metadata map[string]interface{}
}

// InitUploadResponse is returned by POST /v1/files (RAG).
type InitUploadResponse struct {
	FileID    string `json:"file_id"`
	SignedURL string `json:"signed_url"`
	ExpiresAt string `json:"expires_at"`
}

// RagFileStatus represents the processing state of a RAG file.
type RagFileStatus struct {
	FileID             string   `json:"file_id"`
	UploadStatus       string   `json:"upload_status"`
	FileName           string   `json:"file_name"`
	FileType           string   `json:"file_type"`
	MimeType           string   `json:"mime_type"`
	SizeBytes          *int64   `json:"size_bytes,omitempty"`
	AssetURL           *string  `json:"asset_url,omitempty"`
	SignedURL          *string  `json:"signed_url,omitempty"`
	SignedURLExpiresAt *string  `json:"signed_url_expires_at,omitempty"`
	EmbeddingStatus    string   `json:"embedding_status"`
	CreatedAt          string   `json:"created_at"`
	UpdatedAt          string   `json:"updated_at"`
	TotalTokens        *int64   `json:"total_tokens,omitempty"`
	TotalCostUSD       *float64 `json:"total_cost_usd,omitempty"`
	LastErrorCode      *string  `json:"last_error_code,omitempty"`
}

// RagFileListResponse is returned by GET /v1/files (RAG).
type RagFileListResponse struct {
	Files  []RagFileStatus `json:"files"`
	Total  int             `json:"total"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
}

// ListRagFilesParams are the query parameters for GET /v1/files (RAG).
type ListRagFilesParams struct {
	Limit  *int `json:"limit,omitempty"`
	Offset *int `json:"offset,omitempty"`
}

// BulkEmbedRequest triggers embedding jobs for one or more files.
type BulkEmbedRequest struct {
	FileIDs  []string               `json:"file_ids"`
	Wait     *bool                  `json:"wait,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// BulkEmbedResult is the per-file result from POST /v1/files/embed.
type BulkEmbedResult struct {
	FileID          string  `json:"file_id"`
	EmbeddingStatus string  `json:"embedding_status"`
	ChunkCount      *int    `json:"chunk_count,omitempty"`
	Error           *string `json:"error,omitempty"`
}

// BulkEmbedResponse is returned by POST /v1/files/embed.
type BulkEmbedResponse struct {
	Results []BulkEmbedResult `json:"results"`
}

// SearchRequest is the body for POST /v1/files/search.
type SearchRequest struct {
	Query    string                 `json:"query"`
	TopK     *int                   `json:"top_k,omitempty"`
	FileIDs  []string               `json:"file_ids,omitempty"`
	Filter   map[string]interface{} `json:"filter,omitempty"`
	DateFrom *int64                 `json:"date_from,omitempty"`
	DateTo   *int64                 `json:"date_to,omitempty"`
}

// SearchResult is a single vector-search hit.
type SearchResult struct {
	Score      float64                `json:"score"`
	Text       string                 `json:"text"`
	ParentText string                 `json:"parent_text"`
	FileID     *string                `json:"file_id,omitempty"`
	FileName   *string                `json:"file_name,omitempty"`
	FileType   *string                `json:"file_type,omitempty"`
	MimeType   *string                `json:"mime_type,omitempty"`
	ChunkIndex *int                   `json:"chunk_index,omitempty"`
	CreatedAt  *int64                 `json:"created_at,omitempty"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// SearchResponse is returned by POST /v1/files/search.
type SearchResponse struct {
	Results []SearchResult `json:"results"`
}

// ---------------------------------------------------------------------------
// Audio
// ---------------------------------------------------------------------------

type VoiceSettings struct {
	Stability       *float64 `json:"stability,omitempty"`
	SimilarityBoost *float64 `json:"similarity_boost,omitempty"`
	Style           *float64 `json:"style,omitempty"`
	UseSpeakerBoost *bool    `json:"use_speaker_boost,omitempty"`
	Speed           *float64 `json:"speed,omitempty"`
}

type PronunciationDictionaryLocator struct {
	PronunciationDictionaryID string `json:"pronunciation_dictionary_id"`
	VersionID                 string `json:"version_id"`
}

type SpeechParams struct {
	Input                           string                           `json:"input"`
	Model                           *string                          `json:"model,omitempty"`
	Voice                           *string                          `json:"voice,omitempty"`
	Stream                          *bool                            `json:"stream,omitempty"`
	ResponseFormat                  *string                          `json:"response_format,omitempty"`
	LanguageCode                    *string                          `json:"language_code,omitempty"`
	VoiceSettings                   *VoiceSettings                   `json:"voice_settings,omitempty"`
	PronunciationDictionaryLocators []PronunciationDictionaryLocator `json:"pronunciation_dictionary_locators,omitempty"`
	Seed                            *int                             `json:"seed,omitempty"`
	PreviousText                    *string                          `json:"previous_text,omitempty"`
	NextText                        *string                          `json:"next_text,omitempty"`
	PreviousRequestIDs              []string                         `json:"previous_request_ids,omitempty"`
	NextRequestIDs                  []string                         `json:"next_request_ids,omitempty"`
	ApplyTextNormalization          *string                          `json:"apply_text_normalization,omitempty"`
	ApplyLanguageTextNormalization  *bool                            `json:"apply_language_text_normalization,omitempty"`
	UsePvcAsIvc                     *bool                            `json:"use_pvc_as_ivc,omitempty"`
	EnableLogging                   *bool                            `json:"enable_logging,omitempty"`
	OptimizeStreamingLatency        *int                             `json:"optimize_streaming_latency,omitempty"`
	Speaker                         *string                          `json:"speaker,omitempty"`
	TargetLanguageCode              *string                          `json:"target_language_code,omitempty"`
	Pitch                           *float64                         `json:"pitch,omitempty"`
	Pace                            *float64                         `json:"pace,omitempty"`
	Loudness                        *float64                         `json:"loudness,omitempty"`
	SpeechSampleRate                *int                             `json:"speech_sample_rate,omitempty"`
	EnablePreprocessing             *bool                            `json:"enable_preprocessing,omitempty"`
}

type TranscriptionParams struct {
	Model                 string   `json:"model"`
	LanguageCode          *string  `json:"language_code,omitempty"`
	TagAudioEvents        *bool    `json:"tag_audio_events,omitempty"`
	NumSpeakers           *int     `json:"num_speakers,omitempty"`
	TimestampsGranularity *string  `json:"timestamps_granularity,omitempty"`
	Diarize               *bool    `json:"diarize,omitempty"`
	DiarizationThreshold  *float64 `json:"diarization_threshold,omitempty"`
	AdditionalFormats     *string  `json:"additional_formats,omitempty"`
	FileFormat            *string  `json:"file_format,omitempty"`
	CloudStorageURL       *string  `json:"cloud_storage_url,omitempty"`
	SourceURL             *string  `json:"source_url,omitempty"`
	Webhook               *bool    `json:"webhook,omitempty"`
	WebhookID             *string  `json:"webhook_id,omitempty"`
	Temperature           *float64 `json:"temperature,omitempty"`
	Seed                  *int     `json:"seed,omitempty"`
	UseMultiChannel       *bool    `json:"use_multi_channel,omitempty"`
	WebhookMetadata       *string  `json:"webhook_metadata,omitempty"`
	EntityDetection       *string  `json:"entity_detection,omitempty"`
	NoVerbatim            *bool    `json:"no_verbatim,omitempty"`
	DetectSpeakerRoles    *bool    `json:"detect_speaker_roles,omitempty"`
	EntityRedaction       *string  `json:"entity_redaction,omitempty"`
	EntityRedactionMode   *string  `json:"entity_redaction_mode,omitempty"`
	Keyterms              []string `json:"keyterms,omitempty"`
	WithTimestamps        *bool    `json:"with_timestamps,omitempty"`
	DebugMode             *bool    `json:"debug_mode,omitempty"`
}

type TranscriptionTranslateParams struct {
	Model  *string `json:"model,omitempty"`
	Prompt *string `json:"prompt,omitempty"`
}

// AudioTranslationParams is the request body for POST /v1/audio/translations
// (standalone translation endpoint, distinct from /v1/audio/transcriptions/translate).
// Model and File (via fileData/filename in the method call) are required.
type AudioTranslationParams struct {
	// Model is required — the translation model to use.
	Model          string   `json:"model"`
	Prompt         *string  `json:"prompt,omitempty"`
	ResponseFormat *string  `json:"response_format,omitempty"`
	Temperature    *float64 `json:"temperature,omitempty"`
}

type TranscriptionResponse struct {
	Text string `json:"text"`
}

type ListVoicesParams struct {
	NextPageToken     *string  `json:"next_page_token,omitempty"`
	PageSize          *int     `json:"page_size,omitempty"`
	Search            *string  `json:"search,omitempty"`
	Sort              *string  `json:"sort,omitempty"`
	SortDirection     *string  `json:"sort_direction,omitempty"`
	VoiceType         *string  `json:"voice_type,omitempty"`
	Category          *string  `json:"category,omitempty"`
	IncludeTotalCount *bool    `json:"include_total_count,omitempty"`
	VoiceIDs          []string `json:"voice_ids,omitempty"`
}

type Voice struct {
	VoiceID     string `json:"voice_id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description"`
	PreviewURL  string `json:"preview_url"`
	// Labels values are provider-defined and not always strings — decode as
	// arbitrary JSON so a numeric/bool/null label doesn't fail the whole request.
	Labels map[string]interface{} `json:"labels,omitempty"`
}

type VoicesResponse struct {
	Voices []Voice `json:"voices"`
	// Pointers so an omitted has_more / total_count is distinguishable from a
	// real zero value (e.g. paginate while HasMore != nil && *HasMore).
	HasMore       *bool   `json:"has_more,omitempty"`
	TotalCount    *int    `json:"total_count,omitempty"`
	NextPageToken *string `json:"next_page_token"`
}

// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// Video
// ---------------------------------------------------------------------------

// VideoContentItem is a single item in the content array.
type VideoContentItem struct {
	Type      string                 `json:"type"`
	Text      *string                `json:"text,omitempty"`
	ImageURL  map[string]interface{} `json:"image_url,omitempty"`
	VideoURL  map[string]interface{} `json:"video_url,omitempty"`
	AudioURL  map[string]interface{} `json:"audio_url,omitempty"`
	DraftTask map[string]interface{} `json:"draft_task,omitempty"`
	Role      *string                `json:"role,omitempty"`
}

// VideoGenerationParams is the request body for POST /v1/video/generations.
type VideoGenerationParams struct {
	Model                 string             `json:"model"`
	Content               []VideoContentItem `json:"content"`
	CallbackURL           *string            `json:"callback_url,omitempty"`
	ReturnLastFrame       *bool              `json:"return_last_frame,omitempty"`
	ServiceTier           *string            `json:"service_tier,omitempty"`
	ExecutionExpiresAfter *int               `json:"execution_expires_after,omitempty"`
	GenerateAudio         *bool              `json:"generate_audio,omitempty"`
	Draft                 *bool              `json:"draft,omitempty"`
	Resolution            *string            `json:"resolution,omitempty"`
	Ratio                 *string            `json:"ratio,omitempty"`
	Duration              *int               `json:"duration,omitempty"`
	Frames                *int               `json:"frames,omitempty"`
	Seed                  *int               `json:"seed,omitempty"`
	CameraFixed           *bool              `json:"camera_fixed,omitempty"`
	Watermark             *bool              `json:"watermark,omitempty"`
	SafetyIdentifier      *string            `json:"safety_identifier,omitempty"`
	Priority              *int               `json:"priority,omitempty"`
}

// CreateVideoGenerationResponse is the response from POST /v1/video/generations.
type CreateVideoGenerationResponse struct {
	ID string `json:"id"`
}

// VideoTaskError holds error details for a failed video task.
type VideoTaskError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// VideoTaskContent holds the output URLs for a completed video task.
type VideoTaskContent struct {
	VideoURL     *string `json:"video_url,omitempty"`
	LastFrameURL *string `json:"last_frame_url,omitempty"`
}

// VideoTaskUsage holds token usage for a video task.
type VideoTaskUsage struct {
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// VideoTaskResponse is the shape of a single video generation task.
type VideoTaskResponse struct {
	ID                    string            `json:"id"`
	Status                string            `json:"status"`
	Model                 *string           `json:"model,omitempty"`
	Error                 *VideoTaskError   `json:"error,omitempty"`
	CreatedAt             *int64            `json:"created_at,omitempty"`
	UpdatedAt             *int64            `json:"updated_at,omitempty"`
	Content               *VideoTaskContent `json:"content,omitempty"`
	Seed                  *int              `json:"seed,omitempty"`
	Resolution            *string           `json:"resolution,omitempty"`
	Ratio                 *string           `json:"ratio,omitempty"`
	Duration              *int              `json:"duration,omitempty"`
	Frames                *int              `json:"frames,omitempty"`
	FramesPerSecond       *int              `json:"framespersecond,omitempty"`
	GenerateAudio         *bool             `json:"generate_audio,omitempty"`
	SafetyIdentifier      *string           `json:"safety_identifier,omitempty"`
	Priority              *int              `json:"priority,omitempty"`
	Draft                 *bool             `json:"draft,omitempty"`
	DraftTaskID           *string           `json:"draft_task_id,omitempty"`
	ServiceTier           *string           `json:"service_tier,omitempty"`
	ExecutionExpiresAfter *int              `json:"execution_expires_after,omitempty"`
	Usage                 *VideoTaskUsage   `json:"usage,omitempty"`
}

// ListVideoGenerationsParams holds query parameters for GET /v1/video/generations.
type ListVideoGenerationsParams struct {
	Status        *string `json:"status,omitempty"`
	Model         *string `json:"model,omitempty"`
	CreatedAfter  *string `json:"created_after,omitempty"`
	CreatedBefore *string `json:"created_before,omitempty"`
	Limit         *int    `json:"limit,omitempty"`
	Offset        *int    `json:"offset,omitempty"`
}

// VideoTaskListResponse is the response from GET /v1/video/generations.
type VideoTaskListResponse struct {
	Object  string              `json:"object"`
	Data    []VideoTaskResponse `json:"data"`
	HasMore bool                `json:"has_more"`
	Total   int                 `json:"total"`
	Limit   int                 `json:"limit"`
	Offset  int                 `json:"offset"`
}

// ---------------------------------------------------------------------------

type ImageGenerationChunk struct {
	ID      *string     `json:"id,omitempty"`
	Object  *string     `json:"object,omitempty"`
	Created int64       `json:"created"`
	Model   *string     `json:"model,omitempty"`
	Data    []ImageItem `json:"data"`
	Status  *string     `json:"status,omitempty"`
}

// ---------------------------------------------------------------------------
// Moderations — POST /v1/moderations
// ---------------------------------------------------------------------------

type ModerationImageURL struct {
	URL string `json:"url"`
}

type ModerationInputItem struct {
	Type     string              `json:"type"` // "text" | "image_url"
	Text     *string             `json:"text,omitempty"`
	ImageURL *ModerationImageURL `json:"image_url,omitempty"`
}

// ModerationParams is the request body for POST /v1/moderations.
// Input is a string, []string, or []ModerationInputItem. Leave Model nil to use
// the server default ("omni-moderation-latest").
type ModerationParams struct {
	Input interface{} `json:"input"`
	Model *string     `json:"model,omitempty"`
}

type ModerationResult struct {
	Flagged        bool               `json:"flagged"`
	Categories     map[string]bool    `json:"categories"`
	CategoryScores map[string]float64 `json:"category_scores"`
}

type ModerationResponse struct {
	ID      string             `json:"id"`
	Model   string             `json:"model"`
	Results []ModerationResult `json:"results"`
}

// ---------------------------------------------------------------------------
// Web search — POST /v1/web/search
// ---------------------------------------------------------------------------

type WebSearchParams struct {
	Query          string   `json:"query"`
	Model          *string  `json:"model,omitempty"`
	Provider       *string  `json:"provider,omitempty"`     // "native" | "tavily"
	MaxResults     *int     `json:"max_results,omitempty"`  // 1–20, server default 5
	SearchDepth    *string  `json:"search_depth,omitempty"` // "basic" | "advanced"
	IncludeDomains []string `json:"include_domains,omitempty"`
	ExcludeDomains []string `json:"exclude_domains,omitempty"`
	IncludeAnswer  *bool    `json:"include_answer,omitempty"`
}

type WebSearchResultItem struct {
	Title         string   `json:"title"`
	URL           string   `json:"url"`
	Content       string   `json:"content"`
	Score         *float64 `json:"score,omitempty"`
	PublishedDate *string  `json:"published_date,omitempty"`
}

type WebSearchResponse struct {
	Query   string                `json:"query"`
	Answer  *string               `json:"answer,omitempty"`
	Results []WebSearchResultItem `json:"results"`
	// Provider is "native" or "tavily" today; typed as string so an added
	// engine never breaks response decoding for existing SDK versions.
	Provider  string `json:"provider"`
	RequestID string `json:"request_id"`
}

// ---------------------------------------------------------------------------
// Router select — POST /v1/router/select
// ---------------------------------------------------------------------------

type RouterSelectParams struct {
	Messages      []ChatMessage `json:"messages"`
	APIType       *string       `json:"api_type,omitempty"` // "completions" (default) | "responses" | "embeddings"
	ExcludeModels []string      `json:"exclude_models,omitempty"`
}

type AutoRouterMeta struct {
	FallbackUsed   bool    `json:"fallback_used"`
	FallbackReason *string `json:"fallback_reason,omitempty"`
}

type RouterSelectResponse struct {
	Model           string         `json:"model"`
	AutoRouter      AutoRouterMeta `json:"auto_router"`
	ReasoningEffort *string        `json:"reasoning_effort,omitempty"`
}

// ---------------------------------------------------------------------------
// Models — GET /v1/models/search (paginated catalog search)
// ---------------------------------------------------------------------------

// ModelSearchParams holds query parameters for Models.Search. All fields are
// optional; leave a field nil/empty to omit it.
type ModelSearchParams struct {
	Q              *string
	Free           *bool
	Discounted     *bool
	InputModality  []string
	OutputModality []string
	Brand          []string
	Sort           *string // "brand" | "name" | "id" | "context_length"
	Order          *string // "asc" | "desc"
	Limit          *int
	Offset         *int
}

type ModelsPage struct {
	Items  []ModelInfo `json:"items"`
	Total  int         `json:"total"`
	Limit  int         `json:"limit"`
	Offset int         `json:"offset"`
	Brands []string    `json:"brands"`
}

// ---------------------------------------------------------------------------
// Responses — GET /v1/responses (list background jobs) + GET /v1/responses/{id}
// ---------------------------------------------------------------------------

type ResponsesListItem struct {
	ID          string  `json:"id"`
	Object      *string `json:"object,omitempty"`
	Model       *string `json:"model,omitempty"`
	Provider    *string `json:"provider,omitempty"`
	Status      *string `json:"status,omitempty"`
	CreatedAt   *int64  `json:"created_at,omitempty"`
	CompletedAt *int64  `json:"completed_at,omitempty"`
	UsageSynced *bool   `json:"usage_synced,omitempty"`
}

type ResponsesListResponse struct {
	Object  *string             `json:"object,omitempty"`
	Data    []ResponsesListItem `json:"data"`
	HasMore bool                `json:"has_more"`
	FirstID *string             `json:"first_id,omitempty"`
	LastID  *string             `json:"last_id,omitempty"`
}

// ---------------------------------------------------------------------------
// Images — POST /v1/images/edits (edit / inpaint / outpaint / upscale / …)
// ---------------------------------------------------------------------------

// ImageRef is an image reference for the edits endpoint: URL must be a
// data URL (data:image/<fmt>;base64,<b64>) or bare base64 — remote http(s)
// URLs are rejected. You may also pass the base64/data-URL string directly.
type ImageRef struct {
	URL string `json:"url"`
}

// ImageEditParams is the JSON request body for POST /v1/images/edits.
// Image (and Mask / ReferenceImages) accept a base64/data-URL string or an
// ImageRef. Prompt is required for the "edit", "outpaint" and "mix" operations.
type ImageEditParams struct {
	Model           string      `json:"model"`
	Image           interface{} `json:"image"` // string (data-URL/base64) or ImageRef
	Prompt          *string     `json:"prompt,omitempty"`
	Operation       *string     `json:"operation,omitempty"` // edit|inpaint|outpaint|mix|reframe|upscale|remove_background
	Mask            interface{} `json:"mask,omitempty"`
	ReferenceImages interface{} `json:"reference_images,omitempty"` // []string or []ImageRef
	N               *int        `json:"n,omitempty"`
	Size            *string     `json:"size,omitempty"`
	ResponseFormat  *string     `json:"response_format,omitempty"`
	Background      *string     `json:"background,omitempty"`
	UpscaleFactor   *string     `json:"upscale_factor,omitempty"`
	QualityTier     *string     `json:"quality_tier,omitempty"`
	AspectRatio     *string     `json:"aspect_ratio,omitempty"`
	Resolution      *string     `json:"resolution,omitempty"`
	ExpandFactor    interface{} `json:"expand_factor,omitempty"` // string or float64
	MaskFeather     *int        `json:"mask_feather,omitempty"`
}

