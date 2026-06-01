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
├── compare.go         # /v1/compare
├── files.go           # /v1/files (batch file objects)
├── rag.go             # /v1/files RAG endpoints (upload, embed, search)
├── batches.go         # /v1/batches
├── models.go          # /v1/models
├── templates.go       # /v1/templates
├── images.go          # /v1/images/generations
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
| `MESHAPI_SECOND_MODEL` | No | `anthropic/claude-haiku-4-5` | Second model for compare tests |
| `MESHAPI_EMBEDDINGS_MODEL` | No | `openai/text-embedding-3-small` | Model used in embeddings tests |
| `MESHAPI_IMAGE_GEN_MODEL` | No | _(skipped if unset)_ | Image generation model; test skipped if blank |
| `MESHAPI_IMAGE_URL` | No | _(skipped if unset)_ | Publicly accessible image URL for vision tests |

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
