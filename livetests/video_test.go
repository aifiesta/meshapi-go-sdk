package livetest

import (
	"context"
	"testing"
	"time"

	meshapi "meshapi-go-sdk"
)

func TestLive_Videos_CreateAndPoll(t *testing.T) {
	client := newClient(t)
	ctx := context.Background()

	videoModel := liveVideoGenModel()
	if videoModel == "" {
		t.Skip("MESHAPI_VIDEO_GEN_MODEL not set")
	}

	textContent := "A calm ocean wave at sunset."
	duration := 4
	resolution := "480p"
	ratio := "16:9"

	// Create task
	task, err := client.Videos.Create(ctx, meshapi.VideoGenerationParams{
		Model: videoModel,
		Content: []meshapi.VideoContentItem{
			{Type: "text", Text: &textContent},
		},
		Duration:   &duration,
		Resolution: &resolution,
		Ratio:      &ratio,
	})
	if err != nil {
		t.Fatalf("videos.create: %v", err)
	}
	if task.ID == "" {
		t.Fatal("videos.create returned empty task id")
	}
	t.Logf("[PASS] videos.create -> id=%s", task.ID)

	// Poll up to 3 minutes
	deadline := time.Now().Add(3 * time.Minute)
	var result *meshapi.VideoTaskResponse
	for time.Now().Before(deadline) {
		result, err = client.Videos.Get(ctx, task.ID)
		if err != nil {
			t.Fatalf("videos.get: %v", err)
		}

		switch result.Status {
		case "succeeded", "failed", "expired", "cancelled":
			goto done
		}
		time.Sleep(10 * time.Second)
	}

done:
	if result == nil {
		t.Fatal("no result from polling")
	}
	if result.Status != "succeeded" {
		errDetail := ""
		if result.Error != nil {
			errDetail = result.Error.Message
		}
		t.Fatalf("expected status=succeeded, got %q (error=%s)", result.Status, errDetail)
	}
	if result.Content == nil || result.Content.VideoURL == nil || *result.Content.VideoURL == "" {
		t.Fatal("expected video_url on succeeded task")
	}
	t.Logf("[PASS] videos.create+poll -> id=%s video_url=%s", task.ID, *result.Content.VideoURL)
}
