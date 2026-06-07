//go:build !darwin

package metadata

import (
	"syscall"
	"time"
)

func getCreationTime(stat *syscall.Stat_t) time.Time {
	return time.Time{}
}
