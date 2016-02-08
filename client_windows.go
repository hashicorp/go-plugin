package plugin

import (
	"net"
	"time"

	"github.com/Microsoft/go-winio"
)

const defaultConnTimeout = 32 * time.Second

func dialAddress(addr net.Addr) (net.Conn, error) {
	if addr.Network() == "pipe" {
		return winio.DialPipe(addr.String(), defaultConnTimeout)
	}
	return net.Dial(addr.Network(), addr.String())
}
