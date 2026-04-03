package meshapi

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// tryParseSSEFrame unit tests
// ---------------------------------------------------------------------------

func TestTryParseSSEFrame_ValidChunk(t *testing.T) {
	frame := `data: {"id":"x","object":"chat.completion.chunk","created":1,"model":"m","choices":[{"index":0,"delta":{"content":"Hi"},"finish_reason":null}]}` + "\n"
	chunk, done, err := tryParseSSEFrame(frame)
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
	_, done, err := tryParseSSEFrame(frame)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !done {
		t.Error("expected done=true")
	}
}

func TestTryParseSSEFrame_EmptyFrame(t *testing.T) {
	chunk, done, err := tryParseSSEFrame("")
	if err != nil || done || chunk != nil {
		t.Errorf("empty frame: chunk=%v done=%v err=%v", chunk, done, err)
	}
}

func TestTryParseSSEFrame_MalformedJSON(t *testing.T) {
	chunk, done, err := tryParseSSEFrame("data: {not valid}\n")
	if err != nil || done || chunk != nil {
		t.Errorf("malformed: chunk=%v done=%v err=%v", chunk, done, err)
	}
}

func TestTryParseSSEFrame_ErrorFrame(t *testing.T) {
	frame := `data: {"error":{"code":"upstream_error","message":"Provider failed"}}` + "\n"
	_, _, err := tryParseSSEFrame(frame)
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
