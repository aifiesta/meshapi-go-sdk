//go:build integration

package meshapi

import (
	"context"
	"os"
	"testing"
)

func integrationClient(t *testing.T) *Client {
	t.Helper()
	baseURL := os.Getenv("MESHAPI_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8000"
	}
	token := os.Getenv("MESHAPI_TOKEN")
	if token == "" {
		token = "rsk_01KN96KQWDPF2X1E9CP8567JY4"
	}
	return NewClient(Config{BaseURL: baseURL, Token: token})
}

func TestIntegration_ModelsList(t *testing.T) {
	client := integrationClient(t)
	ctx := context.Background()
	models, err := client.Models.List(ctx, ListModelsParams{})
	if err != nil {
		t.Fatalf("models list: %v", err)
	}
	t.Logf("got %d models", len(models))
	for _, m := range models {
		if m.ID == "" {
			t.Error("model with empty ID")
		}
	}
}

func TestIntegration_ModelsFree(t *testing.T) {
	client := integrationClient(t)
	ctx := context.Background()
	models, err := client.Models.Free(ctx)
	if err != nil {
		t.Fatalf("models free: %v", err)
	}
	for _, m := range models {
		if !m.IsFree {
			t.Errorf("paid model in free list: %q", m.ID)
		}
	}
}

func TestIntegration_ChatCreate(t *testing.T) {
	client := integrationClient(t)
	ctx := context.Background()
	model := "openai/gpt-4o-mini"
	resp, err := client.Chat.Completions.Create(ctx, ChatCompletionParams{
		Model:    &model,
		Messages: []ChatMessage{{Role: "user", Content: "Say 'pong' only."}},
	})
	if err != nil {
		t.Fatalf("chat create: %v", err)
	}
	if resp.ID == "" {
		t.Error("empty response ID")
	}
	if len(resp.Choices) == 0 {
		t.Error("no choices")
	}
}

func TestIntegration_ChatStream(t *testing.T) {
	client := integrationClient(t)
	ctx := context.Background()
	model := "openai/gpt-4o-mini"
	chunkCh, errCh := client.Chat.Completions.Stream(ctx, ChatCompletionParams{
		Model:    &model,
		Messages: []ChatMessage{{Role: "user", Content: "Count 1 to 3."}},
	})

	var count int
	for range chunkCh {
		count++
	}
	if err := <-errCh; err != nil {
		t.Fatalf("stream error: %v", err)
	}
	if count == 0 {
		t.Error("expected at least one chunk")
	}
	t.Logf("received %d chunks", count)
}

func TestIntegration_TemplatesCRUD(t *testing.T) {
	client := integrationClient(t)
	ctx := context.Background()
	desc := "Integration test"
	system := "You are a test assistant."

	// Create
	tmpl, err := client.Templates.Create(ctx, CreateTemplateParams{
		Name:        "go-sdk-test-" + os.Getenv("RANDOM_SUFFIX"),
		Description: &desc,
		System:      &system,
	})
	if err != nil {
		t.Fatalf("create template: %v", err)
	}
	t.Cleanup(func() {
		_ = client.Templates.Delete(context.Background(), tmpl.ID)
	})

	// Get
	got, err := client.Templates.Get(ctx, tmpl.ID)
	if err != nil {
		t.Fatalf("get template: %v", err)
	}
	if got.ID != tmpl.ID {
		t.Errorf("ID mismatch: %q vs %q", got.ID, tmpl.ID)
	}

	// Update
	newDesc := "Updated"
	updated, err := client.Templates.Update(ctx, tmpl.ID, UpdateTemplateParams{
		Description: &newDesc,
	})
	if err != nil {
		t.Fatalf("update template: %v", err)
	}
	if updated.Description == nil || *updated.Description != "Updated" {
		t.Errorf("unexpected description: %v", updated.Description)
	}

	// Delete
	if err := client.Templates.Delete(ctx, tmpl.ID); err != nil {
		t.Fatalf("delete template: %v", err)
	}
}
