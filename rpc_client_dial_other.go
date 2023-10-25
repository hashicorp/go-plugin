//go:build !wasm

package plugin

import "net"

func dialRPC(c *Client) (net.Conn, error) {
	conn, err := net.Dial(c.address.Network(), c.address.String())
	if err != nil {
		return nil, err
	}
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		// Make sure to set keep alive so that the connection doesn't die
		tcpConn.SetKeepAlive(true)
	}
	return conn, nil
}
