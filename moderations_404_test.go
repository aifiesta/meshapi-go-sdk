package meshapi

import (
	"errors"
	"fmt"
	"testing"
)

type fakeFieldErr struct {
	StatusCode int
	Message    string
}

func (e *fakeFieldErr) Error() string { return e.Message }

type fakeStatusFieldErr struct {
	Status int
}

func (e fakeStatusFieldErr) Error() string { return fmt.Sprintf("status %d", e.Status) }

type fakeMethodErr struct {
	code int
}

func (e fakeMethodErr) Error() string   { return "method error" }
func (e fakeMethodErr) StatusCode() int { return e.code }

func TestIsModerationNotFound(t *testing.T) {
	var nilPtr *fakeFieldErr

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil error", err: nil, want: false},
		{name: "plain error", err: errors.New("boom"), want: false},
		{name: "field 404", err: &fakeFieldErr{StatusCode: 404, Message: "model not found"}, want: true},
		{name: "field 400", err: &fakeFieldErr{StatusCode: 400}, want: false},
		{name: "status field 404", err: fakeStatusFieldErr{Status: 404}, want: true},
		{name: "method 404", err: fakeMethodErr{code: 404}, want: true},
		{name: "method 500", err: fakeMethodErr{code: 500}, want: false},
		{name: "wrapped 404", err: fmt.Errorf("moderate: %w", &fakeFieldErr{StatusCode: 404}), want: true},
		{name: "double wrapped 404", err: fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", fakeMethodErr{code: 404})), want: true},
		{name: "wrapped 429", err: fmt.Errorf("moderate: %w", fakeMethodErr{code: 429}), want: false},
		{name: "typed nil pointer", err: nilPtr, want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsModerationNotFound(tc.err); got != tc.want {
				t.Fatalf("IsModerationNotFound(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

func TestModerationsNotFoundStatus(t *testing.T) {
	if ModerationsNotFoundStatus != 404 {
		t.Fatalf("ModerationsNotFoundStatus = %d, want 404", ModerationsNotFoundStatus)
	}
}
