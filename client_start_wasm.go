//go:build js && wasm

package plugin

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-plugin/internal/wasmrunner"
	"github.com/hashicorp/go-plugin/runner"
)

// Start the underlying subprocess, communicating with it to negotiate
// a port for RPC connections, and returning the address to connect via RPC.
//
// This method is safe to call multiple times. Subsequent calls have no effect.
// Once a client has been started once, it cannot be started again, even if
// it was killed.
func (c *Client) Start() (addr net.Addr, err error) {
	c.l.Lock()
	defer c.l.Unlock()

	if c.address != nil {
		return c.address, nil
	}

	if c.config.Cmd == nil {
		return nil, fmt.Errorf("Cmd must be set")
	}
	if c.config.Reattach != nil {
		return nil, fmt.Errorf("Reattach is not supported")
	}
	if c.config.RunnerFunc != nil {
		return nil, fmt.Errorf("RunnerFunc is not supported")
	}
	if c.config.SecureConfig != nil {
		return nil, fmt.Errorf("SecureConfig is not supported")
	}

	if c.config.VersionedPlugins == nil {
		c.config.VersionedPlugins = make(map[int]PluginSet)
	}

	// handle all plugins as versioned, using the handshake config as the default.
	version := int(c.config.ProtocolVersion)

	// Make sure we're not overwriting a real version 0. If ProtocolVersion was
	// non-zero, then we have to just assume the user made sure that
	// VersionedPlugins doesn't conflict.
	if _, ok := c.config.VersionedPlugins[version]; !ok && c.config.Plugins != nil {
		c.config.VersionedPlugins[version] = c.config.Plugins
	}

	var versionStrings []string
	for v := range c.config.VersionedPlugins {
		versionStrings = append(versionStrings, strconv.Itoa(v))
	}

	env := []string{
		fmt.Sprintf("%s=%s", c.config.MagicCookieKey, c.config.MagicCookieValue),
		fmt.Sprintf("PLUGIN_MIN_PORT=%d", c.config.MinPort),
		fmt.Sprintf("PLUGIN_MAX_PORT=%d", c.config.MaxPort),
		fmt.Sprintf("PLUGIN_PROTOCOL_VERSIONS=%s", strings.Join(versionStrings, ",")),
	}

	cmd := c.config.Cmd
	if !c.config.SkipHostEnv {
		cmd.Env = append(cmd.Env, os.Environ()...)
	}
	cmd.Env = append(cmd.Env, env...)
	cmd.Stdin = os.Stdin

	// Setup a temporary certificate for client/server mtls, and send the public
	// certificate to the plugin.
	if c.config.AutoMTLS {
		c.logger.Info("configuring client automatic mTLS")
		certPEM, keyPEM, err := generateCert()
		if err != nil {
			c.logger.Error("failed to generate client certificate", "error", err)
			return nil, err
		}
		cert, err := tls.X509KeyPair(certPEM, keyPEM)
		if err != nil {
			c.logger.Error("failed to parse client certificate", "error", err)
			return nil, err
		}

		cmd.Env = append(cmd.Env, fmt.Sprintf("PLUGIN_CLIENT_CERT=%s", certPEM))

		c.config.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
			ClientAuth:   tls.RequireAndVerifyClientCert,
			MinVersion:   tls.VersionTLS12,
			ServerName:   "localhost",
		}
	}

	var runner runner.Runner
	runner, err = wasmrunner.NewWasmRunner(c.logger, cmd)
	if err != nil {
		return nil, err
	}

	c.runner = runner
	startCtx, startCtxCancel := context.WithTimeout(context.Background(), c.config.StartTimeout)
	defer startCtxCancel()
	err = runner.Start(startCtx)
	if err != nil {
		return nil, err
	}

	// Make sure the command is properly cleaned up if there is an error
	defer func() {
		rErr := recover()

		if err != nil || rErr != nil {
			runner.Kill(context.Background())
		}

		if rErr != nil {
			panic(rErr)
		}
	}()

	// Create a context for when we kill
	c.doneCtx, c.ctxCancel = context.WithCancel(context.Background())

	// Start goroutine that logs the stderr
	c.clientWaitGroup.Add(1)
	c.stderrWaitGroup.Add(1)
	// logStderr calls Done()
	go c.logStderr(runner.Name(), runner.Stderr())

	c.clientWaitGroup.Add(1)
	go func() {
		// ensure the context is cancelled when we're done
		defer c.ctxCancel()

		defer c.clientWaitGroup.Done()

		// wait to finish reading from stderr since the stderr pipe reader
		// will be closed by the subsequent call to cmd.Wait().
		c.stderrWaitGroup.Wait()

		// Wait for the command to end.
		err := runner.Wait(context.Background())
		if err != nil {
			c.logger.Error("plugin process exited", "plugin", runner.Name(), "id", runner.ID(), "error", err.Error())
		} else {
			// Log and make sure to flush the logs right away
			c.logger.Info("plugin process exited", "plugin", runner.Name(), "id", runner.ID())
		}

		os.Stderr.Sync()

		// Set that we exited, which takes a lock
		c.l.Lock()
		defer c.l.Unlock()
		c.exited = true
	}()

	// Start a goroutine that is going to be reading the lines
	// out of stdout
	linesCh := make(chan string)
	c.clientWaitGroup.Add(1)
	go func() {
		defer c.clientWaitGroup.Done()
		defer close(linesCh)

		scanner := bufio.NewScanner(runner.Stdout())
		for scanner.Scan() {
			linesCh <- scanner.Text()
		}
		if scanner.Err() != nil {
			c.logger.Error("error encountered while scanning stdout", "error", scanner.Err())
		}
	}()

	// Make sure after we exit we read the lines from stdout forever
	// so they don't block since it is a pipe.
	// The scanner goroutine above will close this, but track it with a wait
	// group for completeness.
	c.clientWaitGroup.Add(1)
	defer func() {
		go func() {
			defer c.clientWaitGroup.Done()
			for range linesCh {
			}
		}()
	}()

	// Some channels for the next step
	timeout := time.After(c.config.StartTimeout)

	// Start looking for the address
	c.logger.Debug("waiting for RPC address", "plugin", runner.Name())
	select {
	case <-timeout:
		err = errors.New("timeout while waiting for plugin to start")
	case <-c.doneCtx.Done():
		err = errors.New("plugin exited before we could connect")
	case line, ok := <-linesCh:
		// Trim the line and split by "|" in order to get the parts of
		// the output.
		line = strings.TrimSpace(line)
		parts := strings.SplitN(line, "|", 6)
		if len(parts) < 4 {
			errText := fmt.Sprintf("Unrecognized remote plugin message: %s", line)
			if !ok {
				errText += "\n" + "Failed to read any lines from plugin's stdout"
			}
			additionalNotes := runner.Diagnose(context.Background())
			if additionalNotes != "" {
				errText += "\n" + additionalNotes
			}
			err = errors.New(errText)
			return
		}

		// Check the core protocol. Wrapped in a {} for scoping.
		{
			var coreProtocol int
			coreProtocol, err = strconv.Atoi(parts[0])
			if err != nil {
				err = fmt.Errorf("Error parsing core protocol version: %s", err)
				return
			}

			if coreProtocol != CoreProtocolVersion {
				err = fmt.Errorf("Incompatible core API version with plugin. "+
					"Plugin version: %s, Core version: %d\n\n"+
					"To fix this, the plugin usually only needs to be recompiled.\n"+
					"Please report this to the plugin author.", parts[0], CoreProtocolVersion)
				return
			}
		}

		// Test the API version
		version, pluginSet, err := c.checkProtoVersion(parts[1])
		if err != nil {
			return addr, err
		}

		// set the Plugins value to the compatible set, so the version
		// doesn't need to be passed through to the ClientProtocol
		// implementation.
		c.config.Plugins = pluginSet
		c.negotiatedVersion = version
		c.logger.Debug("using plugin", "version", version)

		network, address, err := runner.PluginToHost(parts[2], parts[3])
		if err != nil {
			return addr, err
		}

		switch network {
		case "webworker":
			addr, err = ParseWebWorkerAddr(address)
		default:
			err = fmt.Errorf("Unknown address type: %s", address)
		}

		// If we have a server type, then record that. We default to net/rpc
		// for backwards compatibility.
		c.protocol = ProtocolNetRPC
		if len(parts) >= 5 {
			c.protocol = Protocol(parts[4])
		}

		found := false
		for _, p := range c.config.AllowedProtocols {
			if p == c.protocol {
				found = true
				break
			}
		}
		if !found {
			err = fmt.Errorf("Unsupported plugin protocol %q. Supported: %v",
				c.protocol, c.config.AllowedProtocols)
			return addr, err
		}

		// See if we have a TLS certificate from the server.
		// Checking if the length is > 50 rules out catching the unused "extra"
		// data returned from some older implementations.
		if len(parts) >= 6 && len(parts[5]) > 50 {
			err := c.loadServerCert(parts[5])
			if err != nil {
				return nil, fmt.Errorf("error parsing server cert: %s", err)
			}
		}
	}

	c.address = addr
	return
}
