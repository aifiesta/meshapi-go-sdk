package meshapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

// newHeaderCaptureServer returns a client pointed at a test server that
// records the X-Request-Id request header of the last request, counts hits,
// and replies with the given body (JSON) plus a server-side X-Request-Id
// response header.
func newHeaderCaptureServer(t *testing.T, responseRequestID, body string) (*Client, *atomic.Int64, *atomic.Value) {
	t.Helper()
	var hits atomic.Int64
	var lastHeader atomic.Value
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		lastHeader.Store(r.Header.Get("X-Request-Id"))
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Request-Id", responseRequestID)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	client := New(Config{BaseURL: srv.URL, Token: "rsk_test"})
	return client, &hits, &lastHeader
}

func chatBody() string {
	return `{"id":"c1","object":"chat.completion","created":0,"model":"m",` +
		`"choices":[{"index":0,"message":{"role":"assistant","content":"hi"},"finish_reason":"stop"}]}`
}

// ── A) WithRequestID sends the header ────────────────────────────────────────

func TestWithRequestID_SendsHeader_JSONPost(t *testing.T) {
	client, _, lastHeader := newHeaderCaptureServer(t, "req_srv", chatBody())
	ctx := WithRequestID(context.Background(), "my-id_1.2:3")

	model := "openai/gpt-4o-mini"
	if _, err := client.Chat.Completions.Create(ctx, ChatCompletionParams{
		Model:    &model,
		Messages: []ChatMessage{{Role: "user", Content: "hi"}},
	}); err != nil {
		t.Fatalf("chat create: %v", err)
	}
	if got := lastHeader.Load(); got != "my-id_1.2:3" {
		t.Errorf("X-Request-Id header = %v, want %q", got, "my-id_1.2:3")
	}
}

func TestWithRequestID_SendsHeader_Get(t *testing.T) {
	body := `{"id":"tpl_1","name":"t","owner":null,"is_global":false,"created_at":"c","updated_at":"u"}`
	client, _, lastHeader := newHeaderCaptureServer(t, "req_srv", body)
	ctx := WithRequestID(context.Background(), "get-id-42")

	if _, err := client.Templates.Get(ctx, "tpl_1"); err != nil {
		t.Fatalf("templates get: %v", err)
	}
	if got := lastHeader.Load(); got != "get-id-42" {
		t.Errorf("X-Request-Id header = %v, want %q", got, "get-id-42")
	}
}

func TestWithRequestID_SendsHeader_Multipart(t *testing.T) {
	client, _, lastHeader := newHeaderCaptureServer(t, "req_srv", `{"text":"hello"}`)
	ctx := WithRequestID(context.Background(), "multipart-id")

	if _, err := client.Audio.Transcribe(ctx, []byte("audio"), "a.mp3",
		TranscriptionParams{Model: "scribe_v1"}); err != nil {
		t.Fatalf("audio transcribe: %v", err)
	}
	if got := lastHeader.Load(); got != "multipart-id" {
		t.Errorf("X-Request-Id header = %v, want %q", got, "multipart-id")
	}
}

func TestWithRequestID_SendsHeader_Stream(t *testing.T) {
	var lastHeader atomic.Value
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastHeader.Store(r.Header.Get("X-Request-Id"))
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: " + chatBody() + "\n\ndata: [DONE]\n\n"))
	}))
	t.Cleanup(srv.Close)
	client := New(Config{BaseURL: srv.URL, Token: "rsk_test"})

	ctx := WithRequestID(context.Background(), "stream-id")
	model := "openai/gpt-4o-mini"
	chunkCh, errCh := client.Chat.Completions.Stream(ctx, ChatCompletionParams{
		Model:    &model,
		Messages: []ChatMessage{{Role: "user", Content: "hi"}},
	})
	for range chunkCh {
	}
	if err := <-errCh; err != nil {
		t.Fatalf("stream: %v", err)
	}
	if got := lastHeader.Load(); got != "stream-id" {
		t.Errorf("X-Request-Id header = %v, want %q", got, "stream-id")
	}
}

// ── A) invalid id fails before any network I/O ───────────────────────────────

func TestWithRequestID_InvalidID_FailsBeforeSend(t *testing.T) {
	client, hits, _ := newHeaderCaptureServer(t, "req_srv", chatBody())
	model := "openai/gpt-4o-mini"
	params := ChatCompletionParams{
		Model:    &model,
		Messages: []ChatMessage{{Role: "user", Content: "hi"}},
	}

	invalidIDs := []string{
		"",                      // empty
		"has space",             // disallowed character
		"emoji-🚀",               // non-ASCII
		strings.Repeat("a", 65), // too long
		"slash/es",              // disallowed character
	}
	for _, id := range invalidIDs {
		ctx := WithRequestID(context.Background(), id)
		_, err := client.Chat.Completions.Create(ctx, params)
		if err == nil {
			t.Errorf("id %q: expected error, got nil", id)
			continue
		}
		if !strings.Contains(err.Error(), "invalid request id") {
			t.Errorf("id %q: error %q does not mention invalid request id", id, err)
		}
	}
	if n := hits.Load(); n != 0 {
		t.Errorf("server was hit %d times; invalid ids must fail before any network I/O", n)
	}
}

