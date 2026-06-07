package renamer

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func init() {
	time.Local = time.UTC
}

func TestRenamer_Process_Integration(t *testing.T) {
	tempDir := t.TempDir()
	srcDir := filepath.Join(tempDir, "src")
	destDir := filepath.Join(tempDir, "dest")

	if err := os.Mkdir(srcDir, 0755); err != nil {
		t.Fatalf("failed to create src dir: %v", err)
	}

	// Create a dummy image file
	srcFile := filepath.Join(srcDir, "photo.JPG")
	if err := os.WriteFile(srcFile, []byte("dummy jpeg content"), 0644); err != nil {
		t.Fatalf("failed to write dummy photo: %v", err)
	}

	// Set a fixed modification time so the output is deterministic
	fixedTime := time.Date(2025, 6, 7, 10, 30, 0, 0, time.UTC)
	if err := os.Chtimes(srcFile, fixedTime, fixedTime); err != nil {
		t.Fatalf("failed to set time: %v", err)
	}

	opts := Options{
		Destination:        destDir,
		Recursive:          false,
		DryRun:             false,
		LowercaseExtension: true,
		Verbose:            true,
		FolderPattern:      "%Y/%Y-%m/%Y-%m-%d-%s",
		FilePattern:        "%p%Y%m%d-%H%M%S-%s%c%e",
		DefaultPrefix:      "IMG_",
		DefaultSuffix:      "test",
	}

	r := New(opts)
	err := r.Process([]string{srcDir})
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Verify file was moved to the correct destination path
	// Destination subfolder: 2025/2025-06/2025-06-07-test
	// Filename: IMG_20250607-103000-test.jpg (note extension is lowercase .jpg)
	expectedDestPath := filepath.Join(destDir, "2025", "2025-06", "2025-06-07-test", "IMG_20250607-103000-test.jpg")
	if _, err := os.Stat(expectedDestPath); os.IsNotExist(err) {
		t.Errorf("expected destination file to exist at %s, but it does not", expectedDestPath)
	}

	// Verify original file is gone
	if _, err := os.Stat(srcFile); !os.IsNotExist(err) {
		t.Errorf("expected source file %s to be deleted/moved, but it still exists", srcFile)
	}
}

func TestRenamer_DryRun(t *testing.T) {
	tempDir := t.TempDir()
	srcDir := filepath.Join(tempDir, "src")
	destDir := filepath.Join(tempDir, "dest")

	if err := os.Mkdir(srcDir, 0755); err != nil {
		t.Fatalf("failed to create src dir: %v", err)
	}

	srcFile := filepath.Join(srcDir, "photo.jpg")
	if err := os.WriteFile(srcFile, []byte("dummy jpeg content"), 0644); err != nil {
		t.Fatalf("failed to write dummy photo: %v", err)
	}

	opts := Options{
		Destination:        destDir,
		Recursive:          false,
		DryRun:             true, // Dry run enabled
		LowercaseExtension: true,
		Verbose:            true,
		FolderPattern:      "%Y/%Y-%m/%Y-%m-%d-%s",
		FilePattern:        "%p%Y%m%d-%H%M%S-%s%c%e",
		DefaultPrefix:      "IMG_",
		DefaultSuffix:      "test",
	}

	r := New(opts)
	err := r.Process([]string{srcDir})
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Dry run: source file MUST still exist
	if _, err := os.Stat(srcFile); os.IsNotExist(err) {
		t.Errorf("source file should still exist during dry-run")
	}

	// Dry run: destination file MUST NOT exist
	expectedDestPath := filepath.Join(destDir, "2025", "2025-06", "2025-06-07-test", "IMG_20250607-103000-test.jpg")
	if _, err := os.Stat(expectedDestPath); err == nil {
		t.Errorf("destination file should NOT be created during dry-run")
	}
}

func TestRenamer_CollisionCounter(t *testing.T) {
	tempDir := t.TempDir()
	srcDir := filepath.Join(tempDir, "src")
	destDir := filepath.Join(tempDir, "dest")

	if err := os.Mkdir(srcDir, 0755); err != nil {
		t.Fatalf("failed to create src dir: %v", err)
	}

	// Create two files that will have identical times and therefore names
	srcFile1 := filepath.Join(srcDir, "photo1.jpg")
	srcFile2 := filepath.Join(srcDir, "photo2.jpg")

	if err := os.WriteFile(srcFile1, []byte("content 1"), 0644); err != nil {
		t.Fatalf("failed to write photo1: %v", err)
	}
	if err := os.WriteFile(srcFile2, []byte("content 2"), 0644); err != nil {
		t.Fatalf("failed to write photo2: %v", err)
	}

	fixedTime := time.Date(2025, 6, 7, 10, 30, 0, 0, time.UTC)
	if err := os.Chtimes(srcFile1, fixedTime, fixedTime); err != nil {
		t.Fatalf("failed to set time 1: %v", err)
	}
	if err := os.Chtimes(srcFile2, fixedTime, fixedTime); err != nil {
		t.Fatalf("failed to set time 2: %v", err)
	}

	opts := Options{
		Destination:        destDir,
		Recursive:          false,
		DryRun:             false,
		LowercaseExtension: true,
		Verbose:            true,
		FolderPattern:      "%Y/%Y-%m/%Y-%m-%d-%s",
		FilePattern:        "%p%Y%m%d-%H%M%S-%s%c%e",
		DefaultPrefix:      "IMG_",
		DefaultSuffix:      "test",
	}

	r := New(opts)
	// Process sources - order inside directory scan should result in both being processed.
	// Since Process handles one-by-one, the second file processed should trigger the collision suffix "_1".
	err := r.Process([]string{srcDir})
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	expectedPath1 := filepath.Join(destDir, "2025", "2025-06", "2025-06-07-test", "IMG_20250607-103000-test.jpg")
	expectedPath2 := filepath.Join(destDir, "2025", "2025-06", "2025-06-07-test", "IMG_20250607-103000-test_1.jpg")

	if _, err := os.Stat(expectedPath1); os.IsNotExist(err) {
		t.Errorf("expected first file at %s, but not found", expectedPath1)
	}
	if _, err := os.Stat(expectedPath2); os.IsNotExist(err) {
		t.Errorf("expected second file (collided) at %s, but not found", expectedPath2)
	}
}

