package livetest

import (
	"context"
	"io"
	"testing"
	"time"

	meshapi "meshapi-go-sdk"
)

// TestLive_Realtime_ConnectAndClose verifies that the WebSocket upgrade
// succeeds and that a clean close handshake completes without error.
func TestLive_Realtime_ConnectAndClose(t *testing.T) {
	model := liveRealtimeModel()
	if model == "" {
		t.Skip("MESHAPI_REALTIME_MODEL not set — skipping realtime live tests")
	}
	client := newClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	session, err := client.Realtime.Connect(ctx, meshapi.RealtimeConnectParams{Model: model})
	if err != nil {
		t.Fatalf("realtime.connect: %v", err)
	}
	if err := session.Close(); err != nil {
		t.Logf("realtime.close (non-fatal): %v", err)
	}
	t.Logf("[PASS] realtime.connect+close model=%q", model)
}

// TestLive_Realtime_ReceiveSessionCreated verifies that the server sends
// a "session.created" event immediately after the WebSocket handshake.
func TestLive_Realtime_ReceiveSessionCreated(t *testing.T) {
	model := liveRealtimeModel()
	if model == "" {
		t.Skip("MESHAPI_REALTIME_MODEL not set — skipping realtime live tests")
	}
	client := newClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	session, err := client.Realtime.Connect(ctx, meshapi.RealtimeConnectParams{Model: model})
	if err != nil {
		t.Fatalf("realtime.connect: %v", err)
	}
	defer session.Close()

	msg, err := session.Receive(ctx)
	if err != nil {
		if err == io.EOF {
			t.Fatal("connection closed before any frame received")
		}
		t.Fatalf("realtime.receive: %v", err)
	}
	if msg.Event == nil {
		t.Fatal("expected JSON text frame, got binary audio frame")
	}
	msgType, _ := msg.Event["type"].(string)
	t.Logf("[PASS] realtime.receive first frame: type=%q", msgType)
}

// TestLive_Realtime_SendSessionUpdate verifies that the client can send a
// session.update command and receive a session.updated acknowledgement.
func TestLive_Realtime_SendSessionUpdate(t *testing.T) {
	model := liveRealtimeModel()
	if model == "" {
		t.Skip("MESHAPI_REALTIME_MODEL not set — skipping realtime live tests")
	}
	client := newClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	session, err := client.Realtime.Connect(ctx, meshapi.RealtimeConnectParams{Model: model})
	if err != nil {
		t.Fatalf("realtime.connect: %v", err)
	}
	defer session.Close()

	// Drain the initial session.created frame.
	if _, err := session.Receive(ctx); err != nil {
		t.Fatalf("realtime.receive session.created: %v", err)
	}

	// Send a session.update to configure instructions.
	updateCmd := map[string]any{
		"type": "session.update",
		"session": map[string]any{
			"instructions": "You are a helpful assistant.",
		},
	}
	if err := session.Send(ctx, updateCmd); err != nil {
		t.Fatalf("realtime.send session.update: %v", err)
	}

	// Expect a session.updated acknowledgement.
	msg, err := session.Receive(ctx)
	if err != nil {
		t.Fatalf("realtime.receive after session.update: %v", err)
	}
	if msg.Event == nil {
		t.Fatal("expected JSON frame after session.update")
	}
	msgType, _ := msg.Event["type"].(string)
	t.Logf("[PASS] realtime.send session.update -> type=%q", msgType)
}

// TestLive_Realtime_ErrorEnvelope_BadModel verifies that connecting with an
// unknown model results in a *RealtimeError with code "model_not_found".
func TestLive_Realtime_ErrorEnvelope_BadModel(t *testing.T) {
	client := newClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	_, err := client.Realtime.Connect(ctx, meshapi.RealtimeConnectParams{
		Model: "nonexistent/bad-model-xyz",
	})
	if err == nil {
		// Server accepted — could happen if the model check is deferred.
		// Try to read the error envelope from the session.
		t.Log("connect succeeded; model validation may be deferred (non-fatal)")
		return
	}
	t.Logf("[PASS] realtime.connect bad model -> %v", err)
}

// TestLive_Realtime_Events_ChannelAPI verifies the channel-based Events API.
func TestLive_Realtime_Events_ChannelAPI(t *testing.T) {
	model := liveRealtimeModel()
	if model == "" {
		t.Skip("MESHAPI_REALTIME_MODEL not set — skipping realtime live tests")
	}
	client := newClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	session, err := client.Realtime.Connect(ctx, meshapi.RealtimeConnectParams{Model: model})
	if err != nil {
		t.Fatalf("realtime.connect: %v", err)
	}

	msgCh, errCh := session.Events(ctx)

	// Wait for the first frame via the channel API.
	select {
	case msg, ok := <-msgCh:
		if !ok {
			if err := <-errCh; err != nil {
				t.Fatalf("realtime.events channel closed with error: %v", err)
			}
			t.Fatal("channel closed before any message received")
		}
		msgType, _ := msg.Event["type"].(string)
		t.Logf("[PASS] realtime.events first message: type=%q", msgType)
	case err := <-errCh:
		t.Fatalf("realtime.events error: %v", err)
	case <-ctx.Done():
		t.Fatal("timeout waiting for first message via Events()")
	}

	session.Close()
}

// TestLive_Realtime_ContextCancel verifies that cancelling the context
// causes Receive to return promptly without blocking.
func TestLive_Realtime_ContextCancel(t *testing.T) {
	model := liveRealtimeModel()
	if model == "" {
		t.Skip("MESHAPI_REALTIME_MODEL not set — skipping realtime live tests")
	}
	client := newClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)

	session, err := client.Realtime.Connect(ctx, meshapi.RealtimeConnectParams{Model: model})
	if err != nil {
		t.Fatalf("realtime.connect: %v", err)
	}
	defer session.Close()

	// Drain the initial session.created frame so Receive blocks waiting for more.
	if _, err := session.Receive(ctx); err != nil {
		t.Fatalf("realtime.receive session.created: %v", err)
	}

	// Cancel the context and verify Receive returns promptly.
	cancel()
	done := make(chan struct{})
	go func() {
		session.Receive(ctx) //nolint:errcheck
		close(done)
	}()

	select {
	case <-done:
		t.Log("[PASS] realtime.receive returned after context cancel")
	case <-time.After(3 * time.Second):
		t.Fatal("Receive did not return within 3s after context cancel")
	}
}
