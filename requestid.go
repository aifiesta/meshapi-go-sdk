package meshapi

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
)

// requestIDHeader is the header the backend stamps on every response
// (format req_<ULID>), and the header a client may send to supply its own id.
const requestIDHeader = "X-Request-Id"

// requestIDContextKey is the unexported context key WithRequestID stores the
// caller-supplied request id under.
type requestIDContextKey struct{}

// requestIDPattern is the charset/length the backend accepts for a
// client-supplied X-Request-Id. Anything else is silently ignored server-side
// (the backend mints its own id instead), so the SDK rejects an invalid id
// before any network I/O rather than letting it disappear silently.
var requestIDPattern = regexp.MustCompile(`^[A-Za-z0-9._:-]{1,64}$`)

// WithRequestID returns a ctx that makes the client send `X-Request-Id: id`
// on every request issued with it. The backend echoes the id back on the
// response header (readable via the embedded ResponseMeta.RequestID on
// non-streaming responses, and via MeshAPIError.RequestID on errors).
//
// The id must match ^[A-Za-z0-9._:-]{1,64}$ — validation happens at send
// time, and the resource call returns an error before any network I/O when
// the id is invalid.
//
// Streaming methods (Chat.Completions.Stream, Responses.Stream,
// Compare.Stream, Images.Stream) send the header too, but expose no response
// metadata — supply your own id with WithRequestID when you need to correlate
// a stream with server logs.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDContextKey{}, id)
}

// applyRequestID reads a WithRequestID value from ctx, validates it, and sets
// the X-Request-Id request header. It must be called before any network I/O
// so an invalid id fails fast instead of being silently ignored server-side.
func applyRequestID(ctx context.Context, header http.Header) error {
	id, ok := ctx.Value(requestIDContextKey{}).(string)
	if !ok {
		return nil
	}
	if !requestIDPattern.MatchString(id) {
		return fmt.Errorf("meshapi: invalid request id %q: must match ^[A-Za-z0-9._:-]{1,64}$", id)
	}
	header.Set(requestIDHeader, id)
	return nil
}
