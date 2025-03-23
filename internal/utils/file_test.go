package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateFile(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "file_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test creating a file
	filePath := filepath.Join(tempDir, "test.txt")
	file, err := CreateFile(filePath)
	if err != nil {
		t.Fatalf("CreateFile failed: %v", err)
	}
	defer file.Close()

	// Verify file exists
	_, err = os.Stat(filePath)
	if err != nil {
		t.Errorf("File was not created: %v", err)
	}

	// Test creating file in non-existent directory (should create directories)
	deepFilePath := filepath.Join(tempDir, "a", "b", "c", "deep.txt")
	deepFile, err := CreateFile(deepFilePath)
	if err != nil {
		t.Fatalf("CreateFile with deep path failed: %v", err)
	}
	defer deepFile.Close()

	// Verify deep file exists
	_, err = os.Stat(deepFilePath)
	if err != nil {
		t.Errorf("Deep file was not created: %v", err)
	}
}

func TestCreateTempDir(t *testing.T) {
	// Test creating temporary directory
	dirPath, err := CreateTempDir("test_prefix")
	if err != nil {
		t.Fatalf("CreateTempDir failed: %v", err)
	}
	defer os.RemoveAll(dirPath)

	// Verify directory exists
	info, err := os.Stat(dirPath)
	if err != nil {
		t.Errorf("Temp directory was not created: %v", err)
	}

	if !info.IsDir() {
		t.Error("Created path is not a directory")
	}

	// Verify prefix is used
	base := filepath.Base(dirPath)
	if len(base) <= len("test_prefix") || base[:len("test_prefix")] != "test_prefix" {
		t.Errorf("Directory name does not start with prefix: %s", base)
	}
}

func TestCleanupTempDir(t *testing.T) {
	// Create temporary directory
	dirPath, err := os.MkdirTemp("", "cleanup_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create a file inside the directory
	filePath := filepath.Join(dirPath, "test.txt")
	file, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	file.Close()

	// Verify directory exists
	_, err = os.Stat(dirPath)
	if err != nil {
		t.Fatalf("Directory does not exist before cleanup: %v", err)
	}

	// Perform cleanup
	err = CleanupTempDir(dirPath)
	if err != nil {
		t.Errorf("CleanupTempDir failed: %v", err)
	}

	// Verify directory was removed
	_, err = os.Stat(dirPath)
	if err == nil {
		t.Error("Directory still exists after cleanup")
	} else if !os.IsNotExist(err) {
		t.Errorf("Unexpected error checking directory: %v", err)
	}
}
