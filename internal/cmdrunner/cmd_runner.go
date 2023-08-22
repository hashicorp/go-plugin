// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cmdrunner

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin/runner"
)

var (
	_ runner.Runner = (*CmdRunner)(nil)

	// ErrProcessNotFound is returned when a client is instantiated to
	// reattach to an existing process and it isn't found.
	ErrProcessNotFound = errors.New("Reattachment process not found")
)

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

	addrTranslator
}

// NewCmdRunner returns an implementation of runner.Runner for running a plugin
// as a subprocess. It must be passed a cmd that hasn't yet been started.
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
	c.logger.Debug("plugin started", "path", c.path, "pid", c.pid)
	return nil
}

func (c *CmdRunner) Wait() error {
	return c.cmd.Wait()
}

func (c *CmdRunner) Kill() error {
	if c.cmd.Process != nil {
		err := c.cmd.Process.Kill()
		// Swallow ErrProcessDone, we support calling Kill multiple times.
		if !errors.Is(err, os.ErrProcessDone) {
			return err
		}
		return nil
	}

	return nil
}

func (c *CmdRunner) Stdout() io.ReadCloser {
	return c.stdout
}

func (c *CmdRunner) Stderr() io.ReadCloser {
	return c.stderr
}

func (c *CmdRunner) Name() string {
	return c.path
}

func (c *CmdRunner) ID() string {
	return fmt.Sprintf("%d", c.pid)
}
