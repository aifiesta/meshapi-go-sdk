package meshapi

import (
	"strings"
	"time"
)

// ModerationModelLatest is the moderation model id that always points at the
// newest omni moderation model.
const ModerationModelLatest = "omni-moderation-latest"

// moderationModelPrefix is the shared prefix of every accepted moderation model id.
const moderationModelPrefix = "omni-moderation-"

// moderationProviderPrefix is the optional provider prefix an accepted
// moderation model id may carry.
const moderationProviderPrefix = "openai/"

// IsSupportedModerationModel reports whether model is an id that
// POST /v1/moderations accepts.
//
// Accepted ids are `omni-moderation-latest` and dated snapshots of the form
// `omni-moderation-YYYY-MM-DD`, each either bare or prefixed with `openai/`.
// Any other id is rejected by the API with a 404 `model_not_found` error, so
// callers can use this check to fail fast before sending a request.
func IsSupportedModerationModel(model string) bool {
	id := strings.TrimPrefix(model, moderationProviderPrefix)
	if id == ModerationModelLatest {
		return true
	}
	if !strings.HasPrefix(id, moderationModelPrefix) {
		return false
	}
	date := strings.TrimPrefix(id, moderationModelPrefix)
	if len(date) != len("2006-01-02") {
		return false
	}
	_, err := time.Parse("2006-01-02", date)
	return err == nil
}
