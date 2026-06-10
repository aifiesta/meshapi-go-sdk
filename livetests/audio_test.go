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

func TestAudio_STTFromTTS(t *testing.T) {
	client := newClient(t)
	model := ttsModel()
	audioBytes, err := client.Audio.Synthesize(context.Background(), meshapi.SpeechParams{
		Input: "Hello from MeshAPI audio test.",
		Model: &model,
	})
	if err != nil {
		t.Fatalf("TTS step failed: %v", err)
	}
	if len(audioBytes) == 0 {
		t.Fatal("TTS step returned empty bytes; cannot proceed to STT")
	}

	sttModel := os.Getenv("MESHAPI_STT_MODEL")
	if sttModel == "" {
		sttModel = "sarvam/saaras:v3"
	}
	result, err := client.Audio.Transcribe(context.Background(), audioBytes, "tts_output.wav", meshapi.TranscriptionParams{
		Model: sttModel,
	})
	if err != nil {
		t.Fatalf("audio.Transcribe error: %v", err)
	}
	if result == nil || result.Text == "" {
		t.Fatal("audio.Transcribe returned empty text")
	}
	t.Logf("[PASS] audio.Transcribe (via TTS audio) -> %q", result.Text)
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
