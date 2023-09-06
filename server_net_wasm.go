//go:build js && wasm

package plugin

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/hack-pad/go-webworkers/worker"
	"github.com/hack-pad/safejs"
	"github.com/magodo/go-wasmww"
)

var _ net.Listener = &WebWorkerListener{}

// WebWorkerListener implements the net.Listener
type WebWorkerListener struct {
	self    *wasmww.GlobalSelfConn
	ch      <-chan worker.MessageEvent
	closeFn wasmww.WebWorkerCloseFunc

	// acceptCh is a 1 buffered channel, which only allow the 1st receive.
	// Currently, the web worker is only a dedicated one, which means
	// only one client can connect to this web worker at one point.
	acceptCh chan any
}

func NewWebWorkerListener() (net.Listener, error) {
	self, err := wasmww.SelfConn()
	if err != nil {
		return nil, err
	}
	ch, closeFn, err := self.SetupConn()
	if err != nil {
		return nil, err
	}
	acceptCh := make(chan any, 1)
	acceptCh <- struct{}{}
	return &WebWorkerListener{
		self:     self,
		ch:       ch,
		closeFn:  closeFn,
		acceptCh: acceptCh,
	}, nil
}

func (l *WebWorkerListener) Accept() (net.Conn, error) {
	_, ok := <-l.acceptCh
	if !ok {
		return nil, net.ErrClosed
	}

	var name string
	if v, err := l.self.Name(); err == nil {
		name = v
	}
	return NewWebWorkerConnForServer(name, l.ch, l.self.PostMessage, l.acceptCh), nil
}

func (l *WebWorkerListener) Addr() net.Addr {
	var name string
	if v, err := l.self.Name(); err == nil {
		name = v
	}
	return WebWorkerAddr{Name: name}
}

func (l *WebWorkerListener) Close() error {
	return l.closeFn()
}

// WebWorkerAddr implements the net.Addr
type WebWorkerAddr struct{ Name string }

var _ net.Addr = WebWorkerAddr{}

func (WebWorkerAddr) Network() string {
	return "webworker"
}

func (addr WebWorkerAddr) String() string {
	return addr.Name
}

// WebWorkerConn implements the net.Conn
type WebWorkerConn struct {
	name     string
	ch       <-chan worker.MessageEvent
	timerR   *time.Timer
	timerW   *time.Timer
	postFunc postMessageFunc
	readBuf  bytes.Buffer

	// server only
	acceptCh chan any
}

type postMessageFunc func(message safejs.Value, transfers []safejs.Value) error

var _ net.Conn = &WebWorkerConn{}

func NewWebWorkerConnForServer(name string, ch <-chan worker.MessageEvent, postFunc postMessageFunc, acceptCh chan any) *WebWorkerConn {
	return &WebWorkerConn{
		name:     name,
		ch:       ch,
		postFunc: postFunc,
		acceptCh: acceptCh,
	}
}

func NewWebWorkerConnForClient(name string, ch <-chan worker.MessageEvent, postFunc postMessageFunc) *WebWorkerConn {
	return &WebWorkerConn{
		name:     name,
		ch:       ch,
		postFunc: postFunc,
	}
}

func (conn *WebWorkerConn) Close() error {
	if conn.acceptCh != nil {
		conn.acceptCh <- struct{}{}
	}
	return nil
}

func (conn *WebWorkerConn) LocalAddr() net.Addr {
	return WebWorkerAddr{Name: conn.name}
}

func (conn *WebWorkerConn) Read(b []byte) (n int, err error) {
	var (
		event worker.MessageEvent
		ok    bool
	)
	if timeout := conn.timerR; timeout != nil {
		select {
		case <-timeout.C:
			return 0, os.ErrDeadlineExceeded
		default:
		}
	}
	// If there is unread bytes in the buffer, just read them out
	if conn.readBuf.Len() != 0 {
		return io.ReadAtLeast(&conn.readBuf, b, min(len(b), conn.readBuf.Len()))
	}

	if timeout := conn.timerR; timeout != nil {
		select {
		case event, ok = <-conn.ch:
		case <-timeout.C:
			return 0, os.ErrDeadlineExceeded
		}
	} else {
		event, ok = <-conn.ch
	}

	if !ok {
		// Channel closed
		return 0, io.EOF
	}

	data, err := event.Data()
	if err != nil {
		return 0, err
	}
	arrayBufLen, err := data.Length()
	if err != nil {
		return 0, err
	}
	buf := make([]byte, arrayBufLen)
	n, err = safejs.CopyBytesToGo(buf, data)
	if err != nil {
		return 0, err
	}
	if n != arrayBufLen {
		return 0, fmt.Errorf("CopyBytesToGo expect to copy %d bytes, actually %d bytes", arrayBufLen, n)
	}
	n = copy(b, buf)

	// If there are left bytes not read to the target buffer, store them in the readBuf.
	if n < len(buf) {
		if _, err := conn.readBuf.Write(buf[n:]); err != nil {
			return 0, err
		}
	}
	return n, nil
}

func (conn *WebWorkerConn) Write(b []byte) (n int, err error) {
	arraybuf, err := safejs.MustGetGlobal("Uint8Array").New(len(b))
	if err != nil {
		return 0, nil
	}
	n, err = safejs.CopyBytesToJS(arraybuf, b)
	if err != nil {
		return 0, nil
	}
	if n != len(b) {
		return 0, fmt.Errorf("CopyBytesToJS expect to copy %d bytes, actually %d bytes", len(b), n)
	}
	if err := conn.postFunc(arraybuf, nil); err != nil {
		return 0, err
	}
	return len(b), nil
}

func (*WebWorkerConn) RemoteAddr() net.Addr {
	return WebWorkerAddr{Name: "remote"}
}

func (conn *WebWorkerConn) SetDeadline(t time.Time) error {
	if err := conn.SetReadDeadline(t); err != nil {
		return err
	}
	if err := conn.SetWriteDeadline(t); err != nil {
		return err
	}
	return nil
}

func (conn *WebWorkerConn) SetReadDeadline(t time.Time) error {
	conn.timerR = time.NewTimer(time.Until(t))
	return nil
}

func (conn *WebWorkerConn) SetWriteDeadline(t time.Time) error {
	conn.timerW = time.NewTimer(time.Until(t))
	return nil
}
