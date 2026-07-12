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
		var err error
		schema, err = jsonSchemaForType(reflect.TypeOf((*T)(nil)).Elem())
		if err != nil {
			return zero, fmt.Errorf("meshapi: cannot build structured-output schema for %T: %w", zero, err)
		}
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

// jsonSchemaForType derives a JSON schema from a Go type by reflection,
// following encoding/json's conventions: json tags, embedding with the same
// depth/tag dominance rules (an ambiguous JSON name that encoding/json would
// ignore is omitted from the schema), custom decoders, and supported map key
// types. It returns an error for types whose schema would accept JSON that
// json.Unmarshal can never decode (e.g. float- or bool-keyed maps); pass
// WithSchema to supply an explicit schema for such types.
func jsonSchemaForType(t reflect.Type) (map[string]interface{}, error) {
	return schemaFor(t, map[reflect.Type]bool{})
}

func schemaFor(t reflect.Type, seen map[reflect.Type]bool) (map[string]interface{}, error) {
	for t != nil && t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t == nil {
		return map[string]interface{}{}, nil
	}
	if t == timeType {
		return map[string]interface{}{"type": "string", "format": "date-time"}, nil
	}
	// Types with a custom decoder don't follow their reflect kind on the wire.
	// A TextUnmarshaler decodes from a JSON string; a bare json.Unmarshaler can
	// accept any JSON, so leave it unconstrained. (time.Time is handled above so
	// it keeps its date-time format.)
	if reflect.PtrTo(t).Implements(textUnmarshalerType) {
		return map[string]interface{}{"type": "string"}, nil
	}
	if reflect.PtrTo(t).Implements(jsonUnmarshalerType) {
		return map[string]interface{}{}, nil
	}

	switch t.Kind() {
	case reflect.String:
		return map[string]interface{}{"type": "string"}, nil
	case reflect.Bool:
		return map[string]interface{}{"type": "boolean"}, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return map[string]interface{}{"type": "integer"}, nil
	case reflect.Float32, reflect.Float64:
		return map[string]interface{}{"type": "number"}, nil
	case reflect.Slice, reflect.Array:
		if t.Elem().Kind() == reflect.Uint8 { // []byte marshals to a base64 string
			return map[string]interface{}{"type": "string"}, nil
		}
		items, err := schemaFor(t.Elem(), seen)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"type": "array", "items": items}, nil
	case reflect.Map:
		// json.Unmarshal can decode JSON object keys into string, integer, and
		// encoding.TextUnmarshaler key types only. Anything else (float, bool,
		// struct, ...) would yield a schema the model can satisfy but that every
		// decode attempt — including retries — fails on, so refuse up front.
		if !mapKeySupported(t.Key()) {
			return nil, fmt.Errorf(
				"unsupported map key type %s: encoding/json only decodes JSON object keys into "+
					"string, integer, or encoding.TextUnmarshaler key types — change the key type "+
					"or pass WithSchema", t.Key())
		}
		values, err := schemaFor(t.Elem(), seen)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"type":                 "object",
			"additionalProperties": values,
		}, nil
	case reflect.Struct:
		if seen[t] {
			return map[string]interface{}{"type": "object"}, nil // break recursive types
		}
		seen[t] = true
		defer delete(seen, t)
		return structSchema(t, seen)
	default: // Interface, Chan, Func, etc. -> unconstrained
		return map[string]interface{}{}, nil
	}
}

// mapKeySupported mirrors json.Unmarshal's map-key rules: string kinds, integer
// kinds, and types implementing encoding.TextUnmarshaler.
func mapKeySupported(k reflect.Type) bool {
	if k.Implements(textUnmarshalerType) || reflect.PtrTo(k).Implements(textUnmarshalerType) {
		return true
	}
	switch k.Kind() {
	case reflect.String,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	default:
		return false
	}
}

// structField is one candidate for a JSON property, tagged with the embedding
// depth it was found at so conflicts resolve the way encoding/json resolves them.
type structField struct {
	name      string
	tagged    bool
	typ       reflect.Type
	omitempty bool
	depth     int
}

// structSchema builds an object schema from a struct's fields, applying
// encoding/json's field dominance rules for embedded structs: a shallower field
// wins over a deeper one; at equal depth a json-tagged field wins over untagged;
// a genuine tie is ambiguous — encoding/json ignores the value at decode time,
// so the name is omitted from the schema (additionalProperties stays false).
func structSchema(t reflect.Type, seen map[reflect.Type]bool) (map[string]interface{}, error) {
	// Breadth-first over embedded structs so depth is the embedding level.
	var order []string
	byName := map[string][]structField{}
	visited := map[reflect.Type]bool{}
	level := []reflect.Type{t}
	for depth := 0; len(level) > 0; depth++ {
		var next []reflect.Type
		for _, rt := range level {
			if visited[rt] {
				continue
			}
			visited[rt] = true
			for i := 0; i < rt.NumField(); i++ {
				f := rt.Field(i)
				tag := f.Tag.Get("json")
				if tag == "-" {
					continue
				}
				// Untagged anonymous embedded structs flatten into the parent at
				// the next depth (their promoted exported fields count even when
				// the embedded type itself is unexported, as in encoding/json).
				if f.Anonymous && tag == "" {
					ft := f.Type
					for ft.Kind() == reflect.Ptr {
						ft = ft.Elem()
					}
					if ft.Kind() == reflect.Struct {
						next = append(next, ft)
						continue
					}
				}
				if f.PkgPath != "" { // unexported
					continue
				}

				name := f.Name
				tagged := false
				omitempty := false
				if tag != "" {
					parts := strings.Split(tag, ",")
					if parts[0] != "" {
						name = parts[0]
						tagged = true
					}
					for _, p := range parts[1:] {
						if p == "omitempty" {
							omitempty = true
						}
					}
				}
				fld := structField{name: name, tagged: tagged, typ: f.Type, omitempty: omitempty, depth: depth}
				if _, ok := byName[name]; !ok {
					order = append(order, name)
				}
				byName[name] = append(byName[name], fld)
			}
		}
		level = next
	}

	props := map[string]interface{}{}
	required := []string{}
	for _, name := range order {
		group := byName[name]
		minDepth := group[0].depth
		for _, g := range group {
			if g.depth < minDepth {
				minDepth = g.depth
			}
		}
		var atMin []structField
		for _, g := range group {
			if g.depth == minDepth {
				atMin = append(atMin, g)
			}
		}
		winner := atMin[0]
		if len(atMin) > 1 {
			var taggedAtMin []structField
			for _, g := range atMin {
				if g.tagged {
					taggedAtMin = append(taggedAtMin, g)
				}
			}
			if len(taggedAtMin) != 1 {
				continue // ambiguous — encoding/json drops the field, so does the schema
			}
			winner = taggedAtMin[0]
		}

		ps, err := schemaFor(winner.typ, seen)
		if err != nil {
			return nil, err
		}
		props[name] = ps
		if winner.typ.Kind() != reflect.Ptr && !winner.omitempty {
			required = append(required, name)
		}
	}

	schema := map[string]interface{}{
		"type":                 "object",
		"properties":           props,
		"additionalProperties": false,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema, nil
}
