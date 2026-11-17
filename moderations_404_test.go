package meshapi

import "testing"

func TestIsSupportedModerationModel(t *testing.T) {
	tests := []struct {
		name  string
		model string
		want  bool
	}{
		{name: "latest bare", model: "omni-moderation-latest", want: true},
		{name: "latest with provider prefix", model: "openai/omni-moderation-latest", want: true},
		{name: "dated bare", model: "omni-moderation-2024-09-26", want: true},
		{name: "dated with provider prefix", model: "openai/omni-moderation-2024-09-26", want: true},
		{name: "constant", model: ModerationModelLatest, want: true},
		{name: "empty", model: "", want: false},
		{name: "unrelated model", model: "openai/gpt-4o-mini", want: false},
		{name: "other provider prefix", model: "anthropic/omni-moderation-latest", want: false},
		{name: "double provider prefix", model: "openai/openai/omni-moderation-latest", want: false},
		{name: "prefix only", model: "omni-moderation-", want: false},
		{name: "invalid date", model: "omni-moderation-2024-13-40", want: false},
		{name: "short date", model: "omni-moderation-2024-9-26", want: false},
		{name: "trailing text", model: "omni-moderation-latest-x", want: false},
		{name: "wrong case", model: "Omni-Moderation-Latest", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSupportedModerationModel(tt.model); got != tt.want {
				t.Errorf("IsSupportedModerationModel(%q) = %v, want %v", tt.model, got, tt.want)
			}
		})
	}
}
