package download

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewChunk(t *testing.T) {
	// Test creating a new chunk
	url := "https://example.com/test.zip"
	id := 1
	start := int64(1000)
	end := int64(2000)
	tempDir := "/tmp/test"

	chunk := NewChunk(id, url, start, end, tempDir)

	if chunk.ID != id {
		t.Errorf("Expected ID %d, got %d", id, chunk.ID)
	}

	if chunk.URL != url {
		t.Errorf("Expected URL %s, got %s", url, chunk.URL)
	}

	if chunk.Start != start {
		t.Errorf("Expected Start %d, got %d", start, chunk.Start)
	}

	if chunk.End != end {
		t.Errorf("Expected End %d, got %d", end, chunk.End)
	}

	expectedSize := end - start + 1
	if chunk.Size != expectedSize {
		t.Errorf("Expected Size %d, got %d", expectedSize, chunk.Size)
	}

	expectedTempFile := filepath.Join(tempDir, "chunk_1")
	if chunk.TempFile != expectedTempFile {
		t.Errorf("Expected TempFile %s, got %s", expectedTempFile, chunk.TempFile)
	}

	if chunk.Downloaded != 0 {
		t.Errorf("Expected Downloaded to be 0, got %d", chunk.Downloaded)
	}

	if chunk.Completed {
		t.Error("Expected Completed to be false")
	}

	if chunk.Failed {
		t.Error("Expected Failed to be false")
	}

	if chunk.RetryCount != 0 {
		t.Errorf("Expected RetryCount to be 0, got %d", chunk.RetryCount)
	}
}

func TestUpdateProgress(t *testing.T) {
	chunk := NewChunk(1, "https://example.com/test.zip", 0, 1000, "/tmp")

	// Test updating progress
	chunk.UpdateProgress(500)
	if chunk.Downloaded != 500 {
		t.Errorf("Expected Downloaded to be 500, got %d", chunk.Downloaded)
	}
	if chunk.Completed {
		t.Error("Expected Completed to be false")
	}

	// Test completing chunk
	chunk.UpdateProgress(501)
	if chunk.Downloaded != 1001 {
		t.Errorf("Expected Downloaded to be 1001, got %d", chunk.Downloaded)
	}
	if !chunk.Completed {
		t.Error("Expected Completed to be true")
	}
}

func TestMarkFailed(t *testing.T) {
	chunk := NewChunk(1, "https://example.com/test.zip", 0, 1000, "/tmp")

	chunk.MarkFailed()
	if !chunk.Failed {
		t.Error("Expected Failed to be true")
	}
	if chunk.RetryCount != 1 {
		t.Errorf("Expected RetryCount to be 1, got %d", chunk.RetryCount)
	}

	chunk.MarkFailed()
	if chunk.RetryCount != 2 {
		t.Errorf("Expected RetryCount to be 2, got %d", chunk.RetryCount)
	}
}

func TestResetForRetry(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "chunk_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create chunk
	chunk := NewChunk(1, "https://example.com/test.zip", 0, 1000, tempDir)

	// Create temp file
	_, err = os.Create(chunk.TempFile)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// Mark as failed and update progress
	chunk.Downloaded = 500
	chunk.Failed = true

	// Reset
	chunk.ResetForRetry()

	if chunk.Downloaded != 0 {
		t.Errorf("Expected Downloaded to be 0, got %d", chunk.Downloaded)
	}

	if chunk.Failed {
		t.Error("Expected Failed to be false")
	}

	// Verify file was removed
	_, err = os.Stat(chunk.TempFile)
	if !os.IsNotExist(err) {
		t.Error("Expected temp file to be removed")
	}
}

