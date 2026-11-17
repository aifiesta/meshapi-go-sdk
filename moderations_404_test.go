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
		{name: "latest with vendor prefix", model: "openai/omni-moderation-latest", want: true},
		{name: "dated snapshot bare", model: "omni-moderation-2024-09-26", want: true},
		{name: "dated snapshot with vendor prefix", model: "openai/omni-moderation-2024-09-26", want: true},
		{name: "empty", model: "", want: false},
		{name: "arbitrary string", model: "not-a-model", want: false},
		{name: "other vendor prefix", model: "anthropic/omni-moderation-latest", want: false},
		{name: "invalid date", model: "omni-moderation-2024-13-40", want: false},
		{name: "short date", model: "omni-moderation-2024-9-26", want: false},
		{name: "prefix only", model: "omni-moderation-", want: false},
		{name: "trailing garbage", model: "omni-moderation-latest-x", want: false},
		{name: "double vendor prefix", model: "openai/openai/omni-moderation-latest", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAcceptedModerationModel(tt.model); got != tt.want {
				t.Errorf("IsAcceptedModerationModel(%q) = %v, want %v", tt.model, got, tt.want)
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
				t.Errorf("IsModerationModelNotFound(%d, %q) = %v, want %v", tt.status, tt.code, got, tt.want)
			}
		})
	}
}
