package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/godownloader/pkg/downloader"
)

var (
	version = "1.0.0"
)

func main() {
	// Parse command-line flags
	url := flag.String("url", "", "URL to download (required)")
	output := flag.String("output", "", "Output file path (default: filename from URL)")
	threads := flag.Int("threads", runtime.NumCPU(), "Number of download threads (default: number of CPU cores)")
	maxRetries := flag.Int("retries", 3, "Maximum number of retries for failed chunks")
	quiet := flag.Bool("quiet", false, "Suppress output except for errors")
	showVersion := flag.Bool("version", false, "Show version information")

	flag.Parse()

	// Check if user wants version info
	if *showVersion {
		fmt.Printf("Go Downloader v%s\n", version)
		fmt.Printf("Go version: %s\n", runtime.Version())
		os.Exit(0)
	}

	// Check for required URL parameter
	if *url == "" {
		if len(flag.Args()) > 0 {
			// Allow URL as positional argument
			*url = flag.Args()[0]
		} else {
			fmt.Println("Error: URL is required.")
			fmt.Println("\nUsage:")
			flag.PrintDefaults()
			os.Exit(1)
		}
	}

	// Set up signal handling for clean termination
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nDownload canceled. Cleaning up...")
		os.Exit(1)
	}()

	// Create and configure downloader
	dl := downloader.New(*url, *output, *threads)
	dl.SetMaxRetries(*maxRetries)
	dl.SetVerbose(!*quiet)

	// Start download
	err := dl.Download()

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	os.Exit(0)
}
