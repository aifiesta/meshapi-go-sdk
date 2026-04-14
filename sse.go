package meshapi

import (
	"bufio"
	"encoding/json"
	"net/http"
	"strings"
)

// parseSSEFrameData extracts the raw JSON payload from a single SSE frame.
// Returns (data, false, nil) for a data frame,
//         (nil, true, nil)   for [DONE],
//         (nil, false, err)  for an error frame,
//         (nil, false, nil)  for empty/comment frames.
func parseSSEFrameData(frame string) ([]byte, bool, error) {
	var dataLine string
	for _, line := range strings.Split(frame, "\n") {
		if strings.HasPrefix(line, "data: ") {
			dataLine = strings.TrimPrefix(line, "data: ")
		}
	}
	dataLine = strings.TrimSpace(dataLine)
	if dataLine == "" {
		return nil, false, nil
	}
	if dataLine == "[DONE]" {
		return nil, true, nil
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(dataLine), &raw); err != nil {
		// Malformed frame — skip silently
		return nil, false, nil
	}

	// Check for error frame: {"error": {...}}
	if errRaw, ok := raw["error"]; ok {
		var errBody struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}
		_ = json.Unmarshal(errRaw, &errBody)
		code := errBody.Code
		if code == "" {
			code = "upstream_error"
		}
		return nil, false, &MeshAPIError{
			Status:  0,
			Code:    code,
			Message: errBody.Message,
		}
	}

	return []byte(dataLine), false, nil
}

// tryParseSSEFrame parses a single SSE frame into a ChatCompletionChunk.
// Returns (chunk, false, nil) for a data chunk,
//         (nil, true, nil)    for [DONE],
//         (nil, false, err)   for an error frame,
//         (nil, false, nil)   for empty/comment frames.
func tryParseSSEFrame(frame string) (*ChatCompletionChunk, bool, error) {
	data, done, err := parseSSEFrameData(frame)
	if err != nil || done || data == nil {
		return nil, done, err
	}
	var chunk ChatCompletionChunk
	if err := json.Unmarshal(data, &chunk); err != nil {
		return nil, false, nil
	}
	return &chunk, false, nil
}

// parseSSEStreamOf reads an SSE response and sends parsed T chunks to chunkCh
// and any error to errCh. Both channels are always closed when the stream ends.
//
// Streaming errors detected in SSE frames (mid-stream errors) are sent on errCh.
// Connection failures are wrapped as MeshAPIError with code "stream_interrupted".
func parseSSEStreamOf[T any](resp *http.Response, chunkCh chan<- T, errCh chan<- error) {
	defer resp.Body.Close()
	defer close(chunkCh)
	defer close(errCh)

	scanner := bufio.NewScanner(resp.Body)
	var remainder strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		remainder.WriteString(line)
		remainder.WriteByte('\n')

		// SSE frames are delimited by blank lines
		if line != "" {
			continue
		}

		frame := remainder.String()
		remainder.Reset()

		data, done, err := parseSSEFrameData(frame)
		if err != nil {
			errCh <- err
			return
		}
		if done {
			return
		}
		if data == nil {
			continue
		}
		var chunk T
		if err := json.Unmarshal(data, &chunk); err != nil {
			continue
		}
		chunkCh <- chunk
	}

	if err := scanner.Err(); err != nil {
		errCh <- newStreamInterruptedError(err.Error())
	}
}

// parseSSEStream reads a Chat Completions SSE response and sends parsed
// ChatCompletionChunks to chunkCh. It is a specialisation of parseSSEStreamOf.
func parseSSEStream(resp *http.Response, chunkCh chan<- ChatCompletionChunk, errCh chan<- error) {
	parseSSEStreamOf[ChatCompletionChunk](resp, chunkCh, errCh)
}
