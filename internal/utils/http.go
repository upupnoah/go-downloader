package utils

import (
	"fmt"
	"net/http"
	"time"
)

const (
	maxRetries    = 3
	retryInterval = 2 * time.Second
	userAgent     = "Go-Downloader/1.0"
)

// GetContentLength sends a HEAD request to get file size
func GetContentLength(url string) (int64, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Set("User-Agent", userAgent)

	var resp *http.Response
	var retryCount int

	for retryCount < maxRetries {
		resp, err = client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return resp.ContentLength, nil
			}
			err = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		retryCount++
		if retryCount < maxRetries {
			time.Sleep(retryInterval)
		}
	}

	return 0, err
}

// CheckRangeSupport checks if the server supports range requests
func CheckRangeSupport(url string) (bool, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return false, err
	}

	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	acceptRanges := resp.Header.Get("Accept-Ranges")
	return acceptRanges == "bytes", nil
}

// CreateHTTPRequest creates an HTTP request with appropriate headers
func CreateHTTPRequest(method, url string, rangeStart, rangeEnd int64) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", userAgent)

	if rangeStart >= 0 && rangeEnd >= 0 {
		rangeHeader := fmt.Sprintf("bytes=%d-%d", rangeStart, rangeEnd)
		req.Header.Set("Range", rangeHeader)
	}

	return req, nil
}

// DoRequestWithRetry performs an HTTP request with retry logic
func DoRequestWithRetry(client *http.Client, req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error
	var retryCount int

	for retryCount < maxRetries {
		resp, err = client.Do(req)
		if err == nil {
			if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusPartialContent {
				return resp, nil
			}
			resp.Body.Close()
			err = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		retryCount++
		if retryCount < maxRetries {
			time.Sleep(retryInterval)
		}
	}

	return nil, err
}
