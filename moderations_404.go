package meshapi

import (
	"errors"
	"net/http"
)

// ModerationsNotFoundStatus is the HTTP status that POST /v1/moderations
// returns when the requested moderation resource is not available.
//
// As of API v1.0.438, POST /v1/moderations documents a 404 response.
const ModerationsNotFoundStatus = http.StatusNotFound

// statusCoder is implemented by errors that carry an HTTP status code.
type statusCoder interface {
	StatusCode() int
}

// IsModerationsNotFound reports whether err is a 404 response from
// POST /v1/moderations.
//
// It unwraps err and returns true when any error in the chain reports
// ModerationsNotFoundStatus through a StatusCode() int method. It returns
// false for a nil error or for errors that carry no status code.
func IsModerationsNotFound(err error) bool {
	if err == nil {
		return false
	}
	var sc statusCoder
	if errors.As(err, &sc) {
		return sc.StatusCode() == ModerationsNotFoundStatus
	}
	return false
}
