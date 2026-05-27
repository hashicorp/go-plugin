// transport.go
package plugin

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"sync"
	"syscall"
)

var umaskMu sync.Mutex

func secureListen(path string) (net.Listener, error) {
	switch runtime.GOOS {
	case "windows":
		name := fmt.Sprintf("go-plugin-%d", os.Getpid())
		return secureListenWindows(name)
	default:
		return secureListenUnix(path)
	}
}

func secureListenUnix(path string) (net.Listener, error) {
	umaskMu.Lock()
	oldUmask := syscall.Umask(0177)
	l, err := net.Listen("unix", path)
	syscall.Umask(oldUmask)
	umaskMu.Unlock()
	return l, err
}
