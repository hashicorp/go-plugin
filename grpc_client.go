package plugin

import (
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// newGRPCClient creates a new GRPCClient. The Client argument is expected
// to be successfully started already with a lock held.
func newGRPCClient(c *Client) (*GRPCClient, error) {
	// Build dialing options.
	opts := make([]grpc.DialOption, 0, 5)

	// We use a custom dialer so that we can connect over unix domain sockets
	opts = append(opts, grpc.WithDialer(c.dialer))

	// go-plugin expects to block the connection
	opts = append(opts, grpc.WithBlock())

	// Fail right away
	opts = append(opts, grpc.FailOnNonTempDialError(true))

	// If we have no TLS configuration set, we need to explicitly tell grpc
	// that we're connecting with an insecure connection.
	if c.config.TLSConfig == nil {
		opts = append(opts, grpc.WithInsecure())
	} else {
		opts = append(opts, grpc.WithTransportCredentials(
			credentials.NewTLS(c.config.TLSConfig)))
	}

	// Connect. Note the first parameter is unused because we use a custom
	// dialer that has the state to see the address.
	conn, err := grpc.Dial("unused", opts...)
	if err != nil {
		return nil, err
	}

	// Make the plugins
	ps := make(map[string]GRPCPlugin)
	for k, raw := range c.config.Plugins {
		p, ok := raw.(GRPCPlugin)
		if !ok {
			return nil, fmt.Errorf("%q is not a gRPC-compatible plugin", k)
		}

		ps[k] = p
	}

	return &GRPCClient{
		Conn:    conn,
		Plugins: ps,
	}, nil
}

// GRPCClient connects to a GRPCServer over gRPC to dispense plugin types.
type GRPCClient struct {
	Conn    *grpc.ClientConn
	Plugins map[string]GRPCPlugin
}

// ClientProtocol impl.
func (c *GRPCClient) Close() error {
	return c.Conn.Close()
}

// ClientProtocol impl.
func (c *GRPCClient) Dispense(name string) (interface{}, error) {
	p, ok := c.Plugins[name]
	if !ok {
		return nil, fmt.Errorf("unknown plugin type: %s", name)
	}

	return p.GRPCClient(c.Conn)
}

// ClientProtocol impl.
func (c *GRPCClient) Ping() error {
	// TODO
	return nil
}
