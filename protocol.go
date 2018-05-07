package plugin

import (
	"io"
	"net"
)

// Protocol is an enum representing the types of protocols.
type Protocol string

const (
	ProtocolInvalid Protocol = ""
	ProtocolNetRPC  Protocol = "netrpc"
	ProtocolGRPC    Protocol = "grpc"
)

// ServerProtocol is an interface that must be implemented for new plugin
// protocols to be servers.
type ServerProtocol interface {
	// Init is called once to configure and initialize the protocol, but
	// not start listening. This is the point at which all validation should
	// be done and errors returned.
	Init() error

	// Config is extra configuration to be outputted to stdout. This will
	// be automatically base64 encoded to ensure it can be parsed properly.
	// This can be an empty string if additional configuration is not needed.
	Config() string

	// Serve is called to serve connections on the given listener. This should
	// continue until the listener is closed.
	Serve(net.Listener)
}

// ClientProtocol is an interface that must be implemented for new plugin
// protocols to be clients.
type ClientProtocol interface {
	io.Closer

	// Dispense dispenses a new instance of the plugin with the given name.
	Dispense(string) (interface{}, error)

	// Ping checks that the client connection is still healthy.
	Ping() error

	// Reconnect attempts to recreate the connection to a new server
	Reconnect(*Client) error
}

// ClientServer...
type ClientServer interface {
	Serve(io.ReadWriteCloser)
}

// ServerClient [sp]Should this be io.Closer and implement reconnect
type ServerClient interface {
	// Dispense dispenses a new instance of the plugin with the given name.
	Dispense(string) (interface{}, error)
}
