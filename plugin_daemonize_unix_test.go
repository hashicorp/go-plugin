// +build !windows

package plugin

import (
	"syscall"
)

func daemonize() {
	syscall.Umask(0)
}
