// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.32.0
// 	protoc        v4.25.1
// source: api/grpc/v1/user_login.proto

package v1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type UserLogin struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// @inject_tag: json:"driver_id"
	DriverId string `protobuf:"bytes,1,opt,name=DriverId,proto3" json:"driver_id"`
	// @inject_tag: json:"token"
	Token string `protobuf:"bytes,2,opt,name=Token,proto3" json:"token"`
	// @inject_tag: json:"user_id"
	UserId string `protobuf:"bytes,3,opt,name=UserId,proto3" json:"user_id"`
	// 设备token 用于推送到指定设备
	// @inject_tag: json:"driver_token"
	DriverToken string `protobuf:"bytes,4,opt,name=DriverToken,proto3" json:"driver_token"`
	// 客户端类型 例如 Mobile Desktop UnDefined
	// @inject_tag: json:"client_type"
	ClientType string `protobuf:"bytes,5,opt,name=ClientType,proto3" json:"client_type"`
	// 手机厂商 例如 ios、huawei、xiaomi...
	// @inject_tag: json:"platform"
	Platform string `protobuf:"bytes,6,opt,name=Platform,proto3" json:"platform"`
	// 登录时间
	// @inject_tag: json:"login_time"
	LoginTime int64 `protobuf:"varint,7,opt,name=LoginTime,proto3" json:"login_time"`
}

func (x *UserLogin) Reset() {
	*x = UserLogin{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_grpc_v1_user_login_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UserLogin) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UserLogin) ProtoMessage() {}

func (x *UserLogin) ProtoReflect() protoreflect.Message {
	mi := &file_api_grpc_v1_user_login_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UserLogin.ProtoReflect.Descriptor instead.
func (*UserLogin) Descriptor() ([]byte, []int) {
	return file_api_grpc_v1_user_login_proto_rawDescGZIP(), []int{0}
}

func (x *UserLogin) GetDriverId() string {
	if x != nil {
		return x.DriverId
	}
	return ""
}

func (x *UserLogin) GetToken() string {
	if x != nil {
		return x.Token
	}
	return ""
}

func (x *UserLogin) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

func (x *UserLogin) GetDriverToken() string {
	if x != nil {
		return x.DriverToken
	}
	return ""
}

func (x *UserLogin) GetClientType() string {
	if x != nil {
		return x.ClientType
	}
	return ""
}

func (x *UserLogin) GetPlatform() string {
	if x != nil {
		return x.Platform
	}
	return ""
}

func (x *UserLogin) GetLoginTime() int64 {
	if x != nil {
		return x.LoginTime
	}
	return 0
}

type DriverIdAndUserId struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// @inject_tag: json:"driver_id"
	DriverId string `protobuf:"bytes,1,opt,name=DriverId,proto3" json:"driver_id"`
	// @inject_tag: json:"user_id"
	UserId string `protobuf:"bytes,2,opt,name=UserId,proto3" json:"user_id"`
}

func (x *DriverIdAndUserId) Reset() {
	*x = DriverIdAndUserId{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_grpc_v1_user_login_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DriverIdAndUserId) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DriverIdAndUserId) ProtoMessage() {}

func (x *DriverIdAndUserId) ProtoReflect() protoreflect.Message {
	mi := &file_api_grpc_v1_user_login_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DriverIdAndUserId.ProtoReflect.Descriptor instead.
func (*DriverIdAndUserId) Descriptor() ([]byte, []int) {
	return file_api_grpc_v1_user_login_proto_rawDescGZIP(), []int{1}
}

func (x *DriverIdAndUserId) GetDriverId() string {
	if x != nil {
		return x.DriverId
	}
	return ""
}

func (x *DriverIdAndUserId) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

type TokenUpdate struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// @inject_tag: json:"driver_id"
	DriverId string `protobuf:"bytes,1,opt,name=DriverId,proto3" json:"driver_id"`
	// @inject_tag: json:"token"
	Token string `protobuf:"bytes,2,opt,name=Token,proto3" json:"token"`
	// @inject_tag: json:"user_id"
	UserId string `protobuf:"bytes,3,opt,name=UserId,proto3" json:"user_id"`
}

