package metadata

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/abema/go-mp4"
	"github.com/evanoberholster/imagemeta"
)

// GetMediaCreationTime extracts the media creation time from image EXIF headers
// or video mvhd atoms. It falls back to file creation or modification time if metadata is missing.
// It returns the parsed time, a string indicating the source of the time, and any error encountered.
func GetMediaCreationTime(filePath string) (time.Time, string, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	// If it looks like a video, try parsing with go-mp4 first
	if ext == ".mp4" || ext == ".mov" || ext == ".m4v" {
		t, err := parseVideoCreationTime(filePath)
		if err == nil && !t.IsZero() {
			return t, "mvhd", nil
		}
	}

	// Try image EXIF metadata
	t, err := parseImageCreationTime(filePath)
	if err == nil && !t.IsZero() {
		return t, "exif", nil
	}

	// Fallback to file system times
	fi, err := os.Stat(filePath)
	if err != nil {
		return time.Time{}, "", err
	}

	// Try creation time (birth time)
	cTime := getFileCreationTime(fi)
	if !cTime.IsZero() {
		return cTime, "creation", nil
	}

	// Try modification time (mtime)
	mTime := fi.ModTime()
	if !mTime.IsZero() {
		return mTime, "modification", nil
	}

	return time.Time{}, "", errors.New("unable to determine creation or modification time")
}

func parseVideoCreationTime(filePath string) (time.Time, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return time.Time{}, err
	}
	defer file.Close()

	var creationTime time.Time
	errStop := errors.New("stop")

	_, err = mp4.ReadBoxStructure(file, func(h *mp4.ReadHandle) (interface{}, error) {
		if h.BoxInfo.Type == mp4.BoxTypeMvhd() {
			box, _, err := h.ReadPayload()
			if err != nil {
				return nil, err
			}
			if mvhd, ok := box.(*mp4.Mvhd); ok {
				rawTime := mvhd.GetCreationTime()
				if rawTime > 0 {
					// MP4 epoch is 1904-01-01. Offset to Unix epoch 1970-01-01 is 2,082,844,800 seconds.
					unixTime := int64(rawTime) - 2082844800
					creationTime = time.Unix(unixTime, 0)
					return nil, errStop
				}
			}
		}
		return h.Expand()
	})

	if err != nil && !errors.Is(err, errStop) {
		return time.Time{}, err
	}
	return creationTime, nil
}

func parseImageCreationTime(filePath string) (time.Time, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return time.Time{}, err
	}
	defer file.Close()

	ex, err := imagemeta.Decode(file)
	if err != nil {
		return time.Time{}, err
	}

	date := ex.OriginalDate()
	if date.IsZero() {
		return time.Time{}, errors.New("original date is zero")
	}

	return date, nil
}
