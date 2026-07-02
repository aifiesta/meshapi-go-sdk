package meshapi

import (
	"context"
	"net/url"
	"strconv"
)

// DocumentsResource provides access to /v1/documents endpoints.
type DocumentsResource struct {
	http *httpClient
}

// Generate generates a new document (POST /v1/documents/generate).
func (r *DocumentsResource) Generate(ctx context.Context, params GenerateDocumentRequest) (*DocumentResponse, error) {
	var out DocumentResponse
	if err := r.http.post(ctx, "/v1/documents/generate", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns a paginated list of documents (GET /v1/documents).
func (r *DocumentsResource) List(ctx context.Context, params *ListDocumentsParams) (*DocumentListResponse, error) {
	qp := url.Values{}
	if params != nil {
		if params.Limit != nil {
			qp.Set("limit", strconv.Itoa(*params.Limit))
		}
		if params.Offset != nil {
			qp.Set("offset", strconv.Itoa(*params.Offset))
		}
	}
	var out DocumentListResponse
	if err := r.http.get(ctx, "/v1/documents", qp, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get returns a single document by ID (GET /v1/documents/{document_id}).
func (r *DocumentsResource) Get(ctx context.Context, documentID string) (*DocumentResponse, error) {
	var out DocumentResponse
	if err := r.http.get(ctx, "/v1/documents/"+url.PathEscape(documentID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
