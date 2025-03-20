package download

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/godownloader/internal/utils"
)

// Worker represents a download worker
type Worker struct {
	ID        int
	JobQueue  <-chan *Chunk
	Results   chan<- *Result
	WaitGroup *sync.WaitGroup
	Client    *http.Client
}

// Result represents the result of a chunk download
type Result struct {
	Chunk *Chunk
	Error error
}

// NewWorker creates a new worker
func NewWorker(id int, jobQueue <-chan *Chunk, results chan<- *Result, wg *sync.WaitGroup) *Worker {
	return &Worker{
		ID:        id,
		JobQueue:  jobQueue,
		Results:   results,
		WaitGroup: wg,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Start begins the worker's processing loop
func (w *Worker) Start() {
	go func() {
		for chunk := range w.JobQueue {
			result := &Result{
				Chunk: chunk,
			}

			err := w.downloadChunk(chunk)
			if err != nil {
				result.Error = err
				chunk.MarkFailed()
			}

			w.Results <- result
			w.WaitGroup.Done()
		}
	}()
}

// downloadChunk downloads a specific chunk
func (w *Worker) downloadChunk(chunk *Chunk) error {
	// Create the temp file
	file, err := os.Create(chunk.TempFile)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer file.Close()

	// Create the request with range
	req, err := utils.CreateHTTPRequest("GET", chunk.URL, chunk.Start, chunk.End)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Send the request
	resp, err := utils.DoRequestWithRetry(w.Client, req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Verify status code
	if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Create buffered writer for better performance
	buffer := make([]byte, 32*1024) // 32KB buffer

	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			_, writeErr := file.Write(buffer[:n])
			if writeErr != nil {
				return fmt.Errorf("failed to write to file: %w", writeErr)
			}

			// Update progress
			chunk.UpdateProgress(int64(n))
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to read response: %w", err)
		}
	}

	return nil
}

// StartWorkerPool initializes and starts a pool of workers
func StartWorkerPool(numWorkers int, chunks []*Chunk) ([]*Result, error) {
	var wg sync.WaitGroup
	jobQueue := make(chan *Chunk, len(chunks))
	results := make(chan *Result, len(chunks))

	// Create and start workers
	for i := 0; i < numWorkers; i++ {
		worker := NewWorker(i, jobQueue, results, &wg)
		worker.Start()
	}

	// Add jobs to the queue
	for _, chunk := range chunks {
		wg.Add(1)
		jobQueue <- chunk
	}

	// Close the job queue
	close(jobQueue)

	// Wait for all downloads to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var downloadResults []*Result
	for result := range results {
		downloadResults = append(downloadResults, result)
	}

	return downloadResults, nil
}

// RetryFailedChunks attempts to download failed chunks
func RetryFailedChunks(chunks []*Chunk, maxRetries int) error {
	var failedChunks []*Chunk

	// Find failed chunks that haven't exceeded retry limit
	for _, chunk := range chunks {
		if chunk.Failed && chunk.RetryCount < maxRetries {
			chunk.ResetForRetry()
			failedChunks = append(failedChunks, chunk)
		}
	}

	if len(failedChunks) == 0 {
		return nil
	}

	// Retry failed chunks
	results, err := StartWorkerPool(len(failedChunks), failedChunks)
	if err != nil {
		return err
	}

	// Check for errors
	for _, result := range results {
		if result.Error != nil {
			return fmt.Errorf("chunk %d failed: %w", result.Chunk.ID, result.Error)
		}
	}

	return nil
}
