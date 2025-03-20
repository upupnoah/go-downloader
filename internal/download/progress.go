package download

import (
	"fmt"
	"sync"
	"time"
)

// Progress tracks the download progress
type Progress struct {
	TotalSize       int64
	Downloaded      int64
	StartTime       time.Time
	CurrentSpeed    float64
	AverageSpeed    float64
	ETA             time.Duration
	ProgressBar     string
	ProgressPercent float64
	Chunks          []*Chunk
	SpeedSamples    []float64
	mu              sync.Mutex
}

// NewProgress creates a new progress tracker
func NewProgress(totalSize int64, chunks []*Chunk) *Progress {
	return &Progress{
		TotalSize:    totalSize,
		Downloaded:   0,
		StartTime:    time.Now(),
		CurrentSpeed: 0,
		AverageSpeed: 0,
		Chunks:       chunks,
		SpeedSamples: make([]float64, 0, 10),
		mu:           sync.Mutex{},
	}
}

// Update updates the progress information
func (p *Progress) Update() {
	p.mu.Lock()
	defer p.mu.Unlock()

	var downloaded int64
	if p.Chunks != nil {
		for _, chunk := range p.Chunks {
			downloaded += chunk.Downloaded
		}
	} else {
		downloaded = p.Downloaded
	}

	// Calculate speed
	elapsed := time.Since(p.StartTime).Seconds()
	if elapsed > 0 {
		currentSpeed := float64(downloaded) / elapsed
		p.SpeedSamples = append(p.SpeedSamples, currentSpeed)

		// Keep only last 10 samples
		if len(p.SpeedSamples) > 10 {
			p.SpeedSamples = p.SpeedSamples[1:]
		}

		// Calculate average speed from samples
		var totalSpeed float64
		for _, speed := range p.SpeedSamples {
			totalSpeed += speed
		}
		p.AverageSpeed = totalSpeed / float64(len(p.SpeedSamples))
		p.CurrentSpeed = currentSpeed

		// Calculate ETA
		if p.AverageSpeed > 0 {
			remaining := float64(p.TotalSize - downloaded)
			etaSeconds := remaining / p.AverageSpeed
			p.ETA = time.Duration(etaSeconds) * time.Second
		}
	}

	p.Downloaded = downloaded

	// Calculate percentage
	if p.TotalSize > 0 {
		p.ProgressPercent = float64(downloaded) * 100 / float64(p.TotalSize)
	}

	// Create progress bar
	p.createProgressBar()
}

// createProgressBar generates a text-based progress bar
func (p *Progress) createProgressBar() {
	width := 50
	completed := int(float64(width) * float64(p.Downloaded) / float64(p.TotalSize))

	bar := "["
	for i := 0; i < width; i++ {
		if i < completed {
			bar += "="
		} else if i == completed && completed < width {
			bar += ">"
		} else {
			bar += " "
		}
	}
	bar += "]"

	p.ProgressBar = bar
}

// Print displays the current progress
func (p *Progress) Print() {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Format the output with colors if on a terminal
	fmt.Printf("\r%s %.2f%% %.2f MB/%.2f MB (%.2f MB/s) ETA: %s",
		p.ProgressBar,
		p.ProgressPercent,
		float64(p.Downloaded)/(1024*1024),
		float64(p.TotalSize)/(1024*1024),
		p.CurrentSpeed/(1024*1024),
		p.ETA.Round(time.Second),
	)
}

// StartTracking begins tracking and displaying progress
func (p *Progress) StartTracking(updateInterval time.Duration, stopChan <-chan struct{}) {
	ticker := time.NewTicker(updateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.Update()
			p.Print()
		case <-stopChan:
			p.Update()
			p.Print()
			fmt.Println() // Add a newline after the progress bar
			return
		}
	}
}

// PrintSummary prints a summary of the download
func (p *Progress) PrintSummary() {
	p.mu.Lock()
	defer p.mu.Unlock()

	totalTime := time.Since(p.StartTime).Round(time.Second)
	averageSpeedMB := p.AverageSpeed / (1024 * 1024)

	fmt.Printf("\nDownload Summary:\n")
	fmt.Printf("Total size: %.2f MB\n", float64(p.TotalSize)/(1024*1024))
	fmt.Printf("Time taken: %s\n", totalTime)
	fmt.Printf("Average speed: %.2f MB/s\n", averageSpeedMB)
}
