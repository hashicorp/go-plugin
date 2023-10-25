//go:build !wasm

package plugin

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"os/signal"

	"github.com/hashicorp/go-hclog"
)

// Serve serves the plugins given by ServeConfig.
//
// Serve doesn't return until the plugin is done being executed. Any
// fixable errors will be output to os.Stderr and the process will
// exit with a status code of 1. Serve will panic for unexpected
// conditions where a user's fix is unknown.
//
// This is the method that plugins should call in their main() functions.
func Serve(opts *ServeConfig) {
	exitCode := -1
	// We use this to trigger an `os.Exit` so that we can execute our other
	// deferred functions. In test mode, we just output the err to stderr
	// and return.
	defer func() {
		if opts.Test == nil && exitCode >= 0 {
			os.Exit(exitCode)
		}

		if opts.Test != nil && opts.Test.CloseCh != nil {
			close(opts.Test.CloseCh)
		}
	}()

	if opts.Test == nil {
		// Validate the handshake config
		if opts.MagicCookieKey == "" || opts.MagicCookieValue == "" {
			fmt.Fprintf(os.Stderr,
				"Misconfigured ServeConfig given to serve this plugin: no magic cookie\n"+
					"key or value was set. Please notify the plugin author and report\n"+
					"this as a bug.\n")
			exitCode = 1
			return
		}

		// First check the cookie
		if os.Getenv(opts.MagicCookieKey) != opts.MagicCookieValue {
			fmt.Fprintf(os.Stderr,
				"This binary is a plugin. These are not meant to be executed directly.\n"+
					"Please execute the program that consumes these plugins, which will\n"+
					"load any plugins automatically\n")
			exitCode = 1
			return
		}
	}

	// negotiate the version and plugins
	// start with default version in the handshake config
	protoVersion, protoType, pluginSet := protocolVersion(opts)

	logger := opts.Logger
	if logger == nil {
		// internal logger to os.Stderr
		logger = hclog.New(&hclog.LoggerOptions{
			Level:      hclog.Trace,
			Output:     os.Stderr,
			JSONFormat: true,
		})
	}

	// Register a listener so we can accept a connection
	listener, err := serverListener(unixSocketConfigFromEnv())
	if err != nil {
		logger.Error("plugin init error", "error", err)
		return
	}

	// Close the listener on return. We wrap this in a func() on purpose
	// because the "listener" reference may change to TLS.
	defer func() {
		listener.Close()
	}()

	var tlsConfig *tls.Config
	if opts.TLSProvider != nil {
		tlsConfig, err = opts.TLSProvider()
		if err != nil {
			logger.Error("plugin tls init", "error", err)
			return
		}
	}

	var serverCert string
	clientCert := os.Getenv("PLUGIN_CLIENT_CERT")
	// If the client is configured using AutoMTLS, the certificate will be here,
	// and we need to generate our own in response.
	if tlsConfig == nil && clientCert != "" {
		logger.Info("configuring server automatic mTLS")
		clientCertPool := x509.NewCertPool()
		if !clientCertPool.AppendCertsFromPEM([]byte(clientCert)) {
			logger.Error("client cert provided but failed to parse", "cert", clientCert)
		}

		certPEM, keyPEM, err := generateCert()
		if err != nil {
			logger.Error("failed to generate server certificate", "error", err)
			panic(err)
		}

		cert, err := tls.X509KeyPair(certPEM, keyPEM)
		if err != nil {
			logger.Error("failed to parse server certificate", "error", err)
			panic(err)
		}

		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
			ClientAuth:   tls.RequireAndVerifyClientCert,
			ClientCAs:    clientCertPool,
			MinVersion:   tls.VersionTLS12,
			RootCAs:      clientCertPool,
			ServerName:   "localhost",
		}

		// We send back the raw leaf cert data for the client rather than the
		// PEM, since the protocol can't handle newlines.
		serverCert = base64.RawStdEncoding.EncodeToString(cert.Certificate[0])
	}

	// Create the channel to tell us when we're done
	doneCh := make(chan struct{})

	// Create our new stdout, stderr files. These will override our built-in
	// stdout/stderr so that it works across the stream boundary.
	var stdout_r, stderr_r io.Reader
	stdout_r, stdout_w, err := os.Pipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error preparing plugin: %s\n", err)
		os.Exit(1)
	}
	stderr_r, stderr_w, err := os.Pipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error preparing plugin: %s\n", err)
		os.Exit(1)
	}

	// If we're in test mode, we tee off the reader and write the data
	// as-is to our normal Stdout and Stderr so that they continue working
	// while stdio works. This is because in test mode, we assume we're running
	// in `go test` or some equivalent and we want output to go to standard
	// locations.
	if opts.Test != nil {
		// TODO(mitchellh): This isn't super ideal because a TeeReader
		// only works if the reader side is actively read. If we never
		// connect via a plugin client, the output still gets swallowed.
		stdout_r = io.TeeReader(stdout_r, os.Stdout)
		stderr_r = io.TeeReader(stderr_r, os.Stderr)
	}

	// Build the server type
	var server ServerProtocol
	switch protoType {
	case ProtocolNetRPC:
		// If we have a TLS configuration then we wrap the listener
		// ourselves and do it at that level.
		if tlsConfig != nil {
			listener = tls.NewListener(listener, tlsConfig)
		}

		// Create the RPC server to dispense
		server = &RPCServer{
			Plugins: pluginSet,
			Stdout:  stdout_r,
			Stderr:  stderr_r,
			DoneCh:  doneCh,
		}

	case ProtocolGRPC:
		// Create the gRPC server
		server = &GRPCServer{
			Plugins: pluginSet,
			Server:  opts.GRPCServer,
			TLS:     tlsConfig,
			Stdout:  stdout_r,
			Stderr:  stderr_r,
			DoneCh:  doneCh,
			logger:  logger,
		}

	default:
		panic("unknown server protocol: " + protoType)
	}

	// Initialize the servers
	if err := server.Init(); err != nil {
		logger.Error("protocol init", "error", err)
		return
	}

	logger.Debug("plugin address", "network", listener.Addr().Network(), "address", listener.Addr().String())

	// Output the address and service name to stdout so that the client can
	// bring it up. In test mode, we don't do this because clients will
	// attach via a reattach config.
	if opts.Test == nil {
		fmt.Printf("%d|%d|%s|%s|%s|%s\n",
			CoreProtocolVersion,
			protoVersion,
			listener.Addr().Network(),
			listener.Addr().String(),
			protoType,
			serverCert)
		os.Stdout.Sync()
	} else if ch := opts.Test.ReattachConfigCh; ch != nil {
		// Send back the reattach config that can be used. This isn't
		// quite ready if they connect immediately but the client should
		// retry a few times.
		ch <- &ReattachConfig{
			Protocol:        protoType,
			ProtocolVersion: protoVersion,
			Addr:            listener.Addr(),
			Pid:             os.Getpid(),
			Test:            true,
		}
	}

	// Eat the interrupts. In test mode we disable this so that go test
	// can be cancelled properly.
	if opts.Test == nil {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt)
		go func() {
			count := 0
			for {
				<-ch
				count++
				logger.Trace("plugin received interrupt signal, ignoring", "count", count)
			}
		}()
	}

	// Set our stdout, stderr to the stdio stream that clients can retrieve
	// using ClientConfig.SyncStdout/err. We only do this for non-test mode
	// or if the test mode explicitly requests it.
	//
	// In test mode, we use a multiwriter so that the data continues going
	// to the normal stdout/stderr so output can show up in test logs. We
	// also send to the stdio stream so that clients can continue working
	// if they depend on that.
	if opts.Test == nil || opts.Test.SyncStdio {
		if opts.Test != nil {
			// In test mode we need to maintain the original values so we can
			// reset it.
			defer func(out, err *os.File) {
				os.Stdout = out
				os.Stderr = err
			}(os.Stdout, os.Stderr)
		}
		os.Stdout = stdout_w
		os.Stderr = stderr_w
	}

	// Accept connections and wait for completion
	go server.Serve(listener)

	ctx := context.Background()
	if opts.Test != nil && opts.Test.Context != nil {
		ctx = opts.Test.Context
	}
	select {
	case <-ctx.Done():
		// Cancellation. We can stop the server by closing the listener.
		// This isn't graceful at all but this is currently only used by
		// tests and its our only way to stop.
		listener.Close()

		// If this is a grpc server, then we also ask the server itself to
		// end which will kill all connections. There isn't an easy way to do
		// this for net/rpc currently but net/rpc is more and more unused.
		if s, ok := server.(*GRPCServer); ok {
			s.Stop()
		}

		// Wait for the server itself to shut down
		<-doneCh

	case <-doneCh:
		// Note that given the documentation of Serve we should probably be
		// setting exitCode = 0 and using os.Exit here. That's how it used to
		// work before extracting this library. However, for years we've done
		// this so we'll keep this functionality.
	}
}
