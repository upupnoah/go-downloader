package download

import (
	"testing"
	"time"
)

func TestNewProgress(t *testing.T) {
	totalSize := int64(1024 * 1024) // 1MB
	chunks := []*Chunk{
		NewChunk(1, "https://example.com/test.zip", 0, 500, "/tmp"),
		NewChunk(2, "https://example.com/test.zip", 501, 1000, "/tmp"),
	}

	progress := NewProgress(totalSize, chunks)

	if progress.TotalSize != totalSize {
		t.Errorf("Expected TotalSize %d, got %d", totalSize, progress.TotalSize)
	}

	if progress.Downloaded != 0 {
		t.Errorf("Expected Downloaded to be 0, got %d", progress.Downloaded)
	}

	if progress.Chunks == nil || len(progress.Chunks) != 2 {
		t.Errorf("Expected 2 chunks, got %v", progress.Chunks)
	}

	// Test without chunks
	progress = NewProgress(totalSize, nil)
	if progress.TotalSize != totalSize {
		t.Errorf("Expected TotalSize %d, got %d", totalSize, progress.TotalSize)
	}
}

func TestUpdate(t *testing.T) {
	totalSize := int64(1000)

	// Test with chunks
	chunks := []*Chunk{
		NewChunk(1, "https://example.com/test.zip", 0, 499, "/tmp"),
		NewChunk(2, "https://example.com/test.zip", 500, 999, "/tmp"),
	}

	progress := NewProgress(totalSize, chunks)

	// Update progress on chunks
	chunks[0].Downloaded = 250
	chunks[1].Downloaded = 300

	// Update progress
	progress.Update()

	expectedDownloaded := int64(550) // 250 + 300
	if progress.Downloaded != expectedDownloaded {
		t.Errorf("Expected Downloaded to be %d, got %d", expectedDownloaded, progress.Downloaded)
	}

	expectedPercent := 55.0 // 550/1000 * 100
	if progress.ProgressPercent != expectedPercent {
		t.Errorf("Expected ProgressPercent to be %f, got %f", expectedPercent, progress.ProgressPercent)
	}

	// Test without chunks
	progress = NewProgress(totalSize, nil)
	progress.Downloaded = 750

	progress.Update()

	expectedPercent = 75.0 // 750/1000 * 100
	if progress.ProgressPercent != expectedPercent {
		t.Errorf("Expected ProgressPercent to be %f, got %f", expectedPercent, progress.ProgressPercent)
	}
}

func TestCreateProgressBar(t *testing.T) {
	totalSize := int64(100)
	progress := NewProgress(totalSize, nil)

	// Test empty progress bar
	progress.Downloaded = 0
	progress.createProgressBar()

	// Progress bar should be 50 chars wide + 2 for brackets
	expectedLength := 52
	if len(progress.ProgressBar) != expectedLength {
		t.Errorf("Expected progress bar length %d, got %d", expectedLength, len(progress.ProgressBar))
	}

	// Bar should start with [ and end with ]
	if progress.ProgressBar[0] != '[' || progress.ProgressBar[expectedLength-1] != ']' {
		t.Errorf("Progress bar should start with [ and end with ], got %s", progress.ProgressBar)
	}

	// Test half-filled progress bar
	progress.Downloaded = 50
	progress.createProgressBar()

	// First half should be filled with =
	for i := 1; i <= 25; i++ {
		if progress.ProgressBar[i] != '=' {
			t.Errorf("Expected = at position %d, got %c", i, progress.ProgressBar[i])
			break
		}
	}

	// There should be a > at position 26
	if progress.ProgressBar[26] != '>' {
		t.Errorf("Expected > at position 26, got %c", progress.ProgressBar[26])
	}

	// Test full progress bar
	progress.Downloaded = 100
	progress.createProgressBar()

	// All 50 positions should be filled with =
	for i := 1; i <= 50; i++ {
		if progress.ProgressBar[i] != '=' {
			t.Errorf("Expected = at position %d, got %c", i, progress.ProgressBar[i])
			break
		}
	}
}

func TestProgressETACalculation(t *testing.T) {
	totalSize := int64(1000)
	progress := NewProgress(totalSize, nil)

	// Set downloaded to half the total
	progress.Downloaded = 500

	// Set start time in the past (5 seconds ago)
	progress.StartTime = time.Now().Add(-5 * time.Second)

	// Update progress (calculates speed and ETA)
	progress.Update()

	// Speed should be about 100 bytes per second (500 bytes / 5 seconds)
	expectedSpeed := float64(100)
	tolerance := float64(10) // Allow some tolerance due to timing variations

	if progress.AverageSpeed < expectedSpeed-tolerance || progress.AverageSpeed > expectedSpeed+tolerance {
		t.Errorf("Expected speed around %f, got %f", expectedSpeed, progress.AverageSpeed)
	}

	// ETA should be about 5 seconds (500 remaining bytes at 100 bytes/sec)
	expectedETA := 5 * time.Second
	etaTolerance := 1 * time.Second

	if progress.ETA < expectedETA-etaTolerance || progress.ETA > expectedETA+etaTolerance {
		t.Errorf("Expected ETA around %v, got %v", expectedETA, progress.ETA)
	}
}

func TestStartTracking(t *testing.T) {
	// This is more of an integration test and might be hard to test
	// directly without mocking time.Ticker, so we'll do a simple test

	totalSize := int64(100)
	progress := NewProgress(totalSize, nil)
	progress.Downloaded = 100 // Mark as complete to ensure it finishes quickly

	// Create stop channel and close immediately to ensure the function returns
	stopChan := make(chan struct{})
	close(stopChan)

	// This should return immediately
	progress.StartTracking(100*time.Millisecond, stopChan)

	// If we reach here, the function returned as expected
}

func TestPrintSummary(t *testing.T) {
	// This is mainly testing that the function doesn't panic
	// It's hard to test actual console output without mocking fmt

	totalSize := int64(1000)
	progress := NewProgress(totalSize, nil)
	progress.Downloaded = 1000
	progress.StartTime = time.Now().Add(-10 * time.Second)
	progress.AverageSpeed = 100 // 100 bytes per second

	// This should not panic
	progress.PrintSummary()
}
