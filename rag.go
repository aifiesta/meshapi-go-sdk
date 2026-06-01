package meshapi

import (
	"context"
	"fmt"
	"net/url"
)

// RagResource provides access to the RAG /v1/files endpoints.
type RagResource struct {
	http *httpClient
}

// InitUpload initialises a RAG file upload and returns a signed URL for the
// actual file content. After calling this, PUT the file bytes to SignedURL,
// then call Embed to trigger embedding.
func (r *RagResource) InitUpload(ctx context.Context, params InitUploadRequest) (*InitUploadResponse, error) {
	var out InitUploadResponse
	if err := r.http.post(ctx, "/v1/files", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns a paginated list of RAG files owned by the authenticated user.
func (r *RagResource) List(ctx context.Context, params ListRagFilesParams) (*RagFileListResponse, error) {
	query := url.Values{}
	if params.Limit != nil {
		query.Set("limit", fmt.Sprintf("%d", *params.Limit))
	}
	if params.Offset != nil {
		query.Set("offset", fmt.Sprintf("%d", *params.Offset))
	}
	var out RagFileListResponse
	if err := r.http.get(ctx, "/v1/files", query, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get returns the current status of a single RAG file.
func (r *RagResource) Get(ctx context.Context, fileID string) (*RagFileStatus, error) {
	var out RagFileStatus
	if err := r.http.get(ctx, "/v1/files/"+fileID, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Embed enqueues embedding jobs for one or more files. Each file must have
// upload_status=ready and embedding_status=pending or failed.
func (r *RagResource) Embed(ctx context.Context, params BulkEmbedRequest) (*BulkEmbedResponse, error) {
	var out BulkEmbedResponse
	if err := r.http.post(ctx, "/v1/files/embed", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Search performs a vector similarity search over embedded files.
func (r *RagResource) Search(ctx context.Context, params SearchRequest) (*SearchResponse, error) {
	var out SearchResponse
	if err := r.http.post(ctx, "/v1/files/search", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
