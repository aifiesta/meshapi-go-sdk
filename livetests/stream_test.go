package livetest

import (
	"context"
	"strings"
	"testing"

	meshapi "meshapi-go-sdk"
)

func TestLive_Stream_Basic(t *testing.T) {
	client := newClient(t)
	ctx := context.Background()

	chunkCh, errCh := client.Chat.Completions.Stream(ctx, meshapi.ChatCompletionParams{
		Model:     strPtr(liveModel()),
		Messages:  []meshapi.ChatMessage{{Role: "user", Content: "Count from 1 to 5, one number per line."}},
		MaxTokens: intPtr(40),
	})

	var chunks []meshapi.ChatCompletionChunk
	for chunk := range chunkCh {
		chunks = append(chunks, chunk)
	}
	if err := <-errCh; err != nil {
		t.Fatalf("stream error: %v", err)
	}
	if len(chunks) == 0 {
		t.Fatal("received 0 chunks")
	}

	var sb strings.Builder
	for _, c := range chunks {
		if len(c.Choices) > 0 && c.Choices[0].Delta != nil && c.Choices[0].Delta.Content != nil {
			sb.WriteString(*c.Choices[0].Delta.Content)
		}
	}
	text := sb.String()
	t.Logf("[PASS] stream → %d chunks, text=%q", len(chunks), text)
	if text == "" {
		t.Error("reconstructed text is empty")
	}
}

func TestLive_Stream_ChunkStructure(t *testing.T) {
	client := newClient(t)
	ctx := context.Background()

	chunkCh, errCh := client.Chat.Completions.Stream(ctx, meshapi.ChatCompletionParams{
		Model:     strPtr(liveModel()),
		Messages:  []meshapi.ChatMessage{{Role: "user", Content: "Say hello."}},
		MaxTokens: intPtr(10),
	})

	first := true
	for chunk := range chunkCh {
		if chunk.ID == "" {
			t.Error("chunk.id is empty")
		}
		if chunk.Model == "" {
			t.Error("chunk.model is empty")
		}
		if first && len(chunk.Choices) > 0 && chunk.Choices[0].Delta != nil {
			t.Logf("first chunk: id=%q model=%q role=%v content=%v",
				chunk.ID, chunk.Model,
				chunk.Choices[0].Delta.Role,
				chunk.Choices[0].Delta.Content)
			first = false
		}
	}
	if err := <-errCh; err != nil {
		t.Fatalf("stream error: %v", err)
	}
	t.Log("[PASS] stream chunk structure valid")
}

func TestLive_Stream_Cancel(t *testing.T) {
	client := newClient(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	chunkCh, errCh := client.Chat.Completions.Stream(ctx, meshapi.ChatCompletionParams{
		Model:     strPtr(liveModel()),
		Messages:  []meshapi.ChatMessage{{Role: "user", Content: "Write a very long story about a dragon."}},
		MaxTokens: intPtr(200),
	})

	received := 0
	for range chunkCh {
		received++
		if received >= 3 {
			cancel()
			break
		}
	}
	for range chunkCh {
	}
	<-errCh // may be context.Canceled or nil — both are acceptable

	t.Logf("[PASS] stream cancel after %d chunks — no panic/deadlock", received)
}

func TestLive_Stream_ErrorAuth(t *testing.T) {
	skipIfNoBackend(t)
	badClient := meshapi.New(meshapi.Config{
		BaseURL: defaultBaseURL,
		Token:   "rsk_invalid_token",
	})
	ctx := context.Background()

	chunkCh, errCh := badClient.Chat.Completions.Stream(ctx, meshapi.ChatCompletionParams{
		Model:    strPtr(liveModel()),
		Messages: []meshapi.ChatMessage{{Role: "user", Content: "Hello"}},
	})

	for range chunkCh {
	}
	err := <-errCh
	if err == nil {
		t.Fatal("expected auth error, got nil")
	}
	t.Logf("[PASS] stream with bad token → %v", err)
}
