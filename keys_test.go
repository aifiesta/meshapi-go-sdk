package meshapi

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestKeyPath(t *testing.T) {
	tests := []struct {
		name    string
		keyID   string
		want    string
		wantErr bool
	}{
		{name: "simple id", keyID: "key_123", want: "/v1/keys/key_123"},
		{name: "uuid id", keyID: "3f2a9c1e-0000-4000-8000-000000000001", want: "/v1/keys/3f2a9c1e-0000-4000-8000-000000000001"},
		{name: "slash is escaped", keyID: "a/b", want: "/v1/keys/a%2Fb"},
		{name: "empty id", keyID: "", wantErr: true},
		{name: "whitespace id", keyID: "   ", wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := keyPath(tc.keyID)
			if tc.wantErr {
				if !errors.Is(err, errEmptyKeyID) {
					t.Fatalf("expected errEmptyKeyID, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("keyPath(%q) = %q, want %q", tc.keyID, got, tc.want)
			}
		})
	}
}

func TestKeysGet_EmptyIDFailsBeforeRequest(t *testing.T) {
	r := &KeysResource{}
	for _, id := range []string{"", " "} {
		if _, err := r.Get(context.Background(), id); !errors.Is(err, errEmptyKeyID) {
			t.Fatalf("Get(%q): expected errEmptyKeyID, got %v", id, err)
		}
	}
}

func TestAPIKeyDetailsUnmarshal(t *testing.T) {
	tests := []struct {
		name  string
		body  string
		check func(t *testing.T, d APIKeyDetails)
	}{
		{
			name: "full payload",
			body: `{"id":"key_1","org_id":"org_1","label":"ci","masked_key":"rsk_...abcd","scopes":["inference"],"allowed_models":["openai/gpt-4o-mini"],"blocked_models":["anthropic/claude-haiku-4.5"],"model_limits":{"openai/gpt-4o-mini":{"rpm":10}},"effective_limits":{"rpm":10}}`,
			check: func(t *testing.T, d APIKeyDetails) {
				if d.ID != "key_1" || d.OrgID != "org_1" {
					t.Fatalf("unexpected ids: %+v", d)
				}
				if d.Label == nil || *d.Label != "ci" {
					t.Fatalf("label not decoded: %v", d.Label)
				}
				if d.MaskedKey == nil || *d.MaskedKey != "rsk_...abcd" {
					t.Fatalf("masked_key not decoded: %v", d.MaskedKey)
				}
				if len(d.Scopes) != 1 || d.Scopes[0] != "inference" {
					t.Fatalf("scopes not decoded: %v", d.Scopes)
				}
				if len(d.AllowedModels) != 1 || len(d.BlockedModels) != 1 {
					t.Fatalf("model access not decoded: %+v", d)
				}
				if len(d.ModelLimits) == 0 || len(d.EffectiveLimits) == 0 {
					t.Fatalf("limits not decoded: %+v", d)
				}
			},
		},
		{
			name: "null label and masked key",
			body: `{"id":"key_2","label":null,"masked_key":null}`,
			check: func(t *testing.T, d APIKeyDetails) {
				if d.Label != nil || d.MaskedKey != nil {
					t.Fatalf("expected nil label and masked key, got %+v", d)
				}
			},
		},
		{
			name: "unknown fields are ignored",
			body: `{"id":"key_3","future_field":true}`,
			check: func(t *testing.T, d APIKeyDetails) {
				if d.ID != "key_3" {
					t.Fatalf("unexpected id %q", d.ID)
				}
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var d APIKeyDetails
			if err := json.Unmarshal([]byte(tc.body), &d); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			tc.check(t, d)
		})
	}
}
