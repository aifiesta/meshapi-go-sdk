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
