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
		{name: "ok", status: http.StatusOK, want: false},
		{name: "bad request", status: http.StatusBadRequest, want: false},
		{name: "unauthorized", status: http.StatusUnauthorized, want: false},
		{name: "not found", status: http.StatusNotFound, want: true},
		{name: "unprocessable entity", status: http.StatusUnprocessableEntity, want: false},
		{name: "too many requests", status: http.StatusTooManyRequests, want: false},
		{name: "internal server error", status: http.StatusInternalServerError, want: false},
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
