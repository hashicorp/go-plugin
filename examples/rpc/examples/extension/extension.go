package main

import (
	"os"
	"time"

	hclog "github.com/hashicorp/go-hclog"
	"gitlab.com/indis/libs/extension/api"
	"gitlab.com/indis/libs/extension/pkg/server"
	plugin "gitlab.com/indis/libs/third_party/go-plugin"
)

func main() {
	hclog.DefaultOptions = &hclog.LoggerOptions{
		Name:   "extension",
		Output: os.Stderr,
		Level:  hclog.Info,
	}
	logger := hclog.L()
	logger.Info("Started Extension")
	cfg := &plugin.ServeConfig{
		HandshakeConfig: api.Handshake,
		Plugins: map[string]plugin.Plugin{
			"extension": &server.RPC{Impl: &server.Extension{}},
		},
		AcceptPlugins: map[string]plugin.Plugin{
			"host": &server.RPC{},
		},
		Logger: logger,
	}

	done, c, err := plugin.ServeAndDispense(cfg)
	if err != nil {
		logger.Debug(err.Error())
	}
	time.Sleep(time.Second * 5)
	raw, err := c.Dispense("host")
	if err != nil {
		logger.Debug(err.Error())
	}
	if host, ok := raw.(api.Host); ok {
		go func() {
			ch := time.NewTicker(time.Second * 5).C
			for {
				select {
				case <-ch:
					if err := host.HelloHost("Hello From Extension"); err != nil {
						logger.Debug(err.Error())
					}
				}
			}

		}()
	}
	<-done
}
