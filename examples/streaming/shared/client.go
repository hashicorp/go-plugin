package shared

import (
	"context"
	"fmt"
	"io"

	"github.com/hashicorp/go-plugin/examples/streaming/proto"
)

type StreamerGRPC struct {
	client proto.StreamerServiceClient
}

func (g *StreamerGRPC) Read(ctx context.Context, path string) ([]byte, error) {
	readClient, err := g.client.Read(ctx, &proto.Read_Request{
		Path: path,
	})
	if err != nil {
		panic(err)
	}

	// TODO: receive all byte chunks

	resp, err := readClient.Recv()
	if err == io.EOF {
		return resp.ReadBytes, fmt.Errorf("stream closed EOF")
	} else if err != nil {
		return resp.ReadBytes, err
	}

	if resp.Error != "" {
		return resp.ReadBytes, fmt.Errorf("plugin error: %q", resp.Error)
	}

	return resp.ReadBytes, nil
}

func (g *StreamerGRPC) Write(ctx context.Context, path string, bytesToWrite []byte) error {
	writeClient, err := g.client.Write(ctx)
	if err != nil {
		panic(err)
	}

	// TODO: chunk bytesToWrite

	err = writeClient.Send(&proto.Write_Request{
		Path:         path,
		BytesToWrite: bytesToWrite,
	})
	if err != nil {
		return err
	}

	return nil
}
