package livetest

import (
	"context"
	"strings"
	"testing"
	"time"

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
		Model: strPtr(liveModel()),
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

	maxTokens := 10
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

	maxTokens := 20
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

	maxTokens := 10
	skip := true
	resp, err := client.Compare.Create(ctx, meshapi.CompareParams{
		Models: []string{liveModel(), liveModel()},
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
	client := newClient(t)
	ctx := context.Background()

	maxTokens := 10
	skip := true
	eventCh, errCh := client.Compare.Stream(ctx, meshapi.CompareParams{
		Models: []string{liveModel(), liveModel()},
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

func TestLive_FilesAndBatches_Lifecycle(t *testing.T) {
	client := newClient(t)
	ctx := context.Background()

	tag := uniqueName("go-batch")
	uploaded, err := client.Files.Upload(ctx, meshapi.UploadBatchFileParams{
		Purpose:  "batch",
		Requests: batchRequests(tag),
	})
	if err != nil {
		t.Fatalf("files.upload: %v", err)
	}
	t.Logf("[PASS] files.upload -> id=%q", uploaded.ID)

	fetched, err := client.Files.Get(ctx, uploaded.ID)
	if err != nil {
		t.Fatalf("files.get: %v", err)
	}
	if fetched.ID != uploaded.ID {
		t.Fatalf("files.get mismatch: want %q got %q", uploaded.ID, fetched.ID)
	}
	t.Logf("[PASS] files.get -> status=%v bytes=%v", fetched.Status, fetched.Bytes)

	content, err := client.Files.Content(ctx, uploaded.ID)
	if err != nil {
		t.Fatalf("files.content: %v", err)
	}
	if !strings.Contains(string(content), tag+"-1") {
		t.Fatalf("files.content missing custom_id %q", tag+"-1")
	}
	t.Logf("[PASS] files.content -> %d bytes", len(content))

	batch, err := client.Batches.Create(ctx, meshapi.CreateBatchParams{
		InputFileID:      uploaded.ID,
		Endpoint:         "/v1/chat/completions",
		CompletionWindow: "24h",
		Metadata:         map[string]interface{}{"suite": "go-livetest", "ts": time.Now().Unix()},
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

	if err := client.Files.Delete(ctx, uploaded.ID); err != nil {
		t.Fatalf("files.delete: %v", err)
	}
	t.Log("[PASS] files.delete -> 204 No Content")
}
