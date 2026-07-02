package livetest

import (
	"context"
	"testing"

	meshapi "meshapi-go-sdk"
)

func TestLive_RouterSelect_ReturnsAModel(t *testing.T) {
	client := newClient(t)
	ctx := context.Background()

	resp, err := client.Router.Select(ctx, meshapi.RouterSelectParams{
		Messages: []meshapi.ChatMessage{
			{Role: "user", Content: "Write a Python function to reverse a string."},
		},
	})
	if err != nil {
		skipIfUnavailable(t, err, "auto router (AUTO_ROUTER_ENABLED)")
		t.Fatalf("router.Select: %v", err)
	}
	if resp.Model == "" {
		t.Error("router must always return a model (fail-soft)")
	}
}

func TestLive_RouterSelect_HonorsExclusions(t *testing.T) {
	client := newClient(t)
	ctx := context.Background()

	excluded := "openai/gpt-4o-mini"
	resp, err := client.Router.Select(ctx, meshapi.RouterSelectParams{
		Messages: []meshapi.ChatMessage{
			{Role: "user", Content: "Explain the theory of relativity simply."},
		},
		ExcludeModels: []string{excluded},
	})
	if err != nil {
		skipIfUnavailable(t, err, "auto router (AUTO_ROUTER_ENABLED)")
		t.Fatalf("router.Select: %v", err)
	}
	if resp.Model == "" {
		t.Fatal("router must return a model even with exclusions")
	}
	// Unless it fell back to the default, the excluded model must not be picked.
	if !resp.AutoRouter.FallbackUsed && resp.Model == excluded {
		t.Errorf("excluded model %q should not be selected", excluded)
	}
}
