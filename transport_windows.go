// go-plugin/transport_windows.go
//go:build windows

package plugin

import (
	"fmt"
	"net"

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
