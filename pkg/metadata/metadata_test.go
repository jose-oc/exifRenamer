package metadata

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGetMediaCreationTime_Fallback(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "no_exif_photo.jpg")

	// Create a dummy file with no EXIF headers
	err := os.WriteFile(filePath, []byte("dummy text file representing an image with no exif"), 0644)
	if err != nil {
		t.Fatalf("failed to write dummy file: %v", err)
	}

	// Change file modification and access times
	targetTime := time.Date(2021, 12, 25, 12, 0, 0, 0, time.UTC)
	err = os.Chtimes(filePath, targetTime, targetTime)
	if err != nil {
		t.Fatalf("failed to set file times: %v", err)
	}

	creationTime, source, err := GetMediaCreationTime(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the fallback source is correct
	if source != "creation" && source != "modification" {
		t.Errorf("expected source to be 'creation' or 'modification', got: %s", source)
	}

	// Verify we got a non-zero time
	if creationTime.IsZero() {
		t.Error("expected non-zero creation time")
	}

	// If the source was modification time, verify the time matches targetTime
	if source == "modification" {
		if !creationTime.Equal(targetTime) {
			t.Errorf("expected time %v, got %v", targetTime, creationTime)
		}
	}
}

func TestGetMediaCreationTime_NonExistentFile(t *testing.T) {
	_, _, err := GetMediaCreationTime("non_existent_file.jpg")
	if err == nil {
		t.Error("expected error for non-existent file, got nil")
	}
}
