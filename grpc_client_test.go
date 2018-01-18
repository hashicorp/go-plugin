package plugin

import (
	"context"
	"testing"

	"github.com/hashicorp/go-plugin/test/grpc"
	"google.golang.org/grpc"
)

func TestGRPCClient_App(t *testing.T) {
	client, _ := TestPluginGRPCConn(t, map[string]Plugin{
		"test": new(testInterfacePlugin),
	})
	defer client.Close()

	raw, err := client.Dispense("test")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	impl, ok := raw.(testInterface)
	if !ok {
		t.Fatalf("bad: %#v", raw)
	}

	result := impl.Double(21)
	if result != 42 {
		t.Fatalf("bad: %#v", result)
	}

	err = impl.Bidirectional()
	if err != nil {
		t.Fatal(err)
	}
}

func TestGRPCConn_BidirectionalPing(t *testing.T) {
	conn, _ := TestGRPCConn(t, func(s *grpc.Server) {
		grpctest.RegisterPingPongServer(s, &pingPongServer{})
	})
	defer conn.Close()
	pingPongClient := grpctest.NewPingPongClient(conn)

	pResp, err := pingPongClient.Ping(context.Background(), &grpctest.PingRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if pResp.Msg != "pong" {
		t.Fatal("Bad PingPong")
	}
}

func TestGRPCClient_Ping(t *testing.T) {
	client, server := TestPluginGRPCConn(t, map[string]Plugin{
		"test": new(testInterfacePlugin),
	})
	defer client.Close()

	// Run a couple pings
	if err := client.Ping(); err != nil {
		t.Fatalf("err: %s", err)
	}
	if err := client.Ping(); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Close the remote end
	server.server.Stop()

	// Test ping fails
	if err := client.Ping(); err == nil {
		t.Fatal("should error")
	}
}
