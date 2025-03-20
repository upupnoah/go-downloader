package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// CreateTempDir creates a temporary directory for storing chunks
func CreateTempDir(prefix string) (string, error) {
	tempDir, err := os.MkdirTemp("", prefix)
	if err != nil {
		return "", fmt.Errorf("failed to create temporary directory: %w", err)
	}
	return tempDir, nil
}

// CleanupTempDir removes the temporary directory and all its contents
func CleanupTempDir(path string) error {
	return os.RemoveAll(path)
}

// CreateFile creates a new file at the specified path
func CreateFile(path string) (*os.File, error) {
	// Ensure the directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Create the file
	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}

	return file, nil
}

// MergeFiles merges multiple files into a single output file
func MergeFiles(outputPath string, inputPaths []string) error {
	outFile, err := CreateFile(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	buffer := make([]byte, 32*1024) // 32KB buffer for efficient copying

	for _, inputPath := range inputPaths {
		inFile, err := os.Open(inputPath)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %w", inputPath, err)
		}

		_, err = io.CopyBuffer(outFile, inFile, buffer)
		inFile.Close() // Close each file after copying

		if err != nil {
			return fmt.Errorf("failed to copy from %s: %w", inputPath, err)
		}
	}

	return nil
}

// GetFileSize returns the size of a file in bytes
func GetFileSize(path string) (int64, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return 0, fmt.Errorf("failed to stat file: %w", err)
	}
	return fileInfo.Size(), nil
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
