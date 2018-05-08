package client

import (
	plugin "github.com/sampaioletti/go-plugin/examples/go-plugin"
)

type HostServer struct {
	Impl   *Host
	Broker *plugin.MuxBroker
}

func (c *HostServer) HelloHost(hello string, resp *error) error {
	*resp = c.Impl.HelloHost(hello)
	return *resp
}
