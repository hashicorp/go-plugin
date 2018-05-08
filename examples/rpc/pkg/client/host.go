package client

import (
	hclog "github.com/hashicorp/go-hclog"
	"github.com/sampaioletti/go-plugin/examples/api"
)

var _ api.Host = (*Host)(nil)

type Host struct {
}

func (p *Host) HelloHost(hello string) error {
	hclog.L().Info(hello)
	return nil
}
