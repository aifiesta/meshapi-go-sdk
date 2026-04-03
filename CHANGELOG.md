# Changelog

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
