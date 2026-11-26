package meshapi

import (
	"errors"
	"fmt"
	"testing"
)

type testStatusFieldError struct {
	StatusCode int
}

func (e *testStatusFieldError) Error() string { return fmt.Sprintf("status %d", e.StatusCode) }

type testStatusMethodError struct{ code int }

func (e testStatusMethodError) Error() string        { return "method status" }
func (e testStatusMethodError) HTTPStatusCode() int { return e.code }

func TestIsModerationNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil error", err: nil, want: false},
		{name: "plain error", err: errors.New("boom"), want: false},
		{name: "field 404", err: &testStatusFieldError{StatusCode: 404}, want: true},
		{name: "field 400", err: &testStatusFieldError{StatusCode: 400}, want: false},
		{name: "field 500", err: &testStatusFieldError{StatusCode: 500}, want: false},
		{name: "wrapped field 404", err: fmt.Errorf("moderate: %w", &testStatusFieldError{StatusCode: 404}), want: true},
		{name: "method 404", err: testStatusMethodError{code: 404}, want: true},
		{name: "method 401", err: testStatusMethodError{code: 401}, want: false},
		{name: "wrapped method 404", err: fmt.Errorf("outer: %w", testStatusMethodError{code: 404}), want: true},
		{name: "nil pointer", err: (*testStatusFieldError)(nil), want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsModerationNotFound(tt.err); got != tt.want {
				t.Fatalf("IsModerationNotFound(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestModerationsNotFoundStatus(t *testing.T) {
	if ModerationsNotFoundStatus != 404 {
		t.Fatalf("ModerationsNotFoundStatus = %d, want 404", ModerationsNotFoundStatus)
	}
}