func TestWithRequestID_InvalidID_FailsBeforeStream(t *testing.T) {
	client, hits, _ := newHeaderCaptureServer(t, "req_srv", chatBody())
	ctx := WithRequestID(context.Background(), "bad id")
	model := "openai/gpt-4o-mini"
	chunkCh, errCh := client.Chat.Completions.Stream(ctx, ChatCompletionParams{
		Model:    &model,
		Messages: []ChatMessage{{Role: "user", Content: "hi"}},
	})
	for range chunkCh {
	}
	err := <-errCh
	if err == nil || !strings.Contains(err.Error(), "invalid request id") {
		t.Fatalf("expected invalid request id error, got %v", err)
	}
	if n := hits.Load(); n != 0 {
		t.Errorf("server was hit %d times; invalid ids must fail before any network I/O", n)
	}
}

// ── B) ResponseMeta populated from the response header ──────────────────────

func TestResponseMeta_PopulatedFromHeader(t *testing.T) {
	client, _, _ := newHeaderCaptureServer(t, "req_01ABCDEF", chatBody())
	model := "openai/gpt-4o-mini"
	resp, err := client.Chat.Completions.Create(context.Background(), ChatCompletionParams{
		Model:    &model,
		Messages: []ChatMessage{{Role: "user", Content: "hi"}},
	})
	if err != nil {
		t.Fatalf("chat create: %v", err)
	}
	if resp.RequestID != "req_01ABCDEF" {
		t.Errorf("resp.RequestID = %q, want %q", resp.RequestID, "req_01ABCDEF")
	}
}

func TestResponseMeta_PopulatedOnGet(t *testing.T) {
	body := `{"id":"tpl_1","name":"t","owner":null,"is_global":false,"created_at":"c","updated_at":"u"}`
	client, _, _ := newHeaderCaptureServer(t, "req_get_1", body)
	tmpl, err := client.Templates.Get(context.Background(), "tpl_1")
	if err != nil {
		t.Fatalf("templates get: %v", err)
	}
	if tmpl.RequestID != "req_get_1" {
		t.Errorf("tmpl.RequestID = %q, want %q", tmpl.RequestID, "req_get_1")
	}
}

func TestResponseMeta_PopulatedOnMultipart(t *testing.T) {
	client, _, _ := newHeaderCaptureServer(t, "req_mp_1", `{"text":"hello"}`)
	out, err := client.Audio.Transcribe(context.Background(), []byte("audio"), "a.mp3",
		TranscriptionParams{Model: "scribe_v1"})
	if err != nil {
		t.Fatalf("audio transcribe: %v", err)
	}
	if out.RequestID != "req_mp_1" {
		t.Errorf("out.RequestID = %q, want %q", out.RequestID, "req_mp_1")
	}
}

func TestResponseMeta_EchoesClientSuppliedID(t *testing.T) {
	// The backend echoes a valid client-supplied id back in X-Request-Id;
	// simulate that echo and assert it lands in ResponseMeta.
	var srvURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Request-Id", r.Header.Get("X-Request-Id"))
		_, _ = w.Write([]byte(chatBody()))
	}))
	t.Cleanup(srv.Close)
	srvURL = srv.URL
	client := New(Config{BaseURL: srvURL, Token: "rsk_test"})

	ctx := WithRequestID(context.Background(), "echo-me-123")
	model := "openai/gpt-4o-mini"
	resp, err := client.Chat.Completions.Create(ctx, ChatCompletionParams{
		Model:    &model,
		Messages: []ChatMessage{{Role: "user", Content: "hi"}},
	})
	if err != nil {
		t.Fatalf("chat create: %v", err)
	}
	if resp.RequestID != "echo-me-123" {
		t.Errorf("resp.RequestID = %q, want echoed %q", resp.RequestID, "echo-me-123")
	}
}

// ── B) structs with their own RequestID body field still parse correctly ────

func TestWebSearchResponse_RequestIDFromBody(t *testing.T) {
	// WebSearchResponse does not embed ResponseMeta — its RequestID comes from
	// the response body's "request_id" and must keep parsing from there.
	body := `{"query":"q","results":[],"provider":"tavily","request_id":"req_body_ws"}`
	client, _, _ := newHeaderCaptureServer(t, "req_header_ws", body)
	resp, err := client.Web.Search(context.Background(), WebSearchParams{Query: "q"})
	if err != nil {
		t.Fatalf("web search: %v", err)
	}
	if resp.RequestID != "req_body_ws" {
		t.Errorf("resp.RequestID = %q, want body value %q", resp.RequestID, "req_body_ws")
	}
}

func TestCompareResponse_TopLevelHeaderAndPerModelBodyIDs(t *testing.T) {
	// CompareResponse.RequestID (embedded ResponseMeta) is the compare
	// request's header id; Results[i].RequestID are per-model body ids.
	body := `{"comparison_id":"cmp_1","object":"chat.compare","created":0,` +
		`"models":["m1"],"results":[{"model":"m1","latency_ms":1,"request_id":"req_model_1"}],` +
		`"comparison_fallback_used":false,"total_latency_ms":1,"partial":false,"skip_comparison":true}`
	client, _, _ := newHeaderCaptureServer(t, "req_header_cmp", body)
	resp, err := client.Compare.Create(context.Background(), CompareParams{
		Models:   []string{"m1"},
		Messages: []ChatMessage{{Role: "user", Content: "hi"}},
	})
	if err != nil {
		t.Fatalf("compare create: %v", err)
	}
	if resp.RequestID != "req_header_cmp" {
		t.Errorf("resp.RequestID = %q, want header value %q", resp.RequestID, "req_header_cmp")
	}
	if len(resp.Results) != 1 || resp.Results[0].RequestID != "req_model_1" {
		t.Errorf("Results[0].RequestID = %+v, want %q", resp.Results, "req_model_1")
	}
}
