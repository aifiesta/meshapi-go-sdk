package meshapi

import "testing"

func TestIsSupportedModerationModel(t *testing.T) {
	tests := []struct {
		name  string
		model string
		want  bool
	}{
		{name: "latest bare", model: "omni-moderation-latest", want: true},
		{name: "latest with prefix", model: "openai/omni-moderation-latest", want: true},
		{name: "dated bare", model: "omni-moderation-2024-09-26", want: true},
		{name: "dated with prefix", model: "openai/omni-moderation-2024-09-26", want: true},
		{name: "empty", model: "", want: false},
		{name: "chat model", model: "openai/gpt-4o-mini", want: false},
		{name: "other prefix", model: "anthropic/omni-moderation-latest", want: false},
		{name: "malformed date", model: "omni-moderation-2024-9-26", want: false},
		{name: "trailing text", model: "omni-moderation-latest-x", want: false},
		{name: "leading text", model: "x-omni-moderation-latest", want: false},
		{name: "uppercase", model: "OMNI-MODERATION-LATEST", want: false},
		{name: "bare family", model: "omni-moderation", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSupportedModerationModel(tt.model); got != tt.want {
				t.Errorf("IsSupportedModerationModel(%q) = %v, want %v", tt.model, got, tt.want)
			}
		})
	}
}

func TestModerationModelNotFoundCode(t *testing.T) {
	if ModerationModelNotFoundCode != "model_not_found" {
		t.Errorf("ModerationModelNotFoundCode = %q, want %q", ModerationModelNotFoundCode, "model_not_found")
	}
}