func (x *TokenUpdate) Reset() {
	*x = TokenUpdate{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_grpc_v1_user_login_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TokenUpdate) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TokenUpdate) ProtoMessage() {}

func (x *TokenUpdate) ProtoReflect() protoreflect.Message {
	mi := &file_api_grpc_v1_user_login_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TokenUpdate.ProtoReflect.Descriptor instead.
func (*TokenUpdate) Descriptor() ([]byte, []int) {
	return file_api_grpc_v1_user_login_proto_rawDescGZIP(), []int{2}
}

func (x *TokenUpdate) GetDriverId() string {
	if x != nil {
		return x.DriverId
	}
	return ""
}

func (x *TokenUpdate) GetToken() string {
	if x != nil {
		return x.Token
	}
	return ""
}

func (x *TokenUpdate) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

type GetUserLoginByTokenRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// @inject_tag: json:"token"
	Token string `protobuf:"bytes,1,opt,name=Token,proto3" json:"token"`
}

func (x *GetUserLoginByTokenRequest) Reset() {
	*x = GetUserLoginByTokenRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_grpc_v1_user_login_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetUserLoginByTokenRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetUserLoginByTokenRequest) ProtoMessage() {}

func (x *GetUserLoginByTokenRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_grpc_v1_user_login_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetUserLoginByTokenRequest.ProtoReflect.Descriptor instead.
func (*GetUserLoginByTokenRequest) Descriptor() ([]byte, []int) {
	return file_api_grpc_v1_user_login_proto_rawDescGZIP(), []int{3}
}

func (x *GetUserLoginByTokenRequest) GetToken() string {
	if x != nil {
		return x.Token
	}
	return ""
}

type GetUserDriverTokenByUserIdRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// @inject_tag: json:"user_id"
	UserId string `protobuf:"bytes,1,opt,name=UserId,proto3" json:"user_id"`
}

func (x *GetUserDriverTokenByUserIdRequest) Reset() {
	*x = GetUserDriverTokenByUserIdRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_grpc_v1_user_login_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetUserDriverTokenByUserIdRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetUserDriverTokenByUserIdRequest) ProtoMessage() {}

func (x *GetUserDriverTokenByUserIdRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_grpc_v1_user_login_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetUserDriverTokenByUserIdRequest.ProtoReflect.Descriptor instead.
func (*GetUserDriverTokenByUserIdRequest) Descriptor() ([]byte, []int) {
	return file_api_grpc_v1_user_login_proto_rawDescGZIP(), []int{4}
}

func (x *GetUserDriverTokenByUserIdRequest) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

type GetUserDriverTokenByUserIdResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// @inject_tag: json:"tokens"
	Token []string `protobuf:"bytes,2,rep,name=Token,proto3" json:"tokens"`
}

func (x *GetUserDriverTokenByUserIdResponse) Reset() {
	*x = GetUserDriverTokenByUserIdResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_grpc_v1_user_login_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetUserDriverTokenByUserIdResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetUserDriverTokenByUserIdResponse) ProtoMessage() {}

func (x *GetUserDriverTokenByUserIdResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_grpc_v1_user_login_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetUserDriverTokenByUserIdResponse.ProtoReflect.Descriptor instead.
func (*GetUserDriverTokenByUserIdResponse) Descriptor() ([]byte, []int) {
	return file_api_grpc_v1_user_login_proto_rawDescGZIP(), []int{5}
}

func (x *GetUserDriverTokenByUserIdResponse) GetToken() []string {
	if x != nil {
		return x.Token
	}
	return nil
}

type GetUserLoginByUserIdRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// @inject_tag: json:"user_id"
	UserId string `protobuf:"bytes,1,opt,name=UserId,proto3" json:"user_id"`
}

func (x *GetUserLoginByUserIdRequest) Reset() {
	*x = GetUserLoginByUserIdRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_grpc_v1_user_login_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetUserLoginByUserIdRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetUserLoginByUserIdRequest) ProtoMessage() {}

func (x *GetUserLoginByUserIdRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_grpc_v1_user_login_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetUserLoginByUserIdRequest.ProtoReflect.Descriptor instead.
func (*GetUserLoginByUserIdRequest) Descriptor() ([]byte, []int) {
	return file_api_grpc_v1_user_login_proto_rawDescGZIP(), []int{6}
}

func (x *GetUserLoginByUserIdRequest) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

var File_api_grpc_v1_user_login_proto protoreflect.FileDescriptor

var file_api_grpc_v1_user_login_proto_rawDesc = []byte{
	0x0a, 0x1c, 0x61, 0x70, 0x69, 0x2f, 0x67, 0x72, 0x70, 0x63, 0x2f, 0x76, 0x31, 0x2f, 0x75, 0x73,
	0x65, 0x72, 0x5f, 0x6c, 0x6f, 0x67, 0x69, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x07,
	0x75, 0x73, 0x65, 0x72, 0x5f, 0x76, 0x31, 0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x22, 0xd1, 0x01, 0x0a, 0x09, 0x55, 0x73, 0x65, 0x72, 0x4c, 0x6f, 0x67,
	0x69, 0x6e, 0x12, 0x1a, 0x0a, 0x08, 0x44, 0x72, 0x69, 0x76, 0x65, 0x72, 0x49, 0x64, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x44, 0x72, 0x69, 0x76, 0x65, 0x72, 0x49, 0x64, 0x12, 0x14,
	0x0a, 0x05, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x54,
	0x6f, 0x6b, 0x65, 0x6e, 0x12, 0x16, 0x0a, 0x06, 0x55, 0x73, 0x65, 0x72, 0x49, 0x64, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x55, 0x73, 0x65, 0x72, 0x49, 0x64, 0x12, 0x20, 0x0a, 0x0b,
	0x44, 0x72, 0x69, 0x76, 0x65, 0x72, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x0b, 0x44, 0x72, 0x69, 0x76, 0x65, 0x72, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x12, 0x1e,
	0x0a, 0x0a, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x18, 0x05, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x0a, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x1a,
	0x0a, 0x08, 0x50, 0x6c, 0x61, 0x74, 0x66, 0x6f, 0x72, 0x6d, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x08, 0x50, 0x6c, 0x61, 0x74, 0x66, 0x6f, 0x72, 0x6d, 0x12, 0x1c, 0x0a, 0x09, 0x4c, 0x6f,
	0x67, 0x69, 0x6e, 0x54, 0x69, 0x6d, 0x65, 0x18, 0x07, 0x20, 0x01, 0x28, 0x03, 0x52, 0x09, 0x4c,
	0x6f, 0x67, 0x69, 0x6e, 0x54, 0x69, 0x6d, 0x65, 0x22, 0x47, 0x0a, 0x11, 0x44, 0x72, 0x69, 0x76,
	0x65, 0x72, 0x49, 0x64, 0x41, 0x6e, 0x64, 0x55, 0x73, 0x65, 0x72, 0x49, 0x64, 0x12, 0x1a, 0x0a,
	0x08, 0x44, 0x72, 0x69, 0x76, 0x65, 0x72, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x08, 0x44, 0x72, 0x69, 0x76, 0x65, 0x72, 0x49, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x55, 0x73, 0x65,
	0x72, 0x49, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x55, 0x73, 0x65, 0x72, 0x49,
	0x64, 0x22, 0x57, 0x0a, 0x0b, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65,
	0x12, 0x1a, 0x0a, 0x08, 0x44, 0x72, 0x69, 0x76, 0x65, 0x72, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x08, 0x44, 0x72, 0x69, 0x76, 0x65, 0x72, 0x49, 0x64, 0x12, 0x14, 0x0a, 0x05,
	0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x54, 0x6f, 0x6b,
	0x65, 0x6e, 0x12, 0x16, 0x0a, 0x06, 0x55, 0x73, 0x65, 0x72, 0x49, 0x64, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x06, 0x55, 0x73, 0x65, 0x72, 0x49, 0x64, 0x22, 0x32, 0x0a, 0x1a, 0x47, 0x65,
	0x74, 0x55, 0x73, 0x65, 0x72, 0x4c, 0x6f, 0x67, 0x69, 0x6e, 0x42, 0x79, 0x54, 0x6f, 0x6b, 0x65,
	0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x54, 0x6f, 0x6b, 0x65,
	0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x22, 0x3b,
	0x0a, 0x21, 0x47, 0x65, 0x74, 0x55, 0x73, 0x65, 0x72, 0x44, 0x72, 0x69, 0x76, 0x65, 0x72, 0x54,
	0x6f, 0x6b, 0x65, 0x6e, 0x42, 0x79, 0x55, 0x73, 0x65, 0x72, 0x49, 0x64, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x55, 0x73, 0x65, 0x72, 0x49, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x06, 0x55, 0x73, 0x65, 0x72, 0x49, 0x64, 0x22, 0x3a, 0x0a, 0x22, 0x47,
	0x65, 0x74, 0x55, 0x73, 0x65, 0x72, 0x44, 0x72, 0x69, 0x76, 0x65, 0x72, 0x54, 0x6f, 0x6b, 0x65,
	0x6e, 0x42, 0x79, 0x55, 0x73, 0x65, 0x72, 0x49, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x14, 0x0a, 0x05, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09,
	0x52, 0x05, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x22, 0x35, 0x0a, 0x1b, 0x47, 0x65, 0x74, 0x55, 0x73,
	0x65, 0x72, 0x4c, 0x6f, 0x67, 0x69, 0x6e, 0x42, 0x79, 0x55, 0x73, 0x65, 0x72, 0x49, 0x64, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x55, 0x73, 0x65, 0x72, 0x49, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x55, 0x73, 0x65, 0x72, 0x49, 0x64, 0x32, 0x99,
	0x04, 0x0a, 0x10, 0x55, 0x73, 0x65, 0x72, 0x4c, 0x6f, 0x67, 0x69, 0x6e, 0x53, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x12, 0x3f, 0x0a, 0x0f, 0x49, 0x6e, 0x73, 0x65, 0x72, 0x74, 0x55, 0x73, 0x65,
	0x72, 0x4c, 0x6f, 0x67, 0x69, 0x6e, 0x12, 0x12, 0x2e, 0x75, 0x73, 0x65, 0x72, 0x5f, 0x76, 0x31,
	0x2e, 0x55, 0x73, 0x65, 0x72, 0x4c, 0x6f, 0x67, 0x69, 0x6e, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70,
	0x74, 0x79, 0x22, 0x00, 0x12, 0x50, 0x0a, 0x13, 0x47, 0x65, 0x74, 0x55, 0x73, 0x65, 0x72, 0x4c,
	0x6f, 0x67, 0x69, 0x6e, 0x42, 0x79, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x12, 0x23, 0x2e, 0x75, 0x73,
	0x65, 0x72, 0x5f, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x55, 0x73, 0x65, 0x72, 0x4c, 0x6f, 0x67,
	0x69, 0x6e, 0x42, 0x79, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x12, 0x2e, 0x75, 0x73, 0x65, 0x72, 0x5f, 0x76, 0x31, 0x2e, 0x55, 0x73, 0x65, 0x72, 0x4c,
	0x6f, 0x67, 0x69, 0x6e, 0x22, 0x00, 0x12, 0x53, 0x0a, 0x1f, 0x47, 0x65, 0x74, 0x55, 0x73, 0x65,
	0x72, 0x4c, 0x6f, 0x67, 0x69, 0x6e, 0x42, 0x79, 0x44, 0x72, 0x69, 0x76, 0x65, 0x72, 0x49, 0x64,
	0x41, 0x6e, 0x64, 0x55, 0x73, 0x65, 0x72, 0x49, 0x64, 0x12, 0x1a, 0x2e, 0x75, 0x73, 0x65, 0x72,
	0x5f, 0x76, 0x31, 0x2e, 0x44, 0x72, 0x69, 0x76, 0x65, 0x72, 0x49, 0x64, 0x41, 0x6e, 0x64, 0x55,
	0x73, 0x65, 0x72, 0x49, 0x64, 0x1a, 0x12, 0x2e, 0x75, 0x73, 0x65, 0x72, 0x5f, 0x76, 0x31, 0x2e,
	0x55, 0x73, 0x65, 0x72, 0x4c, 0x6f, 0x67, 0x69, 0x6e, 0x22, 0x00, 0x12, 0x50, 0x0a, 0x1e, 0x55,
	0x70, 0x64, 0x61, 0x74, 0x65, 0x55, 0x73, 0x65, 0x72, 0x4c, 0x6f, 0x67, 0x69, 0x6e, 0x54, 0x6f,
	0x6b, 0x65, 0x6e, 0x42, 0x79, 0x44, 0x72, 0x69, 0x76, 0x65, 0x72, 0x49, 0x64, 0x12, 0x14, 0x2e,
	0x75, 0x73, 0x65, 0x72, 0x5f, 0x76, 0x31, 0x2e, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x55, 0x70, 0x64,
	0x61, 0x74, 0x65, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22, 0x00, 0x12, 0x77, 0x0a,
	0x1a, 0x47, 0x65, 0x74, 0x55, 0x73, 0x65, 0x72, 0x44, 0x72, 0x69, 0x76, 0x65, 0x72, 0x54, 0x6f,
	0x6b, 0x65, 0x6e, 0x42, 0x79, 0x55, 0x73, 0x65, 0x72, 0x49, 0x64, 0x12, 0x2a, 0x2e, 0x75, 0x73,
	0x65, 0x72, 0x5f, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x55, 0x73, 0x65, 0x72, 0x44, 0x72, 0x69,
	0x76, 0x65, 0x72, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x42, 0x79, 0x55, 0x73, 0x65, 0x72, 0x49, 0x64,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x2b, 0x2e, 0x75, 0x73, 0x65, 0x72, 0x5f, 0x76,
	0x31, 0x2e, 0x47, 0x65, 0x74, 0x55, 0x73, 0x65, 0x72, 0x44, 0x72, 0x69, 0x76, 0x65, 0x72, 0x54,
	0x6f, 0x6b, 0x65, 0x6e, 0x42, 0x79, 0x55, 0x73, 0x65, 0x72, 0x49, 0x64, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x52, 0x0a, 0x14, 0x47, 0x65, 0x74, 0x55, 0x73, 0x65,
	0x72, 0x4c, 0x6f, 0x67, 0x69, 0x6e, 0x42, 0x79, 0x55, 0x73, 0x65, 0x72, 0x49, 0x64, 0x12, 0x24,
	0x2e, 0x75, 0x73, 0x65, 0x72, 0x5f, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x55, 0x73, 0x65, 0x72,
	0x4c, 0x6f, 0x67, 0x69, 0x6e, 0x42, 0x79, 0x55, 0x73, 0x65, 0x72, 0x49, 0x64, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x12, 0x2e, 0x75, 0x73, 0x65, 0x72, 0x5f, 0x76, 0x31, 0x2e, 0x55,
	0x73, 0x65, 0x72, 0x4c, 0x6f, 0x67, 0x69, 0x6e, 0x22, 0x00, 0x42, 0x34, 0x5a, 0x32, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x63, 0x6f, 0x73, 0x73, 0x69, 0x6d, 0x2f,
	0x63, 0x6f, 0x73, 0x73, 0x2d, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2f, 0x69, 0x6e, 0x74, 0x65,
	0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x75, 0x73, 0x65, 0x72, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_api_grpc_v1_user_login_proto_rawDescOnce sync.Once
	file_api_grpc_v1_user_login_proto_rawDescData = file_api_grpc_v1_user_login_proto_rawDesc
)

func file_api_grpc_v1_user_login_proto_rawDescGZIP() []byte {
	file_api_grpc_v1_user_login_proto_rawDescOnce.Do(func() {
		file_api_grpc_v1_user_login_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_grpc_v1_user_login_proto_rawDescData)
	})
	return file_api_grpc_v1_user_login_proto_rawDescData
}

