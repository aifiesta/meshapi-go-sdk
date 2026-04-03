package meshapi

import (
	"net/url"
	"testing"
)

func TestBuildURL_NoParams(t *testing.T) {
	cfg := Config{BaseURL: "http://localhost:8000", Token: "tok"}
	h := newHTTPClient(cfg)
	u := h.buildURL("/v1/models", nil)
	if u != "http://localhost:8000/v1/models" {
		t.Errorf("unexpected URL: %q", u)
	}
}

func TestBuildURL_WithParams(t *testing.T) {
	cfg := Config{BaseURL: "http://localhost:8000", Token: "tok"}
	h := newHTTPClient(cfg)
	qs := url.Values{"free": {"true"}}
	u := h.buildURL("/v1/models", qs)
	if u != "http://localhost:8000/v1/models?free=true" {
		t.Errorf("unexpected URL: %q", u)
	}
}

func TestBaseHeaders_Bearer(t *testing.T) {
	cfg := Config{BaseURL: "http://localhost:8000", Token: "rsk_test"}
	h := newHTTPClient(cfg)
	headers := h.baseHeaders()
	auth := headers["Authorization"]
	if auth != "Bearer rsk_test" {
		t.Errorf("expected 'Bearer rsk_test', got %q", auth)
	}
}

func TestBaseHeaders_SDKVersion(t *testing.T) {
	cfg := Config{BaseURL: "http://localhost:8000", Token: "tok"}
	h := newHTTPClient(cfg)
	headers := h.baseHeaders()
	sdk := headers[sdkVersionHeader]
	if sdk == "" {
		t.Error("expected X-MeshAPI-SDK header to be set")
	}
}

func TestListModelsParams_FreeFilter(t *testing.T) {
	freeTrue := true
	params := ListModelsParams{Free: &freeTrue}
	qs := url.Values{}
	if params.Free != nil {
		if *params.Free {
			qs.Set("free", "true")
		} else {
			qs.Set("free", "false")
		}
	}
	if qs.Get("free") != "true" {
		t.Errorf("expected 'true', got %q", qs.Get("free"))
	}
}

func TestListModelsParams_PaidFilter(t *testing.T) {
	freeFalse := false
	params := ListModelsParams{Free: &freeFalse}
	qs := url.Values{}
	if params.Free != nil {
		if *params.Free {
			qs.Set("free", "true")
		} else {
			qs.Set("free", "false")
		}
	}
	if qs.Get("free") != "false" {
		t.Errorf("expected 'false', got %q", qs.Get("free"))
	}
}

func TestListModelsParams_NoFilter(t *testing.T) {
	params := ListModelsParams{}
	qs := url.Values{}
	if params.Free != nil {
		qs.Set("free", "x")
	}
	if len(qs) != 0 {
		t.Error("expected no query params when Free is nil")
	}
}
