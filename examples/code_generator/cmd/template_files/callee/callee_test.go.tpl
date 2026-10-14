package main

import (
	"context"
	"os/exec"
	"testing"

	"github.com/hashicorp/go-plugin"

	pb "{{.FullPackagePath}}"
)

func Test_Callee(t *testing.T) {
	path := "./{{.Package}}_callee"  // todo: `go build` first
	pluginClientConfig := &plugin.ClientConfig{
		HandshakeConfig:  pb.Handshake,
		Cmd:              exec.Command(path),
		Plugins:          map[string]plugin.Plugin{

{{- range $item := .Services }}
			"{{.PluginName}}": &pb.GRPC{{.Name}}{},
{{- end }}
		},
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
	}
	client := plugin.NewClient(pluginClientConfig)
	pluginClientConfig.Reattach = client.ReattachConfig()
	protocol, err := client.Client()
	if err != nil {
		t.Errorf("new client error, err=%+v", err)
		return
	}
{{- range $item := .Services }}
	{
        raw, err := protocol.Dispense("{{.PluginName}}")
        if err != nil {
            t.Errorf("PluginName {{.PluginName}} error, err=%+v", err)
            return
        }
        inst, ok := raw.(pb.{{.Name}}Server)
        if !ok {
            t.Errorf("interface type error")
            return
        }
        // test each method
{{$CurService := .Name}}
{{- range $item := .Methods }}
        {
            rsp, err := inst.{{.Name}}(context.Background(), &pb.{{.InputType}}{/*todo: add param here*/})
            if err != nil {
                t.Errorf("run error, err=%+v", err)
                return
            }
            t.Logf("rsp=%+v", rsp)
        }
{{- end }}
	}
{{- end }}
}
