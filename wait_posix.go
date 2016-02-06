package plugin

import (
	"os"
	"syscall"
	"time"
)

// pidAlive tests whether a process is alive or not by sending it Signal 0
func pidAlive(proc *os.Process) bool {
	err := proc.Signal(syscall.Signal(0))
	return err == nil
}

// Wait waits until a process is no longer alive by polling it every 5 seconds
func Wait(pid int) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if !pidAlive(proc) {
			break
		}
	}
	return nil
}
