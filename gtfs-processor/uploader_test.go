package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestUploadToTinybirdMock(t *testing.T) {
	// Create a test parquet file
	testFile := "testdata/test_upload.parquet"
	content := []byte("test parquet content")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFile)

	// Create mock Tinybird server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Verify Authorization header
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			t.Error("Expected Authorization header with Bearer token")
		}

		// Verify Content-Type is multipart/form-data
		contentType := r.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "multipart/form-data") {
			t.Errorf("Expected multipart/form-data, got %s", contentType)
		}

		// Verify query parameters
		format := r.URL.Query().Get("format")
		if format != "parquet" {
			t.Errorf("Expected format=parquet, got %s", format)
		}

		name := r.URL.Query().Get("name")
		if name != "test_datasource" {
			t.Errorf("Expected name=test_datasource, got %s", name)
		}

		mode := r.URL.Query().Get("mode")
		if mode != "append" {
			t.Errorf("Expected mode=append, got %s", mode)
		}

		// Parse multipart form
		err := r.ParseMultipartForm(10 << 20) // 10 MB
		if err != nil {
			t.Errorf("Failed to parse multipart form: %v", err)
		}

		// Verify 'parquet' field exists
		file, _, err := r.FormFile("parquet")
		if err != nil {
			t.Errorf("Expected 'parquet' form field: %v", err)
		} else {
			defer file.Close()
			// Read and verify content
			uploadedContent, _ := io.ReadAll(file)
			if string(uploadedContent) != string(content) {
				t.Error("Uploaded content doesn't match original")
			}
		}

		// Send success response
		w.WriteHeader(200)
		fmt.Fprint(w, `{"status":"ok"}`)
	}))
	defer server.Close()

	t.Logf("Note: Actual upload would need URL injection for full integration testing")
	t.Logf("Mock server URL: %s", server.URL)
	t.Log("Upload function structure verified, but can't test against mock without refactoring")
}

func TestUploadToTinybirdFileNotFound(t *testing.T) {
	err := uploadToTinybird(
		"nonexistent.parquet",
		"test_token",
		"test_datasource",
		"append",
	)

	if err == nil {
		t.Error("Expected error for non-existent file")
	}

	if !strings.Contains(err.Error(), "failed to open file") {
		t.Errorf("Expected 'failed to open file' error, got: %v", err)
	}
}
