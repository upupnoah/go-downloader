package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCreateHTTPRequest(t *testing.T) {
	// Test normal request
	req, err := CreateHTTPRequest("GET", "https://example.com", -1, -1)
	if err != nil {
		t.Fatalf("CreateHTTPRequest failed: %v", err)
	}

	if req.Method != "GET" {
		t.Errorf("Expected method to be GET, got %s", req.Method)
	}

	if req.URL.String() != "https://example.com" {
		t.Errorf("Expected URL to be https://example.com, got %s", req.URL.String())
	}

	// Test range request
	req, err = CreateHTTPRequest("GET", "https://example.com", 100, 200)
	if err != nil {
		t.Fatalf("CreateHTTPRequest with range failed: %v", err)
	}

	rangeHeader := req.Header.Get("Range")
	expectedRange := "bytes=100-200"
	if rangeHeader != expectedRange {
		t.Errorf("Expected Range header to be %s, got %s", expectedRange, rangeHeader)
	}
}

func TestGetContentLength(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "1000")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Test success case
	length, err := GetContentLength(server.URL)
	if err != nil {
		t.Fatalf("GetContentLength failed: %v", err)
	}

	if length != 1000 {
		t.Errorf("Expected content length to be 1000, got %d", length)
	}

	// Test error case - invalid URL
	_, err = GetContentLength("http://invalid-url-that-does-not-exist.example")
	if err == nil {
		t.Error("Expected GetContentLength to fail with invalid URL")
	}

	// Test error case - server error
	errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer errorServer.Close()

	_, err = GetContentLength(errorServer.URL)
	if err == nil {
		t.Error("Expected GetContentLength to fail with server error")
	}
}

func TestCheckRangeSupport(t *testing.T) {
	// Server that supports range requests
	rangeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Accept-Ranges", "bytes")
		w.WriteHeader(http.StatusOK)
	}))
	defer rangeServer.Close()

	// Test server that supports ranges
	supported, err := CheckRangeSupport(rangeServer.URL)
	if err != nil {
		t.Fatalf("CheckRangeSupport failed: %v", err)
	}

	if !supported {
		t.Error("Expected range support to be true")
	}

	// Server that doesn't support range requests
	noRangeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// No Accept-Ranges header
		w.WriteHeader(http.StatusOK)
	}))
	defer noRangeServer.Close()

	// Test server that doesn't support ranges
	supported, err = CheckRangeSupport(noRangeServer.URL)
	if err != nil {
		t.Fatalf("CheckRangeSupport failed: %v", err)
	}

	if supported {
		t.Error("Expected range support to be false")
	}
}

func TestDoRequestWithRetry(t *testing.T) {
	client := &http.Client{
		Timeout: 1 * time.Second,
	}

	// Server that succeeds
	successServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer successServer.Close()

	// Test successful request
	req, err := http.NewRequest("GET", successServer.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := DoRequestWithRetry(client, req)
	if err != nil {
		t.Fatalf("DoRequestWithRetry failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Server that fails then succeeds (tests retry logic)
	attemptCount := 0
	retryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount <= 2 {
			// Fail for the first 2 attempts
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// Succeed on third attempt
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success after retry"))
	}))
	defer retryServer.Close()

	// Test request with retry
	retryReq, err := http.NewRequest("GET", retryServer.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create retry request: %v", err)
	}

	retryResp, err := DoRequestWithRetry(client, retryReq)
	if err != nil {
		t.Fatalf("DoRequestWithRetry with retry failed: %v", err)
	}
	defer retryResp.Body.Close()

	if retryResp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d after retry, got %d", http.StatusOK, retryResp.StatusCode)
	}

	if attemptCount < 3 {
		t.Errorf("Expected at least 3 attempts, got %d", attemptCount)
	}
}
