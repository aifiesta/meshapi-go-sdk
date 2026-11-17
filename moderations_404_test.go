package meshapi

import (
	"net/http"
	"testing"
)

func TestIsAcceptedModerationModel(t *testing.T) {
	tests := []struct {
		name  string
		model string
		want  bool
	}{
		{name: "latest bare", model: "omni-moderation-latest", want: true},
		{name: "latest prefixed", model: "openai/omni-moderation-latest", want: true},
		{name: "snapshot bare", model: "omni-moderation-2024-09-26", want: true},
		{name: "snapshot prefixed", model: "openai/omni-moderation-2024-09-26", want: true},
		{name: "empty", model: "", want: false},
		{name: "prefix only", model: "openai/", want: false},
		{name: "unknown id", model: "not-a-model", want: false},
		{name: "chat model", model: "openai/gpt-4o-mini", want: false},
		{name: "other provider prefix", model: "anthropic/omni-moderation-latest", want: false},
		{name: "malformed date", model: "omni-moderation-2024-9-26", want: false},
		{name: "trailing text", model: "omni-moderation-latest-x", want: false},
		{name: "double prefix", model: "openai/openai/omni-moderation-latest", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAcceptedModerationModel(tt.model); got != tt.want {
				t.Fatalf("IsAcceptedModerationModel(%q) = %v, want %v", tt.model, got, tt.want)
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
		{name: "404 model_not_found", status: http.StatusNotFound, code: "model_not_found", want: true},
		{name: "404 other code", status: http.StatusNotFound, code: "not_found", want: false},
		{name: "400 model_not_found", status: http.StatusBadRequest, code: "model_not_found", want: false},
		{name: "503 empty code", status: http.StatusServiceUnavailable, code: "", want: false},
		{name: "200", status: http.StatusOK, code: "", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsModerationModelNotFound(tt.status, tt.code); got != tt.want {
				t.Fatalf("IsModerationModelNotFound(%d, %q) = %v, want %v", tt.status, tt.code, got, tt.want)
			}
		})
	}
}
