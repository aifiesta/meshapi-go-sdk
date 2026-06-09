package livetest

import (
	"context"
	"os"
	"testing"

	meshapi "meshapi-go-sdk"
	"github.com/stretchr/testify/require"
)

const defaultTTSModel = "sarvam/bulbul:v2"
const defaultSTTModel = "sarvam/saaras:v3"

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
	require.NoError(t, err)
	require.NotEmpty(t, audioBytes)
	t.Logf("[PASS] audio.Synthesize -> %d bytes", len(audioBytes))
}

func TestAudio_ListVoices(t *testing.T) {
	client := newClient(t)
	pageSize := 5
	voices, err := client.Audio.ListVoices(context.Background(), &meshapi.ListVoicesParams{
		PageSize: &pageSize,
	})
	require.NoError(t, err)
	require.NotNil(t, voices)
	t.Logf("[PASS] audio.ListVoices -> %v", voices)
}
