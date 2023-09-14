//go:build wasm && js

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/go-plugin"
	"github.com/hashicorp/go-plugin/examples/grpc/shared"
)

// Here is a real implementation of KV that writes to a local file with
// the key name and the contents are the value of the key.
type KV struct {
	m map[string]string
}

func (kv KV) Put(key string, value []byte) error {
	fmt.Fprintln(os.Stderr, "PUT (stderr)")
	fmt.Fprintln(os.Stdout, "PUT (stdout)")
	log.Println("PUT (log)")
	value = []byte(fmt.Sprintf("%s\n\nWritten from plugin-go-grpc", string(value)))
	kv.m[key] = string(value)
	return nil
}

func (kv KV) Get(key string) ([]byte, error) {
	fmt.Fprintln(os.Stderr, "GET (stderr)")
	fmt.Fprintln(os.Stdout, "GET (stdout)")
	log.Println("GET (log)")
	v, ok := kv.m[key]
	if !ok {
		return nil, fmt.Errorf("key %q not exist", key)
	}
	return []byte(v), nil
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: shared.Handshake,
		Plugins: map[string]plugin.Plugin{
			"kv": &shared.KVGRPCPlugin{Impl: &KV{m: map[string]string{}}},
		},

		// A non-nil value here enables gRPC serving for this plugin...
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
