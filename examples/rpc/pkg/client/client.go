package client

import (
	hclog "github.com/hashicorp/go-hclog"
	plugin "gitlab.com/indis/libs/third_party/go-plugin"
)

var _ plugin.Plugin = (*RPC)(nil)

type RPC struct {
	Impl   *Host
	logger hclog.Logger
	// rpc    plugin.ClientProtocol
	// Plugin *plugin.Client
	//Extension api.Extender
	// name string
}

func (c *RPC) Client(broker *plugin.MuxBroker, client *plugin.RPCConnection) (interface{}, error) {
	return &ExtClient{broker, client}, nil
}

func (s *RPC) Server(b *plugin.MuxBroker) (interface{}, error) {
	return &HostServer{s.Impl, b}, nil
}

// func (r *RPC) FindAndDispenseExt(name string) (api.Extender, error) {
// 	r.name = name
// 	pluginLogger := r.logger.Named(name)
// 	files, err := Discover(name, r.logger)
// 	if err != nil {
// 		return nil, err
// 	}
// 	for _, v := range files {

// 		r.logger.Debug("Creating", "path", v)
// 		var cl plugin.ClientProtocol
// 		if r.Plugin == nil {
// 			r.Plugin = r.newClient(v, pluginLogger)
// 		} else {
// 			r.Plugin.Config.Cmd = exec.Command(v)
// 			r.Plugin.Config.Logger = pluginLogger
// 		}
// 		r.Plugin = r.newClient(v, pluginLogger)
// 		cl, err = r.Plugin.Client()
// 		if err != nil {
// 			if cl != nil {
// 				cl.Close()
// 			}
// 			r.logger.Debug(err.Error(), "file", v)
// 			continue
// 		}
// 		r.rpc = cl
// 		r.logger.Debug("client created", "client", cl, "err", err)
// 		var raw interface{}
// 		r.logger.Debug("Dispensing")
// 		raw, err = cl.Dispense("plugin")
// 		if err != nil {
// 			r.logger.Debug(err.Error(), "file", v)
// 			cl.Close()
// 			continue
// 		}
// 		pl, ok := raw.(api.Extender)
// 		if !ok {
// 			cl.Close()
// 			err = errors.New("Not of Type Plugin")
// 			r.logger.Debug(err.Error(), "file", v)
// 			continue
// 		}

// 		r.logger.Debug("Plugin Started")
// 		return pl, nil
// 	}
// 	return nil, err
// }

// func (r *RPC) newClient(p string, logger hclog.Logger) *plugin.Client {
// 	return plugin.NewClient(&plugin.ClientConfig{
// 		HandshakeConfig: api.Handshake,
// 		Plugins:         map[string]plugin.Plugin{"plugin": r},
// 		Cmd:             exec.Command(p),
// 		Logger:          logger,
// 		Managed:         true,
// 		Reconnect:       true,
// 	})
// }
// func (r *RPC) Close() error {
// 	if r.Plugin != nil {
// 		r.Plugin.Kill()
// 	}
// 	return nil
// }
