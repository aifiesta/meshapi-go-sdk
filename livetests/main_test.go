package livetest

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

// strictRequiredEnv lists env vars whose absence turns a real feature test into
// a silent skip (image gen, vision input, video). See ../.env.livetest.example.
var strictRequiredEnv = []string{
	"MESHAPI_IMAGE_GEN_MODEL",
	"MESHAPI_IMAGE_URL",
	"MESHAPI_VIDEO_GEN_MODEL",
}

func strictMode() bool {
	switch strings.ToLower(liveEnv("MESHAPI_STRICT_LIVETESTS", "")) {
	case "1", "true", "yes":
		return true
	}
	return false
}

// TestMain enforces the pre-hackathon gate: with MESHAPI_STRICT_LIVETESTS set,
// the run fails fast unless every optional-feature env var is present, so
// skip-by-default tests can't hide behind a green run.
func TestMain(m *testing.M) {
	if strictMode() {
		var missing []string
		for _, name := range strictRequiredEnv {
			if liveEnv(name, "") == "" {
				missing = append(missing, name)
			}
		}
		if len(missing) > 0 {
			fmt.Fprintf(os.Stderr,
				"MESHAPI_STRICT_LIVETESTS is set but these env vars are unset, so their "+
					"feature tests would silently skip:\n  - %s\n"+
					"Set them (see ../.env.livetest.example) or unset MESHAPI_STRICT_LIVETESTS.\n",
				strings.Join(missing, "\n  - "))
			os.Exit(1)
		}
	}
	os.Exit(m.Run())
}
