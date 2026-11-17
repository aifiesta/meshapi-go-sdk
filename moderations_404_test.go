package meshapi

import "testing"

func TestIsAcceptedModerationModel(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want bool
	}{
		{name: "latest bare", id: "omni-moderation-latest", want: true},
		{name: "latest prefixed", id: "openai/omni-moderation-latest", want: true},
		{name: "dated bare", id: "omni-moderation-2024-09-26", want: true},
		{name: "dated prefixed", id: "openai/omni-moderation-2024-09-26", want: true},
		{name: "empty", id: "", want: false},
		{name: "prefix only", id: "openai/", want: false},
		{name: "chat model", id: "openai/gpt-4o-mini", want: false},
		{name: "other provider prefix", id: "anthropic/omni-moderation-latest", want: false},
		{name: "double prefix", id: "openai/openai/omni-moderation-latest", want: false},
		{name: "malformed date", id: "omni-moderation-2024-9-26", want: false},
		{name: "trailing text", id: "omni-moderation-latest-x", want: false},
		{name: "wrong case", id: "Omni-Moderation-Latest", want: false},
		{name: "junk", id: "../../etc", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAcceptedModerationModel(tt.id); got != tt.want {
				t.Errorf("IsAcceptedModerationModel(%q) = %v, want %v", tt.id, got, tt.want)
			}
		})
	}
}

func TestModerationModelNotFoundCode(t *testing.T) {
	if ModerationModelNotFoundCode != "model_not_found" {
		t.Errorf("ModerationModelNotFoundCode = %q, want %q", ModerationModelNotFoundCode, "model_not_found")
	}
}
