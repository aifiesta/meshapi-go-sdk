package meshapi

import (
	"bufio"
	"encoding/json"
	"net/http"
	"strings"
)

// parseSSEStream reads an SSE response and sends parsed ChatCompletionChunks
// to chunkCh and any error to errCh. Both channels are closed when done.
//
// Streaming errors detected in SSE frames (mid-stream errors) are sent on errCh.
// Connection failures are wrapped as MeshAPIError with code "stream_interrupted".
func parseSSEStream(resp *http.Response, chunkCh chan<- ChatCompletionChunk, errCh chan<- error) {
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

		// Empty line = end of frame
		frame := remainder.String()
		remainder.Reset()

		chunk, done, err := tryParseSSEFrame(frame)
		if err != nil {
			errCh <- err
			return
		}
		if done {
			return
		}
		if chunk != nil {
			chunkCh <- *chunk
		}
	}

	if err := scanner.Err(); err != nil {
		errCh <- newStreamInterruptedError(err.Error())
	}
}

// tryParseSSEFrame parses a single SSE frame (may contain multiple lines).
// Returns (chunk, false, nil) for a data chunk,
//         (nil, true, nil)    for [DONE],
//         (nil, false, err)   for an error frame,
//         (nil, false, nil)   for empty/comment frames.
func tryParseSSEFrame(frame string) (*ChatCompletionChunk, bool, error) {
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

	var chunk ChatCompletionChunk
	if err := json.Unmarshal([]byte(dataLine), &chunk); err != nil {
		return nil, false, nil
	}
	return &chunk, false, nil
}

func parseJSONSSEStream[T any](resp *http.Response, chunkCh chan<- T, errCh chan<- error) {
	defer resp.Body.Close()
	defer close(chunkCh)
	defer close(errCh)

	scanner := bufio.NewScanner(resp.Body)
	var remainder strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		remainder.WriteString(line)
		remainder.WriteByte('\n')
		if line != "" {
			continue
		}

		frame := remainder.String()
		remainder.Reset()

		chunk, done, err := tryParseJSONSSEFrame[T](frame)
		if err != nil {
			errCh <- err
			return
		}
		if done {
			return
		}
		if chunk != nil {
			chunkCh <- *chunk
		}
	}

	if err := scanner.Err(); err != nil {
		errCh <- newStreamInterruptedError(err.Error())
	}
}

func tryParseJSONSSEFrame[T any](frame string) (*T, bool, error) {
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
		return nil, false, nil
	}
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

	var chunk T
	if err := json.Unmarshal([]byte(dataLine), &chunk); err != nil {
		return nil, false, nil
	}
	return &chunk, false, nil
}
