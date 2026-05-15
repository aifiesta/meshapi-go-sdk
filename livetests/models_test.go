package livetest

import (
	"context"
	"testing"

	meshapi "meshapi-go-sdk"
)

func TestLive_Models_List(t *testing.T) {
	client := newClient(t)
	ctx := context.Background()

	models, err := client.Models.List(ctx, meshapi.ListModelsParams{})
	if err != nil {
		t.Fatalf("models.list: %v", err)
	}
	t.Logf("models.list() → %d models", len(models))
	for _, m := range models {
		if m.ID == "" {
			t.Errorf("model with empty ID")
		}
	}
}

func TestLive_Models_Free(t *testing.T) {
	client := newClient(t)
	ctx := context.Background()

	models, err := client.Models.Free(ctx)
	if err != nil {
		t.Fatalf("models.free: %v", err)
	}
	t.Logf("[PASS] models.free() → %d models", len(models))
	for _, m := range models {
		if !m.IsFree {
			t.Errorf("paid model in free list: %q", m.ID)
		}
	}
}

func TestLive_Models_Paid(t *testing.T) {
	client := newClient(t)
	ctx := context.Background()

	models, err := client.Models.Paid(ctx)
	if err != nil {
		t.Fatalf("models.paid: %v", err)
	}
	t.Logf("[PASS] models.paid() → %d models", len(models))
	for _, m := range models {
		if m.IsFree {
			t.Errorf("free model in paid list: %q", m.ID)
		}
	}
}

func TestLive_Models_ListWithFilter(t *testing.T) {
	client := newClient(t)
	ctx := context.Background()

	freeTrue := true
	free, err := client.Models.List(ctx, meshapi.ListModelsParams{Free: &freeTrue})
	if err != nil {
		t.Fatalf("models.list(free=true): %v", err)
	}
	for _, m := range free {
		if !m.IsFree {
			t.Errorf("paid model in free-filtered list: %q", m.ID)
		}
	}

	freeFalse := false
	paid, err := client.Models.List(ctx, meshapi.ListModelsParams{Free: &freeFalse})
	if err != nil {
		t.Fatalf("models.list(free=false): %v", err)
	}
	for _, m := range paid {
		if m.IsFree {
			t.Errorf("free model in paid-filtered list: %q", m.ID)
		}
	}
	t.Logf("[PASS] filter: free=%d paid=%d", len(free), len(paid))
}
