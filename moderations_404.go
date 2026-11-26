package meshapi

import (
	"errors"
	"net/http"
)

// ErrModerationNotFound is returned when POST /v1/moderations answers 404 Not Found.
//
// The moderations endpoint can return 404 when the requested moderation
// resource is not available. Match it with errors.Is or IsModerationNotFound
// to handle it apart from other API errors.
var ErrModerationNotFound = errors.New("meshapi: moderation resource not found (404)")

// moderationStatusError wraps ErrModerationNotFound and keeps the HTTP status
// and the original error returned by the transport.
type moderationStatusError struct {
	status int
	cause  error
}

// Error implements the error interface.
func (e *moderationStatusError) Error() string {
	if e.cause != nil {
		return ErrModerationNotFound.Error() + ": " + e.cause.Error()
	}
	return ErrModerationNotFound.Error()
}

// Is reports whether target is ErrModerationNotFound.
func (e *moderationStatusError) Is(target error) bool {
	return target == ErrModerationNotFound
}

// Unwrap returns the original error, if any.
func (e *moderationStatusError) Unwrap() error {
	return e.cause
}

// classifyModerationStatus maps an HTTP status returned by POST /v1/moderations
// to a typed error. It returns nil for any status other than 404 so callers
// can fall back to their usual error handling.
func classifyModerationStatus(status int, cause error) error {
	if status != http.StatusNotFound {
		return nil
	}
	return &moderationStatusError{status: status, cause: cause}
}

// IsModerationNotFound reports whether err indicates that POST /v1/moderations
// answered 404 Not Found.
func IsModerationNotFound(err error) bool {
	return errors.Is(err, ErrModerationNotFound)
}
