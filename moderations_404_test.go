package meshapi

import "testing"

func TestIsModerationsNotFound(t *testing.T) {
	tests := []struct {
		name   string
		status int
		want   bool
	}{
		{name: "not found", status: 404, want: true},
		{name: "ok", status: 200, want: false},
		{name: "bad request", status: 400, want: false},
		{name: "unauthorized", status: 401, want: false},
		{name: "unprocessable", status: 422, want: false},
		{name: "server error", status: 500, want: false},
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

func TestModerationsNotFoundStatusValue(t *testing.T) {
	if ModerationsNotFoundStatus != 404 {
		t.Fatalf("ModerationsNotFoundStatus = %d, want 404", ModerationsNotFoundStatus)
	}
}
