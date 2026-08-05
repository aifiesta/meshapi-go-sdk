package livetest

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	meshapi "meshapi-go-sdk"
)

// TestLive_RequestID_Default asserts the backend stamps every response with a
// server-minted X-Request-Id (req_<ULID>) exposed via the embedded
// ResponseMeta on the decoded response.
func TestLive_RequestID_Default(t *testing.T) {
	client := newClient(t)
	ctx := context.Background()

	page, err := client.Models.Search(ctx, meshapi.ModelSearchParams{Limit: intPtr(1)})
	if err != nil {
		t.Fatalf("models.search: %v", err)
	}
	if !strings.HasPrefix(page.RequestID, "req_") {
		t.Errorf("RequestID = %q, want req_<ULID> prefix", page.RequestID)
	}
	t.Logf("[PASS] models.search → request_id=%q", page.RequestID)
}

// TestLive_RequestID_CustomEcho asserts a valid client-supplied id sent via
// WithRequestID is honored and echoed back on the response.
func TestLive_RequestID_CustomEcho(t *testing.T) {
	client := newClient(t)

	customID := fmt.Sprintf("sdk-go-livetest-%d", time.Now().UnixNano())
	ctx := meshapi.WithRequestID(context.Background(), customID)

	page, err := client.Models.Search(ctx, meshapi.ModelSearchParams{Limit: intPtr(1)})
	if err != nil {
		t.Fatalf("models.search with request id: %v", err)
	}
	if page.RequestID != customID {
		t.Errorf("RequestID = %q, want echoed custom id %q", page.RequestID, customID)
	}
	t.Logf("[PASS] custom X-Request-Id echoed → %q", page.RequestID)
}

// TestLive_RequestID_OnChatCompletion asserts the request id also surfaces on
// an inference response, and that an error response carries it too via
// MeshAPIError.RequestID (covered by errors_test.go for the error side).
func TestLive_RequestID_OnChatCompletion(t *testing.T) {
	client := newClient(t)
	ctx := context.Background()

	resp, err := client.Chat.Completions.Create(ctx, meshapi.ChatCompletionParams{
		Model:     strPtr(liveModel()),
		Messages:  []meshapi.ChatMessage{{Role: "user", Content: "Reply with the single word: pong"}},
		MaxTokens: intPtr(10),
	})
	if err != nil {
		t.Fatalf("chat.create: %v", err)
	}
	if !strings.HasPrefix(resp.RequestID, "req_") {
		t.Errorf("chat RequestID = %q, want req_<ULID> prefix", resp.RequestID)
	}
	t.Logf("[PASS] chat.create → request_id=%q", resp.RequestID)
}

// TestLive_RequestID_InvalidFailsFast asserts an invalid id errors client-side
// before any network I/O (the backend would silently ignore it otherwise).
func TestLive_RequestID_InvalidFailsFast(t *testing.T) {
	client := newClient(t)
	ctx := meshapi.WithRequestID(context.Background(), "not valid: has spaces")

	_, err := client.Models.Search(ctx, meshapi.ModelSearchParams{Limit: intPtr(1)})
	if err == nil {
		t.Fatal("expected client-side error for invalid request id, got nil")
	}
	if !strings.Contains(err.Error(), "invalid request id") {
		t.Errorf("error %q does not mention invalid request id", err)
	}
	t.Logf("[PASS] invalid request id rejected client-side: %v", err)
}
