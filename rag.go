package meshapi

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
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
	if err := r.http.get(ctx, "/v1/files/"+url.PathEscape(fileID), nil, &out); err != nil {
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

// UploadFile is a convenience wrapper that calls InitUpload then PUTs the
// file content to the returned signed URL in one step.
// It returns the same InitUploadResponse so the caller has the FileID.
func (r *RagResource) UploadFile(ctx context.Context, params UploadFileParams) (*InitUploadResponse, error) {
	upload, err := r.InitUpload(ctx, InitUploadRequest{
		FileName: params.FileName,
		MimeType: params.MimeType,
		Embed:    params.Embed,
		Metadata: params.Metadata,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, upload.SignedURL, bytes.NewReader(params.Content))
	if err != nil {
		return nil, fmt.Errorf("rag: build PUT request: %w", err)
	}
	req.Header.Set("Content-Type", params.MimeType)

	// Use the SDK's configured HTTP client (respects Config.HTTPClient and
	// timeout) rather than http.DefaultClient. We deliberately skip
	// r.http.do() here because the signed URL must not carry the Authorization
	// header used for MeshAPI requests.
	resp, err := r.http.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("rag: PUT signed URL: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("rag: PUT signed URL returned HTTP %d", resp.StatusCode)
	}

	return upload, nil
}

// Search performs a vector similarity search over embedded files.
func (r *RagResource) Search(ctx context.Context, params SearchRequest) (*SearchResponse, error) {
	var out SearchResponse
	if err := r.http.post(ctx, "/v1/files/search", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
