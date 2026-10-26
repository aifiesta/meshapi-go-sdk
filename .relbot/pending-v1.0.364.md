# Release v1.0.364

> ⚠️ Breaking changes in this release.

## Schemas changed

- AnthropicMessagesRequest.container
- AnthropicMessagesRequest.mcp_servers
- AnthropicMessagesRequest.provider
- AnthropicMessagesRequest.service_tier
- Body_create_transcription.provider
- Body_create_transcription_translate.provider
- Body_create_translation.provider
- ChatCompletionRequest.max_completion_tokens
- ChatCompletionRequest.provider
- CompareRequest.provider
- CostBreakdownOut.rounding_usd
- CostLineOut.charged
- CostLineOut.modifiers
- CostLineOut.quantity
- CostLineOut.rate_per_1m_usd
- CostLineOut.rate_per_1m_usd: string -> anyOf[null, string]
- CostLineOut.rate_usd
- CostLineOut.tokens
- CostLineOut.tokens: integer -> anyOf[integer, null]
- CostLineOut.unit
- CreateBatchRequest.provider
- ImageGenerationRequest.provider
- ResponsesRequest.provider
- SpeechRequest.provider
- VideoGenerationRequest.provider

_STUB: this is a generated placeholder; full reference, changelog and SDK content is produced by a later automated step._
