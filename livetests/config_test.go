package livetest

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"

	meshapi "meshapi-go-sdk"
)

const (
	defaultBaseURL = "http://localhost:8000"
	defaultToken   = "rsk_01KN96KQWDPF2X1E9CP8567JY4"
	defaultModel   = "openai/gpt-4o-mini"
)

var sharedEnv = loadSharedEnv()

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

func newClient(t *testing.T) *meshapi.Client {
	t.Helper()
	baseURL := os.Getenv("MESHAPI_BASE_URL")
	if baseURL == "" {
		baseURL = sharedEnv["MESHAPI_BASE_URL"]
	}
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
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
