package download

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Chunk represents a portion of the file to be downloaded
type Chunk struct {
	ID         int
	URL        string
	Start      int64
	End        int64
	TempFile   string
	Size       int64
	Downloaded int64
	Completed  bool
	Failed     bool
	RetryCount int
	mu         sync.Mutex
}

// NewChunk creates a new chunk
func NewChunk(id int, url string, start, end int64, tempDir string) *Chunk {
	return &Chunk{
		ID:         id,
		URL:        url,
		Start:      start,
		End:        end,
		Size:       end - start + 1,
		TempFile:   filepath.Join(tempDir, fmt.Sprintf("chunk_%d", id)),
		Downloaded: 0,
		Completed:  false,
		Failed:     false,
		RetryCount: 0,
	}
}

// UpdateProgress updates the downloaded bytes count
func (c *Chunk) UpdateProgress(bytesRead int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Downloaded += bytesRead
	if c.Downloaded >= c.Size {
		c.Completed = true
	}
}

// MarkFailed marks the chunk as failed
func (c *Chunk) MarkFailed() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Failed = true
	c.RetryCount++
}

// ResetForRetry resets the chunk for a retry attempt
func (c *Chunk) ResetForRetry() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Remove the temp file if it exists
	if _, err := os.Stat(c.TempFile); err == nil {
		os.Remove(c.TempFile)
	}

	c.Downloaded = 0
	c.Failed = false
}

// GetProgress returns the current progress as a percentage
func (c *Chunk) GetProgress() float64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Size <= 0 {
		return 0
	}
	return float64(c.Downloaded) * 100 / float64(c.Size)
}

// CalculateChunks divides a file into multiple chunks
func CalculateChunks(url string, fileSize int64, numChunks int, tempDir string) ([]*Chunk, error) {
	if fileSize <= 0 {
		return nil, fmt.Errorf("invalid file size: %d", fileSize)
	}

	if numChunks <= 0 {
		numChunks = 1
	}

	// Ensure the number of chunks doesn't exceed the file size
	if fileSize < int64(numChunks) {
		numChunks = int(fileSize)
	}

	chunkSize := fileSize / int64(numChunks)
	chunks := make([]*Chunk, numChunks)

	var start, end int64
	for i := 0; i < numChunks; i++ {
		start = int64(i) * chunkSize
		if i == numChunks-1 {
			end = fileSize - 1
		} else {
			end = start + chunkSize - 1
		}

		chunks[i] = NewChunk(i, url, start, end, tempDir)
	}

	return chunks, nil
}

// ValidateChunks checks if all chunks are completed
func ValidateChunks(chunks []*Chunk) bool {
	for _, chunk := range chunks {
		if !chunk.Completed || chunk.Failed {
			return false
		}
	}
	return true
}

// GetTempFilePaths returns the paths to all chunk temporary files
func GetTempFilePaths(chunks []*Chunk) []string {
	paths := make([]string, len(chunks))
	for i, chunk := range chunks {
		paths[i] = chunk.TempFile
	}
	return paths
}
