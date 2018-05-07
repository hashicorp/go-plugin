package client

import (
	hclog "github.com/hashicorp/go-hclog"
	"gitlab.com/indis/libs/extension/api"
)

var _ api.Host = (*Host)(nil)

type Host struct {
}

func (p *Host) HelloHost(hello string) error {
	hclog.L().Info(hello)
	return nil
}
