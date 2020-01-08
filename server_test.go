package plugin

import (
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestServer_GRPCCancel(t *testing.T) {
	// Create a temporary dir to store the result file
	td, err := ioutil.TempDir("", "plugin")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(td)

	// Create a path that the helper process will write on cleanup
	path := filepath.Join(td, "output")

	// Test the self timeout/cancel
	process := helperProcess("test-grpc-cancel", path)
	c := NewClient(&ClientConfig{
		Cmd:              process,
		HandshakeConfig:  testHandshake,
		Plugins:          testGRPCPluginMap,
		AllowedProtocols: []Protocol{ProtocolGRPC},
	})

	// Grab the client so the process starts
	if _, err := c.Client(); err != nil {
		c.Kill()
		t.Fatalf("err: %s", err)
	}

	// Wait for the server to cancel itself
	for !c.Exited() {
		time.Sleep(500 * time.Millisecond)
	}

	// Test for the file
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestRmListener_impl(t *testing.T) {
	var _ net.Listener = new(rmListener)
}

func TestRmListener(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	tf, err := ioutil.TempFile("", "plugin")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	path := tf.Name()

	// Close the file
	if err := tf.Close(); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Create the listener and test close
	rmL := &rmListener{
		Listener: l,
		Path:     path,
	}
	if err := rmL.Close(); err != nil {
		t.Fatalf("err: %s", err)
	}

	// File should be goe
	if _, err := os.Stat(path); err == nil || !os.IsNotExist(err) {
		t.Fatalf("err: %s", err)
	}
}

func TestProtocolSelection_no_server(t *testing.T) {
	conf := &ServeConfig{
		HandshakeConfig: testVersionedHandshake,
		VersionedPlugins: map[int]PluginSet{
			2: testGRPCPluginMap,
		},
		GRPCServer:  DefaultGRPCServer,
		TLSProvider: helperTLSProvider,
	}

	_, protocol, _ := protocolVersion(conf)
	if protocol != ProtocolGRPC {
		t.Fatalf("bad protocol %s", protocol)
	}

	conf = &ServeConfig{
		HandshakeConfig: testVersionedHandshake,
		VersionedPlugins: map[int]PluginSet{
			2: testGRPCPluginMap,
		},
		TLSProvider: helperTLSProvider,
	}

	_, protocol, _ = protocolVersion(conf)
	if protocol != ProtocolNetRPC {
		t.Fatalf("bad protocol %s", protocol)
	}

}
