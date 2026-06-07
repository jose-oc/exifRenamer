//go:build !darwin && !windows

package metadata

import (
	"os"
	"time"
)

func getFileCreationTime(fi os.FileInfo) time.Time {
	return time.Time{}
}
