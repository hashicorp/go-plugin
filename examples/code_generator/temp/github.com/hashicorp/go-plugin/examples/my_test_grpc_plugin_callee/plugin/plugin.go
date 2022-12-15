package plugin

import (
	"context"
	"fmt"

	pb "github.com/hashicorp/go-plugin/examples/my_test_grpc_plugin"
)

type PluginService struct{}

// Run Implement the interface of grpc
// don't use error to return any info to caller
func (p *PluginService) Run(ctx context.Context, req *pb.Request) (*pb.Response, error) {
	// todo: add logic here
	return &pb.Response{
		Code: 200,
		Msg:  "req is " + fmt.Sprintf("%+v", req),
	}, nil
}
