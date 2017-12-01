package main

import (
	"encoding/json"
	"io/ioutil"

	"github.com/hashicorp/go-plugin"
	"github.com/hashicorp/go-plugin/examples/grpc-bidirectional/shared"
)

// Here is a real implementation of KV that writes to a local file with
// the key name and the contents are the value of the key.
type KV struct {
	a shared.AddHelper
}

type data struct {
	Value int64
}

func (k *KV) Init(a shared.AddHelper) error {
	k.a = a
	return nil
}

func (k *KV) Put(key string, value int64) error {
	v, _ := k.Get(key)

	r, err := k.a.Sum(v, value)
	if err != nil {
		return err
	}

	buf, err := json.Marshal(&data{r})
	if err != nil {
		return err
	}

	return ioutil.WriteFile("kv_"+key, buf, 0644)
}

func (k *KV) Get(key string) (int64, error) {
	dataRaw, err := ioutil.ReadFile("kv_" + key)
	if err != nil {
		return 0, err
	}

	data := &data{}
	err = json.Unmarshal(dataRaw, data)
	if err != nil {
		return 0, err
	}

	return data.Value, nil
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: shared.Handshake,
		Plugins: map[string]plugin.Plugin{
			"kv": &shared.KVPlugin{Impl: &KV{}},
		},

		// A non-nil value here enables gRPC serving for this plugin...
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
