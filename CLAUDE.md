# MeshAPI Go SDK

Official Go client for the MeshAPI AI model gateway.

- **Module**: `github.com/aifiesta/meshapi-go-sdk`
- **Go version**: 1.21+
- **External dependencies**: none

## Project layout

```
go/
├── client.go          # Client struct and New() constructor
├── types.go           # All request/response types
├── http.go            # HTTP transport, retry, auth
├── errors.go          # MeshAPIError type
├── sse.go             # Server-Sent Events parser
├── chat.go            # /v1/chat/completions
├── responses.go       # /v1/responses
├── embeddings.go      # /v1/embeddings
├── compare.go         # /v1/chat/compare
├── rag.go             # /v1/files RAG endpoints (upload, list, get status, embed, search)
├── batches.go         # /v1/batches
├── models.go          # /v1/models
├── templates.go       # /v1/templates
├── images.go          # /v1/images/generations, /v1/images/edits
├── *_test.go          # Unit / contract / integration tests
└── livetests/         # Live tests against a real backend
```

## Common tasks

### Build / vet

```bash
go build ./...
go vet ./...
```

### Unit, contract, and integration tests

```bash
# All tests in the root package (no network required for contract tests)
go test ./... -v

# Run a specific test
go test -run TestSSE -v
```

### Adding a new resource

1. Add request/response types to `types.go` under a clearly labelled section.
2. Create `<resource>.go` with a `<Resource>Resource` struct holding `http *httpClient`.
3. Add the resource field to `Client` in `client.go` and initialise it in `New()`.
4. Follow the pattern in `templates.go` — `http.get`, `http.post`, `http.patch`, `http.delete`.

---

## Live tests

Live tests hit a real MeshAPI backend. They live in `livetests/` which has its own `go.mod` that `replace`-points at the parent SDK.

### Prerequisites

- A running MeshAPI instance (default `http://localhost:8000`), **or** point at the dev API.
- A valid data-plane API key (`rsk_...`).

### Environment variables

Create `go/.env.livetest` (read automatically by the test harness) or export the variables in your shell before running tests.

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `MESHAPI_BASE_URL` | No | `http://localhost:8000` | Base URL of the MeshAPI gateway |
| `MESHAPI_TOKEN` | **Yes** | hardcoded dev key | Data-plane API key (`rsk_...`) |
| `MESHAPI_MODEL` | No | `openai/gpt-4o-mini` | Primary model used in chat/stream tests |
| `MESHAPI_SECOND_MODEL` | No | `anthropic/claude-haiku-4.5` | Second model for compare tests |
| `MESHAPI_EMBEDDINGS_MODEL` | No | `openai/text-embedding-3-small` | Model used in embeddings tests |
| `MESHAPI_IMAGE_GEN_MODEL` | No | _(skipped if unset)_ | Image generation model; test skipped if blank |
| `MESHAPI_IMAGE_URL` | No | _(skipped if unset)_ | Publicly accessible image URL for vision tests |
| `MESHAPI_REALTIME_MODEL` | No | `openai/gpt-realtime-mini` | Realtime-capable model used in WebSocket live tests |

Example `go/.env.livetest`:

```env
MESHAPI_BASE_URL=https://api-dev.meshapi.ai
MESHAPI_TOKEN=rsk_your_key_here
MESHAPI_MODEL=openai/gpt-4o-mini
MESHAPI_EMBEDDINGS_MODEL=openai/text-embedding-3-small
```

### Run all live tests

```bash
cd livetests
go test ./... -v -timeout 300s
```

### Run a single live test file

```bash
cd livetests
go test -run TestLive_RAG -v -timeout 300s
```

### Available live test files

| File | What it tests |
|------|---------------|
| `chat_test.go` | Chat completions (basic, tools, multi-turn) |
| `stream_test.go` | Streaming chat and responses |
| `models_test.go` | Model listing |
| `templates_test.go` | Template CRUD lifecycle |
| `inference_test.go` | Embeddings, responses |
| `errors_test.go` | 401/404 error handling |
| `feature_matrix_test.go` | Cross-model feature matrix |
| `rag_test.go` | RAG upload → embed → list → search |
| `realtime_test.go` | WebSocket connect/close, session.created, session.update, error envelopes, Events() channel API, context cancel |

### Available live test files (updated)

| File | What it tests |
|------|---------------|
| `chat_test.go` | Chat completions (basic, tools, multi-turn) |
| `stream_test.go` | Streaming chat and responses |
| `models_test.go` | Model listing |
| `templates_test.go` | Template CRUD lifecycle |
| `inference_test.go` | Embeddings, responses |
| `errors_test.go` | 401/404 error handling |
| `feature_matrix_test.go` | Cross-model feature matrix |
| `rag_test.go` | RAG upload → embed → list → search |
| `realtime_test.go` | WebSocket connect/close, session lifecycle |
| `audio_test.go` | TTS synthesize, voice listing |
| `video_test.go` | Video list, generate → retrieve |
| `compare_test.go` | Non-streaming compare, streaming compare |
| `moderations_test.go` | Moderation classify: text and multimodal input |

---

## Contribution checklist

Every SDK change — however small — must include all of the following before merging:

1. **Live tests** — add or update `livetests/<resource>_test.go` to cover the new/changed behaviour.
2. **Unit / contract tests** — if the change affects types or HTTP transport, add a test in `*_test.go` files in the root package.
3. **README** — update `README.md` with a usage example for any new or changed public surface.
4. **meshapi-docs** — open a follow-up PR (or note in the PR description) to update the [meshapi-docs](https://github.com/aifiesta/meshapi-docs) repository so the developer documentation stays in sync.

---

---

## Release

Go modules are released by pushing a `v*` git tag on the `github.com/aifiesta/meshapi-go-sdk` repo. There is no publish workflow — the Go module proxy picks up the tag automatically.

### Release checklist

1. **Commit all changes** (no version file to bump — Go uses git tags):
   ```bash
   git add .
   git commit -m "chore: release v0.1.7"
   ```

2. **Tag and push**:
   ```bash
   git tag v0.1.7
   git push origin main
   git push origin v0.1.7
   ```

3. **Verify** the new version is available on the module proxy (may take ~1 min):
   ```bash
   GOFLAGS=-mod=mod go get github.com/aifiesta/meshapi-go-sdk@v0.1.7
   ```

> Go modules must use a domain-qualified module path (e.g. `github.com/aifiesta/meshapi-go-sdk`). Short aliases like `meshapi-go-sdk` are not valid Go module paths.

### RAG live test notes

`TestLive_RAG_UploadAndSearch` does the following:
1. Calls `client.RAG.InitUpload` with `embed=false`.
2. PUTs the file bytes directly to the returned `SignedURL` via `net/http`.
3. Waits up to 30 s for `upload_status=ready`.
4. Calls `client.RAG.Embed` to trigger embedding.
5. Polls up to 90 s for `embedding_status=ready`.
6. Calls `client.RAG.List` and asserts the file appears.
7. Calls `client.RAG.Search` scoped to the file ID and asserts non-empty results.

If the backend is unreachable the test is automatically skipped (not failed).
