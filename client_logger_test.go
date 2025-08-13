// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package plugin

import (
	"io"
	"sync"
	"testing"

	"github.com/hashicorp/go-hclog"
)

func BenchmarkClientLogging(b *testing.B) {
	// We're not actually going to start the process in this benchmark,
	// so this is just a placeholder to satisfy the ClientConfig.
	process := helperProcess("bad-version")

	tests := map[string]hclog.Level{
		"off":   hclog.Off,
		"error": hclog.Error,
		"trace": hclog.Trace,
	}

	for name, logLevel := range tests {
		b.Run(name, func(b *testing.B) {
			logger := hclog.New(&hclog.LoggerOptions{
				Name:   "test-logger",
				Level:  logLevel,
				Output: io.Discard,
				Mutex:  new(sync.Mutex),
			})

			c := NewClient(&ClientConfig{
				Cmd:             process,
				Stderr:          io.Discard,
				HandshakeConfig: testHandshake,
				Logger:          logger,
				Plugins:         testPluginMap,
			})

			r, w := io.Pipe()

			c.clientWaitGroup.Add(1)
			c.pipesWaitGroup.Add(1)
			// logStderr calls Done() on both waitgroups
			go c.logStderr("test", r)

			fakeLogLine := []byte("{\"@level\":\"debug\",\"@timestamp\":\"2006-01-02T15:04:05.000000Z\",\"@message\":\"hello\",\"extra\":\"hi\",\"numbers\":[1,2,3]}\n")
			for b.Loop() {
				n, err := w.Write(fakeLogLine)
				if err != nil || n != len(fakeLogLine) {
					b.Fatal("failed to write to pipe")
				}
			}
			err := w.Close() // causes the c.logStderr goroutine to exit
			if err != nil {
				b.Fatalf("failed to close write end of pipe: %s", err)
			}
		})
	}
}
