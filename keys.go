package meshapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strings"
)

// ---------------------------------------------------------------------------
// Keys: GET /v1/keys/{key_id}
// ---------------------------------------------------------------------------

// APIKeyDetails describes an API key as returned by GET /v1/keys/{key_id}.
//
// The response never includes the key secret or its hash. Fields whose shape
// may grow over time (ModelLimits, EffectiveLimits) are kept as raw JSON so
// callers can decode them into their own types without an SDK upgrade.
type APIKeyDetails struct {
	// ID is the key identifier.
	ID string `json:"id"`
	// OrgID is the organization that owns the key.
	OrgID string `json:"org_id,omitempty"`
	// Label is the human-readable name given to the key, if any.
	Label *string `json:"label,omitempty"`
	// MaskedKey is a masked display form of the key, if available.
	MaskedKey *string `json:"masked_key,omitempty"`
	// Scopes lists the scopes granted to the key.
	Scopes []string `json:"scopes,omitempty"`
	// AllowedModels lists the model ids the key may call. Empty means no allow-list.
	AllowedModels []string `json:"allowed_models,omitempty"`
	// BlockedModels lists the model ids the key may not call.
	BlockedModels []string `json:"blocked_models,omitempty"`
	// ModelLimits holds per-model limits configured on the key.
	ModelLimits json.RawMessage `json:"model_limits,omitempty"`
	// EffectiveLimits holds the limits that actually apply to the key after
	// combining every tier (the lowest limit wins).
	EffectiveLimits json.RawMessage `json:"effective_limits,omitempty"`
}

// KeysResource provides access to the /v1/keys endpoints.
type KeysResource struct {
	http *httpClient
}

// errEmptyKeyID is returned when a key id argument is blank.
var errEmptyKeyID = errors.New("meshapi: key id must not be empty")

// keyPath builds the request path for a single key, escaping the id.
func keyPath(keyID string) (string, error) {
	if strings.TrimSpace(keyID) == "" {
		return "", errEmptyKeyID
	}
	return "/v1/keys/" + url.PathEscape(keyID), nil
}

// Get returns the details of a single API key: its organization, scopes,
// model access (allowed, blocked and per-model limits) and effective limits.
//
// Calls GET /v1/keys/{key_id}.
//
// A data-plane key (rsk_...) can read only itself; asking for any other id
// returns a 404 MeshAPIError. An admin key or user session with the
// keys:read scope can read any key it reaches; a key outside that reach
// returns a 403 MeshAPIError. An unknown id returns 404.
func (r *KeysResource) Get(ctx context.Context, keyID string) (*APIKeyDetails, error) {
	path, err := keyPath(keyID)
	if err != nil {
		return nil, err
	}
	var out APIKeyDetails
	if err := r.http.get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
