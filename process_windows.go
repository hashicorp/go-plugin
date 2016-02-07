package plugin

import (
	"os"
	"syscall"
)

const STILL_ACTIVE = 259

// _pidAlive tests whether a process is alive or not
func _pidAlive(pid int) bool {
	const da = syscall.STANDARD_RIGHTS_READ |
		syscall.PROCESS_QUERY_INFORMATION |
		syscall.SYNCHRONIZE
	h, e := syscall.OpenProcess(da, false, uint32(pid))
	if e != nil {
		return false
	}

	var ec uint32
	e = syscall.GetExitCodeProcess(h, &ec)
	if e != nil {
		return false
	}

	return ec == STILL_ACTIVE
}
