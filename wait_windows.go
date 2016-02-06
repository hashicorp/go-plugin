package plugin

import (
	"os"
	"syscall"
	"time"
)

const STILL_ACTIVE = 259

// pidAlive tests whether a process is alive or not
func pidAlive(pid int) bool {
	const da = syscall.STANDARD_RIGHTS_READ | syscall.PROCESS_QUERY_INFORMATION | syscall.SYNCHRONIZE
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

// Wait waits until a process is no longer alive by polling it every 5 seconds
func Wait(pid int) error {
	_, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if !pidAlive(pid) {
			break
		}
	}
	return nil

}
