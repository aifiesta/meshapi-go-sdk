package meshapi

import (
	"context"
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// modelsURL points users at their dashboard's model list when a model turns out
// not to support structured outputs.
const modelsURL = "https://app.meshapi.ai/org/<your-org-id>/models"

// ParseOption configures Parse.
type ParseOption func(*parseConfig)

type parseConfig struct {
	maxRetries int
	schema     map[string]interface{}
	schemaName string
}

// WithMaxRetries re-prompts the model with the decode error up to n times
// (default 0). Each retry is a billed inference call.
func WithMaxRetries(n int) ParseOption { return func(c *parseConfig) { c.maxRetries = n } }

// WithSchema overrides the auto-derived JSON schema with an explicit one (a bare
// JSON schema object, e.g. map[string]interface{}{"type": "object", ...}).
func WithSchema(schema map[string]interface{}) ParseOption {
	return func(c *parseConfig) { c.schema = schema }
}

// WithSchemaName sets the json_schema name (default "response").
func WithSchemaName(name string) ParseOption {
	return func(c *parseConfig) { c.schemaName = name }
}

// Parse runs a structured (JSON-schema-constrained) chat completion and decodes
// the reply into T. The JSON schema is derived from T by reflection — define a
// struct with `json` tags and Parse builds the schema and the typed result.
// Override the schema with WithSchema. Non-streaming.
//
// With WithMaxRetries(n), a reply that fails to decode is fed back to the model
// with the error appended, up to n times. Returns *StructuredOutputError when it
// still can't be decoded — most often because the model does not support
// structured outputs (it returned plain text instead of JSON).
//
// Note: Go's json.Unmarshal does not enforce required fields — a JSON object
// missing a field decodes to that field's zero value rather than an error. Type
// mismatches and non-JSON prose are caught (and drive retries / the error).
//
// Because Go methods cannot have type parameters, Parse is a package-level
// function taking the completions resource:
//
//	type Country struct {
//		Country string `json:"country"`
//		Capital string `json:"capital"`
//	}
//	c, err := meshapi.Parse[Country](ctx, client.Chat.Completions, params)
func Parse[T any](
	ctx context.Context,
	r *CompletionsResource,
	params ChatCompletionParams,
	opts ...ParseOption,
) (T, error) {
	var zero T

	cfg := parseConfig{}
	for _, o := range opts {
		o(&cfg)
	}
	name := cfg.schemaName
	if name == "" {
		name = "response"
	}
	schema := cfg.schema
	if schema == nil {
		schema = jsonSchemaForType(reflect.TypeOf((*T)(nil)).Elem())
	}
	params.ResponseFormat = map[string]interface{}{
		"type": "json_schema",
		"json_schema": map[string]interface{}{
			"name":   name,
			"schema": schema,
		},
	}

	model := ""
	if params.Model != nil {
		model = *params.Model
	}

	attempt := 0
	for {
		resp, err := r.Create(ctx, params)
		if err != nil {
			return zero, err // transport / API error — surface as-is
		}
		content := chatContent(resp)

		var result T
		var decErr error
		if strings.TrimSpace(content) == "null" {
			// A literal JSON null decodes into a non-pointer T without error,
			// silently yielding a zero value. Treat it as "no data" and route it
			// through the same failure/retry path as non-JSON output.
			decErr = errNullResult
		} else {
			decErr = json.Unmarshal([]byte(content), &result)
		}
		if decErr == nil {
			return result, nil
		}
		if attempt >= cfg.maxRetries {
			return zero, newStructuredOutputError(model, isNotJSON(decErr), decErr)
		}
		attempt++
		params.Messages = appendMessages(
			params.Messages,
			ChatMessage{Role: "assistant", Content: content},
			ChatMessage{Role: "user", Content: correctionPrompt(decErr)},
		)
	}
}

// appendMessages returns a new slice so retries never mutate the caller's slice.
func appendMessages(base []ChatMessage, extra ...ChatMessage) []ChatMessage {
	out := make([]ChatMessage, len(base), len(base)+len(extra))
	copy(out, base)
	return append(out, extra...)
}

func chatContent(resp *ChatCompletionResponse) string {
	if resp == nil || len(resp.Choices) == 0 {
		return ""
	}
	msg := resp.Choices[0].Message
	if msg == nil || msg.Content == nil {
		return ""
	}
	return *msg.Content
}

// errNullResult marks a reply that was a literal JSON null — syntactically valid
// but carrying no object to decode into T. It is treated as "not JSON / no data".
var errNullResult = errors.New("the model returned a JSON null instead of an object")

// isNotJSON reports whether the decode failed because the content wasn't JSON at
// all (prose / empty / a bare null), as opposed to valid JSON that didn't fit the
// type.
func isNotJSON(err error) bool {
	if errors.Is(err, errNullResult) {
		return true
	}
	var se *json.SyntaxError
	return errors.As(err, &se)
}

func correctionPrompt(err error) string {
	return fmt.Sprintf(
		"Your previous response failed schema validation: %v. Return ONLY a JSON object that "+
			"matches the requested schema, with no prose, markdown, or code fences.",
		err,
	)
}

// ---------------------------------------------------------------------------
// Reflection: Go type -> JSON schema
// ---------------------------------------------------------------------------

var (
	timeType            = reflect.TypeOf(time.Time{})
	textUnmarshalerType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
	jsonUnmarshalerType = reflect.TypeOf((*json.Unmarshaler)(nil)).Elem()
)

// jsonSchemaForType derives a JSON schema from a Go type by reflection. It is
// best-effort: it follows encoding/json's json-tag, embedding, and custom-decoder
// conventions closely but does not reproduce every rule. In particular, when
// embedded structs contribute fields with conflicting JSON names, the winner here
// is not disambiguated the way encoding/json does at (un)marshal time. Pass
// WithSchema to supply an explicit schema for such types.
func jsonSchemaForType(t reflect.Type) map[string]interface{} {
	return schemaFor(t, map[reflect.Type]bool{})
}

func schemaFor(t reflect.Type, seen map[reflect.Type]bool) map[string]interface{} {
	for t != nil && t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t == nil {
		return map[string]interface{}{}
	}
	if t == timeType {
		return map[string]interface{}{"type": "string", "format": "date-time"}
	}
	// Types with a custom decoder don't follow their reflect kind on the wire.
	// A TextUnmarshaler decodes from a JSON string; a bare json.Unmarshaler can
	// accept any JSON, so leave it unconstrained. (time.Time is handled above so
	// it keeps its date-time format.)
	if reflect.PtrTo(t).Implements(textUnmarshalerType) {
		return map[string]interface{}{"type": "string"}
	}
	if reflect.PtrTo(t).Implements(jsonUnmarshalerType) {
		return map[string]interface{}{}
	}

	switch t.Kind() {
	case reflect.String:
		return map[string]interface{}{"type": "string"}
	case reflect.Bool:
		return map[string]interface{}{"type": "boolean"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return map[string]interface{}{"type": "integer"}
	case reflect.Float32, reflect.Float64:
		return map[string]interface{}{"type": "number"}
	case reflect.Slice, reflect.Array:
		if t.Elem().Kind() == reflect.Uint8 { // []byte marshals to a base64 string
			return map[string]interface{}{"type": "string"}
		}
		return map[string]interface{}{"type": "array", "items": schemaFor(t.Elem(), seen)}
	case reflect.Map:
		// Only string-keyed maps decode from a JSON object with arbitrary keys.
		// For non-string keys, additionalProperties would describe values json
		// can't place, so emit a plain object instead.
		if t.Key().Kind() == reflect.String {
			return map[string]interface{}{
				"type":                 "object",
				"additionalProperties": schemaFor(t.Elem(), seen),
			}
		}
		return map[string]interface{}{"type": "object"}
	case reflect.Struct:
		if seen[t] {
			return map[string]interface{}{"type": "object"} // break recursive types
		}
		seen[t] = true
		defer delete(seen, t)
		return structSchema(t, seen)
	default: // Interface, Chan, Func, etc. -> unconstrained
		return map[string]interface{}{}
	}
}

func structSchema(t reflect.Type, seen map[reflect.Type]bool) map[string]interface{} {
	props := map[string]interface{}{}
	required := []string{}

	var walk func(rt reflect.Type)
	walk = func(rt reflect.Type) {
		for i := 0; i < rt.NumField(); i++ {
			f := rt.Field(i)
			tag := f.Tag.Get("json")
			if tag == "-" {
				continue
			}
			// Flatten anonymous embedded structs that have no json tag.
			if f.Anonymous && tag == "" {
				ft := f.Type
				for ft.Kind() == reflect.Ptr {
					ft = ft.Elem()
				}
				if ft.Kind() == reflect.Struct {
					walk(ft)
					continue
				}
			}
			if f.PkgPath != "" { // unexported
				continue
			}

			name := f.Name
			omitempty := false
			if tag != "" {
				parts := strings.Split(tag, ",")
				if parts[0] != "" {
					name = parts[0]
				}
				for _, p := range parts[1:] {
					if p == "omitempty" {
						omitempty = true
					}
				}
			}
			props[name] = schemaFor(f.Type, seen)
			if f.Type.Kind() != reflect.Ptr && !omitempty {
				required = append(required, name)
			}
		}
	}
	walk(t)

	schema := map[string]interface{}{
		"type":                 "object",
		"properties":           props,
		"additionalProperties": false,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}
