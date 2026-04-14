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
}

// ImageURL holds the URL and rendering detail for an image content part.
type ImageURL struct {
	URL    string  `json:"url"`
	Detail *string `json:"detail,omitempty"`
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
	Transforms       []string      `json:"transforms,omitempty"`
	Models           []string      `json:"models,omitempty"`
	User             *string       `json:"user,omitempty"`
}

// UsageInfo holds token counts for a completion.
type UsageInfo struct {
	PromptTokens     int                    `json:"prompt_tokens"`
	CompletionTokens int                    `json:"completion_tokens"`
	TotalTokens      int                    `json:"total_tokens"`
	PromptTokensDetails map[string]interface{} `json:"prompt_tokens_details,omitempty"`
}

// ChatCompletionMessage is a completed message in a non-streaming response.
type ChatCompletionMessage struct {
	Role      string     `json:"role"`
	Content   *string    `json:"content,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
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

// ResponsesChunkDelta is the partial content in a Responses API streaming chunk.
// Unlike ChatCompletionChunkDelta it carries a Reasoning field for models that
// emit chain-of-thought tokens (e.g. openai/o4-mini).
type ResponsesChunkDelta struct {
	Role      *string                    `json:"role,omitempty"`
	Content   *string                    `json:"content,omitempty"`
	ToolCalls []ToolCall                 `json:"tool_calls,omitempty"`
	Reasoning *ResponsesMessageReasoning `json:"reasoning,omitempty"`
}

// ResponsesChunkChoice is one choice in a Responses API streaming chunk.
type ResponsesChunkChoice struct {
	Index        int                  `json:"index"`
	Delta        *ResponsesChunkDelta `json:"delta,omitempty"`
	FinishReason *string              `json:"finish_reason,omitempty"`
}

// ResponsesChunk is a single SSE chunk in a streaming Responses API response.
type ResponsesChunk struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int64                  `json:"created"`
	Model   string                 `json:"model"`
	Choices []ResponsesChunkChoice `json:"choices"`
	Usage   *UsageInfo             `json:"usage,omitempty"`
	Cost    *string                `json:"cost,omitempty"`
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
	Owner       string                   `json:"owner"`
	Description *string                  `json:"description,omitempty"`
	System      *string                  `json:"system,omitempty"`
	Messages    []map[string]interface{} `json:"messages,omitempty"`
	Model       *string                  `json:"model,omitempty"`
	Params      map[string]interface{}   `json:"params,omitempty"`
	Variables   []string                 `json:"variables,omitempty"`
	CreatedAt   string                   `json:"created_at"`
	UpdatedAt   string                   `json:"updated_at"`
}

// ---------------------------------------------------------------------------
// Responses
// ---------------------------------------------------------------------------

// ReasoningConfig controls chain-of-thought depth for supported models.
type ReasoningConfig struct {
	Effort string `json:"effort"` // "minimal" | "low" | "medium" | "high"
}

// ResponsesParams is the request body for POST /v1/responses.
type ResponsesParams struct {
	Input           interface{}      `json:"input"`                     // string or []ChatMessage (required)
	Model           *string          `json:"model,omitempty"`
	Stream          *bool            `json:"stream,omitempty"`
	SessionID       *string          `json:"session_id,omitempty"`
	MaxOutputTokens *int             `json:"max_output_tokens,omitempty"`
	Temperature     *float64         `json:"temperature,omitempty"`
	TopP            *float64         `json:"top_p,omitempty"`
	Seed            *int             `json:"seed,omitempty"`
	Reasoning       *ReasoningConfig `json:"reasoning,omitempty"`
	Tools           []Tool           `json:"tools,omitempty"`
	ToolChoice      interface{}      `json:"tool_choice,omitempty"`
	ResponseFormat  interface{}      `json:"response_format,omitempty"`
	Plugins         []interface{}    `json:"plugins,omitempty"`
	User            *string          `json:"user,omitempty"`
}

// ResponsesMessageReasoning holds the reasoning trace returned by the model.
type ResponsesMessageReasoning struct {
	EncryptedContent *string `json:"encrypted_content,omitempty"`
	Summary          *string `json:"summary,omitempty"`
}

// ResponsesMessage is the assistant message in a Responses API result.
type ResponsesMessage struct {
	Role      string                     `json:"role"`
	Content   *string                    `json:"content,omitempty"`
	ToolCalls []ToolCall                 `json:"tool_calls,omitempty"`
	Reasoning *ResponsesMessageReasoning `json:"reasoning,omitempty"`
}

// ResponsesChoice is one result choice in a Responses API response.
type ResponsesChoice struct {
	Index        int               `json:"index"`
	Message      *ResponsesMessage `json:"message,omitempty"`
	FinishReason *string           `json:"finish_reason,omitempty"`
}

// ResponsesResponse is the full non-streaming response body for POST /v1/responses.
type ResponsesResponse struct {
	ID                string            `json:"id"`
	Object            string            `json:"object"`
	Created           int64             `json:"created"`
	Model             string            `json:"model"`
	Choices           []ResponsesChoice `json:"choices"`
	Usage             *UsageInfo        `json:"usage,omitempty"`
	SystemFingerprint *string           `json:"system_fingerprint,omitempty"`
}
