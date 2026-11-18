package meshapi

import (
	"regexp"
	"strings"
)

// ModerationModelLatest is the moderation model id that always points at the
// current omni moderation model.
const ModerationModelLatest = "omni-moderation-latest"

// ErrCodeModelNotFound is the error code returned with an HTTP 404 when
// POST /v1/moderations is called with a model id it does not accept.
const ErrCodeModelNotFound = "model_not_found"

// moderationSnapshotPattern matches dated snapshots such as
// "omni-moderation-2024-09-26".
var moderationSnapshotPattern = regexp.MustCompile(`^omni-moderation-\d{4}-\d{2}-\d{2}$`)

// IsSupportedModerationModel reports whether model is an id that
// POST /v1/moderations accepts.
//
// Accepted ids are "omni-moderation-latest" and dated snapshots of the form
// "omni-moderation-YYYY-MM-DD", either bare or with an "openai/" prefix.
// An empty model is treated as supported because the server applies its
// default. Any other id is rejected by the API with a 404 whose error code is
// ErrCodeModelNotFound, so callers can use this check to fail fast before
// sending a request.
func IsSupportedModerationModel(model string) bool {
	if model == "" {
		return true
	}
	id := strings.TrimPrefix(model, "openai/")
	if id == ModerationModelLatest {
		return true
	}
	return moderationSnapshotPattern.MatchString(id)
}

// IsModerationModelNotFound reports whether an HTTP status and error code pair
// describe the 404 returned by POST /v1/moderations for an unsupported model.
func IsModerationModelNotFound(status int, code string) bool {
	return status == 404 && code == ErrCodeModelNotFound
}
