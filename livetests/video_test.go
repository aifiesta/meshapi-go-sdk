package livetest

import (
	"context"
	"os"
	"testing"

	meshapi "meshapi-go-sdk"
	"github.com/stretchr/testify/require"
)

func TestVideo_List(t *testing.T) {
	client := newClient(t)
	limit := 5
	listing, err := client.Videos.List(context.Background(), &meshapi.ListVideoGenerationsParams{
		Limit: &limit,
	})
	require.NoError(t, err)
	require.NotNil(t, listing)
	t.Logf("[PASS] Videos.List -> total=%d items=%d", listing.Total, len(listing.Data))
}

func TestVideo_GenerateAndRetrieve(t *testing.T) {
	model := os.Getenv("MESHAPI_VIDEO_GEN_MODEL")
	if model == "" {
		model = "byteplus/dreamina-seedance-2-0"
	}
	client := newClient(t)
	text := "A serene mountain lake at sunrise"
	resp, err := client.Videos.Generate(context.Background(), meshapi.VideoGenerationParams{
		Model: model,
		Content: []meshapi.VideoContentItem{
			{Type: "text", Text: &text},
		},
	})
	require.NoError(t, err)
	require.NotEmpty(t, resp.ID)
	t.Logf("[PASS] Videos.Generate -> task_id=%s", resp.ID)

	task, err := client.Videos.Retrieve(context.Background(), resp.ID)
	require.NoError(t, err)
	require.Equal(t, resp.ID, task.ID)
	t.Logf("[PASS] Videos.Retrieve -> status=%s", task.Status)
}
