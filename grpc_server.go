package plugin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"

	"google.golang.org/grpc"
)

// GRPCServer is a ServerType implementation that serves plugins over
// gRPC. This allows plugins to easily be written for other languages.
//
// The GRPCServer outputs a custom configuration as a base64-encoded
// JSON structure represented by the GRPCServerConfig config structure.
type GRPCServer struct {
	// Plugins are the list of plugins to serve.
	Plugins map[string]Plugin

	// Server is the actual server that will accept connections. This
	// will be used for plugin registration as well.
	Server *grpc.Server

	// DoneCh is the channel that is closed when this server has exited.
	DoneCh chan struct{}

	// Stdout/StderrLis are the readers for stdout/stderr that will be copied
	// to the stdout/stderr connection that is output.
	Stdout io.Reader
	Stderr io.Reader

	config GRPCServerConfig
}

// ServerProtocol impl.
func (s *GRPCServer) Init() error {
	// Register all our plugins onto the gRPC server.
	for k, raw := range s.Plugins {
		p, ok := raw.(GRPCPlugin)
		if !ok {
			return fmt.Errorf("%q is not a GRPC-compatibile plugin", k)
		}

		if err := p.GRPCServer(s.Server); err != nil {
			return fmt.Errorf("error registring %q: %s", k, err)
		}
	}

	return nil
}

// Config is the GRPCServerConfig encoded as JSON then base64.
func (s *GRPCServer) Config() string {
	// Create a buffer that will contain our final contents
	var buf bytes.Buffer

	// Wrap the base64 encoding with JSON encoding.
	if err := json.NewEncoder(&buf).Encode(s.config); err != nil {
		// We panic since ths shouldn't happen under any scenario. We
		// carefully control the structure being encoded here and it should
		// always be successful.
		panic(err)
	}

	return buf.String()
}

func (s *GRPCServer) Serve(lis net.Listener) {
	// Start serving in a goroutine
	go s.Server.Serve(lis)

	// Wait until graceful completion
	<-s.DoneCh
}

// GRPCServerConfig is the extra configuration passed along for consumers
// to facilitate using GRPC plugins.
type GRPCServerConfig struct {
	StdoutAddr string `json:"stdout_addr"`
	StderrAddr string `json:"stderr_addr"`
}
