package meshapi

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

type fakeStatusError struct {
	status int
}

func (e *fakeStatusError) Error() string { return fmt.Sprintf("status %d", e.status) }

func (e *fakeStatusError) StatusCode() int { return e.status }

func TestModerationsNotFoundStatus(t *testing.T) {
	if ModerationsNotFoundStatus != http.StatusNotFound {
		t.Fatalf("ModerationsNotFoundStatus = %d, want %d", ModerationsNotFoundStatus, http.StatusNotFound)
	}
}

func TestIsModerationsNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil error", err: nil, want: false},
		{name: "plain error without status", err: errors.New("boom"), want: false},
		{name: "404 status", err: &fakeStatusError{status: http.StatusNotFound}, want: true},
		{name: "400 status", err: &fakeStatusError{status: http.StatusBadRequest}, want: false},
		{name: "500 status", err: &fakeStatusError{status: http.StatusInternalServerError}, want: false},
		{name: "wrapped 404", err: fmt.Errorf("moderate: %w", &fakeStatusError{status: http.StatusNotFound}), want: true},
		{name: "wrapped 401", err: fmt.Errorf("moderate: %w", &fakeStatusError{status: http.StatusUnauthorized}), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsModerationsNotFound(tt.err); got != tt.want {
				t.Errorf("IsModerationsNotFound(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
