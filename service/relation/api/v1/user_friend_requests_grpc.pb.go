// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.25.1
// source: api/v1/user_friend_requests.proto

package v1

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

const (
	UserFriendRequestService_GetFriendRequestList_FullMethodName                   = "/v1.UserFriendRequestService/GetFriendRequestList"
	UserFriendRequestService_SendFriendRequest_FullMethodName                      = "/v1.UserFriendRequestService/SendFriendRequest"
	UserFriendRequestService_ManageFriendRequest_FullMethodName                    = "/v1.UserFriendRequestService/ManageFriendRequest"
	UserFriendRequestService_GetFriendRequestById_FullMethodName                   = "/v1.UserFriendRequestService/GetFriendRequestById"
	UserFriendRequestService_GetFriendRequestByUserIdAndFriendId_FullMethodName    = "/v1.UserFriendRequestService/GetFriendRequestByUserIdAndFriendId"
	UserFriendRequestService_DeleteFriendRequestByUserIdAndFriendId_FullMethodName = "/v1.UserFriendRequestService/DeleteFriendRequestByUserIdAndFriendId"
)

// UserFriendRequestServiceClient is the client API for UserFriendRequestService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type UserFriendRequestServiceClient interface {
	// 获取好友请求列表
	GetFriendRequestList(ctx context.Context, in *GetFriendRequestListRequest, opts ...grpc.CallOption) (*GetFriendRequestListResponse, error)
	// 发送好友请求
	SendFriendRequest(ctx context.Context, in *SendFriendRequestStruct, opts ...grpc.CallOption) (*SendFriendRequestStructResponse, error)
	// 管理好友请求
	ManageFriendRequest(ctx context.Context, in *ManageFriendRequestStruct, opts ...grpc.CallOption) (*emptypb.Empty, error)
	// 根据请求id获取好友请求
	GetFriendRequestById(ctx context.Context, in *GetFriendRequestByIdRequest, opts ...grpc.CallOption) (*FriendRequestList, error)
	// 根据用户id与好友id获取请求
	GetFriendRequestByUserIdAndFriendId(ctx context.Context, in *GetFriendRequestByUserIdAndFriendIdRequest, opts ...grpc.CallOption) (*FriendRequestList, error)
	// 删除已经处理的请求
	DeleteFriendRequestByUserIdAndFriendId(ctx context.Context, in *DeleteFriendRequestByUserIdAndFriendIdRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type userFriendRequestServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewUserFriendRequestServiceClient(cc grpc.ClientConnInterface) UserFriendRequestServiceClient {
	return &userFriendRequestServiceClient{cc}
}

func (c *userFriendRequestServiceClient) GetFriendRequestList(ctx context.Context, in *GetFriendRequestListRequest, opts ...grpc.CallOption) (*GetFriendRequestListResponse, error) {
	out := new(GetFriendRequestListResponse)
	err := c.cc.Invoke(ctx, UserFriendRequestService_GetFriendRequestList_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userFriendRequestServiceClient) SendFriendRequest(ctx context.Context, in *SendFriendRequestStruct, opts ...grpc.CallOption) (*SendFriendRequestStructResponse, error) {
	out := new(SendFriendRequestStructResponse)
	err := c.cc.Invoke(ctx, UserFriendRequestService_SendFriendRequest_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userFriendRequestServiceClient) ManageFriendRequest(ctx context.Context, in *ManageFriendRequestStruct, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, UserFriendRequestService_ManageFriendRequest_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userFriendRequestServiceClient) GetFriendRequestById(ctx context.Context, in *GetFriendRequestByIdRequest, opts ...grpc.CallOption) (*FriendRequestList, error) {
	out := new(FriendRequestList)
	err := c.cc.Invoke(ctx, UserFriendRequestService_GetFriendRequestById_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userFriendRequestServiceClient) GetFriendRequestByUserIdAndFriendId(ctx context.Context, in *GetFriendRequestByUserIdAndFriendIdRequest, opts ...grpc.CallOption) (*FriendRequestList, error) {
	out := new(FriendRequestList)
	err := c.cc.Invoke(ctx, UserFriendRequestService_GetFriendRequestByUserIdAndFriendId_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userFriendRequestServiceClient) DeleteFriendRequestByUserIdAndFriendId(ctx context.Context, in *DeleteFriendRequestByUserIdAndFriendIdRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, UserFriendRequestService_DeleteFriendRequestByUserIdAndFriendId_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// UserFriendRequestServiceServer is the server API for UserFriendRequestService service.
// All implementations must embed UnimplementedUserFriendRequestServiceServer
// for forward compatibility
type UserFriendRequestServiceServer interface {
	// 获取好友请求列表
	GetFriendRequestList(context.Context, *GetFriendRequestListRequest) (*GetFriendRequestListResponse, error)
	// 发送好友请求
	SendFriendRequest(context.Context, *SendFriendRequestStruct) (*SendFriendRequestStructResponse, error)
	// 管理好友请求
	ManageFriendRequest(context.Context, *ManageFriendRequestStruct) (*emptypb.Empty, error)
	// 根据请求id获取好友请求
	GetFriendRequestById(context.Context, *GetFriendRequestByIdRequest) (*FriendRequestList, error)
	// 根据用户id与好友id获取请求
	GetFriendRequestByUserIdAndFriendId(context.Context, *GetFriendRequestByUserIdAndFriendIdRequest) (*FriendRequestList, error)
	// 删除已经处理的请求
	DeleteFriendRequestByUserIdAndFriendId(context.Context, *DeleteFriendRequestByUserIdAndFriendIdRequest) (*emptypb.Empty, error)
	mustEmbedUnimplementedUserFriendRequestServiceServer()
}

// UnimplementedUserFriendRequestServiceServer must be embedded to have forward compatible implementations.
type UnimplementedUserFriendRequestServiceServer struct {
}

func (UnimplementedUserFriendRequestServiceServer) GetFriendRequestList(context.Context, *GetFriendRequestListRequest) (*GetFriendRequestListResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetFriendRequestList not implemented")
}
func (UnimplementedUserFriendRequestServiceServer) SendFriendRequest(context.Context, *SendFriendRequestStruct) (*SendFriendRequestStructResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendFriendRequest not implemented")
}
func (UnimplementedUserFriendRequestServiceServer) ManageFriendRequest(context.Context, *ManageFriendRequestStruct) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ManageFriendRequest not implemented")
}
func (UnimplementedUserFriendRequestServiceServer) GetFriendRequestById(context.Context, *GetFriendRequestByIdRequest) (*FriendRequestList, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetFriendRequestById not implemented")
}
func (UnimplementedUserFriendRequestServiceServer) GetFriendRequestByUserIdAndFriendId(context.Context, *GetFriendRequestByUserIdAndFriendIdRequest) (*FriendRequestList, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetFriendRequestByUserIdAndFriendId not implemented")
}
func (UnimplementedUserFriendRequestServiceServer) DeleteFriendRequestByUserIdAndFriendId(context.Context, *DeleteFriendRequestByUserIdAndFriendIdRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteFriendRequestByUserIdAndFriendId not implemented")
}
func (UnimplementedUserFriendRequestServiceServer) mustEmbedUnimplementedUserFriendRequestServiceServer() {
}

// UnsafeUserFriendRequestServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to UserFriendRequestServiceServer will
// result in compilation errors.
type UnsafeUserFriendRequestServiceServer interface {
	mustEmbedUnimplementedUserFriendRequestServiceServer()
}

func RegisterUserFriendRequestServiceServer(s grpc.ServiceRegistrar, srv UserFriendRequestServiceServer) {
	s.RegisterService(&UserFriendRequestService_ServiceDesc, srv)
}

func _UserFriendRequestService_GetFriendRequestList_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetFriendRequestListRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserFriendRequestServiceServer).GetFriendRequestList(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UserFriendRequestService_GetFriendRequestList_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserFriendRequestServiceServer).GetFriendRequestList(ctx, req.(*GetFriendRequestListRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserFriendRequestService_SendFriendRequest_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SendFriendRequestStruct)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserFriendRequestServiceServer).SendFriendRequest(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UserFriendRequestService_SendFriendRequest_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserFriendRequestServiceServer).SendFriendRequest(ctx, req.(*SendFriendRequestStruct))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserFriendRequestService_ManageFriendRequest_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ManageFriendRequestStruct)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserFriendRequestServiceServer).ManageFriendRequest(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UserFriendRequestService_ManageFriendRequest_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserFriendRequestServiceServer).ManageFriendRequest(ctx, req.(*ManageFriendRequestStruct))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserFriendRequestService_GetFriendRequestById_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetFriendRequestByIdRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserFriendRequestServiceServer).GetFriendRequestById(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UserFriendRequestService_GetFriendRequestById_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserFriendRequestServiceServer).GetFriendRequestById(ctx, req.(*GetFriendRequestByIdRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserFriendRequestService_GetFriendRequestByUserIdAndFriendId_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetFriendRequestByUserIdAndFriendIdRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserFriendRequestServiceServer).GetFriendRequestByUserIdAndFriendId(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UserFriendRequestService_GetFriendRequestByUserIdAndFriendId_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserFriendRequestServiceServer).GetFriendRequestByUserIdAndFriendId(ctx, req.(*GetFriendRequestByUserIdAndFriendIdRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserFriendRequestService_DeleteFriendRequestByUserIdAndFriendId_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteFriendRequestByUserIdAndFriendIdRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserFriendRequestServiceServer).DeleteFriendRequestByUserIdAndFriendId(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UserFriendRequestService_DeleteFriendRequestByUserIdAndFriendId_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserFriendRequestServiceServer).DeleteFriendRequestByUserIdAndFriendId(ctx, req.(*DeleteFriendRequestByUserIdAndFriendIdRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// UserFriendRequestService_ServiceDesc is the grpc.ServiceDesc for UserFriendRequestService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var UserFriendRequestService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "v1.UserFriendRequestService",
	HandlerType: (*UserFriendRequestServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetFriendRequestList",
			Handler:    _UserFriendRequestService_GetFriendRequestList_Handler,
		},
		{
			MethodName: "SendFriendRequest",
			Handler:    _UserFriendRequestService_SendFriendRequest_Handler,
		},
		{
			MethodName: "ManageFriendRequest",
			Handler:    _UserFriendRequestService_ManageFriendRequest_Handler,
		},
		{
			MethodName: "GetFriendRequestById",
			Handler:    _UserFriendRequestService_GetFriendRequestById_Handler,
		},
		{
			MethodName: "GetFriendRequestByUserIdAndFriendId",
			Handler:    _UserFriendRequestService_GetFriendRequestByUserIdAndFriendId_Handler,
		},
		{
			MethodName: "DeleteFriendRequestByUserIdAndFriendId",
			Handler:    _UserFriendRequestService_DeleteFriendRequestByUserIdAndFriendId_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/v1/user_friend_requests.proto",
}
