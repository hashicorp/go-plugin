package plugin

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync/atomic"

	pluginrpc "github.com/hashicorp/otto/rpc"
)

// The APIVersion is outputted along with the RPC address. The plugin
// client validates this API version and will show an error if it doesn't
// know how to speak it.
const APIVersion = "1"

// The "magic cookie" is used to verify that the user intended to
// actually run this binary. If this cookie isn't present as an
// environmental variable, then we bail out early with an error.
const MagicCookieKey = "OTTO_PLUGIN_MAGIC_COOKIE"
const MagicCookieValue = "11aab7ff21cb9ff7b0e9975d53f17a8dab571eac9b5ff0191730046698f07b7f"

// ServeOpts configures what sorts of plugins are served.
type ServeOpts struct {
	AppFunc pluginrpc.AppFunc
}

// Serve serves the plugins given by ServeOpts.
//
// Serve doesn't return until the plugin is done being executed. Any
// errors will be outputted to the log.
func Serve(opts *ServeOpts) {
	// First check the cookie
	if os.Getenv(MagicCookieKey) != MagicCookieValue {
		fmt.Fprintf(os.Stderr,
			"This binary is an Otto plugin. These are not meant to be\n"+
				"executed directly. Please execute `otto`, which will load\n"+
				"any plugins automatically.\n")
		os.Exit(1)
	}

	// Logging goes to the original stderr
	log.SetOutput(os.Stderr)

	// Create our new stdout, stderr files. These will override our built-in
	// stdout/stderr so that it works across the stream boundary.
	stdout_r, stdout_w, err := os.Pipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error preparing Otto plugin: %s\n", err)
		os.Exit(1)
	}
	stderr_r, stderr_w, err := os.Pipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error preparing Otto plugin: %s\n", err)
		os.Exit(1)
	}

	// Register a listener so we can accept a connection
	listener, err := serverListener()
	if err != nil {
		log.Printf("[ERR] plugin init: %s", err)
		return
	}
	defer listener.Close()

	// Create the RPC server to dispense
	server := &pluginrpc.Server{
		AppFunc: opts.AppFunc,
		Stdout:  stdout_r,
		Stderr:  stderr_r,
	}

	// Output the address and service name to stdout so that core can bring it up.
	log.Printf("Plugin address: %s %s\n",
		listener.Addr().Network(), listener.Addr().String())
	fmt.Printf("%s|%s|%s\n",
		APIVersion,
		listener.Addr().Network(),
		listener.Addr().String())
	os.Stdout.Sync()

	// Eat the interrupts
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	go func() {
		var count int32 = 0
		for {
			<-ch
			newCount := atomic.AddInt32(&count, 1)
			log.Printf(
				"Received interrupt signal (count: %d). Ignoring.",
				newCount)
		}
	}()

	// Set our new out, err
	os.Stdout = stdout_w
	os.Stderr = stderr_w

	// Serve
	server.Accept(listener)
}

func serverListener() (net.Listener, error) {
	if runtime.GOOS == "windows" {
		return serverListener_tcp()
	}

	return serverListener_unix()
}

func serverListener_tcp() (net.Listener, error) {
	minPort, err := strconv.ParseInt(os.Getenv("OTTO_PLUGIN_MIN_PORT"), 10, 32)
	if err != nil {
		return nil, err
	}

	maxPort, err := strconv.ParseInt(os.Getenv("OTTO_PLUGIN_MAX_PORT"), 10, 32)
	if err != nil {
		return nil, err
	}

	for port := minPort; port <= maxPort; port++ {
		address := fmt.Sprintf("127.0.0.1:%d", port)
		listener, err := net.Listen("tcp", address)
		if err == nil {
			return listener, nil
		}
	}

	return nil, errors.New("Couldn't bind plugin TCP listener")
}

func serverListener_unix() (net.Listener, error) {
	tf, err := ioutil.TempFile("", "otto-plugin")
	if err != nil {
		return nil, err
	}
	path := tf.Name()

	// Close the file and remove it because it has to not exist for
	// the domain socket.
	if err := tf.Close(); err != nil {
		return nil, err
	}
	if err := os.Remove(path); err != nil {
		return nil, err
	}

	return net.Listen("unix", path)
}
