package livetest

import (
	"context"
	"testing"

	meshapi "meshapi-go-sdk"
)

func TestLive_Compare_NonStreaming(t *testing.T) {
	client := newClient(t)
	ctx := context.Background()

	maxTokens := 20
	skip := true

	result, err := client.Compare.Create(ctx, meshapi.CompareParams{
		Models: []string{liveModel(), liveSecondModel()},
		Messages: []meshapi.ChatMessage{
			{Role: "user", Content: "What is 2+2? Reply in one word."},
		},
		MaxTokens:      &maxTokens,
		SkipComparison: &skip,
	})
	if err != nil {
		t.Fatalf("compare.create: %v", err)
	}
	if result.ComparisonID == "" {
		t.Error("comparison_id is empty")
	}
	if len(result.Results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result.Results))
	}
	for _, r := range result.Results {
		if r.Content == nil && r.Error == nil {
			t.Errorf("result for %s has neither content nor error", r.Model)
		}
		t.Logf("[PASS] model=%s content=%v", r.Model, r.Content)
	}
}

func TestLive_Compare_Streaming(t *testing.T) {
	client := newClient(t)
	ctx := context.Background()

	maxTokens := 50
	skip := true

	eventCh, errCh := client.Compare.Stream(ctx, meshapi.CompareParams{
		Models: []string{liveModel(), liveSecondModel()},
		Messages: []meshapi.ChatMessage{
			{Role: "user", Content: "Tell me a joke."},
		},
		MaxTokens:      &maxTokens,
		SkipComparison: &skip,
	})

	count := 0
	for range eventCh {
		count++
	}
	if err := <-errCh; err != nil {
		t.Fatalf("compare.stream: %v", err)
	}
	if count == 0 {
		t.Error("expected at least one streaming event")
	}
	t.Logf("[PASS] compare.stream → received %d events", count)
}
