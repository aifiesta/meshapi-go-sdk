package livetest

import (
	"context"
	"os"
	"testing"

	meshapi "meshapi-go-sdk"
)

func TestVideo_List(t *testing.T) {
	client := newClient(t)
	limit := 5
	listing, err := client.Videos.List(context.Background(), &meshapi.ListVideoGenerationsParams{
		Limit: &limit,
	})
	if err != nil {
		t.Fatalf("Videos.List error: %v", err)
	}
	if listing == nil {
		t.Fatal("Videos.List returned nil")
	}
	t.Logf("[PASS] Videos.List -> total=%d items=%d", listing.Total, len(listing.Data))
}

func TestVideo_GenerateAndRetrieve(t *testing.T) {
	model := os.Getenv("MESHAPI_VIDEO_GEN_MODEL")
	if model == "" {
		t.Skip("set MESHAPI_VIDEO_GEN_MODEL to run video generation (costly; skipped in CI by default)")
	}
	client := newClient(t)
	text := "A serene mountain lake at sunrise"
	resp, err := client.Videos.Generate(context.Background(), meshapi.VideoGenerationParams{
		Model: model,
		Content: []meshapi.VideoContentItem{
			{Type: "text", Text: &text},
		},
	})
	if err != nil {
		t.Fatalf("Videos.Generate error: %v", err)
	}
	if resp.ID == "" {
		t.Fatal("Videos.Generate returned empty task ID")
	}
	t.Logf("[PASS] Videos.Generate -> task_id=%s", resp.ID)

	task, err := client.Videos.Retrieve(context.Background(), resp.ID)
	if err != nil {
		t.Fatalf("Videos.Retrieve error: %v", err)
	}
	if task.ID != resp.ID {
		t.Fatalf("task ID mismatch: got %s, want %s", task.ID, resp.ID)
	}
	t.Logf("[PASS] Videos.Retrieve -> status=%s", task.Status)
}
