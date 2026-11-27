package meshapi

import "net/http"

// ModerationsNotFoundStatus is the HTTP status that POST /v1/moderations can
// return when the requested resource is not found.
//
// The 404 response was added to the public contract in v1.0.439. Callers that
// classify moderation failures should treat this status as a non-retryable
// client error rather than a transient upstream failure.
const ModerationsNotFoundStatus = http.StatusNotFound

// IsModerationsNotFound reports whether statusCode is the 404 Not Found status
// documented for POST /v1/moderations.
//
// Use it to separate a missing resource from retryable failures such as 429 or
// 5xx responses:
//
//	if IsModerationsNotFound(status) {
//		// Do not retry; check the request (for example the model id).
//	}
func IsModerationsNotFound(statusCode int) bool {
	return statusCode == ModerationsNotFoundStatus
}
