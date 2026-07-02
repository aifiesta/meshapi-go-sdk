package meshapi

import (
	"encoding/json"
	"os"
	"testing"
)

func loadFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatalf("load fixture %q: %v", name, err)
	}
	return data
}

func TestContract_ChatCompletionResponse(t *testing.T) {
	data := loadFixture(t, "chat_completion_response.json")
	var resp ChatCompletionResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.ID != "chatcmpl-abc123" {
		t.Errorf("expected id 'chatcmpl-abc123', got %q", resp.ID)
	}
	if len(resp.Choices) != 1 {
		t.Fatalf("expected 1 choice, got %d", len(resp.Choices))
	}
	if resp.Choices[0].Message == nil {
		t.Fatal("expected non-nil message")
	}
	if resp.Choices[0].Message.Content == nil || *resp.Choices[0].Message.Content != "2 + 2 equals 4." {
		t.Errorf("unexpected message content: %v", resp.Choices[0].Message.Content)
	}
	if resp.Usage == nil || resp.Usage.TotalTokens != 21 {
		t.Errorf("unexpected usage: %v", resp.Usage)
	}
}

func TestContract_ChatCompletionChunk(t *testing.T) {
	data := loadFixture(t, "chat_completion_chunk.json")
	var chunk ChatCompletionChunk
	if err := json.Unmarshal(data, &chunk); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(chunk.Choices) == 0 {
		t.Fatal("expected at least one choice")
	}
	if chunk.Choices[0].Delta == nil {
		t.Fatal("expected non-nil delta")
	}
	if chunk.Choices[0].Delta.Content == nil || *chunk.Choices[0].Delta.Content != "Hello" {
		t.Errorf("expected delta.content='Hello', got %v", chunk.Choices[0].Delta.Content)
	}
}

func TestContract_ModelList(t *testing.T) {
	data := loadFixture(t, "model_list.json")
	var models []ModelInfo
	if err := json.Unmarshal(data, &models); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(models) != 2 {
		t.Fatalf("expected 2 models, got %d", len(models))
	}
	free := 0
	paid := 0
	for _, m := range models {
		if m.IsFree {
			free++
		} else {
			paid++
		}
	}
	if free != 1 || paid != 1 {
		t.Errorf("expected 1 free + 1 paid, got free=%d paid=%d", free, paid)
	}
}

func TestContract_TemplateSummary(t *testing.T) {
	data := loadFixture(t, "template_summary.json")
	var tmpl TemplateSummary
	if err := json.Unmarshal(data, &tmpl); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if tmpl.ID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("unexpected id: %q", tmpl.ID)
	}
	if tmpl.System == nil || *tmpl.System != "You are a helpful assistant who speaks like a pirate." {
		t.Errorf("unexpected system: %v", tmpl.System)
	}
}

func TestContract_Error401(t *testing.T) {
	data := loadFixture(t, "error_401.json")
	var env apiErrorEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if env.Error.Code != "unauthorized" {
		t.Errorf("expected code 'unauthorized', got %q", env.Error.Code)
	}
	if env.RequestID != "req_01HZXYZ" {
		t.Errorf("expected request_id 'req_01HZXYZ', got %q", env.RequestID)
	}
}

func TestContract_Error429WithRetryAfter(t *testing.T) {
	data := loadFixture(t, "error_429.json")
	var env apiErrorEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if env.Error.Code != "rate_limit_exceeded" {
		t.Errorf("expected code 'rate_limit_exceeded', got %q", env.Error.Code)
	}
	if env.Error.RetryAfterSeconds == nil || *env.Error.RetryAfterSeconds != 5 {
		t.Errorf("expected retry_after_seconds=5, got %v", env.Error.RetryAfterSeconds)
	}
}

func TestContract_EmbeddingVector_Float(t *testing.T) {
	data := loadFixture(t, "embedding_response_float.json")
	var resp EmbeddingsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		t.Fatalf("unmarshal float embedding: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 embedding item, got %d", len(resp.Data))
	}
	item := resp.Data[0]
	if item.Embedding.IsBase64() {
		t.Error("expected float embedding, got base64")
	}
	floats := item.Embedding.Floats()
	if len(floats) != 4 {
		t.Fatalf("expected 4 floats, got %d", len(floats))
	}
	if floats[0] != 0.1 {
		t.Errorf("expected floats[0]=0.1, got %v", floats[0])
	}
}

func TestContract_EmbeddingVector_Base64(t *testing.T) {
	data := loadFixture(t, "embedding_response_base64.json")
	var resp EmbeddingsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		t.Fatalf("unmarshal base64 embedding: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 embedding item, got %d", len(resp.Data))
	}
	item := resp.Data[0]
	if !item.Embedding.IsBase64() {
		t.Error("expected base64 embedding, got float array")
	}
	b64 := item.Embedding.Base64()
	if b64 == "" {
		t.Error("expected non-empty base64 string")
	}
	if item.Embedding.Floats() != nil {
		t.Error("Floats() should return nil for base64 embedding")
	}
}

