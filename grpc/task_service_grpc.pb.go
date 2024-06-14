// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.20.1
// source: services/task_service.proto

package grpc

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

// TaskServiceClient is the client API for TaskService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TaskServiceClient interface {
	Subscribe(ctx context.Context, opts ...grpc.CallOption) (TaskService_SubscribeClient, error)
	Fetch(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error)
	SendNotification(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error)
}

type taskServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewTaskServiceClient(cc grpc.ClientConnInterface) TaskServiceClient {
	return &taskServiceClient{cc}
}

func (c *taskServiceClient) Subscribe(ctx context.Context, opts ...grpc.CallOption) (TaskService_SubscribeClient, error) {
	stream, err := c.cc.NewStream(ctx, &TaskService_ServiceDesc.Streams[0], "/grpc.TaskService/Subscribe", opts...)
	if err != nil {
		return nil, err
	}
	x := &taskServiceSubscribeClient{stream}
	return x, nil
}

type TaskService_SubscribeClient interface {
	Send(*StreamMessage) error
	CloseAndRecv() (*Response, error)
	grpc.ClientStream
}

type taskServiceSubscribeClient struct {
	grpc.ClientStream
}

func (x *taskServiceSubscribeClient) Send(m *StreamMessage) error {
	return x.ClientStream.SendMsg(m)
}

func (x *taskServiceSubscribeClient) CloseAndRecv() (*Response, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(Response)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *taskServiceClient) Fetch(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := c.cc.Invoke(ctx, "/grpc.TaskService/Fetch", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *taskServiceClient) SendNotification(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := c.cc.Invoke(ctx, "/grpc.TaskService/SendNotification", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TaskServiceServer is the server API for TaskService service.
// All implementations must embed UnimplementedTaskServiceServer
// for forward compatibility
type TaskServiceServer interface {
	Subscribe(TaskService_SubscribeServer) error
	Fetch(context.Context, *Request) (*Response, error)
	SendNotification(context.Context, *Request) (*Response, error)
	mustEmbedUnimplementedTaskServiceServer()
}

// UnimplementedTaskServiceServer must be embedded to have forward compatible implementations.
type UnimplementedTaskServiceServer struct {
}

func (UnimplementedTaskServiceServer) Subscribe(TaskService_SubscribeServer) error {
	return status.Errorf(codes.Unimplemented, "method Subscribe not implemented")
}
func (UnimplementedTaskServiceServer) Fetch(context.Context, *Request) (*Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Fetch not implemented")
}
func (UnimplementedTaskServiceServer) SendNotification(context.Context, *Request) (*Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendNotification not implemented")
}
func (UnimplementedTaskServiceServer) mustEmbedUnimplementedTaskServiceServer() {}

// UnsafeTaskServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TaskServiceServer will
// result in compilation errors.
type UnsafeTaskServiceServer interface {
	mustEmbedUnimplementedTaskServiceServer()
}

func RegisterTaskServiceServer(s grpc.ServiceRegistrar, srv TaskServiceServer) {
	s.RegisterService(&TaskService_ServiceDesc, srv)
}

func _TaskService_Subscribe_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(TaskServiceServer).Subscribe(&taskServiceSubscribeServer{stream})
}

type TaskService_SubscribeServer interface {
	SendAndClose(*Response) error
	Recv() (*StreamMessage, error)
	grpc.ServerStream
}

type taskServiceSubscribeServer struct {
	grpc.ServerStream
}

func (x *taskServiceSubscribeServer) SendAndClose(m *Response) error {
	return x.ServerStream.SendMsg(m)
}

func (x *taskServiceSubscribeServer) Recv() (*StreamMessage, error) {
	m := new(StreamMessage)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _TaskService_Fetch_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TaskServiceServer).Fetch(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpc.TaskService/Fetch",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TaskServiceServer).Fetch(ctx, req.(*Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _TaskService_SendNotification_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TaskServiceServer).SendNotification(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpc.TaskService/SendNotification",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TaskServiceServer).SendNotification(ctx, req.(*Request))
	}
	return interceptor(ctx, in, info, handler)
}

// TaskService_ServiceDesc is the grpc.ServiceDesc for TaskService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var TaskService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "grpc.TaskService",
	HandlerType: (*TaskServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Fetch",
			Handler:    _TaskService_Fetch_Handler,
		},
		{
			MethodName: "SendNotification",
			Handler:    _TaskService_SendNotification_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Subscribe",
			Handler:       _TaskService_Subscribe_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "services/task_service.proto",
}
