package main

import (
	"os"
	"os/exec"
	"time"

	"github.com/hashicorp/go-hclog"
	"gitlab.com/indis/libs/extension/api"
	"gitlab.com/indis/libs/extension/pkg/client"
	plugin "gitlab.com/indis/libs/third_party/go-plugin"
)

func main() {
	hclog.DefaultOptions = &hclog.LoggerOptions{
		Name:   "plugin",
		Output: os.Stdout,
		Level:  hclog.Debug,
	}
	logger := hclog.L()
	logger.Info("Host Started")
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: api.Handshake,
		Plugins: map[string]plugin.Plugin{
			"extension": &client.RPC{},
		},
		ServedPlugins: map[string]plugin.Plugin{
			"host": &client.RPC{},
		},
		Cmd:       exec.Command("./extension/extension"),
		Logger:    logger,
		Managed:   true,
		Reconnect: true,
	})
	defer client.Kill()
	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		logger.Debug(err.Error())
		panic(0)
	}

	raw, err := rpcClient.Dispense("extension")
	if err != nil {
		logger.Debug(err.Error())
		panic(0)
	}

	if ext, ok := raw.(api.Extender); ok {
		go func() {
			ch := time.NewTicker(time.Second * 5).C
			for {
				select {
				case <-ch:
					if err := ext.HelloExtension("Hello From Host"); err != nil {
						logger.Debug(err.Error())
					}
				}
			}

		}()
	}
	time.Sleep(time.Second * 1)
	done, err := client.Serve()
	if err != nil {
		logger.Debug(err.Error())
		panic(0)
	}
	<-done
}
