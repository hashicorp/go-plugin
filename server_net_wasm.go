//go:build js && wasm

package plugin

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"github.com/hack-pad/safejs"
	"github.com/magodo/go-wasmww"
	"github.com/magodo/go-webworkers/types"
)

var _ net.Listener = &WebWorkerListener{}

// WebWorkerListener implements the net.Listener
type WebWorkerListener struct {
	self *wasmww.SelfSharedConn
	ch   <-chan *wasmww.SelfSharedConnPort
}

func NewWebWorkerListener() (net.Listener, error) {
	self, err := wasmww.NewSelfSharedConn()
	if err != nil {
		return nil, err
	}
	ch, err := self.SetupConn()
	if err != nil {
		return nil, err
	}
	return &WebWorkerListener{
		self: self,
		ch:   ch,
	}, nil
}

func (l *WebWorkerListener) Accept() (net.Conn, error) {
	port, ok := <-l.ch
	if !ok {
		return nil, net.ErrClosed
	}

	name, err := l.self.Name()
	if err != nil {
		return nil, err
	}
	location, err := l.self.Location()
	if err != nil {
		return nil, err
	}

	ch, err := port.SetupConn()
	if err != nil {
		return nil, err
	}

	return NewWebWorkerConnForServer(name, location.Href, ch, port.PostMessage, port.Close), nil
}

func (l *WebWorkerListener) Addr() net.Addr {
	var name string
	if v, err := l.self.Name(); err == nil {
		name = v
	}
	var url string
	if v, err := l.self.Location(); err == nil {
		url = v.Href
	}
	return WebWorkerAddr{Name: name, URL: url}
}

func (l *WebWorkerListener) Close() error {
	return l.self.Close()
}

// WebWorkerAddr implements the net.Addr
type WebWorkerAddr struct {
	Name string
	URL  string
}

var _ net.Addr = WebWorkerAddr{}

func ParseWebWorkerAddr(addr string) (*WebWorkerAddr, error) {
	name, url, ok := strings.Cut(addr, ":")
	if !ok {
		return nil, fmt.Errorf("malformed address: %s", addr)
	}
	return &WebWorkerAddr{
		Name: name,
		URL:  url,
	}, nil
}

func (WebWorkerAddr) Network() string {
	return "webworker"
}

func (addr WebWorkerAddr) String() string {
	return fmt.Sprintf("%s:%s", addr.Name, addr.URL)
}

// WebWorkerConn implements the net.Conn
type WebWorkerConn struct {
	localAddr  net.Addr
	remoteAddr net.Addr
	ch         <-chan types.MessageEventMessage
	closeFunc  connCloseFunc
	postFunc   connPostMessageFunc

	timerR  *time.Timer
	timerW  *time.Timer
	readBuf bytes.Buffer
}

type connPostMessageFunc func(message safejs.Value, transfers []safejs.Value) error
type connCloseFunc func() error

var _ net.Conn = &WebWorkerConn{}

func NewWebWorkerConnForServer(name, url string, ch <-chan types.MessageEventMessage, postFunc connPostMessageFunc, closeFunc connCloseFunc) *WebWorkerConn {
	return &WebWorkerConn{
		localAddr:  WebWorkerAddr{Name: name, URL: url},
		remoteAddr: WebWorkerAddr{Name: "outside"},
		ch:         ch,
		postFunc:   postFunc,
		closeFunc:  closeFunc,
	}
}

func NewWebWorkerConnForClient(name, url string, ch <-chan types.MessageEventMessage, postFunc connPostMessageFunc, closeFunc connCloseFunc) *WebWorkerConn {
	return &WebWorkerConn{
		localAddr:  WebWorkerAddr{Name: "outside"},
		remoteAddr: WebWorkerAddr{Name: name, URL: url},
		ch:         ch,
		postFunc:   postFunc,
		closeFunc:  closeFunc,
	}
}

func (conn *WebWorkerConn) Close() error {
	return conn.closeFunc()
}

func (conn *WebWorkerConn) LocalAddr() net.Addr {
	return conn.localAddr
}

func (conn *WebWorkerConn) Read(b []byte) (n int, err error) {
	var (
		event types.MessageEventMessage
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

func (conn *WebWorkerConn) RemoteAddr() net.Addr {
	return conn.remoteAddr
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
