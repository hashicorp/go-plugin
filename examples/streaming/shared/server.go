// Copyright IBM Corp. 2016, 2025
// SPDX-License-Identifier: MPL-2.0

package shared

import (
	"bytes"
	"context"
	"io"

	"github.com/hashicorp/go-plugin"
	"github.com/hashicorp/go-plugin/examples/streaming/proto"
	"google.golang.org/grpc"
)

type StreamerGRPCServer struct {
	Impl Streamer

	chunkSize int
	proto.UnimplementedStreamerServiceServer
}

func (s *StreamerGRPCServer) Configure(ctx context.Context, req *proto.Configure_Request) (*proto.Configure_Response, error) {
	s.chunkSize = int(req.ChunkSize)
	return &proto.Configure_Response{}, s.Impl.Configure(ctx, req.Path, req.ChunkSize)
}

func (s *StreamerGRPCServer) Read(req *proto.Read_Request, srv proto.StreamerService_ReadServer) error {
	b, err := s.Impl.Read(srv.Context())
	if err != nil {
		return err
	}

	// send it by chunks
	buf := bytes.NewBuffer(b)
	for nextChunk := buf.Next(s.chunkSize); len(nextChunk) > 0; nextChunk = buf.Next(s.chunkSize) {
		err = srv.Send(&proto.Read_ResponseChunk{
			ReadBytes: nextChunk,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *StreamerGRPCServer) Write(srv proto.StreamerService_WriteServer) error {
	var buf bytes.Buffer
	// receive all byte chunks
	for {
		writeReq, err := srv.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		_, err = buf.Write(writeReq.BytesToWrite)
		if err != nil {
			return err
		}
	}

	err := s.Impl.Write(srv.Context(), buf.Bytes())
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
	proto.RegisterStreamerServiceServer(s, &StreamerGRPCServer{
		Impl: p.Impl,
	})
	return nil
}

func (p *StreamerPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &StreamerGRPCClient{
		client: proto.NewStreamerServiceClient(c),
	}, nil
}
