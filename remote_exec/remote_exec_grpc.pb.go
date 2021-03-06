// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package remote_exec

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// RemoteExecClient is the client API for RemoteExec service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type RemoteExecClient interface {
	Status(ctx context.Context, in *StatusRequest, opts ...grpc.CallOption) (*StatusResponse, error)
	StartCommand(ctx context.Context, in *StartCommandRequest, opts ...grpc.CallOption) (*CommandStatusResponse, error)
	StopCommand(ctx context.Context, in *StopCommandRequest, opts ...grpc.CallOption) (*StopCommandResponse, error)
	CommandStatus(ctx context.Context, in *CommandStatusRequest, opts ...grpc.CallOption) (*CommandStatusResponse, error)
	CommandOutput(ctx context.Context, in *CommandOutputRequest, opts ...grpc.CallOption) (RemoteExec_CommandOutputClient, error)
}

type remoteExecClient struct {
	cc grpc.ClientConnInterface
}

func NewRemoteExecClient(cc grpc.ClientConnInterface) RemoteExecClient {
	return &remoteExecClient{cc}
}

func (c *remoteExecClient) Status(ctx context.Context, in *StatusRequest, opts ...grpc.CallOption) (*StatusResponse, error) {
	out := new(StatusResponse)
	err := c.cc.Invoke(ctx, "/remote_exec.RemoteExec/Status", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *remoteExecClient) StartCommand(ctx context.Context, in *StartCommandRequest, opts ...grpc.CallOption) (*CommandStatusResponse, error) {
	out := new(CommandStatusResponse)
	err := c.cc.Invoke(ctx, "/remote_exec.RemoteExec/StartCommand", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *remoteExecClient) StopCommand(ctx context.Context, in *StopCommandRequest, opts ...grpc.CallOption) (*StopCommandResponse, error) {
	out := new(StopCommandResponse)
	err := c.cc.Invoke(ctx, "/remote_exec.RemoteExec/StopCommand", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *remoteExecClient) CommandStatus(ctx context.Context, in *CommandStatusRequest, opts ...grpc.CallOption) (*CommandStatusResponse, error) {
	out := new(CommandStatusResponse)
	err := c.cc.Invoke(ctx, "/remote_exec.RemoteExec/CommandStatus", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *remoteExecClient) CommandOutput(ctx context.Context, in *CommandOutputRequest, opts ...grpc.CallOption) (RemoteExec_CommandOutputClient, error) {
	stream, err := c.cc.NewStream(ctx, &RemoteExec_ServiceDesc.Streams[0], "/remote_exec.RemoteExec/CommandOutput", opts...)
	if err != nil {
		return nil, err
	}
	x := &remoteExecCommandOutputClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type RemoteExec_CommandOutputClient interface {
	Recv() (*CommandOutputBlock, error)
	grpc.ClientStream
}

type remoteExecCommandOutputClient struct {
	grpc.ClientStream
}

func (x *remoteExecCommandOutputClient) Recv() (*CommandOutputBlock, error) {
	m := new(CommandOutputBlock)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// RemoteExecServer is the server API for RemoteExec service.
// All implementations must embed UnimplementedRemoteExecServer
// for forward compatibility
type RemoteExecServer interface {
	Status(context.Context, *StatusRequest) (*StatusResponse, error)
	StartCommand(context.Context, *StartCommandRequest) (*CommandStatusResponse, error)
	StopCommand(context.Context, *StopCommandRequest) (*StopCommandResponse, error)
	CommandStatus(context.Context, *CommandStatusRequest) (*CommandStatusResponse, error)
	CommandOutput(*CommandOutputRequest, RemoteExec_CommandOutputServer) error
	mustEmbedUnimplementedRemoteExecServer()
}

// UnimplementedRemoteExecServer must be embedded to have forward compatible implementations.
type UnimplementedRemoteExecServer struct {
}

func (UnimplementedRemoteExecServer) Status(context.Context, *StatusRequest) (*StatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Status not implemented")
}
func (UnimplementedRemoteExecServer) StartCommand(context.Context, *StartCommandRequest) (*CommandStatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StartCommand not implemented")
}
func (UnimplementedRemoteExecServer) StopCommand(context.Context, *StopCommandRequest) (*StopCommandResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StopCommand not implemented")
}
func (UnimplementedRemoteExecServer) CommandStatus(context.Context, *CommandStatusRequest) (*CommandStatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CommandStatus not implemented")
}
func (UnimplementedRemoteExecServer) CommandOutput(*CommandOutputRequest, RemoteExec_CommandOutputServer) error {
	return status.Errorf(codes.Unimplemented, "method CommandOutput not implemented")
}
func (UnimplementedRemoteExecServer) mustEmbedUnimplementedRemoteExecServer() {}

// UnsafeRemoteExecServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to RemoteExecServer will
// result in compilation errors.
type UnsafeRemoteExecServer interface {
	mustEmbedUnimplementedRemoteExecServer()
}

func RegisterRemoteExecServer(s grpc.ServiceRegistrar, srv RemoteExecServer) {
	s.RegisterService(&RemoteExec_ServiceDesc, srv)
}

func _RemoteExec_Status_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RemoteExecServer).Status(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/remote_exec.RemoteExec/Status",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RemoteExecServer).Status(ctx, req.(*StatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RemoteExec_StartCommand_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StartCommandRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RemoteExecServer).StartCommand(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/remote_exec.RemoteExec/StartCommand",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RemoteExecServer).StartCommand(ctx, req.(*StartCommandRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RemoteExec_StopCommand_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StopCommandRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RemoteExecServer).StopCommand(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/remote_exec.RemoteExec/StopCommand",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RemoteExecServer).StopCommand(ctx, req.(*StopCommandRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RemoteExec_CommandStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CommandStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RemoteExecServer).CommandStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/remote_exec.RemoteExec/CommandStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RemoteExecServer).CommandStatus(ctx, req.(*CommandStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RemoteExec_CommandOutput_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(CommandOutputRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(RemoteExecServer).CommandOutput(m, &remoteExecCommandOutputServer{stream})
}

type RemoteExec_CommandOutputServer interface {
	Send(*CommandOutputBlock) error
	grpc.ServerStream
}

type remoteExecCommandOutputServer struct {
	grpc.ServerStream
}

func (x *remoteExecCommandOutputServer) Send(m *CommandOutputBlock) error {
	return x.ServerStream.SendMsg(m)
}

// RemoteExec_ServiceDesc is the grpc.ServiceDesc for RemoteExec service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var RemoteExec_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "remote_exec.RemoteExec",
	HandlerType: (*RemoteExecServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Status",
			Handler:    _RemoteExec_Status_Handler,
		},
		{
			MethodName: "StartCommand",
			Handler:    _RemoteExec_StartCommand_Handler,
		},
		{
			MethodName: "StopCommand",
			Handler:    _RemoteExec_StopCommand_Handler,
		},
		{
			MethodName: "CommandStatus",
			Handler:    _RemoteExec_CommandStatus_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "CommandOutput",
			Handler:       _RemoteExec_CommandOutput_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "remote_exec/remote_exec.proto",
}
