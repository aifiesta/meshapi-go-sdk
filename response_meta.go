package meshapi

// ResponseMeta carries per-response metadata that the gateway returns in
// headers rather than in the JSON body. Embedded in every response type, so
// `resp.RequestID` is available on any call's result.
//
// The `json:"-"` tag keeps it out of both directions of serialisation: it is
// never populated from the body and never emitted when a response is
// re-marshalled, so round-tripping a response is unchanged.
type ResponseMeta struct {
	// RequestID is the `X-Request-Id` of the response that produced this value
	// — quote it when contacting support.
	//
	// It identifies THIS call specifically, so it stays correct with any number
	// of requests in flight. Empty if the response carried no such header.
	RequestID string `json:"-"`
}

// setRequestID lets the HTTP layer stamp the id without every response type
// having to be enumerated: embedding ResponseMeta satisfies the interface
// automatically.
func (m *ResponseMeta) setRequestID(id string) { m.RequestID = id }

// requestIDSetter is satisfied by anything embedding ResponseMeta.
type requestIDSetter interface{ setRequestID(string) }
