// go-plugin/transport.go
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

// secureListen creates a platform-appropriate secure listener.
// Windows: named pipe restricted to owner via ACL.
// Unix: unix socket created atomically with 0600 permissions.
func secureListen(name string) (net.Listener, error) {
	switch runtime.GOOS {
	case "windows":
		return secureListenWindows(name)
	default:
		return secureListenUnix(name)
	}
}

func secureListenUnix(name string) (net.Listener, error) {
	path := fmt.Sprintf("/tmp/%s.sock", name)
	os.Remove(path)

	umaskMu.Lock()
	oldUmask := syscall.Umask(0177)
	l, err := net.Listen("unix", path)
	syscall.Umask(oldUmask)
	umaskMu.Unlock()

	return l, err
}
