package meshapi

import (
	"regexp"
	"strings"
)

// ModerationModelNotFoundCode is the error code returned with HTTP 404 by
// POST /v1/moderations when the requested model is not an accepted
// moderation model.
const ModerationModelNotFoundCode = "model_not_found"

// moderationModelPattern matches the moderation model ids accepted by
// POST /v1/moderations: omni-moderation-latest or a dated
// omni-moderation-YYYY-MM-DD snapshot.
var moderationModelPattern = regexp.MustCompile(`^omni-moderation-(latest|[0-9]{4}-[0-9]{2}-[0-9]{2})$`)

// IsAcceptedModerationModel reports whether id is a model that
// POST /v1/moderations accepts.
//
// Accepted ids are omni-moderation-latest and omni-moderation-YYYY-MM-DD,
// either bare or with the openai/ prefix. Any other id is rejected by the
// API with a 404 and the model_not_found error code, and no usage is
// recorded for the request.
//
// Use this to check a model id before sending a moderation request.
func IsAcceptedModerationModel(id string) bool {
	id = strings.TrimPrefix(id, "openai/")
	return moderationModelPattern.MatchString(id)
}
