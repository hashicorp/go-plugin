package server

import (
	hclog "github.com/hashicorp/go-hclog"
	"gitlab.com/indis/libs/extension/api"
)

var _ api.Extender = (*Extension)(nil)

type Extension struct {
}

func (p *Extension) HelloExtension(hello string) error {
	hclog.L().Info(hello)
	return nil
}
