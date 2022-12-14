package my_test_grpc_plugin

import (
	"context"

	"github.com/hashicorp/go-plugin"

	"google.golang.org/grpc"
)

var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "BASIC_PLUGIN",
	MagicCookieValue: "hello",
}

type GRPC struct {
	plugin.Plugin
	Impl MyTestGrpcPluginServer
}

func (p *GRPC) GRPCServer(broker *plugin.GRPCBroker, server *grpc.Server) error {
	RegisterMyTestGrpcPluginServer(server, &GPRCPluginServerWrapper{impl: p.Impl})
	return nil
}

func (p *GRPC) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, conn *grpc.ClientConn) (interface{}, error) {
	return &GRPCPluginClientWrapper{client: NewMyTestGrpcPluginClient(conn)}, nil
}

type GPRCPluginServerWrapper struct {
	impl MyTestGrpcPluginServer
	UnimplementedMyTestGrpcPluginServer
}

func (g *GPRCPluginServerWrapper) Run(ctx context.Context, req *Request) (*Response, error) {
	return g.impl.Run(ctx, req)
}

type GRPCPluginClientWrapper struct {
	client MyTestGrpcPluginClient
}

func (g *GRPCPluginClientWrapper) Run(ctx context.Context, req *Request) (*Response, error) {
	return g.client.Run(ctx, req)
}
