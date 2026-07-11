package meshapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"
)

type parseCountry struct {
	Country string `json:"country"`
	Capital string `json:"capital"`
}

func chatPayload(content string) string {
	b, _ := json.Marshal(map[string]interface{}{
		"id": "c1", "object": "chat.completion", "created": 0, "model": "m",
		"choices": []map[string]interface{}{{
			"index":         0,
			"message":       map[string]interface{}{"role": "assistant", "content": content},
			"finish_reason": "stop",
		}},
	})
	return string(b)
}

// newParseServer serves the given bodies in order and records request bodies.
func newParseServer(t *testing.T, bodies ...string) (*Client, *[]map[string]interface{}) {
	t.Helper()
	var mu sync.Mutex
	calls := []map[string]interface{}{}
	idx := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&reqBody)
		mu.Lock()
		calls = append(calls, reqBody)
		i := idx
		idx++
		mu.Unlock()
		if i >= len(bodies) {
			http.Error(w, "no more mock responses", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(bodies[i]))
	}))
	t.Cleanup(srv.Close)
	client := New(Config{BaseURL: srv.URL, Token: "rsk_test"})
	return client, &calls
}

func parseParams() ChatCompletionParams {
	m := "openai/gpt-4o-mini"
	return ChatCompletionParams{Model: &m, Messages: []ChatMessage{{Role: "user", Content: "?"}}}
}

// ── reflector ─────────────────────────────────────────────────────────────────

func TestJSONSchemaForType(t *testing.T) {
	type addr struct {
		City string `json:"city"`
	}
	type person struct {
		Name    string   `json:"name"`
		Age     int      `json:"age"`
		Nick    *string  `json:"nick,omitempty"` // pointer -> optional
		Tags    []string `json:"tags"`
		Address addr     `json:"address"` // nested
		hidden  string   // unexported -> skipped
	}
	s := jsonSchemaForType(reflect.TypeOf(person{}))
	if s["type"] != "object" {
		t.Fatalf("type = %v", s["type"])
	}
	if s["additionalProperties"] != false {
		t.Fatalf("additionalProperties = %v", s["additionalProperties"])
	}
	props := s["properties"].(map[string]interface{})
	if props["name"].(map[string]interface{})["type"] != "string" {
		t.Errorf("name type = %v", props["name"])
	}
	if props["age"].(map[string]interface{})["type"] != "integer" {
		t.Errorf("age type = %v", props["age"])
	}
	if props["tags"].(map[string]interface{})["type"] != "array" {
		t.Errorf("tags type = %v", props["tags"])
	}
	if props["address"].(map[string]interface{})["type"] != "object" {
		t.Errorf("address type = %v", props["address"])
	}
	if _, ok := props["hidden"]; ok {
		t.Error("unexported field leaked into schema")
	}
	req := s["required"].([]string)
	// name, age, tags, address required; nick (pointer/omitempty) not.
	joined := strings.Join(req, ",")
	for _, want := range []string{"name", "age", "tags", "address"} {
		if !strings.Contains(joined, want) {
			t.Errorf("required missing %q (got %v)", want, req)
		}
	}
	if strings.Contains(joined, "nick") {
		t.Errorf("pointer field should be optional, got required %v", req)
	}
}

// ── Parse ─────────────────────────────────────────────────────────────────────

func TestParse_SuccessSendsSchema(t *testing.T) {
	client, calls := newParseServer(t, chatPayload(`{"country":"France","capital":"Paris"}`))
	got, err := Parse[parseCountry](context.Background(), client.Chat.Completions, parseParams())
	if err != nil {
		t.Fatal(err)
	}
	if got.Capital != "Paris" || got.Country != "France" {
		t.Fatalf("got %+v", got)
	}
	rf := (*calls)[0]["response_format"].(map[string]interface{})
	if rf["type"] != "json_schema" {
		t.Fatalf("response_format.type = %v", rf["type"])
	}
	js := rf["json_schema"].(map[string]interface{})
	if js["name"] != "response" {
		t.Errorf("schema name = %v", js["name"])
	}
	if _, ok := js["schema"].(map[string]interface{})["properties"]; !ok {
		t.Error("derived schema missing properties")
	}
}

func TestParse_ProseHintsModelSupport(t *testing.T) {
	client, _ := newParseServer(t, chatPayload("Sure! The capital of France is Paris."))
	_, err := Parse[parseCountry](context.Background(), client.Chat.Completions, parseParams())
	var soe *StructuredOutputError
	if !errors.As(err, &soe) {
		t.Fatalf("want *StructuredOutputError, got %T: %v", err, err)
	}
	if !strings.Contains(soe.Message, "does not support structured outputs") {
		t.Errorf("message missing model-support hint: %s", soe.Message)
	}
	if !strings.Contains(soe.Message, "app.meshapi.ai") || !strings.Contains(soe.Message, "/models") {
		t.Errorf("message missing Models link: %s", soe.Message)
	}
	var se *json.SyntaxError
	if !errors.As(err, &se) {
		t.Errorf("cause should be *json.SyntaxError, got %T", soe.Cause)
	}
}

func TestParse_ShapeMismatch(t *testing.T) {
	client, _ := newParseServer(t, chatPayload(`{"country":123,"capital":"Paris"}`)) // country: int, want string
	_, err := Parse[parseCountry](context.Background(), client.Chat.Completions, parseParams())
	var soe *StructuredOutputError
	if !errors.As(err, &soe) {
		t.Fatalf("want *StructuredOutputError, got %T", err)
	}
	if !strings.Contains(soe.Message, "did not match the requested type") {
		t.Errorf("want shape-mismatch message, got: %s", soe.Message)
	}
}

func TestParse_DefaultNoRetry(t *testing.T) {
	client, calls := newParseServer(t, chatPayload("not json"))
	_, err := Parse[parseCountry](context.Background(), client.Chat.Completions, parseParams())
	if err == nil {
		t.Fatal("want error")
	}
	if len(*calls) != 1 {
		t.Fatalf("want 1 call, got %d", len(*calls))
	}
}

func TestParse_RetryRecoversAndAppendsCorrection(t *testing.T) {
	client, calls := newParseServer(t,
		chatPayload("not json"),                               // bad
		chatPayload(`{"country":"France","capital":"Paris"}`), // good
	)
	got, err := Parse[parseCountry](context.Background(), client.Chat.Completions, parseParams(), WithMaxRetries(1))
	if err != nil {
		t.Fatal(err)
	}
	if got.Capital != "Paris" {
		t.Fatalf("got %+v", got)
	}
	if len(*calls) != 2 {
		t.Fatalf("want 2 calls, got %d", len(*calls))
	}
	msgs := (*calls)[1]["messages"].([]interface{})
	if len(msgs) != 3 { // original user + assistant(bad) + user(correction)
		t.Fatalf("want 3 messages on retry, got %d", len(msgs))
	}
}

func TestParse_WithSchemaOverride(t *testing.T) {
	type box struct {
		X int `json:"x"`
	}
	client, calls := newParseServer(t, chatPayload(`{"x":7}`))
	got, err := Parse[box](
		context.Background(), client.Chat.Completions, parseParams(),
		WithSchema(map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{"x": map[string]interface{}{"type": "integer"}},
		}),
		WithSchemaName("custom"),
	)
	if err != nil {
		t.Fatal(err)
	}
	if got.X != 7 {
		t.Fatalf("got %+v", got)
	}
	js := (*calls)[0]["response_format"].(map[string]interface{})["json_schema"].(map[string]interface{})
	if js["name"] != "custom" {
		t.Errorf("schema name = %v", js["name"])
	}
}
