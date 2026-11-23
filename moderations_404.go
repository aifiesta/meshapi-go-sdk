package meshapi

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrModerationNotFound is matched by errors.Is when POST /v1/moderations
// responds with HTTP 404 Not Found.
//
// Use it to tell a 404 from the moderations endpoint apart from other
// failures:
//
//	if errors.Is(err, meshapi.ErrModerationNotFound) {
//		// handle the 404 response
//	}
var ErrModerationNotFound = errors.New("meshapi: moderations returned 404 not found")

// ModerationNotFoundError describes a 404 Not Found response from
// POST /v1/moderations.
//
// It matches ErrModerationNotFound through errors.Is and exposes the
// underlying transport error, if any, through errors.Unwrap.
type ModerationNotFoundError struct {
	// StatusCode is always http.StatusNotFound.
	StatusCode int
	// Message is the error message returned by the API, if any.
	Message string
	// Err is the underlying error, if any.
	Err error
}

// Error implements the error interface.
func (e *ModerationNotFoundError) Error() string {
	if e == nil {
		return ErrModerationNotFound.Error()
	}
	if e.Message != "" {
		return fmt.Sprintf("meshapi: moderations returned %d not found: %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("meshapi: moderations returned %d not found", e.StatusCode)
}

// Unwrap returns the underlying error, if any.
func (e *ModerationNotFoundError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// Is reports whether target is ErrModerationNotFound.
func (e *ModerationNotFoundError) Is(target error) bool {
	return target == ErrModerationNotFound
}

// classifyModerationStatus returns a *ModerationNotFoundError when status is
// 404 and cause unchanged otherwise. It is used by the moderations resource to
// surface the documented 404 response of POST /v1/moderations.
func classifyModerationStatus(status int, message string, cause error) error {
	if status != http.StatusNotFound {
		return cause
	}
	return &ModerationNotFoundError{
		StatusCode: status,
		Message:    message,
		Err:        cause,
	}
}

// IsModerationNotFound reports whether err is, or wraps, a 404 response from
// POST /v1/moderations.
func IsModerationNotFound(err error) bool {
	return errors.Is(err, ErrModerationNotFound)
}
