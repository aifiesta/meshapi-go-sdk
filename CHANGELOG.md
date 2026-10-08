# Changelog

## [Unreleased]

### Added

- `WebSearchParams.IncludePageContent` returns each result's extracted page text in
  the new `WebSearchResultItem.PageContent`, so one call replaces
  search-then-fetch. `PageContentTruncated` says when that text was cut at the
  server's per-result ceiling, so a truncated page is never served as a whole one.
  Only the tinyfish engine returns page text and the native engine is tried first,
  so pin `Provider` to `"tinyfish"` when you need it.
- A nil `IncludePageContent` sends no new key, so existing calls are byte-identical
  on the wire.

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
