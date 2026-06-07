//go:build darwin

package metadata

import (
	"syscall"
	"time"
)

func getCreationTime(stat *syscall.Stat_t) time.Time {
	sec := stat.Birthtimespec.Sec
	nsec := stat.Birthtimespec.Nsec
	if sec > 0 {
		return time.Unix(sec, nsec)
	}
	return time.Time{}
}
