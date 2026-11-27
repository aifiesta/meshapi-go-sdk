package meshapi

import (
	"net/http"
	"testing"
)

func TestModerationsNotFoundStatus(t *testing.T) {
	if ModerationsNotFoundStatus != 404 {
		t.Fatalf("ModerationsNotFoundStatus = %d, want 404", ModerationsNotFoundStatus)
	}
}

func TestIsModerationsNotFound(t *testing.T) {
	tests := []struct {
		name   string
		status int
		want   bool
	}{
		{name: "not found", status: http.StatusNotFound, want: true},
		{name: "ok", status: http.StatusOK, want: false},
		{name: "bad request", status: http.StatusBadRequest, want: false},
		{name: "unauthorized", status: http.StatusUnauthorized, want: false},
		{name: "unprocessable", status: http.StatusUnprocessableEntity, want: false},
		{name: "server error", status: http.StatusInternalServerError, want: false},
		{name: "zero", status: 0, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsModerationsNotFound(tt.status); got != tt.want {
				t.Errorf("IsModerationsNotFound(%d) = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}

func TestIsModerationsRetryable(t *testing.T) {
	tests := []struct {
		name   string
		status int
		want   bool
	}{
		{name: "not found is not retryable", status: http.StatusNotFound, want: false},
		{name: "bad request is not retryable", status: http.StatusBadRequest, want: false},
		{name: "rate limited is retryable", status: http.StatusTooManyRequests, want: true},
		{name: "server error is retryable", status: http.StatusInternalServerError, want: true},
		{name: "service unavailable is retryable", status: http.StatusServiceUnavailable, want: true},
		{name: "ok is not retryable", status: http.StatusOK, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsModerationsRetryable(tt.status); got != tt.want {
				t.Errorf("IsModerationsRetryable(%d) = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}
