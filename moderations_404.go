package meshapi

import (
	"errors"
	"net/http"
	"reflect"
)

// ModerationsNotFoundStatus is the HTTP status that POST /v1/moderations
// returns when the requested resource, such as the moderation model, is not
// found.
const ModerationsNotFoundStatus = http.StatusNotFound

// IsModerationNotFound reports whether err is a 404 Not Found response from
// POST /v1/moderations.
//
// It walks the wrapped error chain, so errors wrapped with fmt.Errorf("%w")
// are recognised. A nil error returns false.
//
// Example:
//
//	resp, err := client.Moderations.Create(ctx, req)
//	if meshapi.IsModerationNotFound(err) {
//		// The requested model is not in the catalog. Check GET /v1/models.
//	}
func IsModerationNotFound(err error) bool {
	status, ok := moderationErrorStatus(err)
	return ok && status == ModerationsNotFoundStatus
}

// moderationErrorStatus extracts an HTTP status code from err or any error it
// wraps. It accepts errors that expose a StatusCode() int method, or structs
// carrying an integer StatusCode, Status or HTTPStatus field.
func moderationErrorStatus(err error) (int, bool) {
	for e := err; e != nil; e = errors.Unwrap(e) {
		if s, ok := e.(interface{ StatusCode() int }); ok {
			return s.StatusCode(), true
		}
		v := reflect.ValueOf(e)
		for v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
			if v.IsNil() {
				break
			}
			v = v.Elem()
		}
		if v.Kind() != reflect.Struct {
			continue
		}
		for _, name := range [...]string{"StatusCode", "Status", "HTTPStatus"} {
			f := v.FieldByName(name)
			if f.IsValid() && f.CanInt() {
				return int(f.Int()), true
			}
		}
	}
	return 0, false
}
