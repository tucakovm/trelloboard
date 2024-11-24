// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.28.3
// source: users.proto

package users

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
	UsersService_RegisterHandler_FullMethodName      = "/UsersService/RegisterHandler"
	UsersService_VerifyHandler_FullMethodName        = "/UsersService/VerifyHandler"
	UsersService_LoginUserHandler_FullMethodName     = "/UsersService/LoginUserHandler"
	UsersService_GetUserByUsername_FullMethodName    = "/UsersService/GetUserByUsername"
	UsersService_DeleteUserByUsername_FullMethodName = "/UsersService/DeleteUserByUsername"
	UsersService_ChangePassword_FullMethodName       = "/UsersService/ChangePassword"
)

// UsersServiceClient is the client API for UsersService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type UsersServiceClient interface {
	RegisterHandler(ctx context.Context, in *RegisterReq, opts ...grpc.CallOption) (*EmptyResponse, error)
	VerifyHandler(ctx context.Context, in *VerifyReq, opts ...grpc.CallOption) (*EmptyResponse, error)
	LoginUserHandler(ctx context.Context, in *LoginReq, opts ...grpc.CallOption) (*LoginRes, error)
	GetUserByUsername(ctx context.Context, in *GetUserByUsernameReq, opts ...grpc.CallOption) (*GetUserByUsernameRes, error)
	DeleteUserByUsername(ctx context.Context, in *GetUserByUsernameReq, opts ...grpc.CallOption) (*EmptyResponse, error)
	ChangePassword(ctx context.Context, in *ChangePasswordReq, opts ...grpc.CallOption) (*EmptyResponse, error)
}

type usersServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewUsersServiceClient(cc grpc.ClientConnInterface) UsersServiceClient {
	return &usersServiceClient{cc}
}

