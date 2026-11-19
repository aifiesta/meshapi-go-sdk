package meshapi

import (
	"errors"
	"fmt"
	"testing"
)

type fakeStatusError struct {
	StatusCode int
	Message    string
}

func (e *fakeStatusError) Error() string { return e.Message }

func TestIsModerationsNotFound(t *testing.T) {
	notFound := &fakeStatusError{StatusCode: 404, Message: "not found"}
	badRequest := &fakeStatusError{StatusCode: 400, Message: "bad request"}

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil error", err: nil, want: false},
		{name: "plain error", err: errors.New("boom"), want: false},
		{name: "direct 404", err: notFound, want: true},
		{name: "direct 400", err: badRequest, want: false},
		{name: "wrapped 404", err: fmt.Errorf("moderate: %w", notFound), want: true},
		{name: "double wrapped 404", err: fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", notFound)), want: true},
		{name: "wrapped 400", err: fmt.Errorf("moderate: %w", badRequest), want: false},
		{name: "joined with 404", err: errors.Join(errors.New("other"), notFound), want: true},
		{name: "joined without 404", err: errors.Join(errors.New("other"), badRequest), want: false},
		{name: "typed nil pointer", err: (*fakeStatusError)(nil), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsModerationsNotFound(tt.err); got != tt.want {
				t.Fatalf("IsModerationsNotFound(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestModerationsNotFoundStatus(t *testing.T) {
	if ModerationsNotFoundStatus != 404 {
		t.Fatalf("ModerationsNotFoundStatus = %d, want 404", ModerationsNotFoundStatus)
	}
}
