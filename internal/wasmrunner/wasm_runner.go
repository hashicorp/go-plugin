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
	logger   hclog.Logger
	wwConn   *wasmww.WasmSharedWebWorkerConn
	mgmtConn *wasmww.WasmSharedWebWorkerMgmtConn

	stdout io.ReadCloser
	stderr io.ReadCloser

	addrTranslator
}

func NewWasmRunner(logger hclog.Logger, cmd *exec.Cmd) (*WasmRunner, error) {
	ww := &wasmww.WasmSharedWebWorkerConn{
		Path: cmd.Path,
		Args: cmd.Args,
		Env:  cmd.Env,
	}

	return &WasmRunner{
		logger: logger,
		wwConn: ww,
	}, nil
}

func (c *WasmRunner) Start(_ context.Context) error {
	c.logger.Debug("starting plugin", "path", c.wwConn.Path, "args", c.wwConn.Args)
	mgmtConn, err := c.wwConn.Start()
	if err != nil {
		return err
	}
	c.mgmtConn = mgmtConn

	c.logger.Debug("plugin started", "path", c.wwConn.Path, "name", c.wwConn.Name)
	return nil
}

func (c *WasmRunner) Wait(_ context.Context) error {
	c.wwConn.Wait()
	return nil
}

func (c *WasmRunner) Kill(_ context.Context) error {
	return c.wwConn.Close()
}

func (c *WasmRunner) Stdout() io.ReadCloser {
	return c.mgmtConn.Stdout()
}

func (c *WasmRunner) Stderr() io.ReadCloser {
	return c.mgmtConn.Stderr()
}

func (c *WasmRunner) Name() string {
	return c.wwConn.Path
}

func (c *WasmRunner) ID() string {
	return c.wwConn.Name
}

func (c *WasmRunner) Diagnose(ctx context.Context) string {
	return fmt.Sprintf(unrecognizedRemotePluginMessage, c.wwConn.Path)
}

func (c *WasmRunner) WebWorker() *wasmww.WasmSharedWebWorkerConn {
	return c.wwConn
}
