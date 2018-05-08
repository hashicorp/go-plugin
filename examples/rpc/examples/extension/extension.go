package main

import (
	"os"
	"time"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/sampaioletti/go-plugin/examples/api"
	plugin "github.com/sampaioletti/go-plugin/examples/go-plugin"
	"github.com/sampaioletti/go-plugin/examples/pkg/server"
)

func main() {
	hclog.DefaultOptions = &hclog.LoggerOptions{
		Name:   "extension",
		Output: os.Stderr,
		Level:  hclog.Info,
	}
	logger := hclog.L()
	logger.Info("Started Extension")

	//create config with added Accepted Plugins field, this will be the interfaces we will accept from host
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

	//serveanddispense creates a non blocking server and returns a "ServerClient" interface if the protocol implements it
	done, c, err := plugin.ServeAndDispense(cfg)
	if err != nil {
		logger.Debug(err.Error())
	}
	//hack to give time for connection to be established before we dispense which requires the active connections
	time.Sleep(time.Second * 1)

	//this works similarly to the client.Dispense method, except it doesn't load files and dispenses over the existing connection
	//must currently be ran after serve and dispense or the connection doesn't exit
	raw, err := c.Dispense("host")
	if err != nil {
		logger.Debug(err.Error())
	}

	//this works same as normal "dispensed" plugin
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
	//done blocks until the server exits
	<-done
}