func TestGetProgress(t *testing.T) {
	chunk := NewChunk(1, "https://example.com/test.zip", 0, 1000, "/tmp")

	// Test 0 progress
	progress := chunk.GetProgress()
	if progress != 0 {
		t.Errorf("Expected progress to be 0, got %f", progress)
	}

	// Test partial progress
	chunk.Downloaded = 250
	progress = chunk.GetProgress()
	if progress != 25.0 {
		t.Errorf("Expected progress to be 25.0, got %f", progress)
	}

	// Test complete progress
	chunk.Downloaded = 1001
	progress = chunk.GetProgress()
	if progress != 100.1 {
		t.Errorf("Expected progress to be 100.1, got %f", progress)
	}
}

func TestCalculateChunks(t *testing.T) {
	url := "https://example.com/test.zip"
	fileSize := int64(10000)
	numChunks := 4
	tempDir := "/tmp"

	chunks, err := CalculateChunks(url, fileSize, numChunks, tempDir)
	if err != nil {
		t.Fatalf("CalculateChunks failed: %v", err)
	}

	if len(chunks) != numChunks {
		t.Errorf("Expected %d chunks, got %d", numChunks, len(chunks))
	}

	// Check first chunk
	if chunks[0].Start != 0 {
		t.Errorf("Expected first chunk to start at 0, got %d", chunks[0].Start)
	}

	chunkSize := fileSize / int64(numChunks)
	if chunks[0].End != chunkSize-1 {
		t.Errorf("Expected first chunk to end at %d, got %d", chunkSize-1, chunks[0].End)
	}

	// Check last chunk
	lastChunk := chunks[numChunks-1]
	if lastChunk.End != fileSize-1 {
		t.Errorf("Expected last chunk to end at %d, got %d", fileSize-1, lastChunk.End)
	}

	// Test invalid file size
	_, err = CalculateChunks(url, 0, numChunks, tempDir)
	if err == nil {
		t.Error("Expected error for file size 0")
	}

	// Test numChunks <= 0
	chunks, err = CalculateChunks(url, fileSize, 0, tempDir)
	if err != nil {
		t.Fatalf("CalculateChunks failed with numChunks=0: %v", err)
	}
	if len(chunks) != 1 {
		t.Errorf("Expected 1 chunk for numChunks=0, got %d", len(chunks))
	}
}

func TestValidateChunks(t *testing.T) {
	// Create some test chunks
	chunk1 := NewChunk(1, "https://example.com/test.zip", 0, 1000, "/tmp")
	chunk2 := NewChunk(2, "https://example.com/test.zip", 1001, 2000, "/tmp")
	chunk3 := NewChunk(3, "https://example.com/test.zip", 2001, 3000, "/tmp")

	chunks := []*Chunk{chunk1, chunk2, chunk3}

	// Test with all incomplete
	if ValidateChunks(chunks) {
		t.Error("Expected ValidateChunks to return false for incomplete chunks")
	}

	// Test with some complete
	chunk1.Completed = true
	if ValidateChunks(chunks) {
		t.Error("Expected ValidateChunks to return false for partially complete chunks")
	}

	// Test with all complete
	chunk2.Completed = true
	chunk3.Completed = true
	if !ValidateChunks(chunks) {
		t.Error("Expected ValidateChunks to return true for all complete chunks")
	}

	// Test with one failed
	chunk2.Failed = true
	if ValidateChunks(chunks) {
		t.Error("Expected ValidateChunks to return false for failed chunk")
	}
}

func TestGetTempFilePaths(t *testing.T) {
	// Create some test chunks
	chunk1 := NewChunk(1, "https://example.com/test.zip", 0, 1000, "/tmp")
	chunk2 := NewChunk(2, "https://example.com/test.zip", 1001, 2000, "/tmp")

	chunks := []*Chunk{chunk1, chunk2}

	paths := GetTempFilePaths(chunks)

	if len(paths) != 2 {
		t.Errorf("Expected 2 paths, got %d", len(paths))
	}

	if paths[0] != chunk1.TempFile {
		t.Errorf("Expected first path to be %s, got %s", chunk1.TempFile, paths[0])
	}

	if paths[1] != chunk2.TempFile {
		t.Errorf("Expected second path to be %s, got %s", chunk2.TempFile, paths[1])
	}
}
