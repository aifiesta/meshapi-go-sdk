# meshapi-go-sdk

Go client for the MeshAPI AI model gateway.

## Requirements

- Go ≥ 1.21
- Zero external dependencies

## Installation

```bash
go get github.com/aifiesta/meshapi-go-sdk@v0.1.3
```

## Quick Start

```go
import meshapi "github.com/aifiesta/meshapi-go-sdk"

client := meshapi.New(meshapi.Config{
    BaseURL: "https://api.meshapi.ai",
    Token:   "rsk_...",
})

model := "openai/gpt-4o-mini"
resp, err := client.Chat.Completions.Create(ctx, meshapi.ChatCompletionParams{
    Model:    &model,
    Messages: []meshapi.ChatMessage{{Role: "user", Content: "What is 2+2?"}},
})
```

## Chat completions

```go
// Non-streaming
resp, err := client.Chat.Completions.Create(ctx, meshapi.ChatCompletionParams{
    Model:    &model,
    Messages: []meshapi.ChatMessage{{Role: "user", Content: "Hello!"}},
})
fmt.Println(resp.Choices[0].Message.Content)

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
```

## Responses API (reasoning models)

```go
respModel := "openai/o3-mini"
resp, _ := client.Responses.Create(ctx, meshapi.ResponsesParams{
    Model: &respModel,
    Input: "Solve for X: 2x + 5 = 15",
})
```

## Embeddings

```go
embModel := "openai/text-embedding-3-small"
emb, _ := client.Embeddings.Create(ctx, meshapi.EmbeddingsParams{
    Model: &embModel,
    Input: []string{"The quick brown fox"},
})
```

## Image generation

```go
imgModel := "openai/gpt-image-1"
imgN := 1
imgSize := "1024x1024"
img, _ := client.Images.Generate(ctx, meshapi.ImageGenerationParams{
    Model:  &imgModel,
    Prompt: "A watercolor of a fox in a snowy forest",
    N:      &imgN,
    Size:   &imgSize,
})

// Streaming
chunkCh, errCh := client.Images.Stream(ctx, params)
for chunk := range chunkCh { ... }
```

## Compare (multi-model)

```go
compCh, errCh := client.Compare.Stream(ctx, meshapi.CompareParams{
    Models:   []string{"openai/gpt-4o-mini", "anthropic/claude-haiku-4-5"},
    Messages: []meshapi.ChatMessage{{Role: "user", Content: "Hello"}},
})
for event := range compCh { ... }
```

## Batches

Batch jobs accept inline requests — no separate file upload step required.

```go
batch, _ := client.Batches.Create(ctx, meshapi.CreateBatchParams{
    Requests: []meshapi.BatchRequestItem{
        {
            CustomID: "req-1",
            Body: map[string]interface{}{
                "model":    "openai/gpt-5-nano",
                "messages": []map[string]interface{}{{"role": "user", "content": "Hello"}},
            },
        },
    },
    Metadata: map[string]interface{}{"job": "my-batch"},
})

// Poll
got, _ := client.Batches.Get(ctx, batch.ID)
fmt.Println(got.Status)

// Cancel
client.Batches.Cancel(ctx, batch.ID)
```

## RAG (Retrieval-Augmented Generation)

Upload files, embed them, and run vector search.

```go
import "net/http"

// 1. Initialise upload — get a signed URL
upload, _ := client.RAG.InitUpload(ctx, meshapi.InitUploadRequest{
    FileName: "handbook.pdf",
    MimeType: "application/pdf",
})

// 2a. PUT file bytes to the signed URL yourself…
req, _ := http.NewRequestWithContext(ctx, http.MethodPut, upload.SignedURL, fileReader)
req.Header.Set("Content-Type", "application/pdf")
http.DefaultClient.Do(req)

// 2b. …or use the convenience wrapper that does both steps:
upload, _ = client.RAG.UploadFile(ctx, meshapi.UploadFileParams{
    FileName: "handbook.pdf",
    MimeType: "application/pdf",
    Content:  fileBytes,
})

// 3. Trigger embedding
client.RAG.Embed(ctx, meshapi.BulkEmbedRequest{
    FileIDs: []string{upload.FileID},
})

// 4. Poll until ready
for {
    s, _ := client.RAG.Get(ctx, upload.FileID)
    if s.EmbeddingStatus == "ready" { break }
    time.Sleep(3 * time.Second)
}

// 5. Search
topK := 5
results, _ := client.RAG.Search(ctx, meshapi.SearchRequest{
    Query: "onboarding process",
    TopK:  &topK,
})
for _, r := range results.Results {
    fmt.Printf("%.4f  %s\n", r.Score, r.Text)
}

// List files
list, _ := client.RAG.List(ctx, meshapi.ListRagFilesParams{Limit: intPtr(50)})
```

## Realtime (Speech-to-Speech WebSocket)

```go
session, err := client.Realtime.Connect(ctx, meshapi.RealtimeConnectParams{
    Model: "openai/gpt-4o-realtime-preview",
})
if err != nil {
    log.Fatal(err)
}
defer session.Close()

// Send a JSON event
session.Send(ctx, map[string]any{
    "type":    "session.update",
    "session": map[string]any{"instructions": "You are a helpful assistant."},
})

// Send raw audio bytes
session.SendAudio(ctx, pcmBytes)

// Receive frames one at a time
msg, err := session.Receive(ctx)
if err != nil {
    var re *meshapi.RealtimeError
    if errors.As(err, &re) {
        fmt.Println(re.Code) // "insufficient_quota", "idle_timeout", …
    }
}
fmt.Println(msg.Event["type"]) // "session.created", "response.done", …
if msg.Audio != nil { /* binary audio frame */ }

// Or use the channel-based pump
msgCh, errCh := session.Events(ctx)
for msg := range msgCh {
    fmt.Println(msg.Event["type"])
}
if err := <-errCh; err != nil {
    log.Fatal(err)
}
```

Auth is sent via `Sec-WebSocket-Protocol: openai-realtime, Bearer <token>`. The session is safe to send and receive from separate goroutines simultaneously. No external dependencies — the WebSocket client is implemented in the standard library.

## Models

```go
models, _ := client.Models.List(ctx, meshapi.ListModelsParams{})
free, _   := client.Models.Free(ctx)
```

## Templates

```go
tmpl, _ := client.Templates.Create(ctx, meshapi.CreateTemplateParams{Name: "my-tpl"})
client.Templates.Delete(ctx, tmpl.ID)
```

## Error handling

```go
resp, err := client.Chat.Completions.Create(ctx, params)
if err != nil {
    var svcErr *meshapi.MeshAPIError
    if errors.As(err, &svcErr) {
        fmt.Println(svcErr.Status)    // HTTP status
        fmt.Println(svcErr.Code)      // "unauthorized", "rate_limit_exceeded", …
        fmt.Println(svcErr.RequestID) // req_<ULID>
    }
}
```

## Retry / backoff

Retries on 429/502/503/504 with exponential backoff (default 3 retries, 500 ms base, 30 s max, ±20% jitter). Respects `Retry-After`.

```go
maxRetries := 5
client := meshapi.New(meshapi.Config{MaxRetries: &maxRetries})
```

**Streams do not retry.** On connection failure the error channel receives a `MeshAPIError` with `Code="stream_interrupted"`.

## Running tests

```bash
# Unit + contract tests (no server needed)
go test ./...

# Live tests (requires a running backend)
cd livetests
MESHAPI_TOKEN=rsk_... go test ./... -v -timeout 300s
```

## Versioning

```go
fmt.Println(meshapi.Version) // "0.1.0"
```
