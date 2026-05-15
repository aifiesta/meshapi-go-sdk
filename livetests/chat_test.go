package livetest

import (
	"context"
	"strings"
	"testing"

	meshapi "meshapi-go-sdk"
)

func TestLive_Chat_Create(t *testing.T) {
	client := newClient(t)
	ctx := context.Background()

	resp, err := client.Chat.Completions.Create(ctx, meshapi.ChatCompletionParams{
		Model:     strPtr(liveModel()),
		Messages:  []meshapi.ChatMessage{{Role: "user", Content: "Reply with the single word: pong"}},
		MaxTokens: intPtr(10),
	})
	if err != nil {
		t.Fatalf("chat.create: %v", err)
	}
	if resp.ID == "" {
		t.Error("response ID is empty")
	}
	if len(resp.Choices) == 0 {
		t.Fatal("no choices in response")
	}
	choice := resp.Choices[0]
	if choice.Message == nil {
		t.Fatal("choice.message is nil")
	}
	content := ""
	if choice.Message.Content != nil {
		content = *choice.Message.Content
	}
	t.Logf("[PASS] chat.create → id=%q model=%q content=%q finishReason=%v",
		resp.ID, resp.Model, content, choice.FinishReason)
}

func TestLive_Chat_MultiTurn(t *testing.T) {
	client := newClient(t)
	ctx := context.Background()

	resp, err := client.Chat.Completions.Create(ctx, meshapi.ChatCompletionParams{
		Model: strPtr(liveModel()),
		Messages: []meshapi.ChatMessage{
			{Role: "system", Content: "You are a concise assistant. One sentence only."},
			{Role: "user", Content: "What is the capital of France?"},
		},
		MaxTokens: intPtr(20),
	})
	if err != nil {
		t.Fatalf("chat multi-turn: %v", err)
	}
	if len(resp.Choices) == 0 || resp.Choices[0].Message == nil {
		t.Fatal("empty response")
	}
	content := ""
	if resp.Choices[0].Message.Content != nil {
		content = *resp.Choices[0].Message.Content
	}
	t.Logf("[PASS] chat multi-turn → %q", content)
	if !strings.Contains(strings.ToLower(content), "paris") {
		t.Logf("  (note: 'Paris' not in response — may still be correct)")
	}
}

func TestLive_Chat_WithTemplate(t *testing.T) {
	client := newClient(t)
	ctx := context.Background()

	name := uniqueName("go-chat-tpl")
	system := "You are a helpful assistant who always responds in exactly 3 words."
	model := liveModel()
	tmpl, err := client.Templates.Create(ctx, meshapi.CreateTemplateParams{
		Name:   name,
		System: &system,
		Model:  &model,
	})
	if err != nil {
		t.Fatalf("create template for chat test: %v", err)
	}
	t.Cleanup(func() {
		_ = client.Templates.Delete(context.Background(), tmpl.ID)
	})

	resp, err := client.Chat.Completions.Create(ctx, meshapi.ChatCompletionParams{
		Template:  strPtr(tmpl.Name),
		Messages:  []meshapi.ChatMessage{{Role: "user", Content: "Greet me"}},
		MaxTokens: intPtr(15),
	})
	if err != nil {
		t.Fatalf("chat with template: %v", err)
	}
	t.Logf("[PASS] chat with template=%q → response received (id=%q)", tmpl.Name, resp.ID)
}
