//go:build wasm && js

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/hashicorp/go-plugin/examples/grpc/shared"
)

func main() {
	// We don't want to see the plugin logs.
	log.SetOutput(ioutil.Discard)

	// When using WASM, we are not reassigning the os.Stdout/Stderr, but reimplement the writeSync used underlying.
	// This means:
	// - Previously, the `log` package (and the `hclog`) used in the provider will write to the plugin process's original stderr.
	//   Only those explicit write to stdout/stderr (e.g. via fmt.Fprint()), routes to the client via RPC.
	// - Now, no matter using `log` or explicit write, the logs all route to the client via RPC.
	//
	// Hence, we'll need a logger below to filter the non-interesting trace logs.
	// Also, we'll need to define the SyncStdout/Stderr in the client option, to avoid routing them to io.Discard.
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "plugin",
		Output: os.Stdout,
		// Debug level is used to filter the plugin trace logs
		Level: hclog.Debug,
	})

	// We're a host. Start by launching the plugin process.
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: shared.Handshake,
		Plugins:         shared.PluginMap,
		Cmd:             exec.Command("kv-go-grpc.wasm"),
		AllowedProtocols: []plugin.Protocol{
			plugin.ProtocolNetRPC, plugin.ProtocolGRPC,
		},
		Logger:     logger,
		SyncStdout: os.Stdout,
		SyncStderr: os.Stderr,
	})
	defer client.Kill()

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}

	// Request the plugin
	raw, err := rpcClient.Dispense("kv_grpc")
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}

	kv := raw.(shared.KV)
	if err := kv.Put("hello", []byte("world")); err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}

	result, err := kv.Get("hello")
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}

	fmt.Println(string(result))

	// Give the grpc stdio copyChan go routines some time to finish copying
	time.Sleep(time.Millisecond * 100)

	os.Exit(0)
}
