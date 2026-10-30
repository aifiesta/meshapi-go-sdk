package meshapi

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestResponsesInputTokensParams_Marshal(t *testing.T) {
	instructions := "Answer in one word."
	prev := "resp_123"
	cases := []struct {
		name     string
		params   ResponsesInputTokensParams
		contains []string
		absent   []string
	}{
		{
			name:     "string input",
			params:   ResponsesInputTokensParams{Model: "openai/gpt-4o-mini", Input: "Hello"},
			contains: []string{`"model":"openai/gpt-4o-mini"`, `"input":"Hello"`},
			absent:   []string{`"stream"`, `"instructions"`, `"previous_response_id"`, `"provider"`, `"tools"`},
		},
		{
			name: "optional fields are sent when set",
			params: ResponsesInputTokensParams{
				Model:              "openai/gpt-4o-mini",
				Input:              []any{map[string]any{"role": "user", "content": "Hi"}},
				Instructions:       &instructions,
				PreviousResponseID: &prev,
			},
			contains: []string{`"input":[`, `"instructions":"Answer in one word."`, `"previous_response_id":"resp_123"`},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := json.Marshal(tc.params)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			s := string(b)
			for _, want := range tc.contains {
				if !strings.Contains(s, want) {
					t.Errorf("body %s missing %s", s, want)
				}
			}
			for _, bad := range tc.absent {
				if strings.Contains(s, bad) {
					t.Errorf("body %s should not contain %s", s, bad)
				}
			}
		})
	}
}

func TestResponsesInputTokensResponse_Unmarshal(t *testing.T) {
	cases := []struct {
		name       string
		body       string
		wantObject string
		wantTokens int
	}{
		{name: "with object", body: `{"object":"response.input_tokens","input_tokens":17}`, wantObject: "response.input_tokens", wantTokens: 17},
		{name: "tokens only", body: `{"input_tokens":3}`, wantObject: "", wantTokens: 3},
	}
	for _, tc := range cases {
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
