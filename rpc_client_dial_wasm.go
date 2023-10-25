//go:build wasm && js

package plugin

import (
	"net"
)

func dialRPC(c *Client) (net.Conn, error) {
	return NewWebWorkerConnForClient(c.address.(WebWorkerAddr))
}
