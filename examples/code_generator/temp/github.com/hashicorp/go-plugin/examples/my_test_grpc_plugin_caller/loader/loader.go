package loader

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/hashicorp/go-plugin"

	pb "github.com/hashicorp/go-plugin/examples/my_test_grpc_plugin"
)

type Process struct {
	*ProcessBase
}

func (p *Process) Run(ctx context.Context, req *pb.Request) (*pb.Response, error) {
	p.LastRun = time.Now().Unix()
	service := p.Entry.(pb.MyTestGrpcPluginServer)
	return service.Run(ctx, req)
}

type DynamicPlugins struct {
	*DynamicPluginsBase
}

func NewLoader() *DynamicPlugins {
	out := &DynamicPlugins{
		DynamicPluginsBase: NewDynamicPlugins(),
	}
	return out
}

func (d *DynamicPlugins) Load(name string, path string) (*Process, error) {
	if p, ok := d.Plugins.Load(name); ok {
		return p.(*Process), nil
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("[%s]plugin file %s not exists", SourceCodeLoc(1), path)
	}
	d.Lock.Lock()
	defer d.Lock.Unlock()
	if p, ok := d.Plugins.Load(name); ok {
		return p.(*Process), nil
	}
	pluginClientConfig := &plugin.ClientConfig{
		HandshakeConfig:  pb.Handshake,
		Cmd:              exec.Command(path),
		Plugins:          map[string]plugin.Plugin{"main": &pb.GRPC{}},
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
	}
	client := plugin.NewClient(pluginClientConfig)
	pluginClientConfig.Reattach = client.ReattachConfig()
	protocol, err := client.Client()
	if err != nil {
		return nil, WrapError(err, "start process error")
	}
	raw, err := protocol.Dispense("main")
	if err != nil {
		return nil, WrapError(err, "find plugin error")
	}
	p := &Process{
		ProcessBase: &ProcessBase{
			Client:   client,
			Protocol: protocol,
			Entry:    raw,
			LastRun:  0,
		},
	}
	d.Plugins.Store(name, p)
	return p, nil
}
