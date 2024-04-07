// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.25.1
// source: api/grpc/v1/group.proto

package v1

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

const (
	GroupService_GetGroupInfoByGid_FullMethodName      = "/group_v1.GroupService/GetGroupInfoByGid"
	GroupService_GetBatchGroupInfoByIDs_FullMethodName = "/group_v1.GroupService/GetBatchGroupInfoByIDs"
	GroupService_UpdateGroup_FullMethodName            = "/group_v1.GroupService/UpdateGroup"
	GroupService_CreateGroup_FullMethodName            = "/group_v1.GroupService/CreateGroup"
	GroupService_CreateGroupRevert_FullMethodName      = "/group_v1.GroupService/CreateGroupRevert"
	GroupService_DeleteGroup_FullMethodName            = "/group_v1.GroupService/DeleteGroup"
	GroupService_DeleteGroupRevert_FullMethodName      = "/group_v1.GroupService/DeleteGroupRevert"
)

// GroupServiceClient is the client API for GroupService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type GroupServiceClient interface {
	// 通过群组ID获取群组信息
	GetGroupInfoByGid(ctx context.Context, in *GetGroupInfoRequest, opts ...grpc.CallOption) (*Group, error)
	// 批量通过群组ID获取群组信息
	GetBatchGroupInfoByIDs(ctx context.Context, in *GetBatchGroupInfoRequest, opts ...grpc.CallOption) (*GetBatchGroupInfoResponse, error)
	// 更新群组信息
	UpdateGroup(ctx context.Context, in *UpdateGroupRequest, opts ...grpc.CallOption) (*Group, error)
	// 创建群组
	CreateGroup(ctx context.Context, in *CreateGroupRequest, opts ...grpc.CallOption) (*Group, error)
	// CreateGroup回滚操作
	CreateGroupRevert(ctx context.Context, in *CreateGroupRequest, opts ...grpc.CallOption) (*Group, error)
	// 删除群组
	DeleteGroup(ctx context.Context, in *DeleteGroupRequest, opts ...grpc.CallOption) (*EmptyResponse, error)
	// 删除群组回滚操作
	DeleteGroupRevert(ctx context.Context, in *DeleteGroupRequest, opts ...grpc.CallOption) (*EmptyResponse, error)
}

type groupServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewGroupServiceClient(cc grpc.ClientConnInterface) GroupServiceClient {
	return &groupServiceClient{cc}
}

