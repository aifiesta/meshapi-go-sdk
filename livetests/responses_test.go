package livetest

import (
	"context"
	"errors"
	"testing"

	meshapi "meshapi-go-sdk"
)

func TestLive_Responses_ListShape(t *testing.T) {
	client := newClient(t)
	ctx := context.Background()

	limit := 5
	page, err := client.Responses.List(ctx, meshapi.ResponsesListParams{Limit: &limit})
	if err != nil {
		t.Fatalf("responses.list: %v", err)
	}
	if page.Object != nil && *page.Object != "list" {
		t.Errorf("expected object=list, got %q", *page.Object)
	}
	if len(page.Data) > 5 {
		t.Errorf("page exceeded limit: %d items", len(page.Data))
	}
	for _, item := range page.Data {
		if item.ID == "" {
			t.Error("job item missing id")
		}
	}
}

func TestLive_Responses_GetUnknownID(t *testing.T) {
	client := newClient(t)
	ctx := context.Background()

	_, err := client.Responses.Get(ctx, "resp_does_not_exist_000000000000")
	if err == nil {
		t.Fatal("expected an error for a non-existent response id")
	}
	var apiErr *meshapi.MeshAPIError
	if errors.As(err, &apiErr) {
		if apiErr.Status != 400 && apiErr.Status != 404 {
			t.Errorf("expected 400/404, got %d", apiErr.Status)
		}
	} else {
		t.Fatalf("expected *MeshAPIError, got %T: %v", err, err)
	}
}
