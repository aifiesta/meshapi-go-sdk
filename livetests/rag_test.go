package livetest

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	meshapi "meshapi-go-sdk"
)

// ragTestContent is the document uploaded in every RAG live test.
// It contains a unique phrase we search for to verify indexing.
const ragTestContent = `MeshAPI SDK live test document.
This file is used to verify RAG upload, embedding, and vector search.
The document contains the unique phrase "meshapi rag livetest" so search results are deterministic.
`

// putFile uploads raw bytes to a signed URL via HTTP PUT.
func putFile(t *testing.T, signedURL, mimeType string, body []byte) {
	t.Helper()
	req, err := http.NewRequest(http.MethodPut, signedURL, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("build PUT request: %v", err)
	}
	req.Header.Set("Content-Type", mimeType)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PUT signed URL: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		t.Fatalf("PUT signed URL returned HTTP %d", resp.StatusCode)
	}
}

// pollEmbedding waits up to maxWait for a RAG file to reach embedding_status="ready".
// ctx comes immediately after t, before client, following Go convention for test helpers.
func pollEmbedding(t *testing.T, ctx context.Context, client *meshapi.Client, fileID string, maxWait time.Duration) *meshapi.RagFileStatus {
	t.Helper()
	deadline := time.Now().Add(maxWait)
	for time.Now().Before(deadline) {
		status, err := client.RAG.Get(ctx, fileID)
		if err != nil {
			t.Fatalf("rag.get(%s): %v", fileID, err)
		}
		t.Logf("  polling embedding_status=%q for file %s", status.EmbeddingStatus, fileID)
		if status.EmbeddingStatus == "ready" {
			return status
		}
		if status.EmbeddingStatus == "failed" {
			errCode := ""
			if status.LastErrorCode != nil {
				errCode = *status.LastErrorCode
			}
			t.Fatalf("embedding failed for file %s: error_code=%q", fileID, errCode)
		}
		time.Sleep(3 * time.Second)
	}
	t.Fatalf("embedding did not reach 'ready' within %v for file %s", maxWait, fileID)
	return nil
}

// findFileInList paginates through all RAG files until fileID is found or the
// list is exhausted. Returns true if found.
func findFileInList(t *testing.T, ctx context.Context, client *meshapi.Client, fileID string) bool {
	t.Helper()
	const pageSize = 50
	offset := 0
	for {
		page, err := client.RAG.List(ctx, meshapi.ListRagFilesParams{
			Limit:  intPtr(pageSize),
			Offset: intPtr(offset),
		})
		if err != nil {
			t.Fatalf("rag.list (offset=%d): %v", offset, err)
		}
		for _, f := range page.Files {
			if f.FileID == fileID {
				return true
			}
		}
		offset += len(page.Files)
		if offset >= page.Total {
			return false
		}
	}
}

func TestLive_RAG_UploadAndSearch(t *testing.T) {
	client := newClient(t)
	ctx := context.Background()
	mimeType := "text/plain"
	content := []byte(ragTestContent)

	// ── Step 1: InitUpload (embed=false so we test the embed endpoint explicitly) ──
	embedFalse := false
	upload, err := client.RAG.InitUpload(ctx, meshapi.InitUploadRequest{
		FileName: fmt.Sprintf("go-livetest-%d.txt", time.Now().UnixMilli()),
		MimeType: mimeType,
		Embed:    &embedFalse,
	})
	if err != nil {
		t.Fatalf("rag.initUpload: %v", err)
	}
	t.Logf("[PASS] rag.initUpload → file_id=%q", upload.FileID)

	// Note: the RAG API has no DELETE endpoint, so uploaded files cannot be
	// cleaned up programmatically. Each test run leaves one file in the account.

	// ── Step 2: PUT file content to signed URL ──
	putFile(t, upload.SignedURL, mimeType, content)
	t.Log("[PASS] PUT file content to signed URL")

	// ── Step 3: Get — wait for upload_status=ready ──
	var uploadReady *meshapi.RagFileStatus
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		s, err := client.RAG.Get(ctx, upload.FileID)
		if err != nil {
			t.Fatalf("rag.get (upload poll): %v", err)
		}
		if s.UploadStatus == "ready" {
			uploadReady = s
			break
		}
		time.Sleep(2 * time.Second)
	}
	if uploadReady == nil {
		t.Fatal("upload_status did not reach 'ready' within 30s")
	}
	t.Logf("[PASS] rag.get → upload_status=%q embedding_status=%q", uploadReady.UploadStatus, uploadReady.EmbeddingStatus)

	// ── Step 4: Embed ──
	embedResp, err := client.RAG.Embed(ctx, meshapi.BulkEmbedRequest{
		FileIDs: []string{upload.FileID},
	})
	if err != nil {
		t.Fatalf("rag.embed: %v", err)
	}
	if len(embedResp.Results) == 0 {
		t.Fatal("embed returned no results")
	}
	t.Logf("[PASS] rag.embed → status=%q", embedResp.Results[0].EmbeddingStatus)

	// ── Step 5: Poll until embedding_status=ready ──
	pollEmbedding(t, ctx, client, upload.FileID, 90*time.Second)
	t.Logf("[PASS] embedding complete for file %q", upload.FileID)

	// ── Step 6: List — paginate until file is found or all pages exhausted ──
	if !findFileInList(t, ctx, client, upload.FileID) {
		t.Errorf("uploaded file %q not found in list", upload.FileID)
	} else {
		t.Logf("[PASS] rag.list → uploaded file present")
	}

	// ── Step 7: Search ──
	topK := 5
	search, err := client.RAG.Search(ctx, meshapi.SearchRequest{
		Query:   "meshapi rag livetest",
		TopK:    &topK,
		FileIDs: []string{upload.FileID},
	})
	if err != nil {
		t.Fatalf("rag.search: %v", err)
	}
	if len(search.Results) == 0 {
		t.Error("search returned no results")
	} else {
		t.Logf("[PASS] rag.search → %d results, top score=%.4f", len(search.Results), search.Results[0].Score)
	}
}
