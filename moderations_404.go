package meshapi

import (
	"net/http"
	"strings"
	"time"
)

// ModerationModelLatest is the moving alias accepted by POST /v1/moderations.
const ModerationModelLatest = "omni-moderation-latest"

// ErrorCodeModelNotFound is the error code returned with a 404 when a request
// names a model the endpoint does not accept.
const ErrorCodeModelNotFound = "model_not_found"

const (
	moderationModelPrefix   = "omni-moderation-"
	moderationVendorPrefix  = "openai/"
	moderationSnapshotDates = "2006-01-02"
)

// IsAcceptedModerationModel reports whether model is an id that
// POST /v1/moderations accepts.
//
// Accepted ids are "omni-moderation-latest" and dated snapshots of the form
// "omni-moderation-YYYY-MM-DD", either bare or prefixed with "openai/".
// Any other id is rejected by the API with a 404 "model_not_found" error.
// Use this to validate input locally before sending a request.
func IsAcceptedModerationModel(model string) bool {
	id := strings.TrimPrefix(model, moderationVendorPrefix)
	if id == ModerationModelLatest {
		return true
	}
	if !strings.HasPrefix(id, moderationModelPrefix) {
		return false
	}
	date := strings.TrimPrefix(id, moderationModelPrefix)
	if len(date) != len(moderationSnapshotDates) {
		return false
	}
	_, err := time.Parse(moderationSnapshotDates, date)
	return err == nil
}

// IsModerationModelNotFound reports whether an HTTP status and error code
// returned by POST /v1/moderations mean the requested model is not accepted.
//
// The API answers 404 with code "model_not_found" for such requests. No usage
// is recorded for them.
func IsModerationModelNotFound(status int, code string) bool {
	return status == http.StatusNotFound && code == ErrorCodeModelNotFound
}
