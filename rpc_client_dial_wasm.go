//go:build wasm && js

package plugin

import (
	"net"

	"github.com/hashicorp/go-plugin/internal/wasmrunner"
)

func dialRPC(c *Client) (net.Conn, error) {
	ww := c.runner.(*wasmrunner.WasmRunner).WebWorker()
	conn := NewWebWorkerConnForClient(ww.Name, ww.EventChannel(), ww.PostMessage)
	return conn, nil
}
