//go:generate protoc -I ./ ./grpc_controller.proto --go_out=plugins=grpc:.
package plugin

import (
	"context"
)

// GRPCControllerServer handles shutdown calls to terminate the server when the
// plugin client is closed.
type grpcControllerServer struct {
	server *GRPCServer
}

// Shutdown stops the grpc server. It first will attempt a graceful stop, then a
// full stop on the server.
func (s *grpcControllerServer) Shutdown(ctx context.Context, _ *Empty) (*Empty, error) {
	resp := &Empty{}

	// TODO: figure out why GracefullStop doesn't work.
	s.server.Stop()
	return resp, nil
}
