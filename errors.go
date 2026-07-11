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
	Code              string                 `json:"code"`
	Message           string                 `json:"message"`
	Details           []interface{}          `json:"details"`
	ProviderError     map[string]interface{} `json:"provider_error"`
	RetryAfterSeconds *int                   `json:"retry_after_seconds"`
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

	// Fall back to a status-derived code and a FastAPI-style "detail" message
	// when the body isn't the standard {"error": {...}} envelope (e.g.
	// GET /v1/models/{id} 404s return {"detail": "..."}), so callers never get
	// an empty Code/Message.
	code := envelope.Error.Code
	if code == "" {
		code = statusErrorCode(status)
	}
	message := envelope.Error.Message
	if message == "" {
		var alt struct {
			Detail interface{} `json:"detail"`
		}
		if json.Unmarshal(body, &alt) == nil {
			if s, ok := alt.Detail.(string); ok && s != "" {
				message = s
			}
		}
		if message == "" {
			message = fmt.Sprintf("HTTP %d", status)
		}
	}

	return &MeshAPIError{
		Status:            status,
		Code:              code,
		RequestID:         requestID,
		Message:           message,
		Details:           envelope.Error.Details,
		ProviderError:     envelope.Error.ProviderError,
		RetryAfterSeconds: envelope.Error.RetryAfterSeconds,
	}
}

// statusErrorCode maps an HTTP status to a machine-readable code slug, used
// when the response body doesn't carry one.
func statusErrorCode(status int) string {
	switch status {
	case 400:
		return "invalid_request"
	case 401:
		return "unauthorized"
	case 402:
		return "spend_limit_exceeded"
	case 403:
		return "forbidden"
	case 404:
		return "not_found"
	case 409:
		return "conflict"
	case 422:
		return "validation_error"
	case 429:
		return "rate_limit_exceeded"
	case 500, 502, 503, 504:
		return "upstream_error"
	default:
		return "http_error"
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

// StructuredOutputError is returned by Parse when the model's response cannot be
// decoded into the requested type.
//
// The most common cause is that the model does not support structured outputs
// (response_format): the gateway forwards the field, the provider ignores it,
// and the model returns plain text instead of JSON. The underlying json error
// is available via Unwrap (errors.Unwrap / errors.As).
type StructuredOutputError struct {
	// Model is the requested model, when known.
	Model string
	// Cause is the underlying *json.SyntaxError / *json.UnmarshalTypeError.
	Cause error
	// Message is the human-readable description (with a Models-page pointer).
	Message string
}

func (e *StructuredOutputError) Error() string { return e.Message }

func (e *StructuredOutputError) Unwrap() error { return e.Cause }

func newStructuredOutputError(model string, notJSON bool, cause error) *StructuredOutputError {
	where := ""
	if model != "" {
		where = fmt.Sprintf(" from model %q", model)
	}
	var msg string
	if notJSON {
		msg = fmt.Sprintf(
			"could not parse a structured response%s: the model returned text that is not valid JSON, "+
				"which usually means it does not support structured outputs (response_format). Check the "+
				"model's support on the Models page (%s) or the supports_structured_output flag from "+
				"GET /v1/models, and prefer a model with first-class support (e.g. openai/* or google/gemini-*). "+
				"Original error: %v",
			where, modelsURL, cause,
		)
	} else {
		msg = fmt.Sprintf(
			"could not parse a structured response%s: the response was valid JSON but did not match the "+
				"requested type. Retry with a higher WithMaxRetries, or confirm the model supports structured "+
				"outputs on the Models page (%s). Original error: %v",
			where, modelsURL, cause,
		)
	}
	return &StructuredOutputError{Model: model, Cause: cause, Message: msg}
}
