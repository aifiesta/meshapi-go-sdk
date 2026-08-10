package meshapi

import (
	"os/exec"
	"strings"
	"testing"
)

// TestSDKVersionMatchesLatestTag guards the X-MeshAPI-SDK header against the
// drift that has now bitten two SDKs: meshapi-node-sdk 1.0.4 shipped reporting
// node/0.1.3, and meshapi-python-sdk 0.1.12 reported python/0.1.11. Both had a
// version bump that updated the manifest and missed the constant.
//
// Skipped outside a git checkout (e.g. `go get` consumers) rather than failing.
func TestSDKVersionMatchesLatestTag(t *testing.T) {
	out, err := exec.Command("git", "tag", "--sort=-v:refname").Output()
	if err != nil {
		t.Skip("not a git checkout; nothing to compare against")
	}
	tags := strings.Fields(string(out))
	if len(tags) == 0 {
		t.Skip("no tags yet")
	}
	want := "go/" + strings.TrimPrefix(tags[0], "v")
	if sdkVersionValue != want {
		t.Errorf("sdkVersionValue = %q, latest tag implies %q — bump both together",
			sdkVersionValue, want)
	}
}
