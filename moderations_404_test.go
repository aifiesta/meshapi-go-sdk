package meshapi

import "testing"

func TestIsModerationModelAccepted(t *testing.T) {
	tests := []struct {
		name  string
		model string
		want  bool
	}{
		{"bare latest", "omni-moderation-latest", true},
		{"prefixed latest", "openai/omni-moderation-latest", true},
		{"bare dated", "omni-moderation-2024-09-26", true},
		{"prefixed dated", "openai/omni-moderation-2024-09-26", true},
		{"empty", "", false},
		{"stem only", "omni-moderation-", false},
		{"chat model", "openai/gpt-4o-mini", false},
		{"other provider prefix", "anthropic/omni-moderation-latest", false},
		{"short date", "omni-moderation-2024-9-26", false},
		{"letters in date", "omni-moderation-2024-0a-26", false},
		{"wrong separator", "omni-moderation-2024/09/26", false},
		{"trailing text", "omni-moderation-latest-x", false},
		{"uppercase", "OMNI-MODERATION-LATEST", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsModerationModelAccepted(tt.model); got != tt.want {
				t.Errorf("IsModerationModelAccepted(%q) = %v, want %v", tt.model, got, tt.want)
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
		{"404 model_not_found", 404, "model_not_found", true},
		{"404 other code", 404, "not_found", false},
		{"503 model_not_found", 503, "model_not_found", false},
		{"200 empty", 200, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsModerationModelNotFound(tt.status, tt.code); got != tt.want {
				t.Errorf("IsModerationModelNotFound(%d, %q) = %v, want %v", tt.status, tt.code, got, tt.want)
			}
		})
	}
}
