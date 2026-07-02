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
	if resp.Provider != "native" && resp.Provider != "tavily" {
		t.Errorf("unexpected provider %q", resp.Provider)
	}
	for _, hit := range resp.Results {
		if hit.Title == "" || hit.URL == "" {
			t.Errorf("result missing title/url: %+v", hit)
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
