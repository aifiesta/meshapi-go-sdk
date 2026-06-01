package meshapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

// RealtimeConnectParams holds parameters for opening a realtime session.
type RealtimeConnectParams struct {
	// Model is the realtime-capable model ID, e.g. "openai/gpt-4o-realtime-preview".
	Model string
}

// RealtimeMessage is a single frame received from the server.
//
// Exactly one of Text or Audio is non-empty per message.
type RealtimeMessage struct {
	// Text is the raw JSON string for text frames.
	Text string
	// Audio is the raw bytes for binary audio frames.
	Audio []byte
	// Event is the parsed JSON map for text frames; nil for audio frames.
	Event map[string]any
}

// RealtimeError is delivered by the server inside a {"type":"error",...} frame
// before the socket is closed. It implements the error interface.
type RealtimeError struct {
	// Code is the snake_case error code, e.g. "invalid_api_key", "insufficient_quota".
	Code string `json:"code"`
	// Message is a human-readable description.
	Message string `json:"message"`
	// Param is the offending parameter, if any.
	Param string `json:"param,omitempty"`
	// RequestID is the server-assigned session request ID for log correlation.
	RequestID string `json:"request_id,omitempty"`
}

func (e *RealtimeError) Error() string {
	return fmt.Sprintf("realtime[%s]: %s", e.Code, e.Message)
}

// RealtimeSession is an active WebSocket session with the MeshAPI realtime endpoint.
//
// Send and Receive may be called concurrently from separate goroutines.
// Close must be called exactly once when the session is no longer needed.
type RealtimeSession struct {
	conn   *wsConn
	sendMu sync.Mutex // guards concurrent Sends
}

// Send marshals event as JSON and sends it to the server as a text frame.
func (s *RealtimeSession) Send(_ context.Context, event any) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("realtime: marshal event: %w", err)
	}
	s.sendMu.Lock()
	defer s.sendMu.Unlock()
	return s.conn.WriteText(data)
}

// SendAudio sends raw audio bytes to the server as a binary frame.
func (s *RealtimeSession) SendAudio(_ context.Context, audio []byte) error {
	s.sendMu.Lock()
	defer s.sendMu.Unlock()
	return s.conn.WriteBinary(audio)
}

// Receive reads the next frame from the server.
//
// Returns *RealtimeError when the server delivers an error envelope
// ({"type":"error",...}). Returns io.EOF on a clean server-initiated close.
// Context cancellation interrupts an in-progress read.
func (s *RealtimeSession) Receive(ctx context.Context) (RealtimeMessage, error) {
	// Run the blocking read in a goroutine so we can respect ctx.
	type result struct {
		f   wsFrame
		err error
	}
	ch := make(chan result, 1)
	go func() {
		f, err := s.conn.ReadFrame()
		ch <- result{f, err}
	}()

	select {
	case <-ctx.Done():
		return RealtimeMessage{}, ctx.Err()
	case r := <-ch:
		if r.err != nil {
			if isClosedErr(r.err) {
				return RealtimeMessage{}, io.EOF
			}
			return RealtimeMessage{}, r.err
		}
		return decodeFrame(r.f)
	}
}

// Events starts a goroutine that pumps server frames into the returned channels.
//
// msgCh closes when the session ends or ctx is done; errCh receives at most
// one terminal error. Drain msgCh before reading errCh.
func (s *RealtimeSession) Events(ctx context.Context) (<-chan RealtimeMessage, <-chan error) {
	msgCh := make(chan RealtimeMessage, 64)
	errCh := make(chan error, 1)
	go func() {
		defer close(msgCh)
		defer close(errCh)
		for {
			msg, err := s.Receive(ctx)
			if err != nil {
				if err != io.EOF && err != ctx.Err() {
					errCh <- err
				}
				return
			}
			select {
			case msgCh <- msg:
			case <-ctx.Done():
				return
			}
		}
	}()
	return msgCh, errCh
}

// Close closes the WebSocket connection with a normal closure.
// It is safe to call Close more than once.
func (s *RealtimeSession) Close() error {
	return s.conn.Close(wsStatusNormal, "")
}

// RealtimeResource provides access to the MeshAPI WebSocket realtime endpoint.
//
// Accessible as client.Realtime.
type RealtimeResource struct {
	http *httpClient
}

// Connect opens a WebSocket session to the realtime endpoint for the given model.
//
// Authentication is delivered via the Sec-WebSocket-Protocol header following
// the wire contract: "openai-realtime, Bearer <token>".
//
// The returned session is ready for bidirectional frame exchange immediately.
// Cancel ctx to abort the connection attempt; for an established session use
// session.Close().
func (r *RealtimeResource) Connect(_ context.Context, params RealtimeConnectParams) (*RealtimeSession, error) {
	wsURL := realtimeWSURL(r.http.cfg.BaseURL, params.Model, r.http.cfg.Token)

	headers := http.Header{}
	headers.Set("Sec-WebSocket-Protocol", "openai-realtime")
	headers.Set(sdkVersionHeader, sdkVersionValue)

	conn, err := dialWS(wsURL, headers)
	if err != nil {
		return nil, fmt.Errorf("realtime: connect: %w", err)
	}
	return &RealtimeSession{conn: conn}, nil
}

// realtimeWSURL rewrites the http/https scheme to ws/wss and appends the
// realtime path and model query parameter.
func realtimeWSURL(baseURL, model, token string) string {
	base := strings.TrimRight(baseURL, "/")
	base = strings.Replace(base, "https://", "wss://", 1)
	base = strings.Replace(base, "http://", "ws://", 1)
	return base + "/v1/realtime?model=" + url.QueryEscape(model) + "&api_key=" + url.QueryEscape(token)
}

// decodeFrame turns a raw wsFrame into a typed RealtimeMessage.
// Text frames are JSON-decoded; error envelopes return *RealtimeError.
func decodeFrame(f wsFrame) (RealtimeMessage, error) {
	if f.opcode == wsOpBinary {
		return RealtimeMessage{Audio: f.payload}, nil
	}
	if f.opcode == wsOpClose {
		return RealtimeMessage{}, io.EOF
	}
	// Text frame.
	msg := RealtimeMessage{Text: string(f.payload)}
	var evt map[string]any
	if err := json.Unmarshal(f.payload, &evt); err == nil {
		msg.Event = evt
		if evt["type"] == "error" {
			return RealtimeMessage{}, extractRealtimeError(evt)
		}
	}
	return msg, nil
}

// extractRealtimeError parses the error fields from a {"type":"error",...} envelope.
func extractRealtimeError(evt map[string]any) *RealtimeError {
	re := &RealtimeError{Code: "unknown", Message: "realtime error"}
	if errMap, ok := evt["error"].(map[string]any); ok {
		if v, ok := errMap["code"].(string); ok {
			re.Code = v
		}
		if v, ok := errMap["message"].(string); ok {
			re.Message = v
		}
		if v, ok := errMap["param"].(string); ok {
			re.Param = v
		}
	}
	if v, ok := evt["request_id"].(string); ok {
		re.RequestID = v
	}
	return re
}

// isClosedErr reports whether err indicates a connection that was closed.
func isClosedErr(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "use of closed network connection") ||
		strings.Contains(s, "EOF") ||
		strings.Contains(s, "connection reset")
}
