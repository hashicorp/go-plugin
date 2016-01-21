package plugin

import (
	"fmt"
	"log"
	"net/rpc"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/hashicorp/otto/app"
	pluginrpc "github.com/hashicorp/otto/rpc"
)

// testAPIVersion is the ProtocolVersion we use for testing.
const testAPIVersion uint = 1

// testInterface is the test interface we use for plugins.
type testInterface interface {
	Double(int) int
}

// testInterfacePlugin is the implementation of Plugin to create
// RPC client/server implementations for testInterface.
type testInterfacePlugin struct{}

func (p *testInterfacePlugin) Server(b *MuxBroker) (interface{}, error) {
	return &testInterfaceServer{Impl: new(testInterfaceImpl)}, nil
}

func (p *testInterfacePlugin) Client(b *MuxBroker, c *rpc.Client) (interface{}, error) {
	return &testInterfaceClient{Client: c}, nil
}

// testInterfaceImpl implements testInterface concretely
type testInterfaceImpl struct{}

func (i *testInterfaceImpl) Double(v int) int { return v * 2 }

// testInterfaceClient implements testInterface to communicate over RPC
type testInterfaceClient struct {
	Client *rpc.Client
}

func (impl *testInterfaceClient) Double(v int) int {
	var resp int
	err := impl.Client.Call("Plugin.Double", v, &resp)
	if err != nil {
		panic(err)
	}

	return resp
}

// testInterfaceServer is the RPC server for testInterfaceClient
type testInterfaceServer struct {
	Broker *MuxBroker
	Impl   testInterface
}

func (s *testInterfaceServer) Double(arg int, resp *int) error {
	*resp = s.Impl.Double(arg)
	return nil
}

func helperProcess(s ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--"}
	cs = append(cs, s...)
	env := []string{
		"GO_WANT_HELPER_PROCESS=1",
		"OTTO_PLUGIN_MIN_PORT=10000",
		"OTTO_PLUGIN_MAX_PORT=25000",
	}

	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = append(env, os.Environ()...)
	return cmd
}

// This is not a real test. This is just a helper process kicked off by
// tests.
func TestHelperProcess(*testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	defer os.Exit(0)

	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}

		args = args[1:]
	}

	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "No command\n")
		os.Exit(2)
	}

	cmd, args := args[0], args[1:]
	switch cmd {
	case "bad-version":
		fmt.Printf("%d1|tcp|:1234\n", testAPIVersion)
		<-make(chan int)
	case "invalid-rpc-address":
		fmt.Println("lolinvalid")
	case "mock":
		fmt.Printf("%d|tcp|:1234\n", testAPIVersion)
		<-make(chan int)
	case "start-timeout":
		time.Sleep(1 * time.Minute)
		os.Exit(1)
	case "stderr":
		fmt.Printf("%d|tcp|:1234\n", testAPIVersion)
		log.Println("HELLO")
		log.Println("WORLD")
	case "stdin":
		fmt.Printf("%d|tcp|:1234\n", testAPIVersion)
		data := make([]byte, 5)
		if _, err := os.Stdin.Read(data); err != nil {
			log.Printf("stdin read error: %s", err)
			os.Exit(100)
		}

		if string(data) == "hello" {
			os.Exit(0)
		}

		os.Exit(1)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %q\n", cmd)
		os.Exit(2)
	}
}

func testAppFixed(p app.App) pluginrpc.AppFunc {
	return func() app.App {
		return p
	}
}
