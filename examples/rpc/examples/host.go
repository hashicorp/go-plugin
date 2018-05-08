package main

import (
	"os"
	"os/exec"
	"os/signal"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/sampaioletti/go-plugin/examples/api"
	plugin "github.com/sampaioletti/go-plugin/examples/go-plugin"
	"github.com/sampaioletti/go-plugin/examples/pkg/client"
)

func main() {
	hclog.DefaultOptions = &hclog.LoggerOptions{
		Name:   "plugin",
		Output: os.Stdout,
		Level:  hclog.Debug,
	}
	logger := hclog.L()
	logger.Info("Host Started")

	//Try and intercept ^C etc and do a cleanup
	c := make(chan os.Signal, 1)
	signal.Notify(c)
	go func() {
		for range c {
			logger.Info("Recieved sig, cleanup")
			plugin.CleanupClients()
			logger.Info("Recieved sig, panic")
			panic(1)
		}
	}()

	//create a client adding the ServedPlugins map adn Reconnect value
	//ServedPlugins will be made available after a call to client.Server(..)
	//Reconnect should probably be called Resrart and could optionall be a string
	//that parses some reconnect logic like Reconnect: "1s" or enum
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

	//Just in case
	defer func() {
		if r := recover(); r != nil {
			logger.Info("Recovered from Panic")
		}
		plugin.CleanupClients()
	}()

	//Normal Client
	rpcClient, err := client.Client()
	if err != nil {
		logger.Debug(err.Error())
		panic(0)
	}
	//Normal Dispense
	raw, err := rpcClient.Dispense("extension")
	if err != nil {
		logger.Debug(err.Error())
		panic(0)
	}
	//Normal Access to dispensed plugin
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
	//Start a Server on the existing mux
	done, err := client.Serve()
	if err != nil {
		logger.Debug(err.Error())
		panic(0)
	}

	//wait for the server to exit
	<-done
}
