package livetest

import (
	"context"
	"os"
	"testing"

	meshapi "meshapi-go-sdk"
)

const defaultTTSModel = "sarvam/bulbul:v2"

func ttsModel() string {
	if m := os.Getenv("MESHAPI_TTS_MODEL"); m != "" {
		return m
	}
	return defaultTTSModel
}

func TestAudio_Synthesize(t *testing.T) {
	client := newClient(t)
	model := ttsModel()
	audioBytes, err := client.Audio.Synthesize(context.Background(), meshapi.SpeechParams{
		Input: "Hello from MeshAPI audio test.",
		Model: &model,
	})
	if err != nil {
		t.Fatalf("audio.Synthesize error: %v", err)
	}
	if len(audioBytes) == 0 {
		t.Fatal("audio.Synthesize returned empty bytes")
	}
	t.Logf("[PASS] audio.Synthesize -> %d bytes", len(audioBytes))
}

func TestAudio_ListVoices(t *testing.T) {
	client := newClient(t)
	pageSize := 5
	voices, err := client.Audio.ListVoices(context.Background(), &meshapi.ListVoicesParams{
		PageSize: &pageSize,
	})
	if err != nil {
		t.Fatalf("audio.ListVoices error: %v", err)
	}
	if voices == nil {
		t.Fatal("audio.ListVoices returned nil")
	}
	t.Logf("[PASS] audio.ListVoices -> %v", voices)
}
