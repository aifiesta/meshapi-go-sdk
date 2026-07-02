package livetest

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	meshapi "meshapi-go-sdk"
)

func weatherTool() meshapi.Tool {
	return meshapi.Tool{
		Type: "function",
		Function: meshapi.ToolFunction{
			Name:        "get_weather",
			Description: strPtr("Get the current weather for a city."),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"city": map[string]interface{}{"type": "string"},
				},
				"required": []string{"city"},
			},
		},
	}
}

// skipIfToolsUnsupported skips when the model rejects tool calling (400/501
// capability error) rather than failing the suite.
func skipIfToolsUnsupported(t *testing.T, err error) {
	t.Helper()
	var apiErr *meshapi.MeshAPIError
	if errors.As(err, &apiErr) && (apiErr.Status == 400 || apiErr.Status == 501) {
		switch apiErr.Code {
		case "not_implemented", "model_capability_not_supported":
			t.Skipf("model does not support tool calling: %s", apiErr.Code)
		}
	}
}

func TestLive_ToolCalling_RoundTrip(t *testing.T) {
	client := newClient(t)
	ctx := context.Background()

	messages := []meshapi.ChatMessage{{Role: "user", Content: "What is the weather in Paris?"}}
	resp, err := client.Chat.Completions.Create(ctx, meshapi.ChatCompletionParams{
		Model:    strPtr(liveModel()),
		Messages: messages,
		Tools:    []meshapi.Tool{weatherTool()},
		// Force the call so the round-trip is deterministic.
		ToolChoice: meshapi.ToolChoiceObject{
			Type:     "function",
			Function: meshapi.ToolChoiceFunction{Name: "get_weather"},
		},
		MaxTokens: intPtr(100),
	})
	if err != nil {
		skipIfToolsUnsupported(t, err)
		t.Fatalf("chat.create (tools): %v", err)
	}

	if len(resp.Choices) == 0 || resp.Choices[0].Message == nil {
		t.Fatal("no message in response")
	}
	msg := resp.Choices[0].Message
	if len(msg.ToolCalls) == 0 {
		t.Fatal("forced tool_choice must produce a tool_calls array")
	}
	call := msg.ToolCalls[0]
	if call.ID == "" || call.Function.Name != "get_weather" {
		t.Fatalf("unexpected tool call: id=%q name=%q", call.ID, call.Function.Name)
	}
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(call.Function.Arguments), &args); err != nil {
		t.Fatalf("tool call arguments not valid JSON: %v", err)
	}
	if _, ok := args["city"]; !ok {
		t.Errorf("expected a 'city' argument, got %v", args)
	}

	// Feed the tool result back and expect a final natural-language answer.
	assistantMsg := meshapi.ChatMessage{Role: "assistant", ToolCalls: msg.ToolCalls}
	if msg.Content != nil {
		assistantMsg.Content = *msg.Content
	}
	messages = append(messages, assistantMsg, meshapi.ChatMessage{
		Role:       "tool",
		ToolCallID: strPtr(call.ID),
		Content:    `{"temperature": 22, "unit": "celsius", "description": "Sunny"}`,
	})

	final, err := client.Chat.Completions.Create(ctx, meshapi.ChatCompletionParams{
		Model:     strPtr(liveModel()),
		Messages:  messages,
		Tools:     []meshapi.Tool{weatherTool()},
		MaxTokens: intPtr(100),
	})
	if err != nil {
		t.Fatalf("chat.create (final): %v", err)
	}
	if len(final.Choices) == 0 || final.Choices[0].Message == nil || final.Choices[0].Message.Content == nil || *final.Choices[0].Message.Content == "" {
		t.Fatal("expected a final assistant answer after the tool result")
	}
}

func TestLive_ToolCalling_AutoWellFormed(t *testing.T) {
	client := newClient(t)
	ctx := context.Background()

	resp, err := client.Chat.Completions.Create(ctx, meshapi.ChatCompletionParams{
		Model:      strPtr(liveModel()),
		Messages:   []meshapi.ChatMessage{{Role: "user", Content: "What is the weather in Tokyo?"}},
		Tools:      []meshapi.Tool{weatherTool()},
		ToolChoice: "auto",
		MaxTokens:  intPtr(100),
	})
	if err != nil {
		skipIfToolsUnsupported(t, err)
		t.Fatalf("chat.create (auto tools): %v", err)
	}
	if len(resp.Choices) == 0 || resp.Choices[0].Message == nil {
		t.Fatal("no message in response")
	}
	msg := resp.Choices[0].Message
	if len(msg.ToolCalls) > 0 {
		for _, call := range msg.ToolCalls {
			if call.ID == "" || call.Function.Name == "" {
				t.Errorf("malformed tool call: %+v", call)
			}
			var args map[string]interface{}
			if err := json.Unmarshal([]byte(call.Function.Arguments), &args); err != nil {
				t.Errorf("tool call arguments not valid JSON: %v", err)
			}
		}
	} else if msg.Content == nil || *msg.Content == "" {
		t.Error("if no tool was called, the model must reply with content")
	}
}
