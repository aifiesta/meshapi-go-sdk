package meshapi

import (
	"encoding/json"
	"testing"
)

// ── ChatCompletionParams.Timeout serialisation ────────────────────────────────

func TestChatParams_TimeoutSerialisedIntoBody(t *testing.T) {
	v := 600.0
	params := ChatCompletionParams{
		Messages: []ChatMessage{{Role: "user", Content: "hi"}},
		Timeout:  &v,
	}
	data, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	got, ok := m["timeout"]
	if !ok {
		t.Fatal("expected 'timeout' key in serialised body")
	}
	if got != 600.0 {
		t.Errorf("expected timeout=600, got %v", got)
	}
}

func TestChatParams_TimeoutAbsentWhenNil(t *testing.T) {
	params := ChatCompletionParams{
		Messages: []ChatMessage{{Role: "user", Content: "hi"}},
	}
	data, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, ok := m["timeout"]; ok {
		t.Error("'timeout' key must not be present when nil")
	}
}

// ── ResponsesParams.Timeout serialisation ────────────────────────────────────

func TestResponsesParams_TimeoutSerialisedIntoBody(t *testing.T) {
	v := 900.0
	params := ResponsesParams{
		Input:   "hello",
		Timeout: &v,
	}
	data, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	got, ok := m["timeout"]
	if !ok {
		t.Fatal("expected 'timeout' key in serialised body")
	}
	if got != 900.0 {
		t.Errorf("expected timeout=900, got %v", got)
	}
}

func TestResponsesParams_TimeoutAbsentWhenNil(t *testing.T) {
	params := ResponsesParams{Input: "hello"}
	data, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, ok := m["timeout"]; ok {
		t.Error("'timeout' key must not be present when nil")
	}
}

// ── SSE: gateway_timeout error frame ─────────────────────────────────────────

func TestTryParseSSEFrame_GatewayTimeoutFrame(t *testing.T) {
	// This is the exact frame the backend emits when the upstream provider
	// exceeds the server's 300 s provider timeout.
	frame := `data: {"error":{"code":"gateway_timeout","message":"Upstream provider did not respond in time."}}` + "\n"
	_, _, err := tryParseSSEFrame(frame)
	if err == nil {
		t.Fatal("expected error from gateway_timeout frame")
	}
	apiErr, ok := err.(*MeshAPIError)
	if !ok {
		t.Fatalf("expected *MeshAPIError, got %T: %v", err, err)
	}
	if apiErr.Code != "gateway_timeout" {
		t.Errorf("expected code 'gateway_timeout', got %q", apiErr.Code)
	}
}

func TestParseSSEStream_GatewayTimeoutAfterPartialContent(t *testing.T) {
	// The customer scenario: tokens arrive, then the backend times out mid-stream.
	timeoutFrame := `data: {"error":{"code":"gateway_timeout","message":"Upstream provider did not respond in time."}}` + "\n\n"
	frames := []string{
		makeChunkFrame("Hello "),
		makeChunkFrame("world"),
		timeoutFrame,
	}
	resp := makeSSEBody(frames)

	chunkCh := make(chan ChatCompletionChunk)
	errCh := make(chan error, 1)
	go parseSSEStream(resp, chunkCh, errCh)

	var chunks []ChatCompletionChunk
	for chunk := range chunkCh {
		chunks = append(chunks, chunk)
	}
	err := <-errCh

	if len(chunks) != 2 {
		t.Errorf("expected 2 content chunks before error, got %d", len(chunks))
	}
	if err == nil {
		t.Fatal("expected gateway_timeout error, got nil")
	}
	apiErr, ok := err.(*MeshAPIError)
	if !ok {
		t.Fatalf("expected *MeshAPIError, got %T: %v", err, err)
	}
	if apiErr.Code != "gateway_timeout" {
		t.Errorf("expected code 'gateway_timeout', got %q", apiErr.Code)
	}
}

func TestParseSSEStream_GatewayTimeoutNoContent(t *testing.T) {
	// Timeout fires before any content is streamed.
	timeoutFrame := `data: {"error":{"code":"gateway_timeout","message":"Upstream provider did not respond in time."}}` + "\n\n"
	resp := makeSSEBody([]string{timeoutFrame})

	chunks, err := collectChunks(resp)
	if err == nil {
		t.Fatalf("expected error, got %d chunks", len(chunks))
	}
	if len(chunks) != 0 {
		t.Errorf("expected 0 chunks, got %d", len(chunks))
	}
	apiErr, ok := err.(*MeshAPIError)
	if !ok {
		t.Fatalf("expected *MeshAPIError, got %T", err)
	}
	if apiErr.Code != "gateway_timeout" {
		t.Errorf("expected code 'gateway_timeout', got %q", apiErr.Code)
	}
}
