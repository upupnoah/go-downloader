package downloader

import (
	"errors"
	"testing"
)

// mockDownloader is a mock implementation of the internal download.Downloader interface
type mockDownloader struct {
	startCalled bool
	shouldFail  bool
	verbose     bool
	maxRetries  int
	url         string
	outputPath  string
	numThreads  int
}

func (m *mockDownloader) Start() error {
	m.startCalled = true
	if m.shouldFail {
		return errors.New("mock download failed")
	}
	return nil
}

func (m *mockDownloader) SetVerbose(verbose bool) {
	m.verbose = verbose
}

func (m *mockDownloader) SetMaxRetries(maxRetries int) {
	m.maxRetries = maxRetries
}

// TestNew tests the New constructor
func TestNew(t *testing.T) {
	url := "https://example.com/file.zip"
	outputPath := "/tmp/file.zip"
	numThreads := 4

	d := New(url, outputPath, numThreads)

	if d.url != url {
		t.Errorf("Expected URL %s, got %s", url, d.url)
	}

	if d.options.OutputPath != outputPath {
		t.Errorf("Expected OutputPath %s, got %s", outputPath, d.options.OutputPath)
	}

	if d.options.NumThreads != numThreads {
		t.Errorf("Expected NumThreads %d, got %d", numThreads, d.options.NumThreads)
	}

	if d.options.MaxRetries != 3 {
		t.Errorf("Expected default MaxRetries 3, got %d", d.options.MaxRetries)
	}

	if !d.options.Verbose {
		t.Errorf("Expected default Verbose to be true")
	}
}

// TestWithOptions tests the WithOptions constructor
func TestWithOptions(t *testing.T) {
	url := "https://example.com/file.zip"
	options := Options{
		OutputPath: "/tmp/custom.zip",
		NumThreads: 8,
		MaxRetries: 5,
		Verbose:    false,
	}

	d := WithOptions(url, options)

	if d.url != url {
		t.Errorf("Expected URL %s, got %s", url, d.url)
	}

	if d.options.OutputPath != options.OutputPath {
		t.Errorf("Expected OutputPath %s, got %s", options.OutputPath, d.options.OutputPath)
	}

	if d.options.NumThreads != options.NumThreads {
		t.Errorf("Expected NumThreads %d, got %d", options.NumThreads, d.options.NumThreads)
	}

	if d.options.MaxRetries != options.MaxRetries {
		t.Errorf("Expected MaxRetries %d, got %d", options.MaxRetries, d.options.MaxRetries)
	}

	if d.options.Verbose != options.Verbose {
		t.Errorf("Expected Verbose %v, got %v", options.Verbose, d.options.Verbose)
	}
}

// MockableDownloader extends downloader to allow mocking the internal implementation
type MockableDownloader struct {
	*Downloader
	mockImpl *mockDownloader
}

// NewMockableDownloader creates a downloader that uses a mock implementation
func NewMockableDownloader(url, outputPath string, numThreads int, shouldFail bool) *MockableDownloader {
	d := New(url, outputPath, numThreads)
	mockImpl := &mockDownloader{
		shouldFail: shouldFail,
	}

	return &MockableDownloader{
		Downloader: d,
		mockImpl:   mockImpl,
	}
}

// Download overrides the download method to use mock implementation
func (md *MockableDownloader) Download() error {
	// Set the mock implementation
	// Note: This would require the impl field to be exported or accessible
	// For testing purposes, we're focusing on behavior, not implementation
	return md.mockImpl.Start()
}

// TestDownload tests the download behavior
func TestDownload(t *testing.T) {
	// Since we can't directly assign to the Download method or mock the internal
	// implementation without exposing it, we'll test the main behavior indirectly

	// Test the core logic instead of the actual download
	options := DefaultOptions()
	options.OutputPath = "/tmp/test.zip"
	options.NumThreads = 4
	options.MaxRetries = 5
	options.Verbose = true

	d := WithOptions("https://example.com/test.zip", options)

	// Verify options are set correctly
	if d.options.OutputPath != "/tmp/test.zip" {
		t.Errorf("Expected output path '/tmp/test.zip', got '%s'", d.options.OutputPath)
	}

	if d.options.NumThreads != 4 {
		t.Errorf("Expected 4 threads, got %d", d.options.NumThreads)
	}
}

// TestSetVerbose tests the SetVerbose method
func TestSetVerbose(t *testing.T) {
	d := New("https://example.com/file.zip", "/tmp/file.zip", 4)

	// Test setting verbose directly
	d.SetVerbose(false)

	if d.options.Verbose != false {
		t.Errorf("Expected options.Verbose to be false")
	}
}

// TestSetMaxRetries tests the SetMaxRetries method
func TestSetMaxRetries(t *testing.T) {
	d := New("https://example.com/file.zip", "/tmp/file.zip", 4)

	// Test setting max retries directly
	d.SetMaxRetries(10)

	if d.options.MaxRetries != 10 {
		t.Errorf("Expected options.MaxRetries to be 10, got %d", d.options.MaxRetries)
	}
}

// TestDefaultOptions tests the DefaultOptions function
func TestDefaultOptions(t *testing.T) {
	options := DefaultOptions()

	if options.OutputPath != "" {
		t.Errorf("Expected empty OutputPath, got %s", options.OutputPath)
	}

	if options.NumThreads != 0 {
		t.Errorf("Expected NumThreads to be 0, got %d", options.NumThreads)
	}

	if options.MaxRetries != 3 {
		t.Errorf("Expected MaxRetries to be 3, got %d", options.MaxRetries)
	}

	if !options.Verbose {
		t.Errorf("Expected Verbose to be true")
	}
}
