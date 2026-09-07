package livetest

// Live checks that the version this SDK pins is one the gateway actually serves.
//
// The unit tests prove the header is *sent*. Only a real gateway can prove it is
// ACCEPTED — and that is the failure that matters: MeshAPI answers a version it does
// not serve with 400 invalid_api_version rather than falling back, so a stale
// meshapi.APIVersion in a published release breaks every request that release makes.

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	meshapi "meshapi-go-sdk"
)

const apiVersionHeader = "X-Mesh-Version"

type apiVersionEntry struct {
	Label    string  `json:"label"`
	Status   string  `json:"status"`
	SunsetOn *string `json:"sunset_on"`
	Baseline bool    `json:"baseline"`
	NotesURL *string `json:"notes_url"`
}

// get issues a plain authenticated GET, bypassing the SDK so the raw response
// headers and status are observable.
func get(t *testing.T, path string, extra map[string]string) *http.Response {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, liveBaseURL()+path, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+liveEnv("MESHAPI_TOKEN", defaultToken))
	for k, v := range extra {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	return resp
}

// servedVersions returns the gateway's own list of pinnable versions, or nil if this
// deployment predates the endpoint.
//
// GET /v1/api-versions landed in routersvc #1119 and reaches a deployment only on a
// v*.*.* tag, so a 404 means "older than the endpoint" — not a failure of this SDK.
func servedVersions(t *testing.T) []apiVersionEntry {
	t.Helper()
	resp := get(t, "/v1/api-versions", nil)
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /v1/api-versions → %d", resp.StatusCode)
	}
	var entries []apiVersionEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		t.Fatalf("decode api-versions: %v", err)
	}
	return entries
}

// echoed reports the version the gateway says it served.
//
// Read under either name on purpose. routersvc renamed the header to X-Mesh-Version
// (#1110), but that reaches a deployment only on a tag — as of 2026-08-12 api-dev
// echoes x-mesh-version and prod still echoes mesh-version. Both accept either name
// on the REQUEST, so this SDK's pin is honoured on both; only the echo lags. Being
// strict about the echo name here would report a green SDK as broken against an
// untagged prod.
func echoed(resp *http.Response) string {
	if v := resp.Header.Get(apiVersionHeader); v != "" {
		return v
	}
	return resp.Header.Get("Mesh-Version")
}

func TestLive_APIVersion_PinnedVersionIsServed(t *testing.T) {
	// The whole point: a real request carrying this SDK's pin must succeed. A failure
	// with invalid_api_version means the constant is stale and this release cannot
	// talk to the gateway at all.
	client := newClient(t)

	models, err := client.Models.List(context.Background(), meshapi.ListModelsParams{})
	if err != nil {
		t.Fatalf("a request pinned to %s was rejected: %v", meshapi.APIVersion, err)
	}
	if len(models) == 0 {
		t.Error("expected at least one model")
	}
}

func TestLive_APIVersion_GatewayListsOurPin(t *testing.T) {
	// Catches a stale SDK the moment a version is retired, rather than when a
	// customer reports a 400.
	skipIfNoBackend(t)

	entries := servedVersions(t)
	if entries == nil {
		t.Skipf("%s predates GET /v1/api-versions (routersvc #1119)", liveBaseURL())
	}

	labels := make([]string, 0, len(entries))
	for _, e := range entries {
		labels = append(labels, e.Label)
		if e.Label == meshapi.APIVersion {
			return
		}
	}
	t.Errorf("this SDK pins %s, which %s does not serve; served: %v",
		meshapi.APIVersion, liveBaseURL(), labels)
}

func TestLive_APIVersion_OurPinIsNotSunset(t *testing.T) {
	// A version can still be listed while on its way out. A release pinning a sunset
	// version is already broken, it just has not failed yet.
	skipIfNoBackend(t)

	entries := servedVersions(t)
	if entries == nil {
		t.Skip("endpoint not deployed")
	}
	for _, e := range entries {
		if e.Label == meshapi.APIVersion && e.Status == "sunset" {
			t.Errorf("this SDK pins %s, which is sunset (sunset_on=%v)", e.Label, e.SunsetOn)
		}
	}
}

func TestLive_APIVersion_ResponseEchoesTheVersionServed(t *testing.T) {
	skipIfNoBackend(t)

	resp := get(t, "/v1/models", map[string]string{apiVersionHeader: meshapi.APIVersion})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /v1/models → %d", resp.StatusCode)
	}
	if got := echoed(resp); got != meshapi.APIVersion {
		t.Errorf("echoed version = %q, want %q", got, meshapi.APIVersion)
	}
}

func TestLive_APIVersion_UnservedVersionIsRejectedLoudly(t *testing.T) {
	// Confirms the gateway does NOT silently fall back — the property the whole
	// pinning scheme rests on. If this returned 200, a typo'd pin would leave a
	// caller believing they were pinned when they were not.
	skipIfNoBackend(t)

	resp := get(t, "/v1/models", map[string]string{apiVersionHeader: "1999-01"})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("an unserved version returned %d, want 400", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	var envelope struct {
		Error struct{ Code string } `json:"error"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	if envelope.Error.Code != "invalid_api_version" {
		t.Errorf("error code = %q, want invalid_api_version", envelope.Error.Code)
	}
}

func TestLive_APIVersion_UnpinnedRequestGetsTheBaseline(t *testing.T) {
	// No header means the gateway's baseline, and it says which one it used. This is
	// what Config.APIVersion = &"" opts into.
	skipIfNoBackend(t)

	entries := servedVersions(t)
	if entries == nil {
		t.Skip("endpoint not deployed")
	}
	baseline := ""
	for _, e := range entries {
		if e.Baseline {
			baseline = e.Label
		}
	}
	if baseline == "" {
		t.Fatal("the gateway must mark exactly one version as the baseline")
	}

	resp := get(t, "/v1/models", nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /v1/models → %d", resp.StatusCode)
	}
	if got := echoed(resp); got != baseline {
		t.Errorf("unpinned request served %q, want the baseline %q", got, baseline)
	}
}
