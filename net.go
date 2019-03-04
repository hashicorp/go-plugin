package plugin

import (
	"context"
	"net"
	"time"
)

type connWithCancel struct {
	conn net.Conn
	context.CancelFunc
}

func (wrapper *connWithCancel) Read(p []byte) (int, error) {
	return wrapper.conn.Read(p)
}

func (wrapper *connWithCancel) Write(p []byte) (int, error) {
	return wrapper.conn.Write(p)
}

func (wrapper *connWithCancel) Close() error {
	err := wrapper.conn.Close()
	wrapper.CancelFunc()
	return err
}

func (wrapper *connWithCancel) LocalAddr() net.Addr {
	return wrapper.conn.LocalAddr()
}

func (wrapper *connWithCancel) RemoteAddr() net.Addr {
	return wrapper.conn.RemoteAddr()
}
func (wrapper *connWithCancel) SetDeadline(t time.Time) error {
	return wrapper.conn.SetDeadline(t)
}

func (wrapper *connWithCancel) SetReadDeadline(t time.Time) error {
	return wrapper.conn.SetReadDeadline(t)
}

func (wrapper *connWithCancel) SetWriteDeadline(t time.Time) error {
	return wrapper.conn.SetWriteDeadline(t)
}

// NewConnWithCancel returns a net.Conn that signals the specified
// context.CancelFunc when the net.Conn object is closed.
func NewConnWithCancel(
	conn net.Conn,
	cancelFunc context.CancelFunc) (
	wrapper net.Conn) {

	wrapper = &connWithCancel{
		conn:       conn,
		CancelFunc: cancelFunc,
	}

	return
}
