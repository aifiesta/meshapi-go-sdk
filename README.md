# meshapi-go-sdk

Go client for the MeshAPI AI model gateway.

## Requirements

- Go ≥ 1.21
- Zero external dependencies

## Installation

```bash
go get meshapi-go-sdk  # local module
# or after publishing:
go get github.com/yourorg/meshapi-go-sdk@v0.1.0
```

## Quick Start

```go
import "meshapi-go-sdk"

client := meshapi.New(meshapi.Config{
    BaseURL: "http://localhost:8000",
    Token:   "rsk_...",
})

// Non-streaming
model := "openai/gpt-4o-mini"
resp, err := client.Chat.Completions.Create(ctx, meshapi.ChatCompletionParams{
    Model:    &model,
    Messages: []meshapi.ChatMessage{{Role: "user", Content: "What is 2+2?"}},
})

// Streaming
chunkCh, errCh := client.Chat.Completions.Stream(ctx, params)
for chunk := range chunkCh {
    if len(chunk.Choices) > 0 && chunk.Choices[0].Delta != nil {
        fmt.Print(*chunk.Choices[0].Delta.Content)
    }
}
if err := <-errCh; err != nil {
    log.Fatal(err)
}

// Models
models, _ := client.Models.List(ctx, meshapi.ListModelsParams{})
free, _   := client.Models.Free(ctx)

// Templates
tmpl, _ := client.Templates.Create(ctx, meshapi.CreateTemplateParams{Name: "my-tpl"})
client.Templates.Delete(ctx, tmpl.ID)

// Responses (Reasoning)
respModel := "openai/o3-mini"
resp, _ := client.Responses.Create(ctx, meshapi.ResponsesParams{
    Model: &respModel,
    Input: "Solve for X: 2x + 5 = 15",
})

// Embeddings
embModel := "openai/text-embedding-3-small"
emb, _ := client.Embeddings.Create(ctx, meshapi.EmbeddingsParams{
    Model: &embModel,
    Input: []string{"The quick brown fox"},
})

// Compare (Multi-model)
compCh, errCh := client.Compare.Stream(ctx, meshapi.CompareParams{
    Models: []string{"openai/gpt-4o-mini", "anthropic/claude-3-haiku"},
    Messages: []meshapi.ChatMessage{{Role: "user", Content: "Hello"}},
})
for range compCh {
}
if err := <-errCh; err != nil {
    log.Fatal(err)
}

// Files & Batches
file, _ := client.Files.Upload(ctx, meshapi.UploadBatchFileParams{
    Purpose: "batch",
    Requests: []meshapi.BatchRequestItem{
        {
            CustomID: "req-1",
            Body: map[string]interface{}{
                "model": "openai/gpt-4o-mini",
                "messages": []map[string]interface{}{{"role": "user", "content": "Hello"}},
            },
        },
    },
})
batch, _ := client.Batches.Create(ctx, meshapi.CreateBatchParams{
    InputFileID:      file.ID,
    Endpoint:         "/v1/chat/completions",
    CompletionWindow: "24h",
})
```

## Error Handling

```go
resp, err := client.Chat.Completions.Create(ctx, params)
if err != nil {
    var svcErr *meshapi.MeshAPIError
    if errors.As(err, &svcErr) {
        fmt.Println(svcErr.Status)     // HTTP status
        fmt.Println(svcErr.Code)       // "unauthorized", "rate_limit_exceeded", etc.
        fmt.Println(svcErr.RequestID)  // req_<ULID>
    }
}
```

## Retry / Backoff

Retries on 429/502/503/504 with exponential backoff (default 3 retries, 500 ms base, 30 s max, ±20% jitter). Respects `Retry-After` header.

```go
maxRetries := 5
client := meshapi.New(meshapi.Config{
    MaxRetries: &maxRetries,
})
```

## Streaming Failure Recovery

**Streams do not retry.** On connection failure, the error channel receives a `MeshAPIError` with `Code="stream_interrupted"`. Restart a new `Stream` call to reconnect.

## Running Tests

```bash
# Unit + contract tests
go test ./...

# Integration tests (requires localhost:8000)
MESHAPI_BASE_URL=http://localhost:8000 \
MESHAPI_TOKEN=rsk_... \
go test -tags integration ./...
```

## Versioning

```go
fmt.Println(meshapi.Version)  // "0.1.0"
```
