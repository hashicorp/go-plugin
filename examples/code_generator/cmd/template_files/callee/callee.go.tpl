package main

import (
	"runtime"

	"github.com/hashicorp/go-plugin"

	pb "{{.FullPackagePath}}"

	imp "{{.FullPackagePath}}_callee/plugin"
)

func StartPluginCallee() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: pb.Handshake,
		Plugins: map[string]plugin.Plugin{
{{- range $item := .Services }}
			"{{.PluginName}}": &pb.GRPC{{.Name}}{Impl: &imp.{{.Name}}Service{}},
{{- end }}
		},
		// A non-nil value here enables gRPC serving for this plugin...
		GRPCServer: plugin.DefaultGRPCServer,
	})
}

func main() {
	runtime.GOMAXPROCS(1)
	StartPluginCallee()
}
