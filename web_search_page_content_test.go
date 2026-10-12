package meshapi

import (
	"encoding/json"
	"testing"
)

// The wire half: this is an additive flag, so a caller who never sets it must
// send exactly the body it sent before.

func TestWebSearchParams_UnsetPageContentIsAbsentFromThePayload(t *testing.T) {
	data, err := json.Marshal(WebSearchParams{Query: "mars rovers"})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var body map[string]any
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, present := body["include_page_content"]; present {
		t.Errorf("an unset flag must not appear on the wire, got %s", data)
	}
}

func TestWebSearchParams_RequestingPageContentSendsTheFlag(t *testing.T) {
	want := true
	data, err := json.Marshal(WebSearchParams{Query: "mars rovers", IncludePageContent: &want})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var body map[string]any
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got, ok := body["include_page_content"].(bool); !ok || !got {
		t.Errorf("expected include_page_content true, got %s", data)
	}
}

func TestWebSearchParams_EveryDocumentedEngineCanBePinned(t *testing.T) {
	for _, engine := range []string{"native", "tavily", "tinyfish"} {
		e := engine
		data, err := json.Marshal(WebSearchParams{Query: "q", Provider: &e})
		if err != nil {
			t.Fatalf("marshal %q: %v", engine, err)
		}
		var body map[string]any
		if err := json.Unmarshal(data, &body); err != nil {
			t.Fatalf("unmarshal %q: %v", engine, err)
		}
		if body["provider"] != engine {
			t.Errorf("provider %q did not round-trip, got %v", engine, body["provider"])
		}
	}
}

// The response half.

func TestWebSearchResponse_PageContentAndTruncationFlagAreDecoded(t *testing.T) {
	raw := []byte(`{
      "query": "mars rovers",
      "provider": "tinyfish",
      "request_id": "req_01J",
      "results": [
        {
          "title": "Perseverance",
          "url": "https://example.com/p",
          "content": "a short snippet",
          "score": 0.9,
          "page_content": "the full extracted page text",
          "page_content_truncated": true
        }
      ]
    }`)
	var resp WebSearchResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resp.Results))
	}
	hit := resp.Results[0]
	if hit.PageContent == nil {
		t.Fatal("expected page content, got nil")
	}
	if *hit.PageContent != "the full extracted page text" {
		t.Errorf("unexpected page content: %q", *hit.PageContent)
	}
	if !hit.PageContentTruncated {
		t.Error("expected PageContentTruncated true")
	}
	// The snippet is a separate field; page text never replaces it.
	if hit.Content != "a short snippet" {
		t.Errorf("snippet changed: %q", hit.Content)
	}
}

func TestWebSearchResultItem_NullPageContentStaysNilNotEmptyString(t *testing.T) {
	// "no text available" and "empty text" are different answers; a pointer is
	// what keeps them apart.
	var item WebSearchResultItem
	raw := []byte(`{"title":"t","url":"https://example.com","content":"snippet","page_content":null,"page_content_truncated":false}`)
	if err := json.Unmarshal(raw, &item); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if item.PageContent != nil {
		t.Errorf("expected nil page content, got %q", *item.PageContent)
	}
	if item.PageContentTruncated {
		t.Error("expected PageContentTruncated false")
	}
}

func TestWebSearchResultItem_AResponseWithoutTheNewKeysStillDecodes(t *testing.T) {
	// An older gateway omits them entirely; that must not be a decode error.
	var item WebSearchResultItem
	raw := []byte(`{"title":"t","url":"https://example.com","content":"snippet"}`)
	if err := json.Unmarshal(raw, &item); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if item.PageContent != nil {
		t.Error("expected nil page content")
	}
	if item.PageContentTruncated {
		t.Error("expected PageContentTruncated false")
	}
}
