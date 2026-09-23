package meshapi

import (
	"os/exec"
	"strconv"
	"strings"
	"testing"
)

// TestSDKVersionNotBehindLatestTag guards the X-MeshAPI-SDK header against the
// drift that has now bitten two sibling SDKs: meshapi-node-sdk 1.0.4 shipped
// reporting node/0.1.3, and meshapi-python-sdk 0.1.12 reported python/0.1.11.
// Both had a release that updated the manifest and missed the constant.
//
// Asserts "not behind" rather than "equal" deliberately. The constant is bumped
// in a PR and the tag is created after merge, so on a release PR the constant is
// legitimately AHEAD of the newest tag. Requiring equality would turn every
// release PR red, which is a faster way to get a guard deleted than to have no
// guard at all. Being BEHIND is the actual failure: a version was tagged and
// shipped while the header still advertised the previous one.
//
// CI must fetch tags for this to do anything — see fetch-tags in
// .github/workflows/live-tests.yml. Skipped outside a git checkout so `go get`
// consumers are unaffected.
func TestSDKVersionNotBehindLatestTag(t *testing.T) {
	out, err := exec.Command("git", "tag", "--sort=-v:refname").Output()
	if err != nil {
		t.Skip("not a git checkout; nothing to compare against")
	}
	tags := strings.Fields(string(out))
	if len(tags) == 0 {
		t.Skip("no tags available (shallow clone without fetch-tags)")
	}

	constant := strings.TrimPrefix(sdkVersionValue, "go/")
	latest := strings.TrimPrefix(tags[0], "v")

	if cmpSemver(constant, latest) < 0 {
		t.Errorf("sdkVersionValue is %q but %s is already tagged — every request is "+
			"reporting a stale SDK version; bump sdkVersionValue in http.go",
			sdkVersionValue, tags[0])
	}
}

// cmpSemver returns -1, 0 or 1 comparing dotted numeric versions. Unparseable
// segments sort as 0, so a pre-release suffix degrades to a component compare
// rather than a panic.
func cmpSemver(a, b string) int {
	as, bs := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < len(as) || i < len(bs); i++ {
		var ai, bi int
		if i < len(as) {
			ai, _ = strconv.Atoi(as[i])
		}
		if i < len(bs) {
			bi, _ = strconv.Atoi(bs[i])
		}
		if ai != bi {
			if ai < bi {
				return -1
			}
			return 1
		}
	}
	return 0
}