func (c *groupServiceClient) GetGroupInfoByGid(ctx context.Context, in *GetGroupInfoRequest, opts ...grpc.CallOption) (*Group, error) {
	out := new(Group)
	err := c.cc.Invoke(ctx, GroupService_GetGroupInfoByGid_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *groupServiceClient) GetBatchGroupInfoByIDs(ctx context.Context, in *GetBatchGroupInfoRequest, opts ...grpc.CallOption) (*GetBatchGroupInfoResponse, error) {
	out := new(GetBatchGroupInfoResponse)
	err := c.cc.Invoke(ctx, GroupService_GetBatchGroupInfoByIDs_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *groupServiceClient) UpdateGroup(ctx context.Context, in *UpdateGroupRequest, opts ...grpc.CallOption) (*Group, error) {
	out := new(Group)
	err := c.cc.Invoke(ctx, GroupService_UpdateGroup_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *groupServiceClient) CreateGroup(ctx context.Context, in *CreateGroupRequest, opts ...grpc.CallOption) (*Group, error) {
	out := new(Group)
	err := c.cc.Invoke(ctx, GroupService_CreateGroup_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *groupServiceClient) CreateGroupRevert(ctx context.Context, in *CreateGroupRequest, opts ...grpc.CallOption) (*Group, error) {
	out := new(Group)
	err := c.cc.Invoke(ctx, GroupService_CreateGroupRevert_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *groupServiceClient) DeleteGroup(ctx context.Context, in *DeleteGroupRequest, opts ...grpc.CallOption) (*EmptyResponse, error) {
	out := new(EmptyResponse)
	err := c.cc.Invoke(ctx, GroupService_DeleteGroup_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *groupServiceClient) DeleteGroupRevert(ctx context.Context, in *DeleteGroupRequest, opts ...grpc.CallOption) (*EmptyResponse, error) {
	out := new(EmptyResponse)
	err := c.cc.Invoke(ctx, GroupService_DeleteGroupRevert_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GroupServiceServer is the server API for GroupService service.
// All implementations must embed UnimplementedGroupServiceServer
// for forward compatibility
type GroupServiceServer interface {
	// 通过群组ID获取群组信息
	GetGroupInfoByGid(context.Context, *GetGroupInfoRequest) (*Group, error)
	// 批量通过群组ID获取群组信息
	GetBatchGroupInfoByIDs(context.Context, *GetBatchGroupInfoRequest) (*GetBatchGroupInfoResponse, error)
	// 更新群组信息
	UpdateGroup(context.Context, *UpdateGroupRequest) (*Group, error)
	// 创建群组
	CreateGroup(context.Context, *CreateGroupRequest) (*Group, error)
	// CreateGroup回滚操作
	CreateGroupRevert(context.Context, *CreateGroupRequest) (*Group, error)
	// 删除群组
	DeleteGroup(context.Context, *DeleteGroupRequest) (*EmptyResponse, error)
	// 删除群组回滚操作
	DeleteGroupRevert(context.Context, *DeleteGroupRequest) (*EmptyResponse, error)
	mustEmbedUnimplementedGroupServiceServer()
}

// UnimplementedGroupServiceServer must be embedded to have forward compatible implementations.
type UnimplementedGroupServiceServer struct {
}

func (UnimplementedGroupServiceServer) GetGroupInfoByGid(context.Context, *GetGroupInfoRequest) (*Group, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetGroupInfoByGid not implemented")
}
func (UnimplementedGroupServiceServer) GetBatchGroupInfoByIDs(context.Context, *GetBatchGroupInfoRequest) (*GetBatchGroupInfoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetBatchGroupInfoByIDs not implemented")
}
func (UnimplementedGroupServiceServer) UpdateGroup(context.Context, *UpdateGroupRequest) (*Group, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateGroup not implemented")
}
func (UnimplementedGroupServiceServer) CreateGroup(context.Context, *CreateGroupRequest) (*Group, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateGroup not implemented")
}
func (UnimplementedGroupServiceServer) CreateGroupRevert(context.Context, *CreateGroupRequest) (*Group, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateGroupRevert not implemented")
}
func (UnimplementedGroupServiceServer) DeleteGroup(context.Context, *DeleteGroupRequest) (*EmptyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteGroup not implemented")
}
func (UnimplementedGroupServiceServer) DeleteGroupRevert(context.Context, *DeleteGroupRequest) (*EmptyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteGroupRevert not implemented")
}
func (UnimplementedGroupServiceServer) mustEmbedUnimplementedGroupServiceServer() {}

// UnsafeGroupServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to GroupServiceServer will
// result in compilation errors.
type UnsafeGroupServiceServer interface {
	mustEmbedUnimplementedGroupServiceServer()
}

func RegisterGroupServiceServer(s grpc.ServiceRegistrar, srv GroupServiceServer) {
	s.RegisterService(&GroupService_ServiceDesc, srv)
}

func _GroupService_GetGroupInfoByGid_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetGroupInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GroupServiceServer).GetGroupInfoByGid(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GroupService_GetGroupInfoByGid_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GroupServiceServer).GetGroupInfoByGid(ctx, req.(*GetGroupInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GroupService_GetBatchGroupInfoByIDs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetBatchGroupInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GroupServiceServer).GetBatchGroupInfoByIDs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GroupService_GetBatchGroupInfoByIDs_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GroupServiceServer).GetBatchGroupInfoByIDs(ctx, req.(*GetBatchGroupInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GroupService_UpdateGroup_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateGroupRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GroupServiceServer).UpdateGroup(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GroupService_UpdateGroup_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GroupServiceServer).UpdateGroup(ctx, req.(*UpdateGroupRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GroupService_CreateGroup_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateGroupRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GroupServiceServer).CreateGroup(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GroupService_CreateGroup_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GroupServiceServer).CreateGroup(ctx, req.(*CreateGroupRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GroupService_CreateGroupRevert_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateGroupRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GroupServiceServer).CreateGroupRevert(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GroupService_CreateGroupRevert_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GroupServiceServer).CreateGroupRevert(ctx, req.(*CreateGroupRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GroupService_DeleteGroup_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteGroupRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GroupServiceServer).DeleteGroup(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GroupService_DeleteGroup_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GroupServiceServer).DeleteGroup(ctx, req.(*DeleteGroupRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GroupService_DeleteGroupRevert_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteGroupRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GroupServiceServer).DeleteGroupRevert(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GroupService_DeleteGroupRevert_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GroupServiceServer).DeleteGroupRevert(ctx, req.(*DeleteGroupRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// GroupService_ServiceDesc is the grpc.ServiceDesc for GroupService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var GroupService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "group_v1.GroupService",
	HandlerType: (*GroupServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetGroupInfoByGid",
			Handler:    _GroupService_GetGroupInfoByGid_Handler,
		},
		{
			MethodName: "GetBatchGroupInfoByIDs",
			Handler:    _GroupService_GetBatchGroupInfoByIDs_Handler,
		},
		{
			MethodName: "UpdateGroup",
			Handler:    _GroupService_UpdateGroup_Handler,
		},
		{
			MethodName: "CreateGroup",
			Handler:    _GroupService_CreateGroup_Handler,
		},
		{
			MethodName: "CreateGroupRevert",
			Handler:    _GroupService_CreateGroupRevert_Handler,
		},
		{
			MethodName: "DeleteGroup",
			Handler:    _GroupService_DeleteGroup_Handler,
		},
		{
			MethodName: "DeleteGroupRevert",
			Handler:    _GroupService_DeleteGroupRevert_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/grpc/v1/group.proto",
}