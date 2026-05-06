package meshapi

import "context"

type FilesResource struct {
	http *httpClient
}

func (r *FilesResource) Upload(ctx context.Context, params UploadBatchFileParams) (*FileObject, error) {
	var out FileObject
	if err := r.http.post(ctx, "/v1/files", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *FilesResource) Get(ctx context.Context, fileID string) (*FileObject, error) {
	var out FileObject
	if err := r.http.get(ctx, "/v1/files/"+fileID, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *FilesResource) Delete(ctx context.Context, fileID string) error {
	return r.http.delete(ctx, "/v1/files/"+fileID)
}

func (r *FilesResource) Content(ctx context.Context, fileID string) ([]byte, error) {
	return r.http.getBytes(ctx, "/v1/files/"+fileID+"/content", nil)
}
