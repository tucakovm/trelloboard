// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.28.3
// source: api-gateway.proto

package api_gateway

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	ProjectService_GetAllProjects_FullMethodName = "/ProjectService/GetAllProjects"
	ProjectService_Create_FullMethodName         = "/ProjectService/Create"
	ProjectService_Delete_FullMethodName         = "/ProjectService/Delete"
	ProjectService_GetById_FullMethodName        = "/ProjectService/GetById"
)

// ProjectServiceClient is the client API for ProjectService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ProjectServiceClient interface {
	GetAllProjects(ctx context.Context, in *GetAllProjectsReq, opts ...grpc.CallOption) (*GetAllProjectsRes, error)
	Create(ctx context.Context, in *CreateProjectReq, opts ...grpc.CallOption) (*EmptyResponse, error)
	Delete(ctx context.Context, in *DeleteProjectReq, opts ...grpc.CallOption) (*EmptyResponse, error)
	GetById(ctx context.Context, in *GetByIdReq, opts ...grpc.CallOption) (*GetByIdRes, error)
}

type projectServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewProjectServiceClient(cc grpc.ClientConnInterface) ProjectServiceClient {
	return &projectServiceClient{cc}
}

func (c *projectServiceClient) GetAllProjects(ctx context.Context, in *GetAllProjectsReq, opts ...grpc.CallOption) (*GetAllProjectsRes, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetAllProjectsRes)
	err := c.cc.Invoke(ctx, ProjectService_GetAllProjects_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *projectServiceClient) Create(ctx context.Context, in *CreateProjectReq, opts ...grpc.CallOption) (*EmptyResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(EmptyResponse)
	err := c.cc.Invoke(ctx, ProjectService_Create_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *projectServiceClient) Delete(ctx context.Context, in *DeleteProjectReq, opts ...grpc.CallOption) (*EmptyResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(EmptyResponse)
	err := c.cc.Invoke(ctx, ProjectService_Delete_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *projectServiceClient) GetById(ctx context.Context, in *GetByIdReq, opts ...grpc.CallOption) (*GetByIdRes, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetByIdRes)
	err := c.cc.Invoke(ctx, ProjectService_GetById_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ProjectServiceServer is the server API for ProjectService service.
// All implementations must embed UnimplementedProjectServiceServer
// for forward compatibility.
type ProjectServiceServer interface {
	GetAllProjects(context.Context, *GetAllProjectsReq) (*GetAllProjectsRes, error)
	Create(context.Context, *CreateProjectReq) (*EmptyResponse, error)
	Delete(context.Context, *DeleteProjectReq) (*EmptyResponse, error)
	GetById(context.Context, *GetByIdReq) (*GetByIdRes, error)
	mustEmbedUnimplementedProjectServiceServer()
}

// UnimplementedProjectServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedProjectServiceServer struct{}

func (UnimplementedProjectServiceServer) GetAllProjects(context.Context, *GetAllProjectsReq) (*GetAllProjectsRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAllProjects not implemented")
}
func (UnimplementedProjectServiceServer) Create(context.Context, *CreateProjectReq) (*EmptyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Create not implemented")
}
func (UnimplementedProjectServiceServer) Delete(context.Context, *DeleteProjectReq) (*EmptyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Delete not implemented")
}
func (UnimplementedProjectServiceServer) GetById(context.Context, *GetByIdReq) (*GetByIdRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetById not implemented")
}
func (UnimplementedProjectServiceServer) mustEmbedUnimplementedProjectServiceServer() {}
func (UnimplementedProjectServiceServer) testEmbeddedByValue()                        {}

// UnsafeProjectServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ProjectServiceServer will
// result in compilation errors.
type UnsafeProjectServiceServer interface {
	mustEmbedUnimplementedProjectServiceServer()
}

func RegisterProjectServiceServer(s grpc.ServiceRegistrar, srv ProjectServiceServer) {
	// If the following call pancis, it indicates UnimplementedProjectServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&ProjectService_ServiceDesc, srv)
}

func _ProjectService_GetAllProjects_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetAllProjectsReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProjectServiceServer).GetAllProjects(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ProjectService_GetAllProjects_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProjectServiceServer).GetAllProjects(ctx, req.(*GetAllProjectsReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ProjectService_Create_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateProjectReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProjectServiceServer).Create(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ProjectService_Create_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProjectServiceServer).Create(ctx, req.(*CreateProjectReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ProjectService_Delete_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteProjectReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProjectServiceServer).Delete(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ProjectService_Delete_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProjectServiceServer).Delete(ctx, req.(*DeleteProjectReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ProjectService_GetById_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetByIdReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProjectServiceServer).GetById(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ProjectService_GetById_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProjectServiceServer).GetById(ctx, req.(*GetByIdReq))
	}
	return interceptor(ctx, in, info, handler)
}

// ProjectService_ServiceDesc is the grpc.ServiceDesc for ProjectService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ProjectService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "ProjectService",
	HandlerType: (*ProjectServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetAllProjects",
			Handler:    _ProjectService_GetAllProjects_Handler,
		},
		{
			MethodName: "Create",
			Handler:    _ProjectService_Create_Handler,
		},
		{
			MethodName: "Delete",
			Handler:    _ProjectService_Delete_Handler,
		},
		{
			MethodName: "GetById",
			Handler:    _ProjectService_GetById_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api-gateway.proto",
}