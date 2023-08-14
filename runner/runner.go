package runner

import (
	"io"
)

type Runner interface {
	Start() error
	Wait() error
	Kill() error
	Stdout() io.ReadCloser
	Stderr() io.ReadCloser
	Name() string
	ID() string
	AddrTranslator
}

type AddrTranslator interface {
	PluginToHost(pluginNet, pluginAddr string) (hostNet string, hostAddr string, err error)
	HostToPlugin(hostNet, hostAddr string) (pluginNet string, pluginAddr string, err error)
}
