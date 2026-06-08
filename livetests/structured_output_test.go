package livetest

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	meshapi "meshapi-go-sdk"
)

var structuredOutputModels = []string{
	"openai/gpt-4o-mini",
	"google/gemini-3-flash-preview",
}

func TestLive_StructuredOutput_Fields(t *testing.T) {
	client := newClient(t)
	ctx := context.Background()

	schema := map[string]interface{}{
		"type": "json_schema",
		"json_schema": map[string]interface{}{
			"name": "country_info",
			"schema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"capital": map[string]interface{}{"type": "string"},
					"country": map[string]interface{}{"type": "string"},
				},
				"required":             []string{"capital", "country"},
				"additionalProperties": false,
			},
		},
	}

	for _, model := range structuredOutputModels {
		model := model
		t.Run(model, func(t *testing.T) {
			t.Parallel()
			resp, err := client.Chat.Completions.Create(ctx, meshapi.ChatCompletionParams{
				Model: strPtr(model),
				Messages: []meshapi.ChatMessage{
					{Role: "user", Content: "What is the capital of France? Use the provided schema."},
				},
				ResponseFormat: schema,
				MaxTokens:      intPtr(1000),
			})
			if err != nil {
				t.Fatalf("[%s] chat.create: %v", model, err)
			}
			if len(resp.Choices) == 0 || resp.Choices[0].Message == nil || resp.Choices[0].Message.Content == nil {
				t.Fatalf("[%s] empty response", model)
			}

			content := *resp.Choices[0].Message.Content
			var data map[string]interface{}
			if err := json.Unmarshal([]byte(content), &data); err != nil {
				t.Fatalf("[%s] response is not valid JSON: %v\ncontent: %s", model, err, content)
			}

			capital, ok := data["capital"].(string)
			if !ok {
				t.Fatalf("[%s] missing or non-string 'capital' field: %v", model, data)
			}
			if _, ok := data["country"].(string); !ok {
				t.Fatalf("[%s] missing or non-string 'country' field: %v", model, data)
			}
			if !strings.Contains(strings.ToLower(capital), "paris") {
				t.Errorf("[%s] expected Paris as capital, got: %q", model, capital)
			}
			t.Logf("[PASS] %s structured output → %v", model, data)
		})
	}
}

func TestLive_StructuredOutput_FinishReason(t *testing.T) {
	client := newClient(t)
	ctx := context.Background()

	schema := map[string]interface{}{
		"type": "json_schema",
		"json_schema": map[string]interface{}{
			"name": "planet_info",
			"schema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name":              map[string]interface{}{"type": "string"},
					"position_from_sun": map[string]interface{}{"type": "integer"},
				},
				"required":             []string{"name", "position_from_sun"},
				"additionalProperties": false,
			},
		},
	}

	for _, model := range structuredOutputModels {
		model := model
		t.Run(model, func(t *testing.T) {
			t.Parallel()
			resp, err := client.Chat.Completions.Create(ctx, meshapi.ChatCompletionParams{
				Model: strPtr(model),
				Messages: []meshapi.ChatMessage{
					{Role: "user", Content: "Name any planet in our solar system. Use the provided schema."},
				},
				ResponseFormat: schema,
				MaxTokens:      intPtr(1000),
			})
			if err != nil {
				t.Fatalf("[%s] chat.create: %v", model, err)
			}
			if len(resp.Choices) == 0 {
				t.Fatalf("[%s] no choices", model)
			}

			choice := resp.Choices[0]
			if choice.FinishReason == nil || *choice.FinishReason != "stop" {
				t.Errorf("[%s] expected finish_reason 'stop', got %v", model, choice.FinishReason)
			}

			content := ""
			if choice.Message != nil && choice.Message.Content != nil {
				content = *choice.Message.Content
			}
			var data map[string]interface{}
			if err := json.Unmarshal([]byte(content), &data); err != nil {
				t.Fatalf("[%s] response is not valid JSON: %v\ncontent: %s", model, err, content)
			}
			if _, ok := data["name"].(string); !ok {
				t.Fatalf("[%s] missing or non-string 'name' field: %v", model, data)
			}
			if _, ok := data["position_from_sun"].(float64); !ok {
				t.Fatalf("[%s] missing or non-numeric 'position_from_sun' field: %v", model, data)
			}
			t.Logf("[PASS] %s structured output finish_reason → %v", model, data)
		})
	}
}
