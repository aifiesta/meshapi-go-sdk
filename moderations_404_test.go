package meshapi

import "testing"

func TestIsAcceptedModerationModel(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want bool
	}{
		{name: "latest bare", id: "omni-moderation-latest", want: true},
		{name: "latest with openai prefix", id: "openai/omni-moderation-latest", want: true},
		{name: "dated bare", id: "omni-moderation-2024-09-26", want: true},
		{name: "dated with openai prefix", id: "openai/omni-moderation-2024-09-26", want: true},
		{name: "empty", id: "", want: false},
		{name: "chat model", id: "openai/gpt-4o-mini", want: false},
		{name: "other provider prefix", id: "anthropic/omni-moderation-latest", want: false},
		{name: "prefix only", id: "omni-moderation-", want: false},
		{name: "short date", id: "omni-moderation-2024-9-26", want: false},
		{name: "date with letters", id: "omni-moderation-2024-0a-26", want: false},
		{name: "wrong separators", id: "omni-moderation-2024_09_26", want: false},
		{name: "trailing text", id: "omni-moderation-latest-x", want: false},
		{name: "double prefix", id: "openai/openai/omni-moderation-latest", want: false},
		{name: "case differs", id: "Omni-Moderation-Latest", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAcceptedModerationModel(tt.id); got != tt.want {
				t.Errorf("IsAcceptedModerationModel(%q) = %v, want %v", tt.id, got, tt.want)
			}
		})
	}
}

func TestModerationModelNotFoundConstants(t *testing.T) {
	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{name: "status", got: ModerationModelNotFoundStatus, want: 404},
		{name: "code", got: ModerationModelNotFoundCode, want: "model_not_found"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %v, want %v", tt.got, tt.want)
			}
		})
	}
}
