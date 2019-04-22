package plugin

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin/internal/plugin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// GRPCServiceName is the name of the service that the health check should
// return as passing.
const GRPCServiceName = "plugin"

// DefaultGRPCServer can be used with the "GRPCServer" field for Server
// as a default factory method to create a gRPC server with no extra options.
func DefaultGRPCServer(opts []grpc.ServerOption) *grpc.Server {
	return grpc.NewServer(opts...)
}

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
	Server func([]grpc.ServerOption) *grpc.Server

	// TLS should be the TLS configuration if available. If this is nil,
	// the connection will not have transport security.
	TLS *tls.Config

	// DoneCh is the channel that is closed when this server has exited.
	DoneCh chan struct{}

	// Stdout/StderrLis are the readers for stdout/stderr that will be copied
	// to the stdout/stderr connection that is output.
	Stdout io.Reader
	Stderr io.Reader

	stdoutChannel chan []byte
	stderrChannel chan []byte

	config GRPCServerConfig
	server *grpc.Server
	broker *GRPCBroker

	logger hclog.Logger
}

// ServerProtocol impl.
func (s *GRPCServer) Init() error {
	// Create our server
	var opts []grpc.ServerOption
	if s.TLS != nil {
		opts = append(opts, grpc.Creds(credentials.NewTLS(s.TLS)))
	}
	s.server = s.Server(opts)

	// Register the health service
	healthCheck := health.NewServer()
	healthCheck.SetServingStatus(
		GRPCServiceName, grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(s.server, healthCheck)

	// Register the broker service
	brokerServer := newGRPCBrokerServer()
	plugin.RegisterGRPCBrokerServer(s.server, brokerServer)
	s.broker = newGRPCBroker(brokerServer, s.TLS)
	go s.broker.Run()

	// Register the controller
	controllerServer := &grpcControllerServer{
		server: s,
	}
	plugin.RegisterGRPCControllerServer(s.server, controllerServer)

	// Wire up the stdio service
	plugin.RegisterGRPCStdioServer(s.server, &stdioServer{server: s})
	// Avoid getting into a state where we start dropping writes because we're out of "space"
	s.stderrChannel = make(chan []byte, 1000)
	s.stdoutChannel = make(chan []byte, 1000)
	go copyChan(s.Stderr, s.stderrChannel)
	go copyChan(s.Stdout, s.stdoutChannel)

	// Register all our plugins onto the gRPC server.
	for k, raw := range s.Plugins {
		p, ok := raw.(GRPCPlugin)
		if !ok {
			return fmt.Errorf("%q is not a GRPC-compatible plugin", k)
		}

		if err := p.GRPCServer(s.broker, s.server); err != nil {
			return fmt.Errorf("error registering %q: %s", k, err)
		}
	}

	return nil
}

// Stop calls Stop on the underlying grpc.Server
func (s *GRPCServer) Stop() {
	s.server.Stop()
}

// GracefulStop calls GracefulStop on the underlying grpc.Server
func (s *GRPCServer) GracefulStop() {
	s.server.GracefulStop()
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
	defer close(s.DoneCh)
	err := s.server.Serve(lis)
	if err != nil {
		s.logger.Error("grpc server", "error", err)
	}
}

// GRPCServerConfig is the extra configuration passed along for consumers
// to facilitate using GRPC plugins.
type GRPCServerConfig struct {
	StdoutAddr string `json:"stdout_addr"`
	StderrAddr string `json:"stderr_addr"`
}

// Something to consider: How do we deal with I/O that potentially never completes because someone does not write
// a new line?
func splitLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, data[0:i + 1], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}

	// Request more data.
	return 0, nil, nil
}

func copyChan(src io.Reader, dst chan []byte) {
	scanner := bufio.NewScanner(src)

	// Set the split function for the scanning operation.
	scanner.Split(splitLines)
	for scanner.Scan() {
		bytes := scanner.Bytes()
		dst <- bytes
	}
	err := scanner.Err()
	if err != nil {
		panic(err)
	}
}

type stdioServer struct {
	server *GRPCServer
}

func (s *stdioServer) ReadStdio(req *plugin.StdioRequest, server plugin.GRPCStdio_ReadStdioServer) error {
	var reader chan []byte
	switch req.Channel {
	case plugin.StdioRequest_stdout:
		reader = s.server.stdoutChannel
	case plugin.StdioRequest_stderr:
		reader = s.server.stderrChannel
	}
	for {
		select {
		case data := <-reader:
			err := server.Send(&plugin.StdioResponse{Data: data})
			if err != nil {
				return err
			}
		case <-server.Context().Done():
			return nil
		}
	}
}
