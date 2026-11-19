package meshapi

import (
	"errors"
	"net/http"
	"reflect"
)

// ModerationsNotFoundStatus is the HTTP status POST /v1/moderations returns
// when the requested resource (for example the moderation model) cannot be
// found. The API documents this response as of spec v1.0.429.
const ModerationsNotFoundStatus = http.StatusNotFound

// IsModerationsNotFound reports whether err, or any error it wraps, carries
// the 404 status that POST /v1/moderations can return.
//
// It inspects the error chain (including errors joined with errors.Join) for
// a value exposing an integer StatusCode field, such as the SDK's API error
// type. A nil error always returns false.
//
//	res, err := client.Moderations.Create(ctx, req)
//	if meshapi.IsModerationsNotFound(err) {
//	    // choose another model id from GET /v1/models and retry
//	}
func IsModerationsNotFound(err error) bool {
	return statusInChain(err, ModerationsNotFoundStatus)
}

// statusInChain walks the error chain looking for an integer StatusCode
// field equal to want.
func statusInChain(err error, want int) bool {
	if err == nil {
		return false
	}
	if code, ok := statusCodeOf(err); ok && code == want {
		return true
	}
	switch u := err.(type) {
	case interface{ Unwrap() []error }:
		for _, e := range u.Unwrap() {
			if statusInChain(e, want) {
				return true
			}
		}
		return false
	default:
		return statusInChain(errors.Unwrap(err), want)
	}
}

// statusCodeOf returns the value of an integer StatusCode field on err, if
// one exists.
func statusCodeOf(err error) (int, bool) {
	v := reflect.ValueOf(err)
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return 0, false
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return 0, false
	}
	f := v.FieldByName("StatusCode")
	if !f.IsValid() {
		return 0, false
	}
	switch f.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return int(f.Int()), true
	default:
		return 0, false
	}
}
