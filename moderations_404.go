package meshapi

import "net/http"

// ModerationsNotFoundStatus is the HTTP status that POST /v1/moderations
// returns when the requested moderation resource is not available.
//
// As of spec v1.0.440, POST /v1/moderations documents a 404 response in
// addition to its existing statuses. Callers should treat a 404 from this
// endpoint as a non-retryable "not found" result rather than a transport
// failure.
const ModerationsNotFoundStatus = http.StatusNotFound

// IsModerationsNotFound reports whether an HTTP status returned by
// POST /v1/moderations is the documented 404 response.
//
// Use it with the status code carried on a MeshAPIError returned from a
// moderation call:
//
//	if meshapi.IsModerationsNotFound(status) {
//		// The moderation resource is not available; do not retry.
//	}
func IsModerationsNotFound(status int) bool {
	return status == ModerationsNotFoundStatus
}
