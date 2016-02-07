package plugin

import (
	"fmt"
	"log"
	"net/rpc"
	"os"
	"os/exec"
	"testing"
	"time"
)

// testAPIVersion is the ProtocolVersion we use for testing.
var testHandshake = HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "TEST_MAGIC_COOKIE",
	MagicCookieValue: "test",
}

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

// testPluginMap can be used for tests as a plugin map
var testPluginMap = map[string]Plugin{
	"test": new(testInterfacePlugin),
}

func helperProcess(s ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--"}
	cs = append(cs, s...)
	env := []string{
		"GO_WANT_HELPER_PROCESS=1",
		"PLUGIN_MIN_PORT=10000",
		"PLUGIN_MAX_PORT=25000",
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
		fmt.Printf("%d|%d1|tcp|:1234\n", CoreProtocolVersion, testHandshake.ProtocolVersion)
		<-make(chan int)
	case "invalid-rpc-address":
		fmt.Println("lolinvalid")
	case "mock":
		fmt.Printf("%d|%d|tcp|:1234\n", CoreProtocolVersion, testHandshake.ProtocolVersion)
		<-make(chan int)
	case "start-timeout":
		time.Sleep(1 * time.Minute)
		os.Exit(1)
	case "stderr":
		fmt.Printf("%d|%d|tcp|:1234\n", CoreProtocolVersion, testHandshake.ProtocolVersion)
		log.Println("HELLO")
		log.Println("WORLD")
	case "stdin":
		fmt.Printf("%d|%d|tcp|:1234\n", CoreProtocolVersion, testHandshake.ProtocolVersion)
		data := make([]byte, 5)
		if _, err := os.Stdin.Read(data); err != nil {
			log.Printf("stdin read error: %s", err)
			os.Exit(100)
		}

		if string(data) == "hello" {
			os.Exit(0)
		}

		os.Exit(1)
	case "test-interface":
		Serve(&ServeConfig{
			HandshakeConfig: testHandshake,
			Plugins:         testPluginMap,
		})

		// Shouldn't reach here but make sure we exit anyways
		os.Exit(0)
	case "test-interface-daemon":
		// Serve!
		Serve(&ServeConfig{
			HandshakeConfig: testHandshake,
			Plugins:         testPluginMap,
		})

		// Shouldn't reach here but make sure we exit anyways
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %q\n", cmd)
		os.Exit(2)
	}
}
