package client

import (
	"gitlab.com/indis/libs/extension/api"
	plugin "gitlab.com/indis/libs/third_party/go-plugin"
)

var _ api.Extender = (*ExtClient)(nil)

type ExtClient struct {
	broker *plugin.MuxBroker
	client *plugin.RPCConnection
}

func (s *ExtClient) HelloExtension(name string) error {
	var rErr error
	lErr := s.client.Call("Plugin.HelloExtension", name, &rErr)
	if lErr != nil {
		return lErr
	}
	if rErr != nil {
		return rErr
	}
	return nil
}
