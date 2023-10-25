//go:build js && wasm

package plugin

import (
	"net"
	"time"
)

func netAddrDialer(addr net.Addr) func(string, time.Duration) (net.Conn, error) {
	return func(_ string, _ time.Duration) (net.Conn, error) {
		// NOTE: This only works when the addr is an address in the web worker.
		conn, err := NewWebWorkerConnForClient(addr.(WebWorkerAddr))
		if err != nil {
			return nil, err
		}
		return conn, nil
	}
}
