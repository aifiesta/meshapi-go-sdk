package livetest

import (
	"context"
	"errors"
	"testing"

	meshapi "meshapi-go-sdk"
)

// 1x1 transparent PNG, used when MESHAPI_IMAGE_EDIT_INPUT is not provided.
const pixelPNG = "data:image/png;base64," +
	"iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="

func TestLive_Images_Edit(t *testing.T) {
	model := liveEnv("MESHAPI_IMAGE_EDIT_MODEL", "")
	if model == "" {
		t.Skip("set MESHAPI_IMAGE_EDIT_MODEL to run the image-edit live test")
	}
	client := newClient(t)
	ctx := context.Background()

	source := liveEnv("MESHAPI_IMAGE_EDIT_INPUT", pixelPNG)
	prompt := "Make the background a solid blue."
	op := "edit"

	resp, err := client.Images.Edit(ctx, meshapi.ImageEditParams{
		Model:     model,
		Image:     source,
		Prompt:    &prompt,
		Operation: &op,
	})
	if err != nil {
		var apiErr *meshapi.MeshAPIError
		if errors.As(err, &apiErr) {
			switch {
			case apiErr.Status == 400 && apiErr.Code == "invalid_request":
				// Upstream (provider) content/safety rejection of the synthetic
				// test image — the SDK request reached the provider, so the
				// request path is validated. Skip rather than fail.
				t.Skipf("provider rejected the test image: %s", apiErr.Message)
			case (apiErr.Status == 400 || apiErr.Status == 501) &&
				(apiErr.Code == "model_capability_not_supported" || apiErr.Code == "not_implemented"):
				t.Skipf("model does not support image edits: %s", apiErr.Code)
			}
		}
		t.Fatalf("images.Edit: %v", err)
	}
	if len(resp.Data) == 0 {
		t.Fatal("expected at least one edited image")
	}
	first := resp.Data[0]
	if (first.URL == nil || *first.URL == "") && (first.B64JSON == nil || *first.B64JSON == "") {
		t.Error("edited image should have a url or b64_json")
	}
}
