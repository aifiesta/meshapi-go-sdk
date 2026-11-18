package meshapi

import "strings"

// ModerationModelNotFoundCode is the error code that POST /v1/moderations
// returns, with HTTP status 404, when the requested model is not an accepted
// moderation model.
const ModerationModelNotFoundCode = "model_not_found"

// moderationModelPrefix is the stem shared by every accepted moderation model id.
const moderationModelPrefix = "omni-moderation-"

// IsModerationModelAccepted reports whether model has the shape of a model id
// that POST /v1/moderations accepts.
//
// Accepted ids are omni-moderation-latest or omni-moderation-YYYY-MM-DD, either
// bare or with the openai/ prefix. Any other id is refused by the API with a
// 404 model_not_found error. Use this as a local pre-check to avoid a round
// trip. It checks the id's shape only; the API remains the source of truth.
func IsModerationModelAccepted(model string) bool {
	id := strings.TrimPrefix(model, "openai/")
	if !strings.HasPrefix(id, moderationModelPrefix) {
		return false
	}
	suffix := strings.TrimPrefix(id, moderationModelPrefix)
	if suffix == "latest" {
		return true
	}
	return isDateStamp(suffix)
}

// IsModerationModelNotFound reports whether an HTTP status and error code pair
// from POST /v1/moderations means the requested model is not accepted.
func IsModerationModelNotFound(status int, code string) bool {
	return status == 404 && code == ModerationModelNotFoundCode
}

// isDateStamp reports whether s has the YYYY-MM-DD shape.
func isDateStamp(s string) bool {
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
