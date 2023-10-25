//go:build js && wasm

package plugin

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"io"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/magodo/chanio"
	"github.com/magodo/go-wasmww"
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
		if exitCode >= 0 {
			os.Exit(exitCode)
		}
	}()

	if opts.Test != nil {
		fmt.Fprintf(os.Stderr, "Test is not supported\n")
		exitCode = 1
		return
	}

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
	listener, err := NewWebWorkerListener()
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
	stdout_r, stdout_w, err := chanio.Pipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error preparing plugin: %s\n", err)
		os.Exit(1)
	}
	stderr_r, stderr_w, err := chanio.Pipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error preparing plugin: %s\n", err)
		os.Exit(1)
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
	// bring it up.
	fmt.Printf("%d|%d|%s|%s|%s|%s\n",
		CoreProtocolVersion,
		protoVersion,
		listener.Addr().Network(),
		listener.Addr().String(),
		protoType,
		serverCert)
	os.Stdout.Sync()

	// Set our stdout, stderr to the stdio stream that clients can retrieve
	// using ClientConfig.SyncStdout/err.
	wasmww.SetWriteSync(
		[]wasmww.MsgWriter{wasmww.NewMsgWriterToIoWriter(stdout_w)},
		[]wasmww.MsgWriter{wasmww.NewMsgWriterToIoWriter(stderr_w)},
	)

	// Accept connections and wait for completion
	go server.Serve(listener)

	// Wait for the server itself to shut down
	<-doneCh
	// Note that given the documentation of Serve we should probably be
	// setting exitCode = 0 and using os.Exit here. That's how it used to
	// work before extracting this library. However, for years we've done
	// this so we'll keep this functionality.
}