func TestRenamer_UnsupportedAndHidden(t *testing.T) {
	tempDir := t.TempDir()
	srcDir := filepath.Join(tempDir, "src")
	destDir := filepath.Join(tempDir, "dest")

	if err := os.Mkdir(srcDir, 0755); err != nil {
		t.Fatalf("failed to create src dir: %v", err)
	}

	// 1. System/hidden file
	hiddenFile := filepath.Join(srcDir, ".DS_Store")
	if err := os.WriteFile(hiddenFile, []byte("hidden stuff"), 0644); err != nil {
		t.Fatalf("failed to write DS_Store: %v", err)
	}

	// 2. Unsupported extension
	unsupportedFile := filepath.Join(srcDir, "document.txt")
	if err := os.WriteFile(unsupportedFile, []byte("unsupported text doc"), 0644); err != nil {
		t.Fatalf("failed to write txt file: %v", err)
	}

	// 3. Supported image file
	supportedFile := filepath.Join(srcDir, "photo.jpg")
	if err := os.WriteFile(supportedFile, []byte("dummy content"), 0644); err != nil {
		t.Fatalf("failed to write jpg file: %v", err)
	}

	opts := Options{
		Destination:        destDir,
		Recursive:          false,
		DryRun:             false,
		LowercaseExtension: true,
		FolderPattern:      "%Y/%Y-%m",
		FilePattern:        "%p%Y%M%D-%h%m%s%S%C%F", // legacy patterns supported
	}

	r := New(opts)
	err := r.Process([]string{srcDir})
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// DS_Store must not be processed/moved (skipped silently)
	if _, err := os.Stat(hiddenFile); os.IsNotExist(err) {
		t.Errorf("hidden system file should not have been moved or deleted")
	}

	// txt file must be skipped
	if _, err := os.Stat(unsupportedFile); os.IsNotExist(err) {
		t.Errorf("unsupported file should not have been moved or deleted")
	}

	// photo.jpg must be processed
	if _, err := os.Stat(supportedFile); !os.IsNotExist(err) {
		t.Errorf("supported image file should have been moved")
	}
}

func TestRenamer_Recursive(t *testing.T) {
	tempDir := t.TempDir()
	srcDir := filepath.Join(tempDir, "src")
	subSrcDir := filepath.Join(srcDir, "sub")
	destDir := filepath.Join(tempDir, "dest")

	if err := os.MkdirAll(subSrcDir, 0755); err != nil {
		t.Fatalf("failed to create nested src dirs: %v", err)
	}

	// File in sub-directory
	nestedFile := filepath.Join(subSrcDir, "photo.jpg")
	if err := os.WriteFile(nestedFile, []byte("nested dummy image"), 0644); err != nil {
		t.Fatalf("failed to write nested file: %v", err)
	}

	// 1. Recursive = false
	optsNonRecursive := Options{
		Destination:        destDir,
		Recursive:          false,
		DryRun:             false,
		LowercaseExtension: true,
		FolderPattern:      "%Y/%Y-%m",
		FilePattern:        "%p%Y%m%d-%H%M%S%s%c%e",
	}

	r1 := New(optsNonRecursive)
	if err := r1.Process([]string{srcDir}); err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Nested file must still exist because recursive was false
	if _, err := os.Stat(nestedFile); os.IsNotExist(err) {
		t.Errorf("nested file should NOT be moved when recursive is false")
	}

	// 2. Recursive = true
	optsRecursive := Options{
		Destination:        destDir,
		Recursive:          true,
		DryRun:             false,
		LowercaseExtension: true,
		FolderPattern:      "%Y/%Y-%m",
		FilePattern:        "%p%Y%m%d-%H%M%S%s%c%e",
	}

	r2 := New(optsRecursive)
	if err := r2.Process([]string{srcDir}); err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Nested file must be moved now
	if _, err := os.Stat(nestedFile); !os.IsNotExist(err) {
		t.Errorf("nested file should have been moved when recursive is true")
	}
}
