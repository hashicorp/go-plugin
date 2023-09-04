//go:build js && wasm

package wasmrunner

import (
	"context"
	"fmt"
	"io"
	"os/exec"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin/runner"
	"github.com/magodo/go-wasmww"
)

const unrecognizedRemotePluginMessage = `This usually means
  the plugin was not compiled for WASM,
  the plugin failed to negotiate the initial go-plugin protocol handshake
%s`

var _ runner.Runner = (*WasmRunner)(nil)

type WasmRunner struct {
	logger hclog.Logger
	ww     *wasmww.WasmWebWorkerConn

	stdout io.ReadCloser
	stderr io.ReadCloser

	addrTranslator
}

func NewWasmRunner(logger hclog.Logger, cmd *exec.Cmd) (*WasmRunner, error) {
	ww := &wasmww.WasmWebWorkerConn{
		Path: cmd.Path,
		Args: cmd.Args,
		Env:  cmd.Env,
	}

	stdout, err := ww.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderr, err := ww.StderrPipe()
	if err != nil {
		return nil, err
	}

	return &WasmRunner{
		logger: logger,
		ww:     ww,
		stdout: stdout,
		stderr: stderr,
	}, nil
}

func (c *WasmRunner) Start(_ context.Context) error {
	c.logger.Debug("starting plugin", "path", c.ww.Path, "args", c.ww.Args)
	err := c.ww.Start()
	if err != nil {
		return err
	}

	c.logger.Debug("plugin started", "path", c.ww.Path, "name", c.ww.Name)
	return nil
}

func (c *WasmRunner) Wait(_ context.Context) error {
	c.ww.Wait()
	return nil
}

func (c *WasmRunner) Kill(_ context.Context) error {
	c.ww.Terminate()
	return nil
}

func (c *WasmRunner) Stdout() io.ReadCloser {
	return c.stdout
}

func (c *WasmRunner) Stderr() io.ReadCloser {
	return c.stderr
}

func (c *WasmRunner) Name() string {
	return c.ww.Path
}

func (c *WasmRunner) ID() string {
	return c.ww.Name
}

func (c *WasmRunner) Diagnose(ctx context.Context) string {
	return fmt.Sprintf(unrecognizedRemotePluginMessage, c.ww.Path)
}
