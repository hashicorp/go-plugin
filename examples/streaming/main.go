// Copyright IBM Corp. 2016, 2025
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/hashicorp/go-plugin/examples/streaming/shared"
	"google.golang.org/grpc"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("expected path to file as an argument")
	}
	path := os.Args[1]

	logger := hclog.New(&hclog.LoggerOptions{
		Level:      hclog.Trace,
		Output:     os.Stderr,
		JSONFormat: true,
	})

	msgSizeLimit := 1000
	chunkSize := 10

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: plugin.HandshakeConfig{
			ProtocolVersion:  1,
			MagicCookieKey:   "BASIC_PLUGIN",
			MagicCookieValue: "hello",
		},
		Plugins: map[string]plugin.Plugin{
			"streamer": &shared.StreamerPlugin{},
		},
		Cmd: exec.Command("./plugin/streamer"),
		AllowedProtocols: []plugin.Protocol{
			plugin.ProtocolGRPC,
		},
		Logger: logger,
		GRPCDialOptions: []grpc.DialOption{
			grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(msgSizeLimit)),
			grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(msgSizeLimit)),
		},
	})
	defer client.Kill()

	logger.Debug("launching a client")

	rpcClient, err := client.Client()
	if err != nil {
		log.Fatal(err)
	}

	raw, err := rpcClient.Dispense("streamer")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	streamer := raw.(shared.Streamer)
	err = streamer.Configure(ctx, path, int64(chunkSize))
	if err != nil {
		log.Fatal(err)
	}

	err = streamer.Write(ctx, []byte("Lorem ipsum dolor sit amet"))
	if err != nil {
		log.Fatal(err)
	}

	logger.Debug("writing finished")

	b, err := streamer.Read(ctx)
	if err != nil {
		log.Fatal(err)
	}
	logger.Debug(fmt.Sprintf("received %d bytes", len(b)), "bytes", string(b))
}
