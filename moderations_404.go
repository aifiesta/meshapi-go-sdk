package meshapi

import (
	"errors"
	"net/http"
	"reflect"
)

// ModerationsNotFoundStatus is the HTTP status that POST /v1/moderations
// returns when the requested resource cannot be found.
//
// The API documents a 404 response for POST /v1/moderations. Callers can
// use IsModerationNotFound to tell this case apart from other failures
// without parsing the error message.
const ModerationsNotFoundStatus = http.StatusNotFound

// httpStatusCoder is implemented by errors that expose their HTTP status
// through a method.
type httpStatusCoder interface {
	HTTPStatusCode() int
}

// IsModerationNotFound reports whether err, or any error it wraps, carries
// the 404 status that POST /v1/moderations returns when the requested
// resource cannot be found.
//
// It returns false for a nil error and for errors that carry no HTTP status.
//
//	res, err := client.Moderations.Create(ctx, req)
//	if meshapi.IsModerationNotFound(err) {
//		// pick another model id from GET /v1/models and retry
//	}
func IsModerationNotFound(err error) bool {
	status, ok := httpStatusOf(err)
	return ok && status == ModerationsNotFoundStatus
}

// httpStatusOf walks the error chain and returns the first HTTP status it
// finds, either from an HTTPStatusCode method or from an int field named
// StatusCode or Status.
func httpStatusOf(err error) (int, bool) {
	for e := err; e != nil; e = errors.Unwrap(e) {
		if sc, ok := e.(httpStatusCoder); ok {
			return sc.HTTPStatusCode(), true
		}
		if status, ok := statusField(e); ok {
			return status, true
		}
	}
	return 0, false
}

// statusField reads an integer StatusCode or Status field from a struct or
// pointer-to-struct error value.
func statusField(e error) (int, bool) {
	v := reflect.ValueOf(e)
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return 0, false
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return 0, false
	}
	for _, name := range []string{"StatusCode", "Status"} {
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
