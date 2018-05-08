package server

import (
	hclog "github.com/hashicorp/go-hclog"
	"github.com/sampaioletti/go-plugin/examples/api"
)

var _ api.Extender = (*Extension)(nil)

type Extension struct {
}

func (p *Extension) HelloExtension(hello string) error {
	hclog.L().Info(hello)
	return nil
}
