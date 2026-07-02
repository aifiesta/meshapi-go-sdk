package livetest

import (
	"context"
	"errors"
	"testing"

	meshapi "meshapi-go-sdk"
)

// skipIfUnavailable skips when the endpoint is disabled/absent on this
// deployment (403/404/501) rather than failing the suite.
func skipIfUnavailable(t *testing.T, err error, feature string) {
	t.Helper()
	var apiErr *meshapi.MeshAPIError
	if errors.As(err, &apiErr) && (apiErr.Status == 403 || apiErr.Status == 404 || apiErr.Status == 501) {
		t.Skipf("%s unavailable on this deployment: %s", feature, apiErr.Code)
	}
}

func TestLive_Moderations_FlagsHarmfulText(t *testing.T) {
	client := newClient(t)
	ctx := context.Background()

	resp, err := client.Moderations.Create(ctx, meshapi.ModerationParams{
		Input: "I want to hurt and kill someone right now.",
	})
	if err != nil {
		skipIfUnavailable(t, err, "moderations")
		t.Fatalf("moderations.Create: %v", err)
	}
	if len(resp.Results) == 0 {
		t.Fatal("expected at least one moderation result")
	}
	if !resp.Results[0].Flagged {
		t.Error("expected harmful text to be flagged")
	}
	if len(resp.Results[0].Categories) == 0 {
		t.Error("expected category booleans")
	}
}

func TestLive_Moderations_PassesBenignText(t *testing.T) {
	client := newClient(t)
	ctx := context.Background()

	resp, err := client.Moderations.Create(ctx, meshapi.ModerationParams{
		Input: "I love sunny days at the park.",
	})
	if err != nil {
		skipIfUnavailable(t, err, "moderations")
		t.Fatalf("moderations.Create: %v", err)
	}
	if len(resp.Results) == 0 || resp.Results[0].Flagged {
		t.Error("expected benign text not to be flagged")
	}
}

func TestLive_Moderations_BatchInput(t *testing.T) {
	client := newClient(t)
	ctx := context.Background()

	resp, err := client.Moderations.Create(ctx, meshapi.ModerationParams{
		Input: []string{"hello friend", "have a nice day"},
	})
	if err != nil {
		skipIfUnavailable(t, err, "moderations")
		t.Fatalf("moderations.Create: %v", err)
	}
	if len(resp.Results) != 2 {
		t.Errorf("expected 2 results, got %d", len(resp.Results))
	}
}
