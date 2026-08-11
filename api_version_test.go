package meshapi

// X-Mesh-Version — the dated API version this SDK was built against (MESH-508).
//
// MeshAPI versions its contract by date, in a request header. An SDK that sends
// nothing is served the gateway's BASELINE — safe today, but it also means the SDK
// never states which response shape it can actually parse. Sending the version
// explicitly is the difference between "whatever the server defaults to" and "the
// shape this release was written for".
//
// It works today only because BASELINE is the OLDEST supported version, so it never
// moves on its own. Pinning turns that from a coincidence into a contract, and puts
// this release into usage_events.api_version so a version can be retired on evidence
// about who still uses it.
//
// Distinct from X-MeshAPI-SDK: one says which SDK build, the other which contract.
// Neither substitutes for the other.

import (
	"context"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"testing"
)

// newHeaderServer records the headers of each request and echoes the version pin
// back the way the gateway does.
func newHeaderServer(t *testing.T, body string) (*Client, *http.Header) {
	t.Helper()
	var seen http.Header
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set(apiVersionHeader, r.Header.Get(apiVersionHeader))
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return New(Config{BaseURL: srv.URL, Token: "rsk_test"}), &seen
}

func TestAPIVersion_ConstantIsADatedLabel(t *testing.T) {
	// Public because a caller pinning explicitly needs a value to pass, and a caller
	// debugging a shape mismatch needs to know what was sent.
	if APIVersion != "2026-08" {
		t.Errorf("APIVersion = %q, want 2026-08", APIVersion)
	}
}

func TestAPIVersion_ConstantIsWellFormed(t *testing.T) {
	// The gateway 400s a malformed label rather than falling back, so a typo here
	// would break every request this SDK makes, not degrade quietly.
	m := regexp.MustCompile(`^(\d{4})-(\d{2})$`).FindStringSubmatch(APIVersion)
	if m == nil {
		t.Fatalf("APIVersion = %q, not YYYY-MM", APIVersion)
	}
	month, _ := strconv.Atoi(m[2])
	if month < 1 || month > 12 {
		t.Errorf("month %d out of range", month)
	}
}

func TestAPIVersion_SentByDefault(t *testing.T) {
	client, seen := newHeaderServer(t, `[]`)

	if _, err := client.Models.List(context.Background(), ListModelsParams{}); err != nil {
		t.Fatalf("List: %v", err)
	}

	if got := seen.Get(apiVersionHeader); got != APIVersion {
		t.Errorf("%s = %q, want %q", apiVersionHeader, got, APIVersion)
	}
}

func TestAPIVersion_DoesNotDisplaceSDKIdentityHeader(t *testing.T) {
	// Two headers with two different jobs.
	client, seen := newHeaderServer(t, `[]`)

	if _, err := client.Models.List(context.Background(), ListModelsParams{}); err != nil {
		t.Fatalf("List: %v", err)
	}

	if got := seen.Get(sdkVersionHeader); got != sdkVersionValue {
		t.Errorf("%s = %q, want %q", sdkVersionHeader, got, sdkVersionValue)
	}
	if got := seen.Get(apiVersionHeader); got != APIVersion {
		t.Errorf("%s = %q, want %q", apiVersionHeader, got, APIVersion)
	}
}

func TestAPIVersion_SentOnMultipartUploads(t *testing.T) {
	// The multipart path sets Content-Type itself before calling do(), which applies
	// base headers only when absent. Worth asserting separately that the version
	// still rides along on that path.
	var seen http.Header
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"text":"ok"}`))
	}))
	t.Cleanup(srv.Close)
	client := New(Config{BaseURL: srv.URL, Token: "rsk_test"})

	if _, err := client.Audio.Transcribe(context.Background(), []byte{1, 2, 3}, "a.mp3",
		TranscriptionParams{Model: "openai/whisper-1"}); err != nil {
		t.Fatalf("Transcribe: %v", err)
	}

	if got := seen.Get(apiVersionHeader); got != APIVersion {
		t.Errorf("%s = %q, want %q", apiVersionHeader, got, APIVersion)
	}
}

func TestAPIVersion_SatisfiesAGatewayThatRejectsUnknownVersions(t *testing.T) {
	// Simulates the real gateway rather than asserting against a permissive mock.
	// MeshAPI 400s invalid_api_version on a label it does not serve, and treats an
	// EMPTY value as a typo'd pin rather than "no pin". A mock that accepts anything
	// would let both mistakes through.
	served := map[string]bool{APIVersion: true, "2026-09": true}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if pinned, ok := r.Header[http.CanonicalHeaderKey(apiVersionHeader)]; ok && !served[pinned[0]] {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":{"code":"invalid_api_version"}}`))
			return
		}
		_, _ = w.Write([]byte(`[]`))
	}))
	t.Cleanup(srv.Close)
	client := New(Config{BaseURL: srv.URL, Token: "rsk_test"})

	if _, err := client.Models.List(context.Background(), ListModelsParams{}); err != nil {
		t.Fatalf("a request carrying this SDK's own pin was rejected: %v", err)
	}
}

func TestAPIVersion_PerClientOverride(t *testing.T) {
	// A customer who has migrated ahead of this SDK release must not be forced back
	// onto the version the SDK was built against.
	var seen http.Header
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))
	t.Cleanup(srv.Close)
	pinned := "2026-09"
	client := New(Config{BaseURL: srv.URL, Token: "rsk_test", APIVersion: &pinned})

	if _, err := client.Models.List(context.Background(), ListModelsParams{}); err != nil {
		t.Fatalf("List: %v", err)
	}

	if got := seen.Get(apiVersionHeader); got != "2026-09" {
		t.Errorf("%s = %q, want 2026-09", apiVersionHeader, got)
	}
}

func TestAPIVersion_EmptyStringSendsNothing(t *testing.T) {
	// Explicit opt-out, distinct from "unset". A nil APIVersion takes the SDK's own
	// version; a pointer to "" omits the header entirely and asks for the gateway's
	// baseline whatever it may become — the pre-MESH-508 behaviour, still reachable
	// on purpose.
	var seen http.Header
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))
	t.Cleanup(srv.Close)
	none := ""
	client := New(Config{BaseURL: srv.URL, Token: "rsk_test", APIVersion: &none})

	if _, err := client.Models.List(context.Background(), ListModelsParams{}); err != nil {
		t.Fatalf("List: %v", err)
	}

	if _, present := seen[http.CanonicalHeaderKey(apiVersionHeader)]; present {
		t.Errorf("%s was sent despite an explicit opt-out", apiVersionHeader)
	}
}
