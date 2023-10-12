// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package grpctest

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// TestClient is the client API for Test service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TestClient interface {
	Double(ctx context.Context, in *TestRequest, opts ...grpc.CallOption) (*TestResponse, error)
	PrintKV(ctx context.Context, in *PrintKVRequest, opts ...grpc.CallOption) (*PrintKVResponse, error)
	Bidirectional(ctx context.Context, in *BidirectionalRequest, opts ...grpc.CallOption) (*BidirectionalResponse, error)
	Stream(ctx context.Context, opts ...grpc.CallOption) (Test_StreamClient, error)
	PrintStdio(ctx context.Context, in *PrintStdioRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type testClient struct {
	cc grpc.ClientConnInterface
}

func NewTestClient(cc grpc.ClientConnInterface) TestClient {
	return &testClient{cc}
}

func (c *testClient) Double(ctx context.Context, in *TestRequest, opts ...grpc.CallOption) (*TestResponse, error) {
	out := new(TestResponse)
	err := c.cc.Invoke(ctx, "/grpctest.Test/Double", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *testClient) PrintKV(ctx context.Context, in *PrintKVRequest, opts ...grpc.CallOption) (*PrintKVResponse, error) {
	out := new(PrintKVResponse)
	err := c.cc.Invoke(ctx, "/grpctest.Test/PrintKV", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *testClient) Bidirectional(ctx context.Context, in *BidirectionalRequest, opts ...grpc.CallOption) (*BidirectionalResponse, error) {
	out := new(BidirectionalResponse)
	err := c.cc.Invoke(ctx, "/grpctest.Test/Bidirectional", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *testClient) Stream(ctx context.Context, opts ...grpc.CallOption) (Test_StreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &Test_ServiceDesc.Streams[0], "/grpctest.Test/Stream", opts...)
	if err != nil {
		return nil, err
	}
	x := &testStreamClient{stream}
	return x, nil
}

type Test_StreamClient interface {
	Send(*TestRequest) error
	Recv() (*TestResponse, error)
	grpc.ClientStream
}

type testStreamClient struct {
	grpc.ClientStream
}

func (x *testStreamClient) Send(m *TestRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *testStreamClient) Recv() (*TestResponse, error) {
	m := new(TestResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *testClient) PrintStdio(ctx context.Context, in *PrintStdioRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/grpctest.Test/PrintStdio", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TestServer is the server API for Test service.
// All implementations must embed UnimplementedTestServer
// for forward compatibility
type TestServer interface {
	Double(context.Context, *TestRequest) (*TestResponse, error)
	PrintKV(context.Context, *PrintKVRequest) (*PrintKVResponse, error)
	Bidirectional(context.Context, *BidirectionalRequest) (*BidirectionalResponse, error)
	Stream(Test_StreamServer) error
	PrintStdio(context.Context, *PrintStdioRequest) (*emptypb.Empty, error)
	mustEmbedUnimplementedTestServer()
}

// UnimplementedTestServer must be embedded to have forward compatible implementations.
type UnimplementedTestServer struct {
}

func (UnimplementedTestServer) Double(context.Context, *TestRequest) (*TestResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Double not implemented")
}
func (UnimplementedTestServer) PrintKV(context.Context, *PrintKVRequest) (*PrintKVResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PrintKV not implemented")
}
func (UnimplementedTestServer) Bidirectional(context.Context, *BidirectionalRequest) (*BidirectionalResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Bidirectional not implemented")
}
func (UnimplementedTestServer) Stream(Test_StreamServer) error {
	return status.Errorf(codes.Unimplemented, "method Stream not implemented")
}
func (UnimplementedTestServer) PrintStdio(context.Context, *PrintStdioRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PrintStdio not implemented")
}
func (UnimplementedTestServer) mustEmbedUnimplementedTestServer() {}

// UnsafeTestServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TestServer will
// result in compilation errors.
type UnsafeTestServer interface {
	mustEmbedUnimplementedTestServer()
}

func RegisterTestServer(s grpc.ServiceRegistrar, srv TestServer) {
	s.RegisterService(&Test_ServiceDesc, srv)
}

func _Test_Double_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TestRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TestServer).Double(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpctest.Test/Double",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TestServer).Double(ctx, req.(*TestRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Test_PrintKV_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PrintKVRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TestServer).PrintKV(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpctest.Test/PrintKV",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TestServer).PrintKV(ctx, req.(*PrintKVRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Test_Bidirectional_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BidirectionalRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TestServer).Bidirectional(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpctest.Test/Bidirectional",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TestServer).Bidirectional(ctx, req.(*BidirectionalRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Test_Stream_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(TestServer).Stream(&testStreamServer{stream})
}

type Test_StreamServer interface {
	Send(*TestResponse) error
	Recv() (*TestRequest, error)
	grpc.ServerStream
}

type testStreamServer struct {
	grpc.ServerStream
}

func (x *testStreamServer) Send(m *TestResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *testStreamServer) Recv() (*TestRequest, error) {
	m := new(TestRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Test_PrintStdio_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PrintStdioRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TestServer).PrintStdio(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpctest.Test/PrintStdio",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TestServer).PrintStdio(ctx, req.(*PrintStdioRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Test_ServiceDesc is the grpc.ServiceDesc for Test service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Test_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "grpctest.Test",
	HandlerType: (*TestServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Double",
			Handler:    _Test_Double_Handler,
		},
		{
			MethodName: "PrintKV",
			Handler:    _Test_PrintKV_Handler,
		},
		{
			MethodName: "Bidirectional",
			Handler:    _Test_Bidirectional_Handler,
		},
		{
			MethodName: "PrintStdio",
			Handler:    _Test_PrintStdio_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Stream",
			Handler:       _Test_Stream_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "test/grpc/test.proto",
}

// PingPongClient is the client API for PingPong service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type PingPongClient interface {
	Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PongResponse, error)
}

type pingPongClient struct {
	cc grpc.ClientConnInterface
}

func NewPingPongClient(cc grpc.ClientConnInterface) PingPongClient {
	return &pingPongClient{cc}
}

func (c *pingPongClient) Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PongResponse, error) {
	out := new(PongResponse)
	err := c.cc.Invoke(ctx, "/grpctest.PingPong/Ping", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PingPongServer is the server API for PingPong service.
// All implementations must embed UnimplementedPingPongServer
// for forward compatibility
type PingPongServer interface {
	Ping(context.Context, *PingRequest) (*PongResponse, error)
	mustEmbedUnimplementedPingPongServer()
}

// UnimplementedPingPongServer must be embedded to have forward compatible implementations.
type UnimplementedPingPongServer struct {
}

func (UnimplementedPingPongServer) Ping(context.Context, *PingRequest) (*PongResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}
func (UnimplementedPingPongServer) mustEmbedUnimplementedPingPongServer() {}

// UnsafePingPongServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to PingPongServer will
// result in compilation errors.
type UnsafePingPongServer interface {
	mustEmbedUnimplementedPingPongServer()
}

func RegisterPingPongServer(s grpc.ServiceRegistrar, srv PingPongServer) {
	s.RegisterService(&PingPong_ServiceDesc, srv)
}

func _PingPong_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PingPongServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpctest.PingPong/Ping",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PingPongServer).Ping(ctx, req.(*PingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// PingPong_ServiceDesc is the grpc.ServiceDesc for PingPong service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var PingPong_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "grpctest.PingPong",
	HandlerType: (*PingPongServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Ping",
			Handler:    _PingPong_Ping_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "test/grpc/test.proto",
}
