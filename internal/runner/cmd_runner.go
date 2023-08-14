package runner

import (
	"fmt"
	"io"
	"os/exec"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin/runner"
)

var _ runner.Runner = (*CmdRunner)(nil)

// CmdRunner implements the Executor interface. It mostly just passes through
// to exec.Cmd methods.
type CmdRunner struct {
	logger hclog.Logger
	cmd    *exec.Cmd

	stdout io.ReadCloser
	stderr io.ReadCloser

	// Cmd info is persisted early, since the process information will be removed
	// after Kill is called.
	path string
	pid  int
}

// NewCmdRunner must be passed a cmd that hasn't yet been started.
func NewCmdRunner(logger hclog.Logger, cmd *exec.Cmd) (*CmdRunner, error) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	return &CmdRunner{
		logger: logger,
		cmd:    cmd,
		stdout: stdout,
		stderr: stderr,
		path:   cmd.Path,
	}, nil
}

func (c *CmdRunner) Start() error {
	c.logger.Debug("starting plugin", "path", c.cmd.Path, "args", c.cmd.Args)
	err := c.cmd.Start()
	if err != nil {
		return err
	}

	c.pid = c.cmd.Process.Pid
	c.logger.Debug("plugin started", "path", c.cmd.Path, "pid", c.cmd.Process.Pid)
	return nil
}

func (c *CmdRunner) Wait() error {
	return c.cmd.Wait()
}

func (c *CmdRunner) Kill() error {
	if c.cmd.Process != nil {
		return c.cmd.Process.Kill()
	}

	return nil
}

func (c *CmdRunner) Stdout() io.ReadCloser {
	return c.stdout
}

func (c *CmdRunner) Stderr() io.ReadCloser {
	return c.stderr
}

func (c *CmdRunner) PluginToHost(pluginNet, pluginAddr string) (string, string, error) {
	return pluginNet, pluginAddr, nil
}

func (c *CmdRunner) HostToPlugin(hostNet, hostAddr string) (string, string, error) {
	return hostNet, hostAddr, nil
}

func (c *CmdRunner) Name() string {
	return c.path
}

func (c *CmdRunner) ID() string {
	return fmt.Sprintf("%d", c.pid)
}