func (c *usersServiceClient) RegisterHandler(ctx context.Context, in *RegisterReq, opts ...grpc.CallOption) (*EmptyResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(EmptyResponse)
	err := c.cc.Invoke(ctx, UsersService_RegisterHandler_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *usersServiceClient) VerifyHandler(ctx context.Context, in *VerifyReq, opts ...grpc.CallOption) (*EmptyResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(EmptyResponse)
	err := c.cc.Invoke(ctx, UsersService_VerifyHandler_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *usersServiceClient) LoginUserHandler(ctx context.Context, in *LoginReq, opts ...grpc.CallOption) (*LoginRes, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(LoginRes)
	err := c.cc.Invoke(ctx, UsersService_LoginUserHandler_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *usersServiceClient) GetUserByUsername(ctx context.Context, in *GetUserByUsernameReq, opts ...grpc.CallOption) (*GetUserByUsernameRes, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetUserByUsernameRes)
	err := c.cc.Invoke(ctx, UsersService_GetUserByUsername_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *usersServiceClient) DeleteUserByUsername(ctx context.Context, in *GetUserByUsernameReq, opts ...grpc.CallOption) (*EmptyResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(EmptyResponse)
	err := c.cc.Invoke(ctx, UsersService_DeleteUserByUsername_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *usersServiceClient) ChangePassword(ctx context.Context, in *ChangePasswordReq, opts ...grpc.CallOption) (*EmptyResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(EmptyResponse)
	err := c.cc.Invoke(ctx, UsersService_ChangePassword_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// UsersServiceServer is the server API for UsersService service.
// All implementations must embed UnimplementedUsersServiceServer
// for forward compatibility.
type UsersServiceServer interface {
	RegisterHandler(context.Context, *RegisterReq) (*EmptyResponse, error)
	VerifyHandler(context.Context, *VerifyReq) (*EmptyResponse, error)
	LoginUserHandler(context.Context, *LoginReq) (*LoginRes, error)
	GetUserByUsername(context.Context, *GetUserByUsernameReq) (*GetUserByUsernameRes, error)
	DeleteUserByUsername(context.Context, *GetUserByUsernameReq) (*EmptyResponse, error)
	ChangePassword(context.Context, *ChangePasswordReq) (*EmptyResponse, error)
	mustEmbedUnimplementedUsersServiceServer()
}

// UnimplementedUsersServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedUsersServiceServer struct{}

func (UnimplementedUsersServiceServer) RegisterHandler(context.Context, *RegisterReq) (*EmptyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RegisterHandler not implemented")
}
func (UnimplementedUsersServiceServer) VerifyHandler(context.Context, *VerifyReq) (*EmptyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method VerifyHandler not implemented")
}
func (UnimplementedUsersServiceServer) LoginUserHandler(context.Context, *LoginReq) (*LoginRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method LoginUserHandler not implemented")
}
func (UnimplementedUsersServiceServer) GetUserByUsername(context.Context, *GetUserByUsernameReq) (*GetUserByUsernameRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUserByUsername not implemented")
}
func (UnimplementedUsersServiceServer) DeleteUserByUsername(context.Context, *GetUserByUsernameReq) (*EmptyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteUserByUsername not implemented")
}
func (UnimplementedUsersServiceServer) ChangePassword(context.Context, *ChangePasswordReq) (*EmptyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ChangePassword not implemented")
}
func (UnimplementedUsersServiceServer) mustEmbedUnimplementedUsersServiceServer() {}
func (UnimplementedUsersServiceServer) testEmbeddedByValue()                      {}

// UnsafeUsersServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to UsersServiceServer will
// result in compilation errors.
type UnsafeUsersServiceServer interface {
	mustEmbedUnimplementedUsersServiceServer()
}

func RegisterUsersServiceServer(s grpc.ServiceRegistrar, srv UsersServiceServer) {
	// If the following call pancis, it indicates UnimplementedUsersServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&UsersService_ServiceDesc, srv)
}

func _UsersService_RegisterHandler_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegisterReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UsersServiceServer).RegisterHandler(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UsersService_RegisterHandler_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UsersServiceServer).RegisterHandler(ctx, req.(*RegisterReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _UsersService_VerifyHandler_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(VerifyReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UsersServiceServer).VerifyHandler(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UsersService_VerifyHandler_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UsersServiceServer).VerifyHandler(ctx, req.(*VerifyReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _UsersService_LoginUserHandler_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LoginReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UsersServiceServer).LoginUserHandler(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UsersService_LoginUserHandler_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UsersServiceServer).LoginUserHandler(ctx, req.(*LoginReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _UsersService_GetUserByUsername_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetUserByUsernameReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UsersServiceServer).GetUserByUsername(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UsersService_GetUserByUsername_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UsersServiceServer).GetUserByUsername(ctx, req.(*GetUserByUsernameReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _UsersService_DeleteUserByUsername_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetUserByUsernameReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UsersServiceServer).DeleteUserByUsername(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UsersService_DeleteUserByUsername_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UsersServiceServer).DeleteUserByUsername(ctx, req.(*GetUserByUsernameReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _UsersService_ChangePassword_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ChangePasswordReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UsersServiceServer).ChangePassword(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UsersService_ChangePassword_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UsersServiceServer).ChangePassword(ctx, req.(*ChangePasswordReq))
	}
	return interceptor(ctx, in, info, handler)
}

// UsersService_ServiceDesc is the grpc.ServiceDesc for UsersService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var UsersService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "UsersService",
	HandlerType: (*UsersServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RegisterHandler",
			Handler:    _UsersService_RegisterHandler_Handler,
		},
		{
			MethodName: "VerifyHandler",
			Handler:    _UsersService_VerifyHandler_Handler,
		},
		{
			MethodName: "LoginUserHandler",
			Handler:    _UsersService_LoginUserHandler_Handler,
		},
		{
			MethodName: "GetUserByUsername",
			Handler:    _UsersService_GetUserByUsername_Handler,
		},
		{
			MethodName: "DeleteUserByUsername",
			Handler:    _UsersService_DeleteUserByUsername_Handler,
		},
		{
			MethodName: "ChangePassword",
			Handler:    _UsersService_ChangePassword_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "users.proto",
}

const (
	ProjectService_UserOnProject_FullMethodName = "/ProjectService/UserOnProject"
)

// ProjectServiceClient is the client API for ProjectService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ProjectServiceClient interface {
	UserOnProject(ctx context.Context, in *UserOnProjectReq, opts ...grpc.CallOption) (*UserOnProjectRes, error)
}

type projectServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewProjectServiceClient(cc grpc.ClientConnInterface) ProjectServiceClient {
	return &projectServiceClient{cc}
}

func (c *projectServiceClient) UserOnProject(ctx context.Context, in *UserOnProjectReq, opts ...grpc.CallOption) (*UserOnProjectRes, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(UserOnProjectRes)
	err := c.cc.Invoke(ctx, ProjectService_UserOnProject_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ProjectServiceServer is the server API for ProjectService service.
// All implementations must embed UnimplementedProjectServiceServer
// for forward compatibility.
type ProjectServiceServer interface {
	UserOnProject(context.Context, *UserOnProjectReq) (*UserOnProjectRes, error)
	mustEmbedUnimplementedProjectServiceServer()
}

// UnimplementedProjectServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedProjectServiceServer struct{}

func (UnimplementedProjectServiceServer) UserOnProject(context.Context, *UserOnProjectReq) (*UserOnProjectRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UserOnProject not implemented")
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

func _ProjectService_UserOnProject_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UserOnProjectReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProjectServiceServer).UserOnProject(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ProjectService_UserOnProject_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProjectServiceServer).UserOnProject(ctx, req.(*UserOnProjectReq))
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
			MethodName: "UserOnProject",
			Handler:    _ProjectService_UserOnProject_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "users.proto",
}