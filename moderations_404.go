package meshapi

import (
	"net/http"
	"regexp"
	"strings"
)

// ModerationModelNotFoundCode is the error code returned with a 404 by
// POST /v1/moderations when the requested model is not an accepted
// moderation model.
const ModerationModelNotFoundCode = "model_not_found"

// moderationModelPattern matches the bare moderation model ids accepted by
// POST /v1/moderations: omni-moderation-latest or omni-moderation-YYYY-MM-DD.
var moderationModelPattern = regexp.MustCompile(`^omni-moderation-(latest|[0-9]{4}-[0-9]{2}-[0-9]{2})$`)

// IsValidModerationModel reports whether model has the shape accepted by
// POST /v1/moderations. Accepted ids are omni-moderation-latest or
// omni-moderation-YYYY-MM-DD, either bare or with a single "openai/" prefix.
//
// Any other id is rejected by the API with a 404 and the
// ModerationModelNotFoundCode error code. Calling this before sending a
// request lets callers fail fast without a network round trip.
func IsValidModerationModel(model string) bool {
	bare := strings.TrimPrefix(model, "openai/")
	return moderationModelPattern.MatchString(bare)
}

// IsModerationModelNotFound reports whether an HTTP status and error code
// pair from POST /v1/moderations means the requested model is not an
// accepted moderation model.
func IsModerationModelNotFound(status int, code string) bool {
	return status == http.StatusNotFound && code == ModerationModelNotFoundCode
}
