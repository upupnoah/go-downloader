package downloader

import (
	"github.com/godownloader/internal/download"
)

// Options configures the downloader
type Options struct {
	// Output file path. If empty, derived from URL
	OutputPath string

	// Number of concurrent downloading threads
	// If <= 0, defaults to number of CPU cores
	NumThreads int

	// Maximum number of retries for failed chunks
	MaxRetries int

	// Verbose output
	Verbose bool
}

// Downloader is the public downloader interface
type Downloader struct {
	url     string
	options Options
	impl    *download.Downloader
}

// New creates a new downloader with the given URL and output path
func New(url, outputPath string, numThreads int) *Downloader {
	return &Downloader{
		url: url,
		options: Options{
			OutputPath: outputPath,
			NumThreads: numThreads,
			MaxRetries: 3,
			Verbose:    true,
		},
	}
}

// WithOptions creates a new downloader with custom options
func WithOptions(url string, options Options) *Downloader {
	return &Downloader{
		url:     url,
		options: options,
	}
}

// Download starts the download process
func (d *Downloader) Download() error {
	d.impl = download.NewDownloader(d.url, d.options.OutputPath, d.options.NumThreads)
	d.impl.SetMaxRetries(d.options.MaxRetries)
	d.impl.SetVerbose(d.options.Verbose)

	return d.impl.Start()
}

// SetVerbose sets the verbose flag
func (d *Downloader) SetVerbose(verbose bool) {
	d.options.Verbose = verbose
	if d.impl != nil {
		d.impl.SetVerbose(verbose)
	}
}

// SetMaxRetries sets the maximum number of retries for failed chunks
func (d *Downloader) SetMaxRetries(maxRetries int) {
	d.options.MaxRetries = maxRetries
	if d.impl != nil {
		d.impl.SetMaxRetries(maxRetries)
	}
}

// DefaultOptions returns the default options
func DefaultOptions() Options {
	return Options{
		OutputPath: "",
		NumThreads: 0, // Will use CPU count
		MaxRetries: 3,
		Verbose:    true,
	}
}
