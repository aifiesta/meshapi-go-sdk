package meshapi

import (
	"net/http"
	"testing"
)

func TestIsValidModerationModel(t *testing.T) {
	tests := []struct {
		name  string
		model string
		want  bool
	}{
		{name: "bare latest", model: "omni-moderation-latest", want: true},
		{name: "prefixed latest", model: "openai/omni-moderation-latest", want: true},
		{name: "bare dated", model: "omni-moderation-2024-09-26", want: true},
		{name: "prefixed dated", model: "openai/omni-moderation-2024-09-26", want: true},
		{name: "empty", model: "", want: false},
		{name: "other provider prefix", model: "anthropic/omni-moderation-latest", want: false},
		{name: "double prefix", model: "openai/openai/omni-moderation-latest", want: false},
		{name: "chat model", model: "openai/gpt-4o-mini", want: false},
		{name: "malformed date", model: "omni-moderation-2024-9-26", want: false},
		{name: "trailing text", model: "omni-moderation-latest-x", want: false},
		{name: "leading text", model: "x-omni-moderation-latest", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidModerationModel(tt.model); got != tt.want {
				t.Errorf("IsValidModerationModel(%q) = %v, want %v", tt.model, got, tt.want)
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
		{name: "404 model_not_found", status: http.StatusNotFound, code: ModerationModelNotFoundCode, want: true},
		{name: "404 other code", status: http.StatusNotFound, code: "not_found", want: false},
		{name: "400 model_not_found", status: http.StatusBadRequest, code: ModerationModelNotFoundCode, want: false},
		{name: "503 empty code", status: http.StatusServiceUnavailable, code: "", want: false},
		{name: "200 empty code", status: http.StatusOK, code: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsModerationModelNotFound(tt.status, tt.code); got != tt.want {
				t.Errorf("IsModerationModelNotFound(%d, %q) = %v, want %v", tt.status, tt.code, got, tt.want)
			}
		})
	}
}
