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
	s, err := jsonSchemaForType(reflect.TypeOf(person{}))
	if err != nil {
		t.Fatalf("jsonSchemaForType: %v", err)
	}
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

func TestSchema_MapKeys(t *testing.T) {
	strMap, err := jsonSchemaForType(reflect.TypeOf(map[string]int{}))
	if err != nil {
		t.Fatalf("map[string]int: %v", err)
	}
	ap, ok := strMap["additionalProperties"].(map[string]interface{})
	if !ok {
		t.Fatalf("map[string]int should have additionalProperties, got %v", strMap)
	}
	if ap["type"] != "integer" {
		t.Errorf("map[string]int values should be integer, got %v", ap)
	}

	// Integer keys decode from JSON object keys, so they keep a value schema.
	intMap, err := jsonSchemaForType(reflect.TypeOf(map[int]string{}))
	if err != nil {
		t.Fatalf("map[int]string: %v", err)
	}
	if _, present := intMap["additionalProperties"]; !present {
		t.Errorf("map[int]string (json-decodable keys) should have additionalProperties, got %v", intMap)
	}

	// A defined type whose underlying kind is string still counts as string-keyed.
	type keyT string
	defMap, err := jsonSchemaForType(reflect.TypeOf(map[keyT]int{}))
	if err != nil {
		t.Fatalf("map[keyT]int: %v", err)
	}
	if _, present := defMap["additionalProperties"]; !present {
		t.Errorf("map[keyT]int (underlying string) should have additionalProperties, got %v", defMap)
	}

	// TextUnmarshaler keys are decodable too.
	tuMap, err := jsonSchemaForType(reflect.TypeOf(map[hexColor]int{}))
	if err != nil {
		t.Fatalf("map[hexColor]int (TextUnmarshaler key): %v", err)
	}
	if _, present := tuMap["additionalProperties"]; !present {
		t.Errorf("TextUnmarshaler-keyed map should have additionalProperties, got %v", tuMap)
	}
}

func TestSchema_UnsupportedMapKeysError(t *testing.T) {
	for _, typ := range []reflect.Type{
		reflect.TypeOf(map[float64]string{}),
		reflect.TypeOf(map[bool]string{}),
	} {
		if _, err := jsonSchemaForType(typ); err == nil {
			t.Errorf("%v: expected error for json-undecodable map key, got none", typ)
		} else if !strings.Contains(err.Error(), "unsupported map key type") {
			t.Errorf("%v: error should name the unsupported key, got %v", typ, err)
		}
	}

	// Nested occurrences surface too.
	type holder struct {
		Data map[float64]string `json:"data"`
	}
	if _, err := jsonSchemaForType(reflect.TypeOf(holder{})); err == nil {
		t.Error("nested unsupported map key should error")
	}
}

func TestParse_UnsupportedMapKeyFailsBeforeRequest(t *testing.T) {
	type bad struct {
		Data map[bool]string `json:"data"`
	}
	client, calls := newParseServer(t, chatPayload(`{}`))
	_, err := Parse[bad](context.Background(), client.Chat.Completions, parseParams())
	if err == nil || !strings.Contains(err.Error(), "unsupported map key type") {
		t.Fatalf("expected schema error, got %v", err)
	}
	if len(*calls) != 0 {
		t.Fatalf("no HTTP request should be made, got %d", len(*calls))
	}
}

// hexColor decodes from a JSON string via encoding.TextUnmarshaler even though its
// reflect kind is struct.
type hexColor struct {
	R, G, B uint8
}

func (c *hexColor) UnmarshalText([]byte) error { return nil }

// rawThing decodes from arbitrary JSON via json.Unmarshaler.
type rawThing struct {
	N int
}

func (r *rawThing) UnmarshalJSON([]byte) error { return nil }

// dualDecoder implements both decoder interfaces; encoding/json prefers
// UnmarshalJSON, so the schema must not constrain it to a string.
type dualDecoder struct {
	N int
}

func (d *dualDecoder) UnmarshalJSON([]byte) error { return nil }
func (d *dualDecoder) UnmarshalText([]byte) error { return nil }

