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

	// We should have a Streamer now! This feels like a normal interface
	// implementation but is in fact over an RPC connection.
	streamer := raw.(shared.Streamer)
	data, err := streamer.Read(context.Background(), path)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("received %d bytes\n", len(data))
}
