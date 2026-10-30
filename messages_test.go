package meshapi

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestMessagesCountTokensParams_Marshal(t *testing.T) {
	maxTokens := 1024
	tier := "auto"
	cases := []struct {
		name     string
		params   MessagesCountTokensParams
		contains []string
		absent   []string
	}{
		{
			name: "minimal body omits optional fields",
			params: MessagesCountTokensParams{
				Model:    "anthropic/claude-haiku-4.5",
				Messages: []any{map[string]any{"role": "user", "content": "Hello"}},
			},
			contains: []string{`"model":"anthropic/claude-haiku-4.5"`, `"messages":[`},
			absent:   []string{`"max_tokens"`, `"system"`, `"tools"`, `"mcp_servers"`, `"container"`, `"service_tier"`, `"provider"`},
		},
		{
			name: "optional fields are sent when set",
			params: MessagesCountTokensParams{
				Model:       "anthropic/claude-haiku-4.5",
				Messages:    []any{map[string]any{"role": "user", "content": "Hi"}},
				System:      "Be brief.",
				MaxTokens:   &maxTokens,
				ServiceTier: &tier,
				MCPServers:  []any{map[string]any{"type": "url", "name": "docs"}},
			},
			contains: []string{`"system":"Be brief."`, `"max_tokens":1024`, `"service_tier":"auto"`, `"mcp_servers":[`},
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

func TestMessagesCountTokensResponse_Unmarshal(t *testing.T) {
	cases := []struct {
		name string
		body string
		want int
	}{
		{name: "count", body: `{"input_tokens":42}`, want: 42},
		{name: "zero", body: `{"input_tokens":0}`, want: 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var out MessagesCountTokensResponse
			if err := json.Unmarshal([]byte(tc.body), &out); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if out.InputTokens != tc.want {
				t.Errorf("InputTokens = %d, want %d", out.InputTokens, tc.want)
			}
		})
	}
}
