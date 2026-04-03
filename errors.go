package meshapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// MeshAPIError is returned when MeshAPI responds with a non-2xx status
// or sends an error frame mid-stream.
type MeshAPIError struct {
	// Status is the HTTP status code (0 for stream/parse errors).
	Status int
	// Code is the machine-readable error code slug (e.g. "unauthorized").
	Code string
	// RequestID is the req_<ULID> tracing identifier from the response.
	RequestID string
	// Message is the human-readable error description.
	Message string
	// Details contains validation error details (non-nil on 422 responses).
	Details []interface{}
	// ProviderError contains upstream provider error details when available.
	ProviderError map[string]interface{}
	// RetryAfterSeconds is set on 429 responses.
	RetryAfterSeconds *int
}

func (e *MeshAPIError) Error() string {
	return fmt.Sprintf("MeshAPIError(status=%d, code=%q, request_id=%q): %s",
		e.Status, e.Code, e.RequestID, e.Message)
}

// apiErrorBody is the inner error object in the response envelope.
type apiErrorBody struct {
	Code              string        `json:"code"`
	Message           string        `json:"message"`
	Details           []interface{} `json:"details"`
	ProviderError     map[string]interface{} `json:"provider_error"`
	RetryAfterSeconds *int          `json:"retry_after_seconds"`
}

// apiErrorEnvelope is the top-level error response shape.
type apiErrorEnvelope struct {
	Error     apiErrorBody `json:"error"`
	RequestID string       `json:"request_id"`
}

// newErrorFromResponse reads and parses a non-2xx HTTP response.
func newErrorFromResponse(resp *http.Response) *MeshAPIError {
	status := resp.StatusCode
	requestID := resp.Header.Get("X-Request-Id")
	contentType := resp.Header.Get("Content-Type")

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 65536))

	if !strings.Contains(contentType, "application/json") {
		msg := strings.TrimSpace(string(body))
		if len(msg) > 500 {
			msg = msg[:500]
		}
		if msg == "" {
			msg = fmt.Sprintf("HTTP %d", status)
		}
		return &MeshAPIError{
			Status:    status,
			Code:      "parse_error",
			RequestID: requestID,
			Message:   msg,
		}
	}

	var envelope apiErrorEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		return &MeshAPIError{
			Status:    status,
			Code:      "parse_error",
			RequestID: requestID,
			Message:   fmt.Sprintf("HTTP %d", status),
		}
	}

	if envelope.RequestID != "" {
		requestID = envelope.RequestID
	}

	return &MeshAPIError{
		Status:            status,
		Code:              envelope.Error.Code,
		RequestID:         requestID,
		Message:           envelope.Error.Message,
		Details:           envelope.Error.Details,
		ProviderError:     envelope.Error.ProviderError,
		RetryAfterSeconds: envelope.Error.RetryAfterSeconds,
	}
}

// newStreamInterruptedError creates an error for a mid-stream connection failure.
func newStreamInterruptedError(cause string) *MeshAPIError {
	return &MeshAPIError{
		Status:  0,
		Code:    "stream_interrupted",
		Message: fmt.Sprintf("stream interrupted: %s", cause),
	}
}
