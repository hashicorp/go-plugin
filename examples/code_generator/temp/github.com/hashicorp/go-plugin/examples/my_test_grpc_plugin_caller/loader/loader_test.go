package loader

import (
	"context"
	"testing"
	"time"

	pb "github.com/hashicorp/go-plugin/examples/my_test_grpc_plugin"
)

func TestProcess_Run(t *testing.T) {
	SetPluginUnloadSeconds(3)
	loader := NewLoader()
	const pluinPath = "../../my_test_grpc_plugin_callee/my_test_grpc_plugin_callee"
	p, err := loader.Load("my_test_grpc_plugins", pluinPath)
	if err != nil {
		t.Errorf("load error, err=%+v", err)
		return
	}
	rsp, err := p.Run(context.Background(), &pb.Request{
		Field1: 1,
		Field2: 2,
		Field3: 3,
		Field4: "4",
		Field5: []byte("5"),
		Field6: []string{"6"},
	})
	if err != nil {
		t.Errorf("run error, err=%+v", err)
		return
	}
	t.Logf("rsp=%+v", rsp)
	{
		p1, err1 := loader.Load("my_test_grpc_plugins", "../my_test_grpc_plugin_callee")
		if err1 != nil {
			t.Errorf("load error, err=%+v", err1)
			return
		}
		rsp1, err1 := p1.Run(context.Background(), &pb.Request{
			Field1: 11,
			Field2: 22,
			Field3: 33,
			Field4: "44",
			Field5: []byte("55"),
			Field6: []string{"66"},
		})
		if err1 != nil {
			t.Errorf("run error, err=%+v", err1)
			return
		}
		t.Logf("rsp=%+v", rsp1)
	}
	time.Sleep(time.Duration(4) * time.Second)
	{
		p2, err2 := loader.Load("my_test_grpc_plugins", pluinPath)
		if err2 != nil {
			t.Errorf("load error, err=%+v", err2)
			return
		}
		rsp2, err2 := p2.Run(context.Background(), &pb.Request{
			Field1: 111,
			Field2: 222,
			Field3: 333,
			Field4: "444",
			Field5: []byte("555"),
			Field6: []string{"666"},
		})
		if err2 != nil {
			t.Errorf("run error, err=%+v", err2)
			return
		}
		t.Logf("rsp=%+v", rsp2)
	}
}
