// go-plugin/transport_other.go
//go:build !windows

package plugin

import "net"

// secureListenWindows is only implemented on Windows
// this stub satisfies the compiler on other platforms
func secureListenWindows(name string) (net.Listener, error) {
	return nil, nil // never called on non-Windows
}
