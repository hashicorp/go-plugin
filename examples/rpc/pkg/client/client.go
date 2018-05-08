package client

import (
	hclog "github.com/hashicorp/go-hclog"
	plugin "github.com/sampaioletti/go-plugin/examples/go-plugin"
)

var _ plugin.Plugin = (*RPC)(nil)

type RPC struct {
	Impl   *Host
	logger hclog.Logger
}

func (c *RPC) Client(broker *plugin.MuxBroker, client *plugin.RPCConnection) (interface{}, error) {
	return &ExtClient{broker, client}, nil
}

func (s *RPC) Server(b *plugin.MuxBroker) (interface{}, error) {
	return &HostServer{s.Impl, b}, nil
}
