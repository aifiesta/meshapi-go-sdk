# meshapi-go-sdk

Go client for the MeshAPI AI model gateway.

## Requirements

- Go ≥ 1.21
- Zero external dependencies

## Installation

```bash
go get github.com/aifiesta/meshapi-go-sdk@latest
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

## Structured outputs

`meshapi.Parse[T]` constrains the model to a JSON schema and decodes the reply
into `T`. Define a struct with `json` tags — the schema is derived from it by
reflection. Because Go methods can't have type parameters, `Parse` is a
package-level function that takes the completions resource.

```go
type Country struct {
    Country            string `json:"country"`
    Capital            string `json:"capital"`
    PopulationMillions float64 `json:"population_millions"`
}

model := "openai/gpt-4o-mini"
country, err := meshapi.Parse[Country](ctx, client.Chat.Completions,
    meshapi.ChatCompletionParams{
        Model:    &model,
        Messages: []meshapi.ChatMessage{{Role: "user", Content: "Give me structured facts about France."}},
    },
)
if err != nil {
    log.Fatal(err)
}
fmt.Println(country.Capital, country.PopulationMillions) // typed
```

Options: `meshapi.WithMaxRetries(n)` re-prompts on a decode failure (default 0,
each retry is a billed call); `meshapi.WithSchema(map[string]any{...})` overrides
the auto-derived schema; `meshapi.WithSchemaName("...")` sets the schema label.

> Go's `json.Unmarshal` does not enforce required fields — a missing field
> decodes to its zero value. Type mismatches and non-JSON prose are caught.

### When the model doesn't support structured output

If decoding fails after any retries, `Parse` returns a `*StructuredOutputError`
(the underlying `json` error is on `.Cause`, reachable via `errors.As`). When the
model returns plain text instead of JSON — usually because it doesn't support
`response_format` — the message points at the model's support:

```go
var soErr *meshapi.StructuredOutputError
if errors.As(err, &soErr) {
    fmt.Println(soErr.Message)
    // "… the model returned text that is not valid JSON … Check the model's
    //  support on the Models page (https://app.meshapi.ai/…/models) …"
}
```

Check a model's `supports_structured_output` flag via `client.Models`, or on the
Models page in your dashboard. `Parse` is non-streaming.

## Responses API (reasoning models)

```go
respModel := "openai/o3-mini"
resp, _ := client.Responses.Create(ctx, meshapi.ResponsesParams{
    Model: &respModel,
    Input: "Solve for X: 2x + 5 = 15",
})

// List background response jobs, or fetch a persisted/background response by id.
// Synchronous create responses are not guaranteed to be retrievable via Get.
limit := 20
jobs, _ := client.Responses.List(ctx, meshapi.ResponsesListParams{Limit: &limit})
job, _ := client.Responses.Get(ctx, "resp_abc123")
```

## Embeddings

```go
embModel := "openai/text-embedding-3-small"
emb, _ := client.Embeddings.Create(ctx, meshapi.EmbeddingsParams{
    Model: &embModel,
    Input: []string{"The quick brown fox"},
})
```

## Audio (TTS, STT, voices)

```go
// Text-to-speech — returns []byte of raw audio
ttsModel := "sarvam/bulbul:v2"
voiceName := "meera"
audioBytes, err := client.Audio.Synthesize(ctx, meshapi.SpeechParams{
    Input: "Hello from MeshAPI.",
    Model: &ttsModel,
    Voice: &voiceName,
})
os.WriteFile("output.wav", audioBytes, 0644)

// Speech-to-text — send raw audio bytes with a filename hint
fileData, _ := os.ReadFile("audio.wav")
result, err := client.Audio.Transcribe(ctx, fileData, "audio.wav", meshapi.TranscriptionParams{
    Model: "sarvam/saaras:v3",
    // Optional: LanguageCode is model-specific (e.g. Sarvam expects "en-IN", not "en").
})
fmt.Println(result.Text)

// Translate audio to English via /v1/audio/transcriptions/translate
translateModel := "sarvam/saaras:v3"
translated, err := client.Audio.Translate(ctx, fileData, "audio.wav", &meshapi.TranscriptionTranslateParams{
    Model: &translateModel,
})
fmt.Println(translated.Text)

// Standalone audio translation via POST /v1/audio/translations
// (distinct endpoint from Translate above)
translated2, err := client.Audio.Translations(ctx, fileData, "audio.wav", meshapi.AudioTranslationParams{
    Model: "openai/whisper-large-v3",
})
fmt.Println(translated2.Text)

