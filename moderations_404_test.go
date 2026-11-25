package meshapi

import (
	"errors"
	"fmt"
	"testing"
)

type moderationsFakeStatusErr struct {
	StatusCode int
	Message    string
}

func (e *moderationsFakeStatusErr) Error() string { return e.Message }

type moderationsFakeAltErr struct {
	HTTPStatus int
}

func (e moderationsFakeAltErr) Error() string { return "alt" }

func TestIsModerationsNotFound(t *testing.T) {
	var nilTyped *moderationsFakeStatusErr

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil error", err: nil, want: false},
		{name: "plain error without status", err: errors.New("boom"), want: false},
		{name: "404 status", err: &moderationsFakeStatusErr{StatusCode: 404, Message: "not found"}, want: true},
		{name: "400 status", err: &moderationsFakeStatusErr{StatusCode: 400, Message: "bad request"}, want: false},
		{name: "500 status", err: &moderationsFakeStatusErr{StatusCode: 500, Message: "server"}, want: false},
		{name: "wrapped 404", err: fmt.Errorf("moderate: %w", &moderationsFakeStatusErr{StatusCode: 404}), want: true},
		{name: "double wrapped 404", err: fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", &moderationsFakeStatusErr{StatusCode: 404})), want: true},
		{name: "value type with HTTPStatus 404", err: moderationsFakeAltErr{HTTPStatus: 404}, want: true},
		{name: "typed nil pointer", err: nilTyped, want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsModerationsNotFound(tc.err); got != tc.want {
				t.Fatalf("IsModerationsNotFound(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

func TestModerationsNotFoundStatus(t *testing.T) {
	if ModerationsNotFoundStatus != 404 {
		t.Fatalf("ModerationsNotFoundStatus = %d, want 404", ModerationsNotFoundStatus)
	}
}
