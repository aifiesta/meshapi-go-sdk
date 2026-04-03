package meshapi

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

// nopReadCloser wraps a Reader with a no-op Close.
func nopReadCloser(r io.Reader) io.ReadCloser {
	return io.NopCloser(r)
}

func makeErrorResponse(status int, body string, contentType string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": {contentType}},
		Body:       nopReadCloser(strings.NewReader(body)),
	}
}

func TestNewErrorFromResponse_401(t *testing.T) {
	body := `{"error":{"code":"unauthorized","message":"Invalid or missing API key."},"request_id":"req_001"}`
	resp := makeErrorResponse(401, body, "application/json")
	err := newErrorFromResponse(resp)
	if err.Status != 401 {
		t.Errorf("expected status 401, got %d", err.Status)
	}
	if err.Code != "unauthorized" {
		t.Errorf("expected code 'unauthorized', got %q", err.Code)
	}
	if err.RequestID != "req_001" {
		t.Errorf("expected request_id 'req_001', got %q", err.RequestID)
	}
}

func TestNewErrorFromResponse_429WithRetryAfter(t *testing.T) {
	body := `{"error":{"code":"rate_limit_exceeded","message":"Rate limit exceeded.","retry_after_seconds":5},"request_id":"req_429"}`
	resp := makeErrorResponse(429, body, "application/json")
	err := newErrorFromResponse(resp)
	if err.Status != 429 {
		t.Errorf("expected status 429, got %d", err.Status)
	}
	if err.RetryAfterSeconds == nil || *err.RetryAfterSeconds != 5 {
		t.Errorf("expected retry_after_seconds=5, got %v", err.RetryAfterSeconds)
	}
}

func TestNewErrorFromResponse_422WithDetails(t *testing.T) {
	body := `{"error":{"code":"validation_error","message":"Request validation failed.","details":[{"type":"missing","loc":["body","messages"],"msg":"Field required"}]},"request_id":"req_422"}`
	resp := makeErrorResponse(422, body, "application/json")
	err := newErrorFromResponse(resp)
	if err.Code != "validation_error" {
		t.Errorf("expected code 'validation_error', got %q", err.Code)
	}
	if len(err.Details) == 0 {
		t.Error("expected non-empty details")
	}
}

func TestNewErrorFromResponse_HTMLBody(t *testing.T) {
	body := "<html><body>Bad Gateway</body></html>"
	resp := makeErrorResponse(502, body, "text/html")
	err := newErrorFromResponse(resp)
	if err.Code != "parse_error" {
		t.Errorf("expected code 'parse_error', got %q", err.Code)
	}
	if !strings.Contains(err.Message, "Bad Gateway") {
		t.Errorf("expected 'Bad Gateway' in message, got %q", err.Message)
	}
}

func TestNewErrorFromResponse_MalformedJSON(t *testing.T) {
	body := "{not json}"
	resp := makeErrorResponse(500, body, "application/json")
	err := newErrorFromResponse(resp)
	if err.Code != "parse_error" {
		t.Errorf("expected code 'parse_error', got %q", err.Code)
	}
}

func TestMeshAPIError_Error(t *testing.T) {
	err := &MeshAPIError{Status: 404, Code: "not_found", RequestID: "req_x", Message: "Not found"}
	s := err.Error()
	if !strings.Contains(s, "404") || !strings.Contains(s, "not_found") {
		t.Errorf("unexpected Error() output: %q", s)
	}
}

func TestStreamInterruptedError(t *testing.T) {
	err := newStreamInterruptedError("connection reset")
	if err.Code != "stream_interrupted" {
		t.Errorf("expected code 'stream_interrupted', got %q", err.Code)
	}
	if err.Status != 0 {
		t.Errorf("expected status 0, got %d", err.Status)
	}
}
