// Package meshapi provides a typed Go client for the MeshAPI AI model gateway.
package meshapi

// ---------------------------------------------------------------------------
// Chat Completions
// ---------------------------------------------------------------------------

// ChatMessage represents a single message in the conversation.
type ChatMessage struct {
	Role       string        `json:"role"`
	Content    interface{}   `json:"content,omitempty"` // string or []ContentPart
	Name       *string       `json:"name,omitempty"`
	ToolCallID *string       `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCall    `json:"tool_calls,omitempty"`
}

// ContentPart is one element of a multimodal message content array.
type ContentPart struct {
	Type     string     `json:"type"`
	Text     *string    `json:"text,omitempty"`
	ImageURL *ImageURL  `json:"image_url,omitempty"`
	InputAudio *InputAudio `json:"input_audio,omitempty"`
}

// ImageURL holds the URL and rendering detail for an image content part.
type ImageURL struct {
	URL    string  `json:"url"`
	Detail *string `json:"detail,omitempty"`
}

type InputAudio struct {
	Data   string `json:"data"`
	Format string `json:"format"`
}

// ToolCall represents a tool invocation in an assistant message.
type ToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function ToolCallFunction `json:"function"`
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
	Type     string              `json:"type"`
	Function ToolChoiceFunction  `json:"function"`
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
	Messages         []ChatMessage `json:"messages"`
	Model            *string       `json:"model,omitempty"`
	Stream           *bool         `json:"stream,omitempty"`
	Template         *string       `json:"template,omitempty"`
	Variables        map[string]string `json:"variables,omitempty"`
	SessionID        *string       `json:"session_id,omitempty"`
	Temperature      *float64      `json:"temperature,omitempty"`
	MaxTokens        *int          `json:"max_tokens,omitempty"`
	TopP             *float64      `json:"top_p,omitempty"`
	FrequencyPenalty *float64      `json:"frequency_penalty,omitempty"`
	PresencePenalty  *float64      `json:"presence_penalty,omitempty"`
	Stop             interface{}   `json:"stop,omitempty"` // string or []string
	Seed             *int          `json:"seed,omitempty"`
	Tools            []Tool        `json:"tools,omitempty"`
	ToolChoice       interface{}   `json:"tool_choice,omitempty"`
	ResponseFormat   map[string]interface{} `json:"response_format,omitempty"`
	Transforms       []string      `json:"transforms,omitempty"`
	Models           []string      `json:"models,omitempty"`
	User             *string       `json:"user,omitempty"`
	Modality         *string       `json:"modality,omitempty"`
	Image            *ImageOptions `json:"image,omitempty"`
	AsyncMode        *bool         `json:"async_mode,omitempty"`
	Modalities       []string      `json:"modalities,omitempty"`
	Audio            *AudioOutputOptions `json:"audio,omitempty"`
	// Timeout overrides the server's upstream-provider timeout (default 300 s).
	// Set this for requests that may take longer than 5 minutes. This is
	// independent of the SDK-level TimeoutMs option on Config, which controls
	// the HTTP client timeout.
	Timeout          *float64      `json:"timeout,omitempty"`
}

// UsageInfo holds token counts for a completion.
type UsageInfo struct {
	PromptTokens     int                    `json:"prompt_tokens"`
	CompletionTokens int                    `json:"completion_tokens"`
	TotalTokens      int                    `json:"total_tokens"`
	PromptTokensDetails map[string]interface{} `json:"prompt_tokens_details,omitempty"`
	CompletionTokensDetails map[string]interface{} `json:"completion_tokens_details,omitempty"`
	ClassifierPromptTokens *int `json:"classifier_prompt_tokens,omitempty"`
	ClassifierCompletionTokens *int `json:"classifier_completion_tokens,omitempty"`
	ClassifierTokens *int `json:"classifier_tokens,omitempty"`
}

// ChatCompletionMessage is a completed message in a non-streaming response.
type ChatCompletionMessage struct {
	Role      string     `json:"role"`
	Content   *string    `json:"content,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
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
	Role      *string    `json:"role,omitempty"`
	Content   *string    `json:"content,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
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

// ModelPricing holds per-token pricing for a model.
type ModelPricing struct {
	PromptUSDPer1K              *string `json:"prompt_usd_per_1k,omitempty"`
	CompletionUSDPer1K          *string `json:"completion_usd_per_1k,omitempty"`
	ImageUSDPerImage            *string `json:"image_usd_per_image,omitempty"`
	DiscountPct                 *string `json:"discount_pct,omitempty"`
	PromptUSDPer1KDiscounted    *string `json:"prompt_usd_per_1k_discounted,omitempty"`
	CompletionUSDPer1KDiscounted *string `json:"completion_usd_per_1k_discounted,omitempty"`
}

// ModelInfo describes an available model.
type ModelInfo struct {
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	ContextLength *int          `json:"context_length,omitempty"`
	IsFree        bool          `json:"is_free"`
	Pricing       *ModelPricing `json:"pricing,omitempty"`
	Description   *string       `json:"description,omitempty"`
	SupportsThinking *bool `json:"supports_thinking,omitempty"`
	SupportsCompletionsAPI *bool `json:"supports_completions_api,omitempty"`
	SupportsResponsesAPI *bool `json:"supports_responses_api,omitempty"`
	ModelType *string `json:"model_type,omitempty"`
	InputModalities []string `json:"input_modalities,omitempty"`
	OutputModalities []string `json:"output_modalities,omitempty"`
}

// ListModelsParams holds optional query parameters for listing models.
type ListModelsParams struct {
	Free *bool // nil = no filter
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

type EmbeddingsParams struct {
	Model          *string     `json:"model,omitempty"`
	Input          interface{} `json:"input"`
	Dimensions     *int        `json:"dimensions,omitempty"`
	EncodingFormat *string     `json:"encoding_format,omitempty"`
	InputType      *string     `json:"input_type,omitempty"`
	Provider       interface{} `json:"provider,omitempty"`
	User           *string     `json:"user,omitempty"`
}

type EmbeddingItem struct {
	Object    string      `json:"object"`
	Index     int         `json:"index"`
	Embedding interface{} `json:"embedding"`
}

type EmbeddingsUsage struct {
	PromptTokens int `json:"prompt_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

type EmbeddingsResponse struct {
	Object string            `json:"object"`
	Data   []EmbeddingItem   `json:"data"`
	Model  string            `json:"model"`
	Usage  *EmbeddingsUsage  `json:"usage,omitempty"`
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
	Model           *string       `json:"model,omitempty"`
	Input           interface{}   `json:"input"`
	Template        *string       `json:"template,omitempty"`
	Variables       map[string]string `json:"variables,omitempty"`
	SessionID       *string       `json:"session_id,omitempty"`
	Stream          *bool         `json:"stream,omitempty"`
	MaxOutputTokens *int          `json:"max_output_tokens,omitempty"`
	Temperature     *float64      `json:"temperature,omitempty"`
	TopP            *float64      `json:"top_p,omitempty"`
	Seed            *int          `json:"seed,omitempty"`
	Reasoning       map[string]interface{} `json:"reasoning,omitempty"`
	Tools           []interface{} `json:"tools,omitempty"`
	ToolChoice      interface{}   `json:"tool_choice,omitempty"`
	ResponseFormat  map[string]interface{} `json:"response_format,omitempty"`
	Plugins         []interface{} `json:"plugins,omitempty"`
	User            *string       `json:"user,omitempty"`
	// Timeout overrides the server's upstream-provider timeout (default 300 s).
	Timeout         *float64      `json:"timeout,omitempty"`
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
	ComparisonID          string               `json:"comparison_id"`
	Object                string               `json:"object"`
	Created               int64                `json:"created"`
	Models                []string             `json:"models"`
	Results               []ModelCompareResult `json:"results"`
	Comparison            *string              `json:"comparison,omitempty"`
	ComparisonModel       *string              `json:"comparison_model,omitempty"`
	ComparisonUsage       *TokenUsage          `json:"comparison_usage,omitempty"`
	ComparisonFallbackUsed bool                `json:"comparison_fallback_used"`
	TotalLatencyMs        int                  `json:"total_latency_ms"`
	Partial               bool                 `json:"partial"`
	SkipComparison        bool                 `json:"skip_comparison"`
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
	ID           string  `json:"id"`
	Object       *string `json:"object,omitempty"`
	Endpoint     *string `json:"endpoint,omitempty"`
	InputFileID  *string `json:"input_file_id,omitempty"`
	OutputFileID *string `json:"output_file_id,omitempty"`
	Status       *string `json:"status,omitempty"`
	Model        *string `json:"model,omitempty"`
	Provider     *string `json:"provider,omitempty"`
	CreatedAt    *int64  `json:"created_at,omitempty"`
	CompletedAt  *int64  `json:"completed_at,omitempty"`
	UsageSynced  *bool   `json:"usage_synced,omitempty"`
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
	Prompt       string  `json:"prompt"`
	Model        *string `json:"model,omitempty"`
	N            *int    `json:"n,omitempty"`
	Size         *string `json:"size,omitempty"`
	Quality        *string `json:"quality,omitempty"`
	ResponseFormat *string `json:"response_format,omitempty"`
	OutputFormat   *string `json:"output_format,omitempty"`
	Stream         *bool   `json:"stream,omitempty"`
}

type ImageItem struct {
	URL           *string `json:"url,omitempty"`
	B64JSON       *string `json:"b64_json,omitempty"`
	RevisedPrompt *string `json:"revised_prompt,omitempty"`
}

type ImageUsage struct {
	PromptTokens           int                    `json:"prompt_tokens"`
	CompletionTokens       int                    `json:"completion_tokens"`
	TotalTokens            int                    `json:"total_tokens"`
	InputTokensDetails     map[string]interface{} `json:"input_tokens_details,omitempty"`
	OutputTokensDetails    map[string]interface{} `json:"output_tokens_details,omitempty"`
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
	FileID              string                 `json:"file_id"`
	UploadStatus        string                 `json:"upload_status"`
	FileName            string                 `json:"file_name"`
	FileType            string                 `json:"file_type"`
	MimeType            string                 `json:"mime_type"`
	SizeBytes           *int64                 `json:"size_bytes,omitempty"`
	AssetURL            *string                `json:"asset_url,omitempty"`
	SignedURL           *string                `json:"signed_url,omitempty"`
	SignedURLExpiresAt  *string                `json:"signed_url_expires_at,omitempty"`
	EmbeddingStatus     string                 `json:"embedding_status"`
	CreatedAt           string                 `json:"created_at"`
	UpdatedAt           string                 `json:"updated_at"`
	TotalTokens         *int64                 `json:"total_tokens,omitempty"`
	TotalCostUSD        *float64               `json:"total_cost_usd,omitempty"`
	LastErrorCode       *string                `json:"last_error_code,omitempty"`
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
