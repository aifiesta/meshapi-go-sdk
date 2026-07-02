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

func TestLive_Compare_WithSynthesis(t *testing.T) {
	client := newClient(t)
	ctx := context.Background()

	maxTokens := 60
	skip := false
	instructions := "Briefly state which answer is clearer."

	result, err := client.Compare.Create(ctx, meshapi.CompareParams{
		Models: []string{liveModel(), liveSecondModel()},
		Messages: []meshapi.ChatMessage{
			{Role: "user", Content: "In one sentence, what is TCP?"},
		},
		ComparisonInstructions: &instructions,
		MaxTokens:              &maxTokens,
		SkipComparison:         &skip,
	})
	if err != nil {
		t.Fatalf("compare.create (synthesis): %v", err)
	}
	if len(result.Results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result.Results))
	}
	anyContent := false
	for _, r := range result.Results {
		if r.Content != nil {
			anyContent = true
		}
	}
	// When a per-model answer succeeded and the comparison model did not fall
	// back, a synthesized comparison must be present with usage.
	if anyContent && !result.ComparisonFallbackUsed {
		if result.Comparison == nil || *result.Comparison == "" {
			t.Error("expected a synthesized comparison when skip_comparison=false")
		}
		if result.ComparisonModel == nil {
			t.Error("expected comparison_model to be reported")
		}
		if result.ComparisonUsage == nil {
			t.Error("expected comparison_usage to be populated")
		}
	}
}

func TestLive_Compare_ModelOverrides(t *testing.T) {
	client := newClient(t)
	ctx := context.Background()

	maxTokens := 20
	skip := true
	temp := 0.0
	overrideMax := 10

	result, err := client.Compare.Create(ctx, meshapi.CompareParams{
		Models: []string{liveModel(), liveSecondModel()},
		Messages: []meshapi.ChatMessage{
			{Role: "user", Content: "Say hi in one word."},
		},
		ModelOverrides: []meshapi.ModelOverride{
			{Model: liveModel(), Temperature: &temp, MaxTokens: &overrideMax},
		},
		MaxTokens:      &maxTokens,
		SkipComparison: &skip,
	})
	if err != nil {
		t.Fatalf("compare.create (overrides): %v", err)
	}
	if len(result.Results) != 2 {
		t.Fatalf("overrides must not drop any model from the fan-out; got %d", len(result.Results))
	}
}
