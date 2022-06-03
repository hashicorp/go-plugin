package plugin

import (
	"io"
	"log"
	"os"
	"syscall"
)

func copyStream(name string, dst io.Writer, src io.Reader) {
	if src == nil {
		panic(name + ": src is nil")
	}
	if dst == nil {
		panic(name + ": dst is nil")
	}
	_, err := io.Copy(dst, src)
	if err != nil && err != io.EOF {
		// Linux kernel return EIO when attempting to read from a master pseudo
		// terminal which no longer has an open slave. So ignore error here.
		// See https://github.com/creack/pty/issues/21
		if pathErr, ok := err.(*os.PathError); ok && pathErr.Err == syscall.EIO {
			return
		}

		log.Printf("[ERR] plugin: stream copy '%s' error: %s", name, err)
	}
}
