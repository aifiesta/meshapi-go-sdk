package meshapi

import "testing"

func TestIsSupportedModerationModel(t *testing.T) {
	tests := []struct {
		name  string
		model string
		want  bool
	}{
		{name: "empty uses server default", model: "", want: true},
		{name: "latest bare", model: "omni-moderation-latest", want: true},
		{name: "latest with prefix", model: "openai/omni-moderation-latest", want: true},
		{name: "dated snapshot bare", model: "omni-moderation-2024-09-26", want: true},
		{name: "dated snapshot with prefix", model: "openai/omni-moderation-2024-09-26", want: true},
		{name: "chat model rejected", model: "openai/gpt-4o-mini", want: false},
		{name: "other provider prefix rejected", model: "anthropic/omni-moderation-latest", want: false},
		{name: "malformed date rejected", model: "omni-moderation-2024-9-26", want: false},
		{name: "trailing text rejected", model: "omni-moderation-latest-x", want: false},
		{name: "arbitrary string rejected", model: "hello world", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSupportedModerationModel(tt.model); got != tt.want {
				t.Errorf("IsSupportedModerationModel(%q) = %v, want %v", tt.model, got, tt.want)
			}
		})
	}
}

func TestIsModerationModelNotFound(t *testing.T) {
	tests := []struct {
		name   string
		status int
		code   string
		want   bool
	}{
		{name: "404 model_not_found", status: 404, code: "model_not_found", want: true},
		{name: "404 other code", status: 404, code: "not_found", want: false},
		{name: "400 model_not_found", status: 400, code: "model_not_found", want: false},
		{name: "503 upstream", status: 503, code: "", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsModerationModelNotFound(tt.status, tt.code); got != tt.want {
				t.Errorf("IsModerationModelNotFound(%d, %q) = %v, want %v", tt.status, tt.code, got, tt.want)
			}
		})
	}
}
