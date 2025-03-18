package shared

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"github.com/hashicorp/go-plugin/examples/streaming/proto"
	"google.golang.org/grpc"
)

type StreamerRPCServer struct {
	Impl Streamer
	proto.UnimplementedStreamerServiceServer
}

func (s *StreamerRPCServer) Read(req *proto.Read_Request, srv proto.StreamerService_ReadServer) error {
	readBytes, err := s.Impl.Read(srv.Context(), req.Path)

	// TODO: chunk readBytes

	resp := &proto.Read_Response{
		ReadBytes: readBytes,
	}
	if err != nil {
		resp.Error = err.Error()
	}

	return srv.Send(resp)
}

func (s *StreamerRPCServer) Write(srv proto.StreamerService_WriteServer) error {
	// TODO: receive all byte chunks

	req, err := srv.Recv()
	if err != nil {
		return err
	}

	err = s.Impl.Write(srv.Context(), req.Path, req.BytesToWrite)
	if err != nil {
		return err
	}

	return nil
}

var _ plugin.GRPCPlugin = &StreamerPlugin{}

type StreamerPlugin struct {
	plugin.NetRPCUnsupportedPlugin
	Impl Streamer
}

func (p *StreamerPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterStreamerServiceServer(s, &StreamerRPCServer{
		Impl: p.Impl,
	})
	return nil
}

func (p *StreamerPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &StreamerGRPC{
		client: proto.NewStreamerServiceClient(c),
	}, nil
}
