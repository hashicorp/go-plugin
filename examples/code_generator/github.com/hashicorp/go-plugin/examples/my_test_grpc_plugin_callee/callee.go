package main

import (
	"runtime"

	"github.com/hashicorp/go-plugin"

	pb "github.com/hashicorp/go-plugin/examples/my_test_grpc_plugin"

	imp "github.com/hashicorp/go-plugin/examples/my_test_grpc_plugin_callee/plugin"
)

func StartPluginCaller() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: pb.Handshake,
		Plugins: map[string]plugin.Plugin{
			"my_plugin_1": &pb.GRPC{Impl: &imp.PluginService{}},
		},
		// A non-nil value here enables gRPC serving for this plugin...
		GRPCServer: plugin.DefaultGRPCServer,
	})
}

func main() {
	runtime.GOMAXPROCS(1)
	StartPluginCaller()
}
