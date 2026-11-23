package meshapi

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

func TestClassifyModerationStatus(t *testing.T) {
	cause := errors.New("transport error")

	tests := []struct {
		name         string
		status       int
		message      string
		cause        error
		wantNotFound bool
		wantSame     bool
		wantMsg      string
	}{
		{name: "404 with message", status: http.StatusNotFound, message: "not found", cause: cause, wantNotFound: true, wantMsg: "meshapi: moderations returned 404 not found: not found"},
		{name: "404 without message", status: http.StatusNotFound, cause: nil, wantNotFound: true, wantMsg: "meshapi: moderations returned 404 not found"},
		{name: "400 passes through", status: http.StatusBadRequest, message: "bad", cause: cause, wantSame: true},
		{name: "500 passes through", status: http.StatusInternalServerError, cause: cause, wantSame: true},
		{name: "200 with nil cause", status: http.StatusOK, cause: nil, wantSame: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifyModerationStatus(tt.status, tt.message, tt.cause)

			if tt.wantSame {
				if got != tt.cause {
					t.Fatalf("got %v, want unchanged cause %v", got, tt.cause)
				}
				if IsModerationNotFound(got) {
					t.Fatalf("IsModerationNotFound(%v) = true, want false", got)
				}
				return
			}

			if !tt.wantNotFound {
				return
			}
			var nf *ModerationNotFoundError
			if !errors.As(got, &nf) {
				t.Fatalf("got %T, want *ModerationNotFoundError", got)
			}
			if nf.StatusCode != http.StatusNotFound {
				t.Errorf("StatusCode = %d, want %d", nf.StatusCode, http.StatusNotFound)
			}
			if nf.Error() != tt.wantMsg {
				t.Errorf("Error() = %q, want %q", nf.Error(), tt.wantMsg)
			}
			if !errors.Is(got, ErrModerationNotFound) {
				t.Errorf("errors.Is(got, ErrModerationNotFound) = false")
			}
			if tt.cause != nil && !errors.Is(got, tt.cause) {
				t.Errorf("errors.Is(got, cause) = false, want true")
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
		{name: "sentinel", err: ErrModerationNotFound, want: true},
		{name: "typed", err: &ModerationNotFoundError{StatusCode: http.StatusNotFound}, want: true},
		{name: "wrapped typed", err: fmt.Errorf("moderate: %w", &ModerationNotFoundError{StatusCode: http.StatusNotFound}), want: true},
		{name: "unrelated", err: errors.New("boom"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsModerationNotFound(tt.err); got != tt.want {
				t.Errorf("IsModerationNotFound(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestModerationNotFoundErrorNilReceiver(t *testing.T) {
	var e *ModerationNotFoundError
	if e.Error() != ErrModerationNotFound.Error() {
		t.Errorf("nil Error() = %q, want %q", e.Error(), ErrModerationNotFound.Error())
	}
	if e.Unwrap() != nil {
		t.Errorf("nil Unwrap() = %v, want nil", e.Unwrap())
	}
}
