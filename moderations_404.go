package meshapi

import (
	"errors"
	"net/http"
	"reflect"
)

// ModerationsNotFoundStatus is the HTTP status that POST /v1/moderations
// returns when the requested moderation model or resource is not found.
//
// The API documents this response as of spec version v1.0.433.
const ModerationsNotFoundStatus = http.StatusNotFound

// IsModerationsNotFound reports whether err is a 404 Not Found response
// from POST /v1/moderations.
//
// It walks the error chain with errors.Unwrap and reads the HTTP status from
// the first error that carries one, such as a *MeshAPIError returned by the
// client. It returns false for a nil error or for an error with no HTTP status.
//
//	resp, err := client.Moderations.Create(ctx, req)
//	if meshapi.IsModerationsNotFound(err) {
//		// The moderation model is not available; choose another from GET /v1/models.
//	}
func IsModerationsNotFound(err error) bool {
	return moderationsErrorStatus(err) == ModerationsNotFoundStatus
}

// moderationsErrorStatus returns the HTTP status carried by the first error in
// the chain that exposes one through a StatusCode, Status or HTTPStatus integer
// field. It returns 0 if no error in the chain carries a status.
func moderationsErrorStatus(err error) int {
	for e := err; e != nil; e = errors.Unwrap(e) {
		if code, ok := moderationsStatusFromValue(reflect.ValueOf(e)); ok {
			return code
		}
	}
	return 0
}

func moderationsStatusFromValue(v reflect.Value) (int, bool) {
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return 0, false
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return 0, false
	}
	for _, name := range []string{"StatusCode", "Status", "HTTPStatus"} {
		f := v.FieldByName(name)
		if !f.IsValid() {
			continue
		}
		switch f.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return int(f.Int()), true
		}
	}
	return 0, false
}
