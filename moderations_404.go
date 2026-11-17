package meshapi

import "strings"

// ModerationModelNotFoundStatus is the HTTP status that POST /v1/moderations
// returns when the requested model is not a moderation model the gateway
// accepts. The request fails before it reaches a provider.
const ModerationModelNotFoundStatus = 404

// ModerationModelNotFoundCode is the error code sent with a 404 from
// POST /v1/moderations when the model id is not accepted.
const ModerationModelNotFoundCode = "model_not_found"

const (
	moderationModelPrefix   = "omni-moderation-"
	moderationModelLatest   = "omni-moderation-latest"
	moderationProviderScope = "openai/"
)

// IsAcceptedModerationModel reports whether id is a model id that
// POST /v1/moderations accepts.
//
// Accepted ids are "omni-moderation-latest" and dated snapshots of the form
// "omni-moderation-YYYY-MM-DD", either bare or with the "openai/" prefix.
// Any other id is answered with a 404 and the code
// ModerationModelNotFoundCode. Use this to check an id before sending a
// request; the gateway remains the source of truth.
func IsAcceptedModerationModel(id string) bool {
	id = strings.TrimPrefix(id, moderationProviderScope)
	if id == moderationModelLatest {
		return true
	}
	if !strings.HasPrefix(id, moderationModelPrefix) {
		return false
	}
	return isDateShape(strings.TrimPrefix(id, moderationModelPrefix))
}

// isDateShape reports whether s has the YYYY-MM-DD shape (digits and dashes).
func isDateShape(s string) bool {
	if len(s) != 10 {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if i == 4 || i == 7 {
			if c != '-' {
				return false
			}
			continue
		}
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
