package main

import (
	"context"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/hashicorp/go-plugin/examples/streaming/shared"
)

type StreamerExample struct {
	logger hclog.Logger
}

func (g *StreamerExample) Read(ctx context.Context, path string) ([]byte, error) {
	g.logger.Debug("message from StreamerExample.Read")
	b, err := os.ReadFile(path)
	return b, err
}

func (g *StreamerExample) Write(ctx context.Context, path string, data []byte) error {
	g.logger.Debug("message from StreamerExample.Write")
	os.WriteFile(path, []byte(data), os.FileMode(os.O_RDWR))
	return nil
}

var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "BASIC_PLUGIN",
	MagicCookieValue: "hello",
}

func main() {
	logger := hclog.New(&hclog.LoggerOptions{
		Level:      hclog.Trace,
		Output:     os.Stderr,
		JSONFormat: true,
	})

	streamer := &StreamerExample{
		logger: logger,
	}
	var pluginMap = map[string]plugin.Plugin{
		"streamer": &shared.StreamerPlugin{
			Impl: streamer,
		},
	}

	logger.Debug("plugin launched, about to be served")

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
		GRPCServer:      plugin.DefaultGRPCServer,
		Logger:          logger,
	})
}
