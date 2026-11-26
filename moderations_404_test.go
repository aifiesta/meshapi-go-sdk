package meshapi

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

func TestClassifyModerationStatus(t *testing.T) {
	cause := errors.New("upstream said not found")

	tests := []struct {
		name         string
		status       int
		cause        error
		wantNil      bool
		wantNotFound bool
		wantCause    bool
	}{
		{name: "200 is not classified", status: http.StatusOK, wantNil: true},
		{name: "400 is not classified", status: http.StatusBadRequest, wantNil: true},
		{name: "401 is not classified", status: http.StatusUnauthorized, wantNil: true},
		{name: "500 is not classified", status: http.StatusInternalServerError, wantNil: true},
		{name: "404 without cause", status: http.StatusNotFound, wantNotFound: true},
		{name: "404 with cause", status: http.StatusNotFound, cause: cause, wantNotFound: true, wantCause: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := classifyModerationStatus(tt.status, tt.cause)
			if tt.wantNil {
				if err != nil {
					t.Fatalf("expected nil, got %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("expected an error, got nil")
			}
			if got := IsModerationNotFound(err); got != tt.wantNotFound {
				t.Errorf("IsModerationNotFound = %v, want %v", got, tt.wantNotFound)
			}
			if got := errors.Is(err, cause); got != tt.wantCause {
				t.Errorf("errors.Is(err, cause) = %v, want %v", got, tt.wantCause)
			}
		})
	}
}

func TestIsModerationNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil", err: nil, want: false},
		{name: "unrelated error", err: errors.New("boom"), want: false},
		{name: "sentinel", err: ErrModerationNotFound, want: true},
		{name: "wrapped sentinel", err: fmt.Errorf("classify: %w", ErrModerationNotFound), want: true},
		{name: "classified 404", err: classifyModerationStatus(http.StatusNotFound, nil), want: true},
		{name: "wrapped classified 404", err: fmt.Errorf("call: %w", classifyModerationStatus(http.StatusNotFound, errors.New("x"))), want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsModerationNotFound(tt.err); got != tt.want {
				t.Errorf("IsModerationNotFound(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestModerationStatusErrorMessage(t *testing.T) {
	tests := []struct {
		name  string
		cause error
		want  string
	}{
		{name: "no cause", cause: nil, want: ErrModerationNotFound.Error()},
		{name: "with cause", cause: errors.New("detail"), want: ErrModerationNotFound.Error() + ": detail"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := classifyModerationStatus(http.StatusNotFound, tt.cause)
			if got := err.Error(); got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}
