package plugin

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type streamer interface {
	Send(*ConnInfo) error
	Recv() (*ConnInfo, error)
}

type sendErr struct {
	i  *ConnInfo
	ch chan error
}

type gRPCBrokerServer struct {
	send chan *sendErr
	recv chan *ConnInfo
	quit chan struct{}
}

func (s *gRPCBrokerServer) NewConn(stream GRPCBroker_NewConnServer) error {
	doneCh := stream.Context().Done()

	go func() {
		for {
			select {
			case <-doneCh:
				return
			case <-s.quit:
				return
			case s := <-s.send:
				err := stream.Send(s.i)
				s.ch <- err
			}
		}
	}()

	for {
		i, err := stream.Recv()
		if err != nil {
			return err
		}
		s.recv <- i
	}

	return nil
}

func (s *gRPCBrokerServer) Send(i *ConnInfo) error {
	ch := make(chan error)
	defer close(ch)

	s.send <- &sendErr{
		i:  i,
		ch: ch,
	}

	return <-ch
}

func (s *gRPCBrokerServer) Recv() (*ConnInfo, error) {
	return <-s.recv, nil
}

type gRPCBrokerClientImpl struct {
	client GRPCBrokerClient
	send   chan *sendErr
	recv   chan *ConnInfo
	quit   chan struct{}
}

func (s *gRPCBrokerClientImpl) NewConn() error {
	ctx, cancelFunc := context.WithCancel(context.Background())

	stream, err := s.client.NewConn(ctx)
	if err != nil {
		return err
	}
	doneCh := stream.Context().Done()

	go func() {
		defer cancelFunc()

		for {
			select {
			case <-doneCh:
				return
			case <-s.quit:
				return
			case s := <-s.send:
				err := stream.Send(s.i)
				s.ch <- err
			}
		}
	}()

	for {
		i, err := stream.Recv()
		if err != nil {
			return err
		}
		s.recv <- i
	}

	return nil
}

func (s *gRPCBrokerClientImpl) Send(i *ConnInfo) error {
	ch := make(chan error)
	defer close(ch)

	s.send <- &sendErr{
		i:  i,
		ch: ch,
	}

	return <-ch
}

func (s *gRPCBrokerClientImpl) Recv() (*ConnInfo, error) {
	return <-s.recv, nil
}

// MuxBroker is responsible for brokering multiplexed connections by unique ID.
//
// It is used by plugins to multiplex multiple RPC connections and data
// streams on top of a single connection between the plugin process and the
// host process.
//
// This allows a plugin to request a channel with a specific ID to connect to
// or accept a connection from, and the broker handles the details of
// holding these channels open while they're being negotiated.
//
// The Plugin interface has access to these for both Server and Client.
// The broker can be used by either (optionally) to reserve and connect to
// new multiplexed streams. This is useful for complex args and return values,
// or anything else you might need a data stream for.
type GRPCBroker struct {
	nextId   uint32
	streamer streamer
	streams  map[uint32]*gRPCBrokerPending
	TLS      *tls.Config

	sync.Mutex
}

type gRPCBrokerPending struct {
	ch     chan *ConnInfo
	doneCh chan struct{}
}

func newGRPCBroker(s streamer) *GRPCBroker {
	return &GRPCBroker{
		streamer: s,
		streams:  make(map[uint32]*gRPCBrokerPending),
	}
}

// Accept accepts a connection by ID.
//
// This should not be called multiple times with the same ID at one time.
func (b *GRPCBroker) Accept(id uint32) (net.Listener, error) {
	listener, err := serverListener()
	if err != nil {
		return nil, err
	}

	// TODO: time this out?
	err = b.streamer.Send(&ConnInfo{
		ServiceId: id,
		Network:   listener.Addr().Network(),
		Address:   listener.Addr().String(),
	})
	if err != nil {
		return nil, err
	}

	return listener, nil
}

// AcceptAndServe is used to accept a specific stream ID and immediately
// serve an RPC server on that stream ID. This is used to easily serve
// complex arguments.
//
// The served interface is always registered to the "Plugin" name.
func (b *GRPCBroker) AcceptAndServe(id uint32, s func([]grpc.ServerOption) *grpc.Server) {
	listener, err := b.Accept(id)
	if err != nil {
		log.Printf("[ERR] plugin: plugin acceptAndServe error: %s", err)
		return
	}
	defer listener.Close()

	var opts []grpc.ServerOption
	if b.TLS != nil {
		opts = []grpc.ServerOption{grpc.Creds(credentials.NewTLS(b.TLS))}
	}

	s(opts).Serve(listener)
}

// Close closes the connection and all sub-connections.
func (b *GRPCBroker) Close() error {
	return nil
}

// Dial opens a connection by ID.
func (b *GRPCBroker) Dial(id uint32) (conn *grpc.ClientConn, err error) {
	var c *ConnInfo

	// Open the stream
	p := b.getStream(id)
	select {
	case c = <-p.ch:
		close(p.doneCh)
	case <-time.After(5 * time.Second):
		b.Lock()
		defer b.Unlock()
		delete(b.streams, id)

		return nil, fmt.Errorf("timeout waiting for accept")
	}

	var addr net.Addr
	switch c.Network {
	case "tcp":
		addr, err = net.ResolveTCPAddr("tcp", c.Address)
	case "unix":
		addr, err = net.ResolveUnixAddr("unix", c.Address)
	default:
		err = fmt.Errorf("Unknown address type: %s", c.Address)
	}
	if err != nil {
		return nil, err
	}

	return dialGRPCConn(b.TLS, netAddrDialer(addr))
}

// NextId returns a unique ID to use next.
//
// It is possible for very long-running plugin hosts to wrap this value,
// though it would require a very large amount of RPC calls. In practice
// we've never seen it happen.
func (m *GRPCBroker) NextId() uint32 {
	return atomic.AddUint32(&m.nextId, 1)
}

// Run starts the brokering and should be executed in a goroutine, since it
// blocks forever, or until the session closes.
//
// Uses of MuxBroker never need to call this. It is called internally by
// the plugin host/client.
func (m *GRPCBroker) Run() {
	for {
		stream, err := m.streamer.Recv()
		if err != nil {
			// Once we receive an error, just exit
			break
		}

		// Initialize the waiter
		p := m.getStream(stream.ServiceId)
		select {
		case p.ch <- stream:
		default:
		}
	}
}

func (m *GRPCBroker) getStream(id uint32) *gRPCBrokerPending {
	m.Lock()
	defer m.Unlock()

	p, ok := m.streams[id]
	if ok {
		return p
	}

	m.streams[id] = &gRPCBrokerPending{
		ch:     make(chan *ConnInfo, 1),
		doneCh: make(chan struct{}),
	}
	return m.streams[id]
}
