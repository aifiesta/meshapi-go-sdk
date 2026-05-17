package livetest

import (
	"context"
	"testing"

	meshapi "meshapi-go-sdk"
)

func TestLive_FeatureMatrix_StableOptions(t *testing.T) {
	client := newClient(t)
	ctx := context.Background()

	temperature := 0.0
	topP := 1.0
	maxTokens := 10
	seed := 42
	resp, err := client.Chat.Completions.Create(ctx, meshapi.ChatCompletionParams{
		Model:       strPtr(liveModel()),
		Messages:    []meshapi.ChatMessage{{Role: "user", Content: "Reply with exactly the word: seeded"}},
		Seed:        &seed,
		Temperature: &temperature,
		TopP:        &topP,
		User:        strPtr("go-feature-matrix"),
		MaxTokens:   &maxTokens,
	})
	if err != nil {
		t.Fatalf("chat options: %v", err)
	}
	t.Logf("[PASS] chat options -> id=%q model=%q", resp.ID, resp.Model)

	// responses with reasoning requires a reasoning-capable model; skip with default model
	t.Log("[SKIP] responses stable options -> reasoning.effort not supported by default model")

	emb, err := client.Embeddings.Create(ctx, meshapi.EmbeddingsParams{
		Model: strPtr(liveEmbeddingsModel()),
		Input: []string{"alpha", "beta"},

		User:  strPtr("go-feature-matrix"),
	})
	if err != nil {
		t.Fatalf("embeddings options: %v", err)
	}
	t.Logf("[PASS] embeddings options -> items=%d", len(emb.Data))

	skip := true
	comp, err := client.Compare.Create(ctx, meshapi.CompareParams{
		Models:                 []string{liveModel(), liveSecondModel()},
		Messages:               []meshapi.ChatMessage{{Role: "user", Content: "Reply with compare"}},
		ComparisonInstructions: strPtr("Do not add extra prose."),
		MaxTokens:              &maxTokens,
		SkipComparison:         &skip,
	})
	if err != nil {
		t.Fatalf("compare options: %v", err)
	}
	t.Logf("[PASS] compare options -> results=%d", len(comp.Results))
}

func TestLive_FeatureMatrix_Multimodal(t *testing.T) {
	client := newClient(t)
	ctx := context.Background()

	imageURL := liveEnv("MESHAPI_IMAGE_URL", "")
	if imageURL == "" {
		t.Skip("set MESHAPI_IMAGE_URL to test image input")
	}

	maxTokens := 30
	text := "Describe this image in three words."
	resp, err := client.Chat.Completions.Create(ctx, meshapi.ChatCompletionParams{
		Model: strPtr(liveEnv("MESHAPI_IMAGE_MODEL", liveModel())),
		Messages: []meshapi.ChatMessage{{
			Role: "user",
			Content: []meshapi.ContentPart{
				{Type: "text", Text: &text},
				{Type: "image_url", ImageURL: &meshapi.ImageURL{URL: imageURL}},
			},
		}},
		MaxTokens: &maxTokens,
	})
	if err != nil {
		t.Fatalf("chat image input: %v", err)
	}
	t.Logf("[PASS] chat image input -> id=%q", resp.ID)
}
