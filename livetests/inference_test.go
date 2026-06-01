package livetest

import (
	"context"
	"testing"

	meshapi "meshapi-go-sdk"
)

func batchRequests(tag string) []meshapi.BatchRequestItem {
	return []meshapi.BatchRequestItem{
		{
			CustomID: tag + "-1",
			Body: map[string]interface{}{
				"model": MODELOrDefault(),
				"messages": []map[string]interface{}{
					{"role": "user", "content": "Reply with the single word: hello"},
				},
				"max_tokens": 10,
			},
		},
		{
			CustomID: tag + "-2",
			Body: map[string]interface{}{
				"model": MODELOrDefault(),
				"messages": []map[string]interface{}{
					{"role": "user", "content": "Reply with the single word: world"},
				},
				"max_tokens": 10,
			},
		},
	}
}

func MODELOrDefault() string {
	return liveModel()
}

func TestLive_Embeddings_Create(t *testing.T) {
	client := newClient(t)
	ctx := context.Background()

	resp, err := client.Embeddings.Create(ctx, meshapi.EmbeddingsParams{
		Model: strPtr(liveEmbeddingsModel()),
		Input: "MeshAPI embeddings smoke test",
	})
	if err != nil {
		t.Fatalf("embeddings.create: %v", err)
	}
	if len(resp.Data) == 0 {
		t.Fatal("embeddings.create returned 0 items")
	}
	t.Logf("[PASS] embeddings.create -> items=%d model=%q", len(resp.Data), resp.Model)
}

func TestLive_Responses_Create(t *testing.T) {
	client := newClient(t)
	ctx := context.Background()

	maxTokens := 16
	resp, err := client.Responses.Create(ctx, meshapi.ResponsesParams{
		Model:           strPtr(liveModel()),
		Input:           "Reply with exactly the word: ok",
		MaxOutputTokens: &maxTokens,
	})
	if err != nil {
		t.Fatalf("responses.create: %v", err)
	}
	t.Logf("[PASS] responses.create -> id=%v status=%v", resp.ID, resp.Status)
}

func TestLive_Responses_Stream(t *testing.T) {
	client := newClient(t)
	ctx := context.Background()

	maxTokens := 32
	eventCh, errCh := client.Responses.Stream(ctx, meshapi.ResponsesParams{
		Model:           strPtr(liveModel()),
		Input:           "Count from 1 to 3.",
		MaxOutputTokens: &maxTokens,
	})

	count := 0
	for range eventCh {
		count++
	}
	if err := <-errCh; err != nil {
		if strings.Contains(err.Error(), "status=501") {
			t.Skip("[SKIP] responses.stream -> 501 Not Implemented (model may not support native responses streaming fallback)")
		}
		t.Fatalf("responses.stream: %v", err)
	}

	if count == 0 {
		t.Fatal("responses.stream returned 0 events")
	}
	t.Logf("[PASS] responses.stream -> %d event(s)", count)
}

func TestLive_Compare_Create(t *testing.T) {
	client := newClient(t)
	ctx := context.Background()

	maxTokens := 16
	skip := true
	resp, err := client.Compare.Create(ctx, meshapi.CompareParams{
		Models: []string{liveModel(), liveSecondModel()},
		Messages: []meshapi.ChatMessage{
			{Role: "user", Content: "Reply with the word: compare"},
		},
		MaxTokens:      &maxTokens,
		SkipComparison: &skip,
	})
	if err != nil {
		t.Fatalf("compare.create: %v", err)
	}
	if len(resp.Results) != 2 {
		t.Fatalf("expected 2 compare results, got %d", len(resp.Results))
	}
	t.Logf("[PASS] compare.create -> results=%d partial=%v", len(resp.Results), resp.Partial)
}

func TestLive_Compare_Stream(t *testing.T) {
	t.Skip("server-side SQLAlchemy session concurrency issue when compare tests run back-to-back")
	client := newClient(t)
	ctx := context.Background()

	maxTokens := 16
	skip := true
	eventCh, errCh := client.Compare.Stream(ctx, meshapi.CompareParams{
		Models: []string{liveModel(), liveSecondModel()},
		Messages: []meshapi.ChatMessage{
			{Role: "user", Content: "Reply with the word: stream"},
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
		t.Fatal("compare.stream returned 0 events")
	}
	t.Logf("[PASS] compare.stream -> %d event(s)", count)
}

func TestLive_Batches_Lifecycle(t *testing.T) {
	client := newClient(t)
	ctx := context.Background()

	tag := uniqueName("go-batch")

	// Create batch with inline requests (no file upload step required)
	batch, err := client.Batches.Create(ctx, meshapi.CreateBatchParams{
		Requests: batchRequests(tag),
		Metadata: map[string]interface{}{"suite": "go-livetest"},
	})
	if err != nil {
		t.Fatalf("batches.create: %v", err)
	}
	t.Logf("[PASS] batches.create -> id=%q status=%v", batch.ID, batch.Status)

	limit := 10
	listed, err := client.Batches.List(ctx, nil, &limit)
	if err != nil {
		t.Fatalf("batches.list: %v", err)
	}
	found := false
	for _, item := range listed.Data {
		if item.ID == batch.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("batch %q not found in list", batch.ID)
	}
	t.Logf("[PASS] batches.list -> count=%d", len(listed.Data))

	gotBatch, err := client.Batches.Get(ctx, batch.ID)
	if err != nil {
		t.Fatalf("batches.get: %v", err)
	}
	if gotBatch.ID != batch.ID {
		t.Fatalf("batches.get mismatch: want %q got %q", batch.ID, gotBatch.ID)
	}
	t.Logf("[PASS] batches.get -> status=%v", gotBatch.Status)

	cancelled, err := client.Batches.Cancel(ctx, batch.ID)
	if err != nil {
		t.Fatalf("batches.cancel: %v", err)
	}
	if cancelled.ID != batch.ID {
		t.Fatalf("batches.cancel mismatch: want %q got %q", batch.ID, cancelled.ID)
	}
	t.Logf("[PASS] batches.cancel -> status=%v", cancelled.Status)
}

func TestLive_Images_Generate(t *testing.T) {
	client := newClient(t)
	ctx := context.Background()

	imageGenModel := liveImageGenModel()
	if imageGenModel == "" {
		t.Skip("MESHAPI_IMAGE_GEN_MODEL not set")
	}

	n := 1
	size := "1024x1024"
	resp, err := client.Images.Generate(ctx, meshapi.ImageGenerationParams{
		Model:  &imageGenModel,
		Prompt: "A small blue square on a white background.",
		N:      &n,
		Size:   &size,
	})
	if err != nil {
		t.Fatalf("images.generate: %v", err)
	}
	if resp.Created == 0 {
		t.Fatal("images.generate returned 0 created timestamp")
	}
	if len(resp.Data) == 0 {
		t.Fatal("images.generate returned 0 images")
	}
	if (resp.Data[0].B64JSON == nil || *resp.Data[0].B64JSON == "") && (resp.Data[0].URL == nil || *resp.Data[0].URL == "") {
		t.Fatal("images.generate returned empty image data")
	}

	t.Logf("[PASS] images.generate -> created=%d images=%d", resp.Created, len(resp.Data))
}

