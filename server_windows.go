package plugin

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/Microsoft/go-winio"
)

func serverListener() (net.Listener, error) {
	minPort := os.Getenv("PLUGIN_MIN_PORT")
	maxPort := os.Getenv("PLUGIN_MAX_PORT")

	if minPort != "" && maxPort != "" {
		return serverListenerTCP(minPort, maxPort)
	}

	return serverListenerNPipe()
}

func serverListenerTCP(minPortString, maxPortString string) (net.Listener, error) {
	minPort, err := strconv.ParseInt(minPortString, 10, 32)
	if err != nil {
		return nil, err
	}

	maxPort, err := strconv.ParseInt(maxPortString, 10, 32)
	if err != nil {
		return nil, err
	}

	for port := minPort; port <= maxPort; port++ {
		address := fmt.Sprintf("127.0.0.1:%d", port)
		listener, err := net.Listen("tcp", address)
		if err == nil {
			return listener, nil
		}
	}

	return nil, errors.New("Couldn't bind plugin TCP listener")
}

func serverListenerNPipe() (net.Listener, error) {
	addr, err := newTempPath("//./pipe/")
	if err != nil {
		return nil, err
	}

	return winio.ListenPipe(addr, "")
}
