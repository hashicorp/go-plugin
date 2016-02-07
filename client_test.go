package plugin

import (
	"bytes"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	process := helperProcess("mock")
	c := NewClient(&ClientConfig{Cmd: process, HandshakeConfig: testHandshake})
	defer c.Kill()

	// Test that it parses the proper address
	addr, err := c.Start()
	if err != nil {
		t.Fatalf("err should be nil, got %s", err)
	}

	if addr.Network() != "tcp" {
		t.Fatalf("bad: %#v", addr)
	}

	if addr.String() != ":1234" {
		t.Fatalf("bad: %#v", addr)
	}

	// Test that it exits properly if killed
	c.Kill()

	if process.ProcessState == nil {
		t.Fatal("should have process state")
	}

	// Test that it knows it is exited
	if !c.Exited() {
		t.Fatal("should say client has exited")
	}
}

func TestClient_testInterface(t *testing.T) {
	process := helperProcess("test-interface")
	c := NewClient(&ClientConfig{
		Cmd:             process,
		HandshakeConfig: testHandshake,
		Plugins:         testPluginMap,
	})
	defer c.Kill()

	// Grab the RPC client
	client, err := c.Client()
	if err != nil {
		t.Fatalf("err should be nil, got %s", err)
	}

	// Grab the impl
	raw, err := client.Dispense("test")
	if err != nil {
		t.Fatalf("err should be nil, got %s", err)
	}

	impl, ok := raw.(testInterface)
	if !ok {
		t.Fatalf("bad: %#v", raw)
	}

	result := impl.Double(21)
	if result != 42 {
		t.Fatalf("bad: %#v", result)
	}

	// Kill it
	c.Kill()

	// Test that it knows it is exited
	if !c.Exited() {
		t.Fatal("should say client has exited")
	}
}

func TestClient_cmdAndReattach(t *testing.T) {
	config := &ClientConfig{
		Cmd:      helperProcess("start-timeout"),
		Reattach: &ReattachConfig{},
	}

	c := NewClient(config)
	defer c.Kill()

	_, err := c.Start()
	if err == nil {
		t.Fatal("err should not be nil")
	}
}

func TestClient_reattachNotFound(t *testing.T) {
	// Find a bad pid
	var pid int = 5000
	for i := pid; i < 32000; i++ {
		if _, err := os.FindProcess(i); err != nil {
			pid = i
			break
		}
	}

	// Addr that won't work
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	addr := l.Addr()
	l.Close()

	// Reattach
	c := NewClient(&ClientConfig{
		Reattach: &ReattachConfig{
			Addr: addr,
			Pid:  pid,
		},
		HandshakeConfig: testHandshake,
		Plugins:         testPluginMap,
	})

	// Start shouldn't error
	if _, err := c.Start(); err == nil {
		t.Fatal("should error")
	} else if err != ErrProcessNotFound {
		t.Fatalf("err: %s", err)
	}
}

func TestClientStart_badVersion(t *testing.T) {
	config := &ClientConfig{
		Cmd:             helperProcess("bad-version"),
		StartTimeout:    50 * time.Millisecond,
		HandshakeConfig: testHandshake,
	}

	c := NewClient(config)
	defer c.Kill()

	_, err := c.Start()
	if err == nil {
		t.Fatal("err should not be nil")
	}
}

func TestClient_Start_Timeout(t *testing.T) {
	config := &ClientConfig{
		Cmd:             helperProcess("start-timeout"),
		StartTimeout:    50 * time.Millisecond,
		HandshakeConfig: testHandshake,
	}

	c := NewClient(config)
	defer c.Kill()

	_, err := c.Start()
	if err == nil {
		t.Fatal("err should not be nil")
	}
}

func TestClient_Stderr(t *testing.T) {
	stderr := new(bytes.Buffer)
	process := helperProcess("stderr")
	c := NewClient(&ClientConfig{
		Cmd:             process,
		Stderr:          stderr,
		HandshakeConfig: testHandshake,
	})
	defer c.Kill()

	if _, err := c.Start(); err != nil {
		t.Fatalf("err: %s", err)
	}

	for !c.Exited() {
		time.Sleep(10 * time.Millisecond)
	}

	if !strings.Contains(stderr.String(), "HELLO\n") {
		t.Fatalf("bad log data: '%s'", stderr.String())
	}

	if !strings.Contains(stderr.String(), "WORLD\n") {
		t.Fatalf("bad log data: '%s'", stderr.String())
	}
}

func TestClient_Stdin(t *testing.T) {
	// Overwrite stdin for this test with a temporary file
	tf, err := ioutil.TempFile("", "terraform")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(tf.Name())
	defer tf.Close()

	if _, err = tf.WriteString("hello"); err != nil {
		t.Fatalf("error: %s", err)
	}

	if err = tf.Sync(); err != nil {
		t.Fatalf("error: %s", err)
	}

	if _, err = tf.Seek(0, 0); err != nil {
		t.Fatalf("error: %s", err)
	}

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	os.Stdin = tf

	process := helperProcess("stdin")
	c := NewClient(&ClientConfig{Cmd: process, HandshakeConfig: testHandshake})
	defer c.Kill()

	_, err = c.Start()
	if err != nil {
		t.Fatalf("error: %s", err)
	}

	for {
		if c.Exited() {
			break
		}

		time.Sleep(50 * time.Millisecond)
	}

	if !process.ProcessState.Success() {
		t.Fatal("process didn't exit cleanly")
	}
}
