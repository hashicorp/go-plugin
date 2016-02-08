// +build !windows

package plugin

import "net"

func dialAddress(addr net.Addr) (net.Conn, error) {
	return net.Dial(addr.Network(), addr.String())
}
