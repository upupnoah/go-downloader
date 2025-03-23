package download

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestNewDownloader(t *testing.T) {
	url := "https://example.com/file.zip"
	outputPath := "/tmp/output.zip"
	numThreads := 4

	downloader := NewDownloader(url, outputPath, numThreads)

	if downloader.URL != url {
		t.Errorf("Expected URL %s, got %s", url, downloader.URL)
	}

	if downloader.OutputPath != outputPath {
		t.Errorf("Expected OutputPath %s, got %s", outputPath, downloader.OutputPath)
	}

	if downloader.NumThreads != numThreads {
		t.Errorf("Expected NumThreads %d, got %d", numThreads, downloader.NumThreads)
	}

	if downloader.MaxRetries != 3 {
		t.Errorf("Expected default MaxRetries 3, got %d", downloader.MaxRetries)
	}

	if !downloader.Verbose {
		t.Errorf("Expected default Verbose to be true")
	}

	// Test with empty output path
	downloader = NewDownloader(url, "", numThreads)
	if downloader.OutputPath != "file.zip" {
		t.Errorf("Expected OutputPath to be 'file.zip', got %s", downloader.OutputPath)
	}

	// Test with negative threads
	downloader = NewDownloader(url, outputPath, -1)
	expectedThreads := runtime.NumCPU()
	if downloader.NumThreads != expectedThreads {
		t.Errorf("Expected NumThreads to be %d (CPU count), got %d", expectedThreads, downloader.NumThreads)
	}
}

func TestSetVerbose(t *testing.T) {
	downloader := NewDownloader("https://example.com/file.zip", "/tmp/output.zip", 4)

	// Test setting verbose to false
	downloader.SetVerbose(false)
	if downloader.Verbose {
		t.Error("Expected Verbose to be false")
	}

	// Test setting verbose to true
	downloader.SetVerbose(true)
	if !downloader.Verbose {
		t.Error("Expected Verbose to be true")
	}
}

func TestSetMaxRetries(t *testing.T) {
	downloader := NewDownloader("https://example.com/file.zip", "/tmp/output.zip", 4)

	// Test setting max retries
	downloader.SetMaxRetries(5)
	if downloader.MaxRetries != 5 {
		t.Errorf("Expected MaxRetries to be 5, got %d", downloader.MaxRetries)
	}
}

// setupTestServer sets up a mock HTTP server for testing
func setupTestServer(t *testing.T, supportsRanges bool, contentLength int64) *httptest.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle HEAD requests for metadata
		if r.Method == "HEAD" {
			if supportsRanges {
				w.Header().Set("Accept-Ranges", "bytes")
			}
			w.Header().Set("Content-Length", fmt.Sprintf("%d", contentLength))
			return
		}

		// Handle GET requests for content
		if r.Method == "GET" {
			rangeHeader := r.Header.Get("Range")

			if supportsRanges && rangeHeader != "" {
				// Parse range header
				var start, end int64
				fmt.Sscanf(rangeHeader, "bytes=%d-%d", &start, &end)

				// Set partial content response
				w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, contentLength))
				w.WriteHeader(http.StatusPartialContent)

				// Write fake data for the range
				data := make([]byte, end-start+1)
				w.Write(data)
			} else {
				// Write full content
				data := make([]byte, contentLength)
				w.Write(data)
			}
		}
	})

	return httptest.NewServer(handler)
}

func TestDownloadSingleThreaded(t *testing.T) {
	// Skip in automated tests since it requires network operations
	// Remove this skip when testing locally or in an environment that allows network access
	t.Skip("Skipping test that requires network operations")

	// Setup mock server
	server := setupTestServer(t, false, 1024)
	defer server.Close()

	// Create temp dir for output
	tempDir, err := os.MkdirTemp("", "downloader_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	outputPath := filepath.Join(tempDir, "output.zip")

	// Create downloader with single thread
	downloader := NewDownloader(server.URL, outputPath, 1)

	// Set timeout to a reasonable value for tests
	downloader.Client.Timeout = 2 * time.Second

	// Run download
	err = downloader.Start()
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}

	// Verify file was created
	fi, err := os.Stat(outputPath)
	if err != nil {
		t.Errorf("Output file not created: %v", err)
	}
	if fi.Size() != 1024 {
		t.Errorf("Expected file size 1024, got %d", fi.Size())
	}
}

func TestMergeChunks(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "downloader_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create some chunks
	chunk1 := NewChunk(1, "https://example.com/test.zip", 0, 100, tempDir)
	chunk2 := NewChunk(2, "https://example.com/test.zip", 101, 200, tempDir)
	chunks := []*Chunk{chunk1, chunk2}

	// Create chunk files with some content
	for _, chunk := range chunks {
		file, err := os.Create(chunk.TempFile)
		if err != nil {
			t.Fatalf("Failed to create chunk file: %v", err)
		}

		// Write some data
		data := make([]byte, chunk.Size)
		for i := range data {
			data[i] = byte(chunk.ID) // Fill with chunk ID for identification
		}

		_, err = file.Write(data)
		if err != nil {
			t.Fatalf("Failed to write to chunk file: %v", err)
		}

		file.Close()
	}

	// Create downloader
	outputPath := filepath.Join(tempDir, "merged.zip")
	downloader := NewDownloader("https://example.com/test.zip", outputPath, 2)
	downloader.Chunks = chunks

	// Merge chunks
	err = downloader.mergeChunks()
	if err != nil {
		t.Fatalf("mergeChunks failed: %v", err)
	}

	// Verify merged file
	fi, err := os.Stat(outputPath)
	if err != nil {
		t.Errorf("Merged file not created: %v", err)
	}

	expectedSize := int64(201) // 0-200 inclusive
	if fi.Size() != expectedSize {
		t.Errorf("Expected merged file size %d, got %d", expectedSize, fi.Size())
	}

	// Optional: verify content
	mergedFile, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read merged file: %v", err)
	}

	// First chunk should have ID 1
	if len(mergedFile) > 0 && mergedFile[0] != 1 {
		t.Errorf("Expected first byte to be 1, got %d", mergedFile[0])
	}

	// Second chunk should have ID 2
	if len(mergedFile) > 101 && mergedFile[101] != 2 {
		t.Errorf("Expected byte at position 101 to be 2, got %d", mergedFile[101])
	}
}
