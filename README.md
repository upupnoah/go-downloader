# Go Downloader

A multi-threaded downloader written in Go for efficient large file downloads.

## Features

- Multi-threaded concurrent downloads for increased speed
- Real-time display of download progress, speed, and estimated time remaining
- Support for resumable downloads and chunked downloading
- Automatic detection of server support for range requests
- Automatic fallback to single-threaded download (when server doesn't support range requests)
- Failure retry mechanism
- Simple and easy-to-use command line interface

## Installation

### From Source

```bash
git clone https://github.com/yourusername/go-downloader.git
cd go-downloader
go build -o godownloader ./cmd/downloader
```

### Using Go Install

```bash
go install github.com/yourusername/go-downloader/cmd/downloader@latest
```

## Usage

### Command Line

```bash
# Basic usage
godownloader -url https://example.com/largefile.zip

# Specify output file
godownloader -url https://example.com/largefile.zip -output myfile.zip

# Specify thread count
godownloader -url https://example.com/largefile.zip -threads 8

# Quiet mode
godownloader -url https://example.com/largefile.zip -quiet

# View help
godownloader -help
```

### As a Library

```go
package main

import (
    "fmt"
    "os"

    "github.com/yourusername/go-downloader/pkg/downloader"
)

func main() {
    // Basic usage
    dl := downloader.New("https://example.com/largefile.zip", "output.zip", 4)
    err := dl.Download()
    if err != nil {
        fmt.Printf("Download failed: %v\n", err)
        os.Exit(1)
    }

    // Using custom options
    options := downloader.Options{
        OutputPath: "custom-output.zip",
        NumThreads: 8,
        MaxRetries: 5,
        Verbose:    true,
    }
    dl = downloader.WithOptions("https://example.com/largefile.zip", options)
    err = dl.Download()
    if err != nil {
        fmt.Printf("Download failed: %v\n", err)
        os.Exit(1)
    }
}
```

## Command Line Parameters

| Parameter  | Description                          | Default                     |
| ---------- | ------------------------------------ | --------------------------- |
| `-url`     | URL to download                      | -                           |
| `-output`  | Output file path                     | Filename extracted from URL |
| `-threads` | Number of download threads           | Number of CPU cores         |
| `-retries` | Number of retry attempts on failure  | 3                           |
| `-quiet`   | Quiet mode, only show error messages | false                       |
| `-version` | Display version information          | false                       |

## How It Works

1. Sends a HEAD request to get file size and range support information
2. If the server supports range requests, splits the file into multiple chunks
3. Creates a worker for each chunk and downloads concurrently
4. Tracks and displays download progress in real-time
5. Merges all chunks into the final file
6. Cleans up temporary files

## License

MIT
