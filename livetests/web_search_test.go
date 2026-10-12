package livetest

import (
	"context"
	"testing"

	meshapi "meshapi-go-sdk"
)

func TestLive_WebSearch_Basic(t *testing.T) {
	client := newClient(t)
	ctx := context.Background()

	maxResults := 3
	resp, err := client.Web.Search(ctx, meshapi.WebSearchParams{
		Query:      "what is the capital of France",
		MaxResults: &maxResults,
	})
	if err != nil {
		skipIfUnavailable(t, err, "web search (WEB_SEARCH_ENABLED)")
		t.Fatalf("web.Search: %v", err)
	}
	if resp.Query == "" {
		t.Error("expected query echoed back")
	}
	if resp.Provider != "native" && resp.Provider != "tavily" && resp.Provider != "tinyfish" {
		t.Errorf("unexpected provider %q", resp.Provider)
	}
	for _, hit := range resp.Results {
		if hit.Title == "" || hit.URL == "" {
			t.Errorf("result missing title/url: %+v", hit)
		}
		// Not asking must never produce page text, whichever engine served.
		if hit.PageContent != nil {
			t.Errorf("page text arrived unasked-for from %q", resp.Provider)
		}
		if hit.PageContentTruncated {
			t.Error("expected PageContentTruncated false when page content was not requested")
		}
	}
}

func TestLive_WebSearch_WithAnswer(t *testing.T) {
	client := newClient(t)
	ctx := context.Background()

	includeAnswer := true
	resp, err := client.Web.Search(ctx, meshapi.WebSearchParams{
		Query:         "who wrote the book Dune",
		IncludeAnswer: &includeAnswer,
	})
	if err != nil {
		skipIfUnavailable(t, err, "web search (WEB_SEARCH_ENABLED)")
		t.Fatalf("web.Search: %v", err)
	}
	// Answer is best-effort; assert the field decodes, not that it is non-nil.
	if resp.Answer != nil {
		t.Logf("answer: %s", *resp.Answer)
	}
}

// TestLive_WebSearch_PageContent pins tinyfish: page text is that engine only,
// and the native engine is tried first, so an unpinned request would prove
// nothing.
func TestLive_WebSearch_PageContent(t *testing.T) {
	client := newClient(t)
	ctx := context.Background()

	engine, withPages, maxResults := "tinyfish", true, 2
	resp, err := client.Web.Search(ctx, meshapi.WebSearchParams{
		Query:              "James Webb telescope earliest galaxy",
		Provider:           &engine,
		IncludePageContent: &withPages,
		MaxResults:         &maxResults,
	})
	if err != nil {
		// Covers web search being off AND this deployment having no tinyfish
		// credential, so the engine is not registered. Both are deployment
		// configuration rather than a defect in this SDK.
		skipIfUnavailable(t, err, "tinyfish web search engine")
		t.Fatalf("web.Search: %v", err)
	}
	if resp.Provider != "tinyfish" {
		t.Fatalf("pinning was ignored, served by %q", resp.Provider)
	}
	if len(resp.Results) == 0 {
		t.Fatal("expected at least one result")
	}

	withText := 0
	for _, hit := range resp.Results {
		if hit.Title == "" || hit.URL == "" {
			t.Errorf("result missing title/url: %+v", hit)
		}
		if hit.PageContent != nil {
			if *hit.PageContent == "" {
				t.Error("PageContent should be nil, never a pointer to empty string")
			}
			withText++
		}
	}
	if withText == 0 {
		// A per-URL fetch failing upstream is normal and silent, so this is a
		// skip rather than a failure — the shape above is what this SDK owns.
		t.Skip("upstream returned no page text for any result on this run")
	}
}
