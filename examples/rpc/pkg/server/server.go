package server

import (
	"gitlab.com/indis/libs/extension/api"
	plugin "gitlab.com/indis/libs/third_party/go-plugin"
)

//var _ api.Plugable = (*PluginRPCServer)(nil)

// func Serve(impl api.Extender) {

// }
func NewRPCServer(impl *Extension) (*RPC, error) {
	return &RPC{Impl: impl}, nil
}

type RPC struct { //server.RPC
	Impl *Extension
}

func (s *RPC) Server(b *plugin.MuxBroker) (interface{}, error) {
	return &ExtServer{s.Impl, b}, nil
}
func (c *RPC) Client(b *plugin.MuxBroker, client *plugin.RPCConnection) (interface{}, error) {
	return &HostClient{b, client}, nil
	return nil, nil
}

func (s *RPC) Serve() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: api.Handshake,
		Plugins:         map[string]plugin.Plugin{"plugin": s},
	})
}