// List available voices
pageSize := 10
voices, err := client.Audio.ListVoices(ctx, &meshapi.ListVoicesParams{PageSize: &pageSize})

// Get a specific voice
voice, err := client.Audio.GetVoice(ctx, "voice-id")
```

## Video generation

```go
// Submit a video generation task
prompt := "A serene mountain lake at sunrise"
task, err := client.Videos.Generate(ctx, meshapi.VideoGenerationParams{
    Model: "byteplus/dreamina-seedance-2-0",
    Content: []meshapi.VideoContentItem{
        {Type: "text", Text: &prompt},
    },
})
fmt.Println("Task ID:", task.ID)

// Poll until complete
for {
    status, _ := client.Videos.Retrieve(ctx, task.ID)
    if status.Status == "succeeded" || status.Status == "failed" {
        break
    }
    time.Sleep(5 * time.Second)
}

// List past generation tasks
limit := 20
listing, err := client.Videos.List(ctx, meshapi.ListVideoGenerationsParams{Limit: &limit})
fmt.Printf("%d total tasks\n", listing.Total)
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

// Editing — Image is a base64/data-URL string (or meshapi.ImageRef);
// remote http(s) URLs are rejected by this endpoint.
editPrompt, editOp := "Replace the background with a beach at sunset", "edit"
edited, _ := client.Images.Edit(ctx, meshapi.ImageEditParams{
    Model:     "openai/gpt-image-1",
    Image:     "data:image/png;base64,<...>",
    Prompt:    &editPrompt,
    Operation: &editOp, // or inpaint / outpaint / mix / reframe / upscale / remove_background
})
```

## Compare (multi-model)

```go
compCh, errCh := client.Compare.Stream(ctx, meshapi.CompareParams{
    Models:   []string{"openai/gpt-4o-mini", "anthropic/claude-haiku-4.5"},
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
limit := 50
list, _ := client.RAG.List(ctx, meshapi.ListRagFilesParams{Limit: &limit})
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

// Send a JSON event (GA session shape)
session.Send(ctx, map[string]any{
    "type": "session.update",
    "session": map[string]any{
        "type":              "realtime",
        "output_modalities": []string{"audio"},
        "instructions":      "You are a helpful assistant.",
        "audio": map[string]any{
            "input":  map[string]any{"format": map[string]any{"type": "audio/pcm", "rate": 24000}},
            "output": map[string]any{"format": map[string]any{"type": "audio/pcm", "rate": 24000}, "voice": "alloy"},
        },
    },
})

// Append input audio (PCM16 24kHz) — sent as base64 input_audio_buffer.append
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

// Paginated catalog search (DB-only, no model cost)
q, limit := "gpt", 10
page, _ := client.Models.Search(ctx, meshapi.ModelSearchParams{Q: &q, Limit: &limit})
fmt.Println(page.Total, page.Brands)

// Fetch one model's detail
gpt4o, _ := client.Models.Get(ctx, "openai/gpt-4o")
```

## Moderations

```go
res, _ := client.Moderations.Create(ctx, meshapi.ModerationParams{Input: "text to classify"})
if len(res.Results) > 0 && res.Results[0].Flagged {
    fmt.Println("flagged:", res.Results[0].Categories)
}
```

## Web search

Gated server-side by `WEB_SEARCH_ENABLED`. Native-first with Tavily fallback;
inspect `res.Provider` to see which engine served the request.

```go
maxResults, includeAnswer := 5, true
res, _ := client.Web.Search(ctx, meshapi.WebSearchParams{
    Query:         "latest news on Mars rovers",
    MaxResults:    &maxResults,
    IncludeAnswer: &includeAnswer,
})
fmt.Println(res.Provider)
for _, hit := range res.Results {
    fmt.Println(hit.Title, hit.URL)
}
```

## Router select

Gated server-side by `AUTO_ROUTER_ENABLED`. Returns the model the Auto Router
*would* pick — without running inference.

```go
sel, _ := client.Router.Select(ctx, meshapi.RouterSelectParams{
    Messages: []meshapi.ChatMessage{{Role: "user", Content: "Prove that 2+2=4."}},
})
fmt.Println(sel.Model, sel.AutoRouter.FallbackUsed)
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
fmt.Println(meshapi.Version) // "0.1.12"
```