func TestSchema_CustomDecoders(t *testing.T) {
	// TextUnmarshaler -> string schema, not a struct object.
	s, err := jsonSchemaForType(reflect.TypeOf(hexColor{}))
	if err != nil {
		t.Fatalf("hexColor: %v", err)
	}
	if s["type"] != "string" {
		t.Fatalf("type implementing TextUnmarshaler should map to string schema, got %v", s)
	}
	if _, ok := s["properties"]; ok {
		t.Errorf("custom decoder should not emit struct properties, got %v", s)
	}

	// json.Unmarshaler -> unconstrained {}.
	r, err := jsonSchemaForType(reflect.TypeOf(rawThing{}))
	if err != nil {
		t.Fatalf("rawThing: %v", err)
	}
	if len(r) != 0 {
		t.Errorf("type implementing json.Unmarshaler should be unconstrained, got %v", r)
	}

	// Both interfaces -> UnmarshalJSON wins (as in encoding/json), so the
	// schema must stay unconstrained, not "string".
	d, err := jsonSchemaForType(reflect.TypeOf(dualDecoder{}))
	if err != nil {
		t.Fatalf("dualDecoder: %v", err)
	}
	if len(d) != 0 {
		t.Errorf("dual-decoder type should follow UnmarshalJSON (unconstrained), got %v", d)
	}
}

// ── embedded field dominance (encoding/json rules) ────────────────────────────

type embA struct {
	Shared string
	OnlyA  string `json:"only_a"`
}

type embB struct {
	Shared string
}

func TestSchema_EmbeddedConflictDropsAmbiguousName(t *testing.T) {
	// Two embedded structs expose "Shared" at the same depth with equal tag
	// status — encoding/json ignores the value, so the schema omits the name.
	type both struct {
		embA
		embB
	}
	s, err := jsonSchemaForType(reflect.TypeOf(both{}))
	if err != nil {
		t.Fatalf("jsonSchemaForType: %v", err)
	}
	props := s["properties"].(map[string]interface{})
	if _, ok := props["Shared"]; ok {
		t.Errorf("ambiguous embedded name should be dropped, got %v", props)
	}
	if _, ok := props["only_a"]; !ok {
		t.Errorf("non-conflicting embedded field should survive, got %v", props)
	}
	if req, ok := s["required"].([]string); ok {
		for _, r := range req {
			if r == "Shared" {
				t.Errorf("dropped name must not be required, got %v", req)
			}
		}
	}
	// Sanity: encoding/json agrees — the ambiguous key is ignored on decode.
	var b both
	if err := json.Unmarshal([]byte(`{"Shared":"x","only_a":"y"}`), &b); err != nil {
		t.Fatal(err)
	}
	if b.embA.Shared != "" || b.embB.Shared != "" {
		t.Fatalf("expected encoding/json to ignore ambiguous key, got %+v", b)
	}
}

func TestSchema_OuterFieldShadowsEmbedded(t *testing.T) {
	// The outer (shallower) field wins regardless of declaration order.
	type outer struct {
		embA
		Shared int // depth 0 beats embA's depth 1
	}
	s, err := jsonSchemaForType(reflect.TypeOf(outer{}))
	if err != nil {
		t.Fatalf("jsonSchemaForType: %v", err)
	}
	props := s["properties"].(map[string]interface{})
	if got := props["Shared"].(map[string]interface{})["type"]; got != "integer" {
		t.Errorf("outer field should shadow embedded (integer), got %v", got)
	}
}

func TestSchema_TaggedBeatsUntaggedAtSameDepth(t *testing.T) {
	// Same JSON name ("Pick") at the same depth; only one is json-tagged.
	// encoding/json lets the tagged field win — the schema must agree.
	type viaTag struct {
		N int `json:"Pick"`
	}
	type viaName struct {
		Pick string
	}
	type mix struct {
		viaTag
		viaName
	}
	s, err := jsonSchemaForType(reflect.TypeOf(mix{}))
	if err != nil {
		t.Fatalf("jsonSchemaForType: %v", err)
	}
	props := s["properties"].(map[string]interface{})
	got, ok := props["Pick"].(map[string]interface{})
	if !ok {
		t.Fatalf("tagged field should win, got %v", props)
	}
	if got["type"] != "integer" {
		t.Errorf("winner should be the tagged int field, got %v", got)
	}
	// Sanity: encoding/json picks the tagged field on decode.
	var m mix
	if err := json.Unmarshal([]byte(`{"Pick":7}`), &m); err != nil {
		t.Fatal(err)
	}
	if m.viaTag.N != 7 || m.viaName.Pick != "" {
		t.Fatalf("expected tagged field to receive the value, got %+v", m)
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

func TestParse_NullResultIsError(t *testing.T) {
	client, calls := newParseServer(t, chatPayload("null"))
	got, err := Parse[parseCountry](context.Background(), client.Chat.Completions, parseParams())
	var soe *StructuredOutputError
	if !errors.As(err, &soe) {
		t.Fatalf("literal JSON null should yield *StructuredOutputError, got %T: %v", err, err)
	}
	if got != (parseCountry{}) {
		t.Errorf("want zero value alongside error, got %+v", got)
	}
	if len(*calls) != 1 {
		t.Fatalf("default (no retry) should make 1 call, got %d", len(*calls))
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
