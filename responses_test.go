package meshapi

import (
	"encoding/json"
	"testing"
)

func strPtr(s string) *string { return &s }

func TestResponsesInputTokensParams_Marshal(t *testing.T) {
	tests := []struct {
		name   string
		params ResponsesInputTokensParams
		want   map[string]any
		absent []string
	}{
		{
			name:   "string input only",
			params: ResponsesInputTokensParams{Model: "openai/gpt-4o-mini", Input: "hello"},
			want:   map[string]any{"model": "openai/gpt-4o-mini", "input": "hello"},
			absent: []string{"instructions", "tools", "tool_choice", "previous_response_id"},
		},
		{
			name: "instructions and previous response",
			params: ResponsesInputTokensParams{
				Model:              "openai/gpt-4o-mini",
				Input:              "next",
				Instructions:       strPtr("be brief"),
				PreviousResponseID: strPtr("resp_123"),
			},
			want: map[string]any{
				"model":                "openai/gpt-4o-mini",
				"input":                "next",
				"instructions":         "be brief",
				"previous_response_id": "resp_123",
			},
			absent: []string{"tools", "tool_choice"},
		},
		{
			name: "tool choice string",
			params: ResponsesInputTokensParams{
				Model:      "openai/gpt-4o-mini",
				Input:      "hi",
				ToolChoice: "auto",
			},
			want:   map[string]any{"tool_choice": "auto"},
			absent: []string{"tools"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b, err := json.Marshal(tc.params)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			var got map[string]any
			if err := json.Unmarshal(b, &got); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			for k, v := range tc.want {
				if got[k] != v {
					t.Errorf("field %q = %v, want %v", k, got[k], v)
				}
			}
			for _, k := range tc.absent {
				if _, ok := got[k]; ok {
					t.Errorf("field %q should be omitted", k)
				}
			}
		})
	}
}

func TestResponsesInputTokensResponse_Unmarshal(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantObject string
		wantTokens int
	}{
		{"typical", `{"object":"response.input_tokens","input_tokens":42}`, "response.input_tokens", 42},
		{"zero", `{"object":"response.input_tokens","input_tokens":0}`, "response.input_tokens", 0},
		{"extra fields ignored", `{"object":"response.input_tokens","input_tokens":7,"x":1}`, "response.input_tokens", 7},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var out ResponsesInputTokensResponse
			if err := json.Unmarshal([]byte(tc.body), &out); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if out.Object != tc.wantObject {
				t.Errorf("Object = %q, want %q", out.Object, tc.wantObject)
			}
			if out.InputTokens != tc.wantTokens {
				t.Errorf("InputTokens = %d, want %d", out.InputTokens, tc.wantTokens)
			}
		})
	}
}
