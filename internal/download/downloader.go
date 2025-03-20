package download

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"runtime"
	"time"

	"github.com/godownloader/internal/utils"
)

// Downloader represents the main downloader
type Downloader struct {
	URL            string
	OutputPath     string
	NumThreads     int
	ChunkSize      int64
	TempDir        string
	ContentLength  int64
	SupportsRanges bool
	Chunks         []*Chunk
	Progress       *Progress
	Client         *http.Client
	MaxRetries     int
	Verbose        bool
}

// NewDownloader creates a new downloader
func NewDownloader(url, outputPath string, numThreads int) *Downloader {
	if numThreads <= 0 {
		numThreads = runtime.NumCPU()
	}

	if outputPath == "" {
		outputPath = filepath.Base(url)
	}

	return &Downloader{
		URL:        url,
		OutputPath: outputPath,
		NumThreads: numThreads,
		TempDir:    "",
		MaxRetries: 3,
		Verbose:    true,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Start begins the download process
func (d *Downloader) Start() error {
	if d.Verbose {
		fmt.Printf("Starting download of %s with %d threads\n", d.URL, d.NumThreads)
	}

	// Create temporary directory
	tempDir, err := utils.CreateTempDir("downloader")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	d.TempDir = tempDir
	defer utils.CleanupTempDir(tempDir)

	// Get content length and check if server supports range requests
	contentLength, err := utils.GetContentLength(d.URL)
	if err != nil {
		return fmt.Errorf("failed to get content length: %w", err)
	}

	supportsRanges, err := utils.CheckRangeSupport(d.URL)
	if err != nil {
		return fmt.Errorf("failed to check range support: %w", err)
	}

	d.ContentLength = contentLength
	d.SupportsRanges = supportsRanges

	// If the server doesn't support range requests or if using single thread,
	// fall back to single-threaded download
	if !supportsRanges || d.NumThreads == 1 {
		return d.downloadSingleThreaded()
	}

	return d.downloadMultiThreaded()
}

// downloadMultiThreaded handles multi-threaded download
func (d *Downloader) downloadMultiThreaded() error {
	if d.Verbose {
		fmt.Println("Using multi-threaded download")
	}

	// Calculate chunks
	chunks, err := CalculateChunks(d.URL, d.ContentLength, d.NumThreads, d.TempDir)
	if err != nil {
		return fmt.Errorf("failed to calculate chunks: %w", err)
	}
	d.Chunks = chunks

	// Create progress tracker
	progress := NewProgress(d.ContentLength, chunks)
	d.Progress = progress

	// Start progress tracking
	stopProgressChan := make(chan struct{})
	go progress.StartTracking(100*time.Millisecond, stopProgressChan)

	// Start worker pool
	results, err := StartWorkerPool(d.NumThreads, chunks)
	if err != nil {
		close(stopProgressChan)
		return fmt.Errorf("download failed: %w", err)
	}

	// Check results
	var hasFailures bool
	for _, result := range results {
		if result.Error != nil {
			hasFailures = true
			if d.Verbose {
				fmt.Printf("Chunk %d failed: %v\n", result.Chunk.ID, result.Error)
			}
		}
	}

	// Retry failed chunks
	if hasFailures {
		if d.Verbose {
			fmt.Println("Retrying failed chunks...")
		}

		err = RetryFailedChunks(d.Chunks, d.MaxRetries)
		if err != nil {
			close(stopProgressChan)
			return fmt.Errorf("retry failed: %w", err)
		}
	}

	// Verify all chunks are complete
	if !ValidateChunks(d.Chunks) {
		close(stopProgressChan)
		return fmt.Errorf("download incomplete, some chunks failed")
	}

	// Signal progress tracking to stop
	close(stopProgressChan)

	// Print summary if verbose
	if d.Verbose {
		d.Progress.PrintSummary()
	}

	// Merge chunks
	if d.Verbose {
		fmt.Println("Merging chunks...")
	}

	err = d.mergeChunks()
	if err != nil {
		return fmt.Errorf("failed to merge chunks: %w", err)
	}

	if d.Verbose {
		fmt.Printf("Download completed: %s\n", d.OutputPath)
	}

	return nil
}

// downloadSingleThreaded downloads the file using a single thread
func (d *Downloader) downloadSingleThreaded() error {
	if d.Verbose {
		if !d.SupportsRanges {
			fmt.Println("Server doesn't support range requests. Using single-threaded download.")
		} else {
			fmt.Println("Using single-threaded download.")
		}
	}

	// Create the request
	req, err := utils.CreateHTTPRequest("GET", d.URL, -1, -1)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Send the request
	resp, err := utils.DoRequestWithRetry(d.Client, req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Create the output file
	file, err := utils.CreateFile(d.OutputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Create progress tracker for single-threaded download
	var totalSize int64 = resp.ContentLength
	if totalSize <= 0 {
		totalSize = d.ContentLength
	}

	progress := NewProgress(totalSize, nil)
	d.Progress = progress

	stopProgressChan := make(chan struct{})
	go progress.StartTracking(100*time.Millisecond, stopProgressChan)

	// Download the file
	buffer := make([]byte, 32*1024) // 32KB buffer
	var downloaded int64

	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			_, writeErr := file.Write(buffer[:n])
			if writeErr != nil {
				close(stopProgressChan)
				return fmt.Errorf("failed to write to file: %w", writeErr)
			}

			downloaded += int64(n)
			progress.Downloaded = downloaded
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			close(stopProgressChan)
			return fmt.Errorf("failed to read response: %w", err)
		}
	}

	// Signal progress tracking to stop
	close(stopProgressChan)

	// Print summary if verbose
	if d.Verbose {
		progress.PrintSummary()
	}

	if d.Verbose {
		fmt.Printf("Download completed: %s\n", d.OutputPath)
	}

	return nil
}

// mergeChunks combines all downloaded chunks into the final file
func (d *Downloader) mergeChunks() error {
	// Get paths to all chunk files
	paths := GetTempFilePaths(d.Chunks)

	// Merge files
	err := utils.MergeFiles(d.OutputPath, paths)
	if err != nil {
		return fmt.Errorf("failed to merge chunks: %w", err)
	}

	return nil
}

// SetVerbose sets the verbose flag
func (d *Downloader) SetVerbose(verbose bool) {
	d.Verbose = verbose
}

// SetMaxRetries sets the maximum number of retries for failed chunks
func (d *Downloader) SetMaxRetries(maxRetries int) {
	d.MaxRetries = maxRetries
}
