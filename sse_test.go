package meshapi

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"testing/iotest"
)

// ---------------------------------------------------------------------------
// tryParseSSEFrame unit tests
// ---------------------------------------------------------------------------

func TestTryParseSSEFrame_ValidChunk(t *testing.T) {
	frame := `data: {"id":"x","object":"chat.completion.chunk","created":1,"model":"m","choices":[{"index":0,"delta":{"content":"Hi"},"finish_reason":null}]}` + "\n"
	chunk, done, err := tryParseSSEFrame(frame, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if done {
		t.Fatal("expected done=false")
	}
	if chunk == nil {
		t.Fatal("expected non-nil chunk")
	}
	if len(chunk.Choices) == 0 {
		t.Fatal("expected at least one choice")
	}
	if chunk.Choices[0].Delta == nil || chunk.Choices[0].Delta.Content == nil {
		t.Fatal("expected delta.content")
	}
	if *chunk.Choices[0].Delta.Content != "Hi" {
		t.Errorf("expected 'Hi', got %q", *chunk.Choices[0].Delta.Content)
	}
}

func TestTryParseSSEFrame_DoneSentinel(t *testing.T) {
	frame := "data: [DONE]\n"
	_, done, err := tryParseSSEFrame(frame, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !done {
		t.Error("expected done=true")
	}
}

func TestTryParseSSEFrame_EmptyFrame(t *testing.T) {
	chunk, done, err := tryParseSSEFrame("", "")
	if err != nil || done || chunk != nil {
		t.Errorf("empty frame: chunk=%v done=%v err=%v", chunk, done, err)
	}
}

func TestTryParseSSEFrame_MalformedJSON(t *testing.T) {
	chunk, done, err := tryParseSSEFrame("data: {not valid}\n", "")
	if err != nil || done || chunk != nil {
		t.Errorf("malformed: chunk=%v done=%v err=%v", chunk, done, err)
	}
}

func TestTryParseSSEFrame_ErrorFrame(t *testing.T) {
	frame := `data: {"error":{"code":"upstream_error","message":"Provider failed"}}` + "\n"
	_, _, err := tryParseSSEFrame(frame, "")
	if err == nil {
		t.Fatal("expected error from error frame")
	}
	svcErr, ok := err.(*MeshAPIError)
	if !ok {
		t.Fatalf("expected *MeshAPIError, got %T", err)
	}
	if svcErr.Code != "upstream_error" {
		t.Errorf("expected code 'upstream_error', got %q", svcErr.Code)
	}
}

// ---------------------------------------------------------------------------
// parseSSEStream integration tests (with mock http.Response)
// ---------------------------------------------------------------------------

func makeSSEBody(frames []string) *http.Response {
	body := strings.Join(frames, "")
	return &http.Response{
		StatusCode: 200,
		Body:       nopReadCloser(strings.NewReader(body)),
		Header:     http.Header{},
	}
}

func collectChunks(resp *http.Response) ([]ChatCompletionChunk, error) {
	chunkCh := make(chan ChatCompletionChunk)
	errCh := make(chan error, 1)
	go parseSSEStream(resp, chunkCh, errCh)

	var chunks []ChatCompletionChunk
	for chunk := range chunkCh {
		chunks = append(chunks, chunk)
	}
	return chunks, <-errCh
}

func makeChunkFrame(content string) string {
	return fmt.Sprintf(
		`data: {"id":"x","object":"chat.completion.chunk","created":1,"model":"m","choices":[{"index":0,"delta":{"content":%q},"finish_reason":null}]}`+"\n\n",
		content,
	)
}

func TestParseSSEStream_SingleChunk(t *testing.T) {
	frames := []string{makeChunkFrame("Hello"), "data: [DONE]\n\n"}
	resp := makeSSEBody(frames)
	chunks, err := collectChunks(resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}
}

func TestParseSSEStream_MultipleChunks(t *testing.T) {
	frames := []string{
		makeChunkFrame("A"),
		makeChunkFrame("B"),
		makeChunkFrame("C"),
		"data: [DONE]\n\n",
	}
	resp := makeSSEBody(frames)
	chunks, err := collectChunks(resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chunks) != 3 {
		t.Fatalf("expected 3 chunks, got %d", len(chunks))
	}
}

func TestParseSSEStream_DoneTerminates(t *testing.T) {
	// Frame after [DONE] must not be yielded
	frames := []string{
		makeChunkFrame("First"),
		"data: [DONE]\n\n",
		makeChunkFrame("NEVER"),
	}
	resp := makeSSEBody(frames)
	chunks, err := collectChunks(resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk (before DONE), got %d", len(chunks))
	}
}

func TestParseSSEStream_MidStreamError(t *testing.T) {
	errorFrame := `data: {"error":{"code":"upstream_error","message":"Server died"}}` + "\n\n"
	frames := []string{makeChunkFrame("Part1"), errorFrame}
	resp := makeSSEBody(frames)

	chunkCh := make(chan ChatCompletionChunk)
	errCh := make(chan error, 1)
	go parseSSEStream(resp, chunkCh, errCh)

	first := <-chunkCh
	if first.Choices[0].Delta == nil || *first.Choices[0].Delta.Content != "Part1" {
		t.Error("expected first chunk to be 'Part1'")
	}

	// Drain remaining chunks
	for range chunkCh {
	}

	err := <-errCh
	if err == nil {
		t.Fatal("expected error from mid-stream error frame")
	}
	svcErr, ok := err.(*MeshAPIError)
	if !ok {
		t.Fatalf("expected *MeshAPIError, got %T", err)
	}
	if svcErr.Code != "upstream_error" {
		t.Errorf("expected code 'upstream_error', got %q", svcErr.Code)
	}
}

// ---------------------------------------------------------------------------
// request_id on mid-stream error frames
//
// Reported by a customer on the Node SDK: an upstream_error mid-stream arrived
// with an empty request id, so the one failure they most needed to report was
// the one they could not identify. Every SDK had the same gap.
// ---------------------------------------------------------------------------

func TestErrorFrameFallsBackToResponseHeader(t *testing.T) {
	frame := `data: {"error":{"code":"upstream_error","message":"boom"}}` + "\n"
	_, _, err := tryParseSSEFrame(frame, "req_hdr")
	svcErr, ok := err.(*MeshAPIError)
	if !ok {
		t.Fatalf("expected *MeshAPIError, got %T", err)
	}
	if svcErr.RequestID != "req_hdr" {
		t.Errorf("RequestID = %q, want %q", svcErr.RequestID, "req_hdr")
	}
}

func TestErrorFramePrefersItsOwnRequestID(t *testing.T) {
	frame := `data: {"error":{"code":"upstream_error","message":"boom"},"request_id":"req_body"}` + "\n"
	_, _, err := tryParseSSEFrame(frame, "req_hdr")
	svcErr := err.(*MeshAPIError)
	if svcErr.RequestID != "req_body" {
		t.Errorf("RequestID = %q, want the frame's own %q", svcErr.RequestID, "req_body")
	}
}

func TestErrorFrameFallsBackOnEmptyFrameID(t *testing.T) {
	frame := `data: {"error":{"code":"upstream_error","message":"boom"},"request_id":""}` + "\n"
	_, _, err := tryParseSSEFrame(frame, "req_hdr")
	svcErr := err.(*MeshAPIError)
	if svcErr.RequestID != "req_hdr" {
		t.Errorf("RequestID = %q, want header fallback %q", svcErr.RequestID, "req_hdr")
	}
}

func TestErrorFrameIDEmptyWhenNeitherSourceHasOne(t *testing.T) {
	frame := `data: {"error":{"code":"upstream_error","message":"boom"}}` + "\n"
	_, _, err := tryParseSSEFrame(frame, "")
	svcErr := err.(*MeshAPIError)
	if svcErr.RequestID != "" {
		t.Errorf("RequestID = %q, want empty", svcErr.RequestID)
	}
}

func TestResponseMetaCarriesRequestID(t *testing.T) {
	// Non-streaming parity: the id lands on the returned value, so concurrent
	// calls stay distinguishable without any shared callback.
	var resp ChatCompletionResponse
	var setter interface{ setRequestID(string) } = &resp
	setter.setRequestID("req_nonstream")
	if resp.RequestID != "req_nonstream" {
		t.Errorf("RequestID = %q, want %q", resp.RequestID, "req_nonstream")
	}
}

// TestEveryDecodeTargetCarriesRequestID guards against the gap Greptile found:
// ResponseMeta was embedded only in types named *Response, so BatchObject,
// RagFileStatus, TemplateSummary, Voice and ModelsPage decoded fine and then
// silently dropped the header — the API promised an id those callers could
// never read.
//
// Enumerating the types by hand is what caused the miss, so this walks the
// source for every decode destination instead.
func TestEveryDecodeTargetCarriesRequestID(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	// `var out X` / `var page X` / … immediately before an h.get/post/patch call
	// is how every resource decodes. Slices are exempt: a []T has nowhere to put
	// the id, exactly as arrays are exempt in the Node SDK.
	decl := regexp.MustCompile(`(?m)^\s*var (?:out|page|resp|result) ([A-Za-z]\w*)$`)
	seen := map[string]bool{}
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range decl.FindAllStringSubmatch(string(src), -1) {
			seen[m[1]] = true
		}
	}
	if len(seen) == 0 {
		t.Fatal("found no decode destinations — the detection regex has rotted")
	}

	types, err := os.ReadFile("types.go")
	if err != nil {
		t.Fatal(err)
	}
	src := string(types)
	for name := range seen {
		if name == "T" { // generic parameter, not a concrete type
			continue
		}
		start := strings.Index(src, "type "+name+" struct {")
		if start == -1 {
			continue // declared elsewhere; not a body-decoded response type
		}
		end := strings.Index(src[start:], "\n}")
		if end == -1 {
			t.Fatalf("could not find end of type %s", name)
		}
		if !strings.Contains(src[start:start+end], "ResponseMeta") {
			t.Errorf("%s is decoded from a response body but does not embed ResponseMeta, "+
				"so its callers cannot read RequestID", name)
		}
	}
}

func TestStreamInterruptedCarriesRequestID(t *testing.T) {
	// A connection that dies mid-stream is exactly as untraceable as a
	// mid-stream error frame: the 200 and its headers are long gone. Both
	// scanner-error paths had the id in hand and dropped it.
	resp := &http.Response{
		Header: http.Header{"X-Request-Id": []string{"req_interrupted"}},
		Body:   io.NopCloser(iotest.ErrReader(errors.New("connection reset"))),
	}
	chunkCh := make(chan ChatCompletionChunk)
	errCh := make(chan error, 1)
	go parseSSEStream(resp, chunkCh, errCh)
	for range chunkCh {
	}

	err := <-errCh
	svcErr, ok := err.(*MeshAPIError)
	if !ok {
		t.Fatalf("expected *MeshAPIError, got %T (%v)", err, err)
	}
	if svcErr.Code != "stream_interrupted" {
		t.Errorf("Code = %q, want stream_interrupted", svcErr.Code)
	}
	if svcErr.RequestID != "req_interrupted" {
		t.Errorf("RequestID = %q, want %q", svcErr.RequestID, "req_interrupted")
	}
}
