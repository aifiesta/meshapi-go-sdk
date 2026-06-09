package meshapi

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// AudioResource provides access to /v1/audio/* endpoints.
type AudioResource struct {
	http *httpClient
}

// Synthesize sends POST /v1/audio/speech and returns raw audio bytes.
func (r *AudioResource) Synthesize(ctx context.Context, params SpeechParams) ([]byte, error) {
	return r.http.postBytes(ctx, "/v1/audio/speech", params)
}

// Transcribe sends POST /v1/audio/transcriptions as a multipart upload.
func (r *AudioResource) Transcribe(ctx context.Context, fileData []byte, filename string, params TranscriptionParams) (*TranscriptionResponse, error) {
	fields := transcriptionParamsToFields(params)
	var out TranscriptionResponse
	if err := r.http.postMultipart(ctx, "/v1/audio/transcriptions", fields, fileData, filename, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetTranscription sends GET /v1/audio/transcriptions/{transcription_id}.
func (r *AudioResource) GetTranscription(ctx context.Context, transcriptionID string) (map[string]any, error) {
	var out map[string]any
	if err := r.http.get(ctx, fmt.Sprintf("/v1/audio/transcriptions/%s", transcriptionID), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Translate sends POST /v1/audio/transcriptions/translate as a multipart upload.
func (r *AudioResource) Translate(ctx context.Context, fileData []byte, filename string, params *TranscriptionTranslateParams) (*TranscriptionResponse, error) {
	fields := map[string]string{}
	if params != nil {
		if params.Model != nil {
			fields["model"] = *params.Model
		}
		if params.Prompt != nil {
			fields["prompt"] = *params.Prompt
		}
	}
	var out TranscriptionResponse
	if err := r.http.postMultipart(ctx, "/v1/audio/transcriptions/translate", fields, fileData, filename, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListVoices sends GET /v1/audio/voices.
func (r *AudioResource) ListVoices(ctx context.Context, params *ListVoicesParams) (map[string]any, error) {
	qp := url.Values{}
	if params != nil {
		if params.NextPageToken != nil {
			qp.Set("next_page_token", *params.NextPageToken)
		}
		if params.PageSize != nil {
			qp.Set("page_size", strconv.Itoa(*params.PageSize))
		}
		if params.Search != nil {
			qp.Set("search", *params.Search)
		}
		if params.Sort != nil {
			qp.Set("sort", *params.Sort)
		}
		if params.SortDirection != nil {
			qp.Set("sort_direction", *params.SortDirection)
		}
		if params.VoiceType != nil {
			qp.Set("voice_type", *params.VoiceType)
		}
		if params.Category != nil {
			qp.Set("category", *params.Category)
		}
		if params.IncludeTotalCount != nil {
			qp.Set("include_total_count", strconv.FormatBool(*params.IncludeTotalCount))
		}
		for _, id := range params.VoiceIDs {
			qp.Add("voice_ids", id)
		}
	}
	var out map[string]any
	if err := r.http.get(ctx, "/v1/audio/voices", qp, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetVoice sends GET /v1/audio/voices/{voice_id}.
func (r *AudioResource) GetVoice(ctx context.Context, voiceID string) (map[string]any, error) {
	var out map[string]any
	if err := r.http.get(ctx, fmt.Sprintf("/v1/audio/voices/%s", voiceID), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func transcriptionParamsToFields(p TranscriptionParams) map[string]string {
	fields := map[string]string{
		"model": p.Model,
	}
	if p.LanguageCode != nil {
		fields["language_code"] = *p.LanguageCode
	}
	if p.TagAudioEvents != nil {
		fields["tag_audio_events"] = strconv.FormatBool(*p.TagAudioEvents)
	}
	if p.NumSpeakers != nil {
		fields["num_speakers"] = strconv.Itoa(*p.NumSpeakers)
	}
	if p.TimestampsGranularity != nil {
		fields["timestamps_granularity"] = *p.TimestampsGranularity
	}
	if p.Diarize != nil {
		fields["diarize"] = strconv.FormatBool(*p.Diarize)
	}
	if p.DiarizationThreshold != nil {
		fields["diarization_threshold"] = strconv.FormatFloat(*p.DiarizationThreshold, 'f', -1, 64)
	}
	if p.AdditionalFormats != nil {
		fields["additional_formats"] = *p.AdditionalFormats
	}
	if p.FileFormat != nil {
		fields["file_format"] = *p.FileFormat
	}
	if p.CloudStorageURL != nil {
		fields["cloud_storage_url"] = *p.CloudStorageURL
	}
	if p.SourceURL != nil {
		fields["source_url"] = *p.SourceURL
	}
	if p.Webhook != nil {
		fields["webhook"] = strconv.FormatBool(*p.Webhook)
	}
	if p.WebhookID != nil {
		fields["webhook_id"] = *p.WebhookID
	}
	if p.Temperature != nil {
		fields["temperature"] = strconv.FormatFloat(*p.Temperature, 'f', -1, 64)
	}
	if p.Seed != nil {
		fields["seed"] = strconv.Itoa(*p.Seed)
	}
	if p.UseMultiChannel != nil {
		fields["use_multi_channel"] = strconv.FormatBool(*p.UseMultiChannel)
	}
	if p.WebhookMetadata != nil {
		fields["webhook_metadata"] = *p.WebhookMetadata
	}
	if p.EntityDetection != nil {
		fields["entity_detection"] = *p.EntityDetection
	}
	if p.NoVerbatim != nil {
		fields["no_verbatim"] = strconv.FormatBool(*p.NoVerbatim)
	}
	if p.DetectSpeakerRoles != nil {
		fields["detect_speaker_roles"] = strconv.FormatBool(*p.DetectSpeakerRoles)
	}
	if p.EntityRedaction != nil {
		fields["entity_redaction"] = *p.EntityRedaction
	}
	if p.EntityRedactionMode != nil {
		fields["entity_redaction_mode"] = *p.EntityRedactionMode
	}
	if p.WithTimestamps != nil {
		fields["with_timestamps"] = strconv.FormatBool(*p.WithTimestamps)
	}
	if p.DebugMode != nil {
		fields["debug_mode"] = strconv.FormatBool(*p.DebugMode)
	}
	return fields
}