func TestContract_BatchObject_WithResults(t *testing.T) {
	data := loadFixture(t, "batch_completed.json")
	var batch BatchObject
	if err := json.Unmarshal(data, &batch); err != nil {
		t.Fatalf("unmarshal batch: %v", err)
	}
	if batch.ID != "batch_abc123" {
		t.Errorf("expected id 'batch_abc123', got %q", batch.ID)
	}
	if batch.Status != "completed" {
		t.Errorf("expected status 'completed', got %q", batch.Status)
	}
	if len(batch.Results) != 2 {
		t.Errorf("expected 2 results, got %d", len(batch.Results))
	}
	if batch.Results[0]["custom_id"] != "req-1" {
		t.Errorf("expected results[0].custom_id='req-1', got %v", batch.Results[0]["custom_id"])
	}
	if batch.CompletionWindow == nil || *batch.CompletionWindow != "24h" {
		t.Errorf("expected completion_window='24h', got %v", batch.CompletionWindow)
	}
	if batch.RequestCounts == nil {
		t.Error("expected non-nil request_counts")
	}
}

// TestContract_AudioTranslationResponse verifies that TranscriptionResponse
// unmarshals correctly — used by both Translate and the new Translations method.
func TestContract_AudioTranslationResponse(t *testing.T) {
	data := loadFixture(t, "audio_translation_response.json")
	var resp TranscriptionResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Text != "Hello, this is a translated audio message." {
		t.Errorf("unexpected text: %q", resp.Text)
	}
}

// TestContract_AudioTranslationParams verifies that AudioTranslationParams
// round-trips correctly and that optional fields are omitted when nil.
func TestContract_AudioTranslationParams(t *testing.T) {
	model := "whisper-1"
	params := AudioTranslationParams{Model: model}
	data, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	// prompt, response_format, temperature must not appear
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal map: %v", err)
	}
	if m["model"] != model {
		t.Errorf("expected model=%q, got %v", model, m["model"])
	}
	for _, k := range []string{"prompt", "response_format", "temperature"} {
		if _, ok := m[k]; ok {
			t.Errorf("key %q should be absent when nil", k)
		}
	}
}

// TestContract_ResponsesParams_NewFields verifies the new pass-2 fields on
// ResponsesParams are present with the correct JSON keys.
func TestContract_ResponsesParams_NewFields(t *testing.T) {
	prevID := "resp_prev"
	instr := "Be concise."
	store := true
	expireAt := int64(1893456000)
	maxTC := 3
	params := ResponsesParams{
		Input:              "Hello",
		PreviousResponseID: &prevID,
		Instructions:       &instr,
		Thinking:           map[string]interface{}{"enabled": true},
		Caching:            map[string]interface{}{"ttl": 300},
		Store:              &store,
		Include:            []interface{}{"usage"},
		ExpireAt:           &expireAt,
		MaxToolCalls:       &maxTC,
		ContextManagement:  map[string]interface{}{"strategy": "truncate"},
	}
	data, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	checks := map[string]interface{}{
		"previous_response_id": prevID,
		"instructions":         instr,
		"store":                true,
		"expire_at":            float64(expireAt),
		"max_tool_calls":       float64(maxTC),
	}
	for k, want := range checks {
		got, ok := m[k]
		if !ok {
			t.Errorf("missing key %q in marshalled JSON", k)
			continue
		}
		if got != want {
			t.Errorf("key %q: expected %v, got %v", k, want, got)
		}
	}
	for _, k := range []string{"thinking", "caching", "include", "context_management"} {
		if _, ok := m[k]; !ok {
			t.Errorf("missing key %q in marshalled JSON", k)
		}
	}
}

// TestContract_ChatCompletionParams_Cache verifies the new cache field.
func TestContract_ChatCompletionParams_Cache(t *testing.T) {
	cacheOn := true
	params := ChatCompletionParams{
		Messages: []ChatMessage{{Role: "user", Content: "hi"}},
		Cache:    &cacheOn,
	}
	data, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if m["cache"] != true {
		t.Errorf("expected cache=true, got %v", m["cache"])
	}
	// nil cache must be omitted
	params2 := ChatCompletionParams{Messages: params.Messages}
	data2, _ := json.Marshal(params2)
	var m2 map[string]interface{}
	json.Unmarshal(data2, &m2)
	if _, ok := m2["cache"]; ok {
		t.Error("cache key should be absent when nil")
	}
}

// TestContract_CreateTemplateParams_TeamID verifies the new team_id field.
func TestContract_CreateTemplateParams_TeamID(t *testing.T) {
	teamID := "team_abc"
	params := CreateTemplateParams{Name: "my-template", TeamID: &teamID}
	data, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if m["team_id"] != teamID {
		t.Errorf("expected team_id=%q, got %v", teamID, m["team_id"])
	}
	// nil team_id must be omitted
	params2 := CreateTemplateParams{Name: "my-template"}
	data2, _ := json.Marshal(params2)
	var m2 map[string]interface{}
	json.Unmarshal(data2, &m2)
	if _, ok := m2["team_id"]; ok {
		t.Error("team_id key should be absent when nil")
	}
}
