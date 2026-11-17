package meshapi

import "regexp"

// ModerationModelNotFoundCode is the error code that POST /v1/moderations
// returns with HTTP 404 when the requested model id is not an accepted
// moderation model.
const ModerationModelNotFoundCode = "model_not_found"

// moderationModelPattern matches the moderation model ids accepted by
// POST /v1/moderations: "omni-moderation-latest" or a dated snapshot
// "omni-moderation-YYYY-MM-DD", either bare or with the "openai/" prefix.
var moderationModelPattern = regexp.MustCompile(`^(openai/)?omni-moderation-(latest|[0-9]{4}-[0-9]{2}-[0-9]{2})$`)

// IsSupportedModerationModel reports whether model is an id that
// POST /v1/moderations accepts.
//
// The API answers any other model id with HTTP 404 and the error code
// ModerationModelNotFoundCode. Call this before sending a request to catch
// an unsupported id without a network round trip. An empty model is
// reported as unsupported; omit the field instead if you want the server
// default.
func IsSupportedModerationModel(model string) bool {
	return moderationModelPattern.MatchString(model)
}
