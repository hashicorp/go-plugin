package plugin

import (
	"context"
	"io/ioutil"
	"net"
	"os"
	"testing"
	"time"
)

func TestServer_testMode(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan *ReattachConfig, 1)
	closeCh := make(chan struct{})
	go Serve(&ServeConfig{
		HandshakeConfig: testHandshake,
		Plugins:         testGRPCPluginMap,
		GRPCServer:      DefaultGRPCServer,
		Test: &ServeTestConfig{
			Context:          ctx,
			ReattachConfigCh: ch,
			CloseCh:          closeCh,
		},
	})

	// We should get a config
	var config *ReattachConfig
	select {
	case config = <-ch:
	case <-time.After(2000 * time.Millisecond):
		t.Fatal("should've received reattach")
	}
	if config == nil {
		t.Fatal("config should not be nil")
	}

	// Connect!
	c := NewClient(&ClientConfig{
		Cmd:              nil,
		HandshakeConfig:  testHandshake,
		Plugins:          testGRPCPluginMap,
		Reattach:         config,
		AllowedProtocols: []Protocol{ProtocolGRPC},
	})
	client, err := c.Client()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Pinging should work
	if err := client.Ping(); err != nil {
		t.Fatalf("should not err: %s", err)
	}

	// Kill which should do nothing
	c.Kill()
	if err := client.Ping(); err != nil {
		t.Fatalf("should not err: %s", err)
	}

	ctx2, cancel2 := context.WithCancel(context.Background())
	defer cancel2()

	ch2 := make(chan *ReattachConfig, 1)
	closeCh2 := make(chan struct{})
	go Serve(&ServeConfig{
		HandshakeConfig: testHandshake,
		Plugins:         testGRPCPluginMap,
		GRPCServer:      DefaultGRPCServer,
		Test: &ServeTestConfig{
			Context:          ctx2,
			ReattachConfigCh: ch2,
			CloseCh:          closeCh2,
		},
	})

	var config2 *ReattachConfig
	select {
	case config2 = <-ch2:
	case <-time.After(2000 * time.Millisecond):
		t.Fatal("should've received reattach")
	}
	if config2 == nil {
		t.Fatal("config should not be nil")
	}

	// Canceling should cause an exit
	cancel()
	<-closeCh
	if err := client.Ping(); err == nil {
		t.Fatal("should error")
	}

	cancel2()
	<-closeCh2

	if os.Stdout.Name() != "/dev/stdout" {
		t.Fatalf("Stdout didn't get reset; is %q", os.Stdout.Name())
	}

	// Try logging, this should show out in tests. We have to manually verify.
	t.Logf("HELLO")
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
