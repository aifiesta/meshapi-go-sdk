package livetest

import (
	"context"
	"errors"
	"testing"

	meshapi "meshapi-go-sdk"
)

func TestLive_Error_Unauthorized(t *testing.T) {
	skipIfNoBackend(t)
	badClient := meshapi.New(meshapi.Config{
		BaseURL: defaultBaseURL,
		Token:   "rsk_badtoken",
	})
	ctx := context.Background()

	_, err := badClient.Chat.Completions.Create(ctx, meshapi.ChatCompletionParams{
		Model:    strPtr(liveModel()),
		Messages: []meshapi.ChatMessage{{Role: "user", Content: "hello"}},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var svcErr *meshapi.MeshAPIError
	if !errors.As(err, &svcErr) {
		t.Fatalf("expected *MeshAPIError, got %T: %v", err, err)
	}
	if svcErr.Status != 401 {
		t.Errorf("expected status 401, got %d", svcErr.Status)
	}
	if svcErr.Code != "unauthorized" {
		t.Errorf("expected code 'unauthorized', got %q", svcErr.Code)
	}
	t.Logf("[PASS] unauthorized → status=%d code=%q requestId=%q",
		svcErr.Status, svcErr.Code, svcErr.RequestID)
}

func TestLive_Error_NotFound(t *testing.T) {
	client := newClient(t)
	ctx := context.Background()

	_, err := client.Templates.Get(ctx, "00000000-0000-0000-0000-000000000000")
	if err == nil {
		t.Fatal("expected 404, got nil")
	}

	var svcErr *meshapi.MeshAPIError
	if !errors.As(err, &svcErr) {
		t.Fatalf("expected *MeshAPIError, got %T: %v", err, err)
	}
	if svcErr.Status != 404 {
		t.Errorf("expected status 404, got %d", svcErr.Status)
	}
	t.Logf("[PASS] not_found → status=%d code=%q", svcErr.Status, svcErr.Code)
}

func TestLive_Error_TypeChain(t *testing.T) {
	skipIfNoBackend(t)
	badClient := meshapi.New(meshapi.Config{
		BaseURL: defaultBaseURL,
		Token:   "rsk_badtoken",
	})
	ctx := context.Background()

	_, err := badClient.Models.List(ctx, meshapi.ListModelsParams{})
	if err == nil {
		t.Fatal("expected error")
	}

	var svcErr *meshapi.MeshAPIError
	if !errors.As(err, &svcErr) {
		t.Fatalf("expected *MeshAPIError in chain, got %T", err)
	}
	t.Logf("[PASS] error type chain verified: *MeshAPIError status=%d code=%q", svcErr.Status, svcErr.Code)
}
