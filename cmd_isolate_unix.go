package plugin

import (
	"os/exec"
	"syscall"
)

// isolateCmd sets the setid for the process
func (c *Client) isolateCmd(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.Setsid = true
}
