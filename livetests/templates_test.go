package livetest

import (
	"context"
	"fmt"
	"testing"
	"time"

	meshapi "meshapi-go-sdk"
)

func uniqueName(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixMilli())
}

func TestLive_Templates_CRUD(t *testing.T) {
	client := newClient(t)
	ctx := context.Background()

	name := uniqueName("go-livetest")
	desc := "Go SDK live test template"
	system := "You are a test assistant."

	// --- Create ---
	tmpl, err := client.Templates.Create(ctx, meshapi.CreateTemplateParams{
		Name:        name,
		Description: &desc,
		System:      &system,
	})
	if err != nil {
		t.Fatalf("templates.create: %v", err)
	}
	owner := ""
	if tmpl.Owner != nil {
		owner = *tmpl.Owner
	}
	t.Logf("[PASS] templates.create → id=%q owner=%q", tmpl.ID, owner)

	t.Cleanup(func() {
		_ = client.Templates.Delete(context.Background(), tmpl.ID)
	})

	// --- List ---
	all, err := client.Templates.List(ctx)
	if err != nil {
		t.Fatalf("templates.list: %v", err)
	}
	found := false
	for _, tpl := range all {
		if tpl.ID == tmpl.ID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("created template %q not found in list (%d items)", tmpl.ID, len(all))
	}
	t.Logf("[PASS] templates.list → %d templates, created template present", len(all))

	// --- Get ---
	got, err := client.Templates.Get(ctx, tmpl.ID)
	if err != nil {
		t.Fatalf("templates.get: %v", err)
	}
	if got.ID != tmpl.ID {
		t.Errorf("get id mismatch: want %q got %q", tmpl.ID, got.ID)
	}
	t.Logf("[PASS] templates.get → name=%q", got.Name)

	// --- Update ---
	newDesc := "Updated by Go SDK live test"
	updated, err := client.Templates.Update(ctx, tmpl.ID, meshapi.UpdateTemplateParams{
		Description: &newDesc,
	})
	if err != nil {
		t.Fatalf("templates.update: %v", err)
	}
	if updated.Description == nil || *updated.Description != newDesc {
		t.Errorf("update description mismatch: got %v", updated.Description)
	}
	t.Logf("[PASS] templates.update → description=%q", *updated.Description)

	// --- Delete ---
	if err := client.Templates.Delete(ctx, tmpl.ID); err != nil {
		t.Fatalf("templates.delete: %v", err)
	}
	t.Log("[PASS] templates.delete → 204 No Content")

	// --- Verify 404 ---
	_, err = client.Templates.Get(ctx, tmpl.ID)
	if err == nil {
		t.Fatal("expected 404 after delete, got nil")
	}
	var svcErr *meshapi.MeshAPIError
	if !errors.As(err, &svcErr) {
		t.Fatalf("expected *MeshAPIError, got %T: %v", err, err)
	}
	if svcErr.Status != 404 {
		t.Errorf("expected status 404, got %d", svcErr.Status)
	}
	t.Logf("[PASS] templates.get(deleted) → 404 %q", svcErr.Code)
}