var file_api_grpc_v1_user_login_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_api_grpc_v1_user_login_proto_goTypes = []interface{}{
	(*UserLogin)(nil),                          // 0: user_v1.UserLogin
	(*DriverIdAndUserId)(nil),                  // 1: user_v1.DriverIdAndUserId
	(*TokenUpdate)(nil),                        // 2: user_v1.TokenUpdate
	(*GetUserLoginByTokenRequest)(nil),         // 3: user_v1.GetUserLoginByTokenRequest
	(*GetUserDriverTokenByUserIdRequest)(nil),  // 4: user_v1.GetUserDriverTokenByUserIdRequest
	(*GetUserDriverTokenByUserIdResponse)(nil), // 5: user_v1.GetUserDriverTokenByUserIdResponse
	(*GetUserLoginByUserIdRequest)(nil),        // 6: user_v1.GetUserLoginByUserIdRequest
	(*emptypb.Empty)(nil),                      // 7: google.protobuf.Empty
}
var file_api_grpc_v1_user_login_proto_depIdxs = []int32{
	0, // 0: user_v1.UserLoginService.InsertUserLogin:input_type -> user_v1.UserLogin
	3, // 1: user_v1.UserLoginService.GetUserLoginByToken:input_type -> user_v1.GetUserLoginByTokenRequest
	1, // 2: user_v1.UserLoginService.GetUserLoginByDriverIdAndUserId:input_type -> user_v1.DriverIdAndUserId
	2, // 3: user_v1.UserLoginService.UpdateUserLoginTokenByDriverId:input_type -> user_v1.TokenUpdate
	4, // 4: user_v1.UserLoginService.GetUserDriverTokenByUserId:input_type -> user_v1.GetUserDriverTokenByUserIdRequest
	6, // 5: user_v1.UserLoginService.GetUserLoginByUserId:input_type -> user_v1.GetUserLoginByUserIdRequest
	7, // 6: user_v1.UserLoginService.InsertUserLogin:output_type -> google.protobuf.Empty
	0, // 7: user_v1.UserLoginService.GetUserLoginByToken:output_type -> user_v1.UserLogin
	0, // 8: user_v1.UserLoginService.GetUserLoginByDriverIdAndUserId:output_type -> user_v1.UserLogin
	7, // 9: user_v1.UserLoginService.UpdateUserLoginTokenByDriverId:output_type -> google.protobuf.Empty
	5, // 10: user_v1.UserLoginService.GetUserDriverTokenByUserId:output_type -> user_v1.GetUserDriverTokenByUserIdResponse
	0, // 11: user_v1.UserLoginService.GetUserLoginByUserId:output_type -> user_v1.UserLogin
	6, // [6:12] is the sub-list for method output_type
	0, // [0:6] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_api_grpc_v1_user_login_proto_init() }
func file_api_grpc_v1_user_login_proto_init() {
	if File_api_grpc_v1_user_login_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_api_grpc_v1_user_login_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UserLogin); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_grpc_v1_user_login_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DriverIdAndUserId); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_grpc_v1_user_login_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TokenUpdate); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_grpc_v1_user_login_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetUserLoginByTokenRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_grpc_v1_user_login_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetUserDriverTokenByUserIdRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_grpc_v1_user_login_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetUserDriverTokenByUserIdResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_grpc_v1_user_login_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetUserLoginByUserIdRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_api_grpc_v1_user_login_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_api_grpc_v1_user_login_proto_goTypes,
		DependencyIndexes: file_api_grpc_v1_user_login_proto_depIdxs,
		MessageInfos:      file_api_grpc_v1_user_login_proto_msgTypes,
	}.Build()
	File_api_grpc_v1_user_login_proto = out.File
	file_api_grpc_v1_user_login_proto_rawDesc = nil
	file_api_grpc_v1_user_login_proto_goTypes = nil
	file_api_grpc_v1_user_login_proto_depIdxs = nil
}
