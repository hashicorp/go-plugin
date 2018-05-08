package server

import (
	plugin "github.com/sampaioletti/go-plugin/examples/go-plugin"
)

type ExtServer struct {
	Impl   *Extension
	Broker *plugin.MuxBroker
}

func (s *ExtServer) HelloExtension(hello string, resp *error) error {
	*resp = s.Impl.HelloExtension(hello)
	return *resp
}
