package meshapi

import (
	"net/http"
	"regexp"
	"strings"
)

// ModerationModelNotFoundCode is the error code that POST /v1/moderations
// returns with HTTP 404 when the requested model is not a supported
// moderation model.
const ModerationModelNotFoundCode = "model_not_found"

// ModerationModelLatest is the moving alias for the current moderation model.
const ModerationModelLatest = "omni-moderation-latest"

// moderationModelPrefix is the optional provider prefix accepted on
// moderation model ids.
const moderationModelPrefix = "openai/"

var moderationSnapshotPattern = regexp.MustCompile(`^omni-moderation-\d{4}-\d{2}-\d{2}$`)

// IsAcceptedModerationModel reports whether model is an id that
// POST /v1/moderations accepts.
//
// Accepted ids are "omni-moderation-latest" and dated snapshots of the form
// "omni-moderation-YYYY-MM-DD", either bare or with the "openai/" prefix.
// Any other id is rejected by the API with HTTP 404 and the error code
// ModerationModelNotFoundCode. Use this helper to validate input before
// sending a request.
func IsAcceptedModerationModel(model string) bool {
	id := strings.TrimPrefix(model, moderationModelPrefix)
	if id == ModerationModelLatest {
		return true
	}
	return moderationSnapshotPattern.MatchString(id)
}

// IsModerationModelNotFound reports whether an HTTP status and error code
// returned by POST /v1/moderations mean the requested model is not a
// supported moderation model.
//
// The API returns this response before running the request, so no usage is
// recorded for it.
func IsModerationModelNotFound(status int, code string) bool {
	return status == http.StatusNotFound && code == ModerationModelNotFoundCode
}
