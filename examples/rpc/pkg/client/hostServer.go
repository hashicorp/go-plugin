package client

import (
	plugin "gitlab.com/indis/libs/third_party/go-plugin"
)

type HostServer struct {
	Impl   *Host
	Broker *plugin.MuxBroker
}

func (c *HostServer) HelloHost(hello string, resp *error) error {
	*resp = c.Impl.HelloHost(hello)
	return *resp
}
