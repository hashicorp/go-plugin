// +build !windows

package plugin

import "net"

func serverListener() (net.Listener, error) {
	path, err := newTempPath("")
	if err != nil {
		return nil, err
	}

	return net.Listen("unix", path)
}
