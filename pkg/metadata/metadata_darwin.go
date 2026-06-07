//go:build darwin

package metadata

import (
	"os"
	"syscall"
	"time"
)

func getFileCreationTime(fi os.FileInfo) time.Time {
	if stat, ok := fi.Sys().(*syscall.Stat_t); ok {
		sec := stat.Birthtimespec.Sec
		nsec := stat.Birthtimespec.Nsec
		if sec > 0 {
			return time.Unix(sec, nsec)
		}
	}
	return time.Time{}
}
