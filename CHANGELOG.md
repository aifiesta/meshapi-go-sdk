# Changelog

## [Unreleased]

### API version

- **This release targets MeshAPI version `2026-08`.** Every request now sends
  `X-Mesh-Version: 2026-08`, exported as `meshapi.APIVersion` — distinct from
  `meshapi.Version`, which identifies the SDK build.
- Pin a different one with `Config.APIVersion`, or set it to a pointer to the empty
  string to send no header and take the gateway's baseline (the behaviour of
  `v0.1.12` and earlier). A `nil` `APIVersion` uses this SDK's version.
- Why it matters: an unpinned client is served whatever the gateway defaults to, so
  it never states which response shape it can parse. Pinning means a future version
  that changes a shape cannot change it underneath this release.

### Fixed

- `ModelPricing` now declares `InputUSDPerUnit` / `OutputUSDPerUnit`. `encoding/json`
  ignores unknown keys, so these were being **discarded**: for models that are not
  token-priced — per-second video, per-image, per-1k-chars — the per-1M fields are
  null by design, which left the SDK reporting a priced model as having no price.
- `ModelPricing.PromptUSDPer1K` / `CompletionUSDPer1K` are documented as retired: the
  gateway stopped returning them in `v1.0.135`, so they are always `nil`. They remain
  declared for backwards compatibility. Their `// Required` comment was wrong.

## [0.1.0] — Initial release

- `NewClient(Config)` with `Chat`, `Models`, `Templates` resources
- Chat completions: `Create` (non-streaming) and `Stream` (channels)
- Models: `List`, `Free`, `Paid`
- Templates: `Create`, `List`, `Get`, `Update`, `Delete`
- `MeshAPIError` with `Status`, `Code`, `RequestID`, `RetryAfterSeconds`
- Retry with exponential backoff (default 3 retries, codes 429/502/503/504)
- SSE parser with blank-line frame delimiter and [DONE] sentinel support
- Streaming fail-fast: no automatic reconnect (documented)
- `X-MeshAPI-SDK: go/0.1.0` header on every request
