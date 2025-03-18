// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package shared

import (
	"bytes"
	"context"
	"io"

	"github.com/hashicorp/go-plugin/examples/streaming/proto"
)

type StreamerGRPCClient struct {
	client        proto.StreamerServiceClient
	chunkByteSize int
}

var _ Streamer = &StreamerGRPCClient{}

func (g *StreamerGRPCClient) Configure(ctx context.Context, path string, chunkSize int64) error {
	g.chunkByteSize = int(chunkSize)
	_, err := g.client.Configure(ctx, &proto.Configure_Request{
		Path:      path,
		ChunkSize: chunkSize,
	})
	return err
}

func (g *StreamerGRPCClient) Read(ctx context.Context) ([]byte, error) {
	readClient, err := g.client.Read(ctx, &proto.Read_Request{})
	if err != nil {
		return nil, err
	}

	// receive all byte chunks
	var buf bytes.Buffer
	for {
		resp, err := readClient.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			return buf.Bytes(), err
		}

		_, err = buf.Write(resp.ReadBytes)
		if err != nil {
			return buf.Bytes(), err
		}
	}

	return buf.Bytes(), nil
}

func (g *StreamerGRPCClient) Write(ctx context.Context, b []byte) error {
	writeClient, err := g.client.Write(ctx)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(b)
	for chunkBytes := buf.Next(g.chunkByteSize); len(chunkBytes) > 0; chunkBytes = buf.Next(g.chunkByteSize) {
		err = writeClient.Send(&proto.Write_RequestChunk{
			BytesToWrite: chunkBytes,
		})
		if err != nil {
			return err
		}
	}

	_, err = writeClient.CloseAndRecv()
	if err != nil && err != io.EOF {
		return err
	}

	return nil
}
