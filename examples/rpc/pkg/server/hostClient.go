package server

import (
	"gitlab.com/indis/libs/extension/api"
	plugin "gitlab.com/indis/libs/third_party/go-plugin"
)

var _ api.Host = (*HostClient)(nil)

type HostClient struct {
	broker *plugin.MuxBroker
	client *plugin.RPCConnection
}

func (h *HostClient) HelloHost(name string) error {
	var rErr error
	lErr := h.client.Call("Plugin.HelloHost", name, &rErr)
	if lErr != nil {
		return lErr
	}
	if rErr != nil {
		return rErr
	}
	return nil
}
