// transport_other.go
//go:build !windows

package plugin

import "net"

func secureListenWindows(name string) (net.Listener, error) {
	return nil, nil // never called on non-Windows
}

func pipeName() string {
	return "" // never called on non-Windows
}
