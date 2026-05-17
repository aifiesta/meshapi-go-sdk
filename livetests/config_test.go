package livetest

import (
	"bufio"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	meshapi "meshapi-go-sdk"
)

const (
	defaultBaseURL = "http://localhost:8000"
	defaultToken   = "rsk_01KN96KQWDPF2X1E9CP8567JY4"
	defaultModel   = "openai/gpt-4o-mini"
)

var (
	sharedEnv        = loadSharedEnv()
	backendChecked   bool
	backendReachable bool
	backendMu        sync.Mutex
)

func checkConnectivity(baseURL string) bool {
	u, err := url.Parse(baseURL)
	if err != nil {
		return false
	}
	host := u.Host
	if !strings.Contains(host, ":") {
		if u.Scheme == "https" {
			host += ":443"
		} else {
			host += ":80"
		}
	}
	conn, err := net.DialTimeout("tcp", host, 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func loadSharedEnv() map[string]string {
	values := map[string]string{}
	envPath := filepath.Join("..", ".env.livetest")

	file, err := os.Open(envPath)
	if err != nil {
		return values
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if key != "" {
			values[key] = value
		}
	}

	return values
}

func skipIfNoBackend(t *testing.T) {
	t.Helper()
	baseURL := liveBaseURL()

	backendMu.Lock()
	if !backendChecked {
		backendReachable = checkConnectivity(baseURL)
		backendChecked = true
	}
	reachable := backendReachable
	backendMu.Unlock()

	if !reachable {
		t.Skipf("Backend %s is not reachable, skipping live test", baseURL)
	}
}

func liveBaseURL() string {
	baseURL := os.Getenv("MESHAPI_BASE_URL")
	if baseURL == "" {
		baseURL = sharedEnv["MESHAPI_BASE_URL"]
	}
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return baseURL
}

func newClient(t *testing.T) *meshapi.Client {
	t.Helper()
	skipIfNoBackend(t)

	baseURL := liveBaseURL()
	token := os.Getenv("MESHAPI_TOKEN")
	if token == "" {
		token = sharedEnv["MESHAPI_TOKEN"]
	}
	if token == "" {
		token = defaultToken
	}
	return meshapi.New(meshapi.Config{
		BaseURL: baseURL,
		Token:   token,
	})
}

func liveModel() string {
	return liveEnv("MESHAPI_MODEL", defaultModel)
}

func liveSecondModel() string {
	fallback := "anthropic/claude-haiku-4-5"
	if liveModel() == fallback {
		fallback = defaultModel
	}
	return liveEnv("MESHAPI_SECOND_MODEL", fallback)
}

func liveEmbeddingsModel() string {
	return liveEnv("MESHAPI_EMBEDDINGS_MODEL", "openai/text-embedding-3-small")
}

func liveImageGenModel() string {
	return liveEnv("MESHAPI_IMAGE_GEN_MODEL", "")
}


func strPtr(s string) *string { return &s }
func intPtr(i int) *int       { return &i }

func liveEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		value = sharedEnv[key]
	}
	if value == "" {
		value = fallback
	}
	return value
}
