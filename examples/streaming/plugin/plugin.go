// Copyright IBM Corp. 2016, 2025
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"errors"
	"io"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/hashicorp/go-plugin/examples/streaming/shared"
)

type FileStreamer struct {
	logger hclog.Logger
	path   string
}

func (fs *FileStreamer) Configure(ctx context.Context, path string, _ int64) error {
	fs.path = path
	return nil
}

func (fs *FileStreamer) Read(ctx context.Context) ([]byte, error) {
	fs.logger.Debug("FileStreamer: Read", "path", fs.path)
	f, err := os.OpenFile(fs.path, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer func() {
		cErr := f.Close()
		err = errors.Join(err, cErr)
	}()
	return io.ReadAll(f)
}

func (fs *FileStreamer) Write(ctx context.Context, b []byte) error {
	fs.logger.Debug("FileStreamer: Write", "path", fs.path)
	f, err := os.OpenFile(fs.path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer func() {
		cErr := f.Close()
		err = errors.Join(err, cErr)
	}()

	n, err := f.Write(b)
	if err != nil {
		return err
	}
	fs.logger.Debug("FileStreamer: Write finished", "bytes written", n)
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

	streamer := &FileStreamer{
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
