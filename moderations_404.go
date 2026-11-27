package meshapi

import "net/http"

// ModerationsNotFoundStatus is the HTTP status code that POST /v1/moderations
// can return when the requested resource is not found.
//
// The status was added to the documented responses of POST /v1/moderations
// in API spec v1.0.437. Callers can treat it as a non-retryable client error.
const ModerationsNotFoundStatus = http.StatusNotFound

// IsModerationsNotFound reports whether statusCode is the 404 Not Found
// status documented for POST /v1/moderations.
//
// Use it to tell a missing resource apart from other moderation failures, for
// example when inspecting the status code carried by an API error:
//
//	if meshapi.IsModerationsNotFound(status) {
//		// do not retry; check the request and try again
//	}
func IsModerationsNotFound(statusCode int) bool {
	return statusCode == ModerationsNotFoundStatus
}

// IsModerationsRetryable reports whether a POST /v1/moderations response with
// statusCode is worth retrying. A 404 Not Found is never retryable.
func IsModerationsRetryable(statusCode int) bool {
	if IsModerationsNotFound(statusCode) {
		return false
	}
	return statusCode == http.StatusTooManyRequests || statusCode >= http.StatusInternalServerError
}
