// transport_windows.go
//go:build windows

package plugin

import (
	"fmt"
	"net"
	"os"

	winio "github.com/Microsoft/go-winio"
)

func secureListenWindows(name string) (net.Listener, error) {
	pipe := fmt.Sprintf(`\\.\pipe\%s`, name)
	return winio.ListenPipe(pipe, &winio.PipeConfig{
		SecurityDescriptor: "D:P(A;;GA;;;OW)", // owner full access only
		MessageMode:        false,
		InputBufferSize:    65536,
		OutputBufferSize:   65536,
	})
}

// pipeName returns the named pipe path for the client to connect to
func pipeName() string {
	return fmt.Sprintf(`\\.\pipe\go-plugin-%d`, os.Getpid())
}
