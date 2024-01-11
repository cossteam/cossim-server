// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.32.0
// 	protoc        v3.15.8
// source: api/v1/dialog.proto

package v1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type CreateDialogRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// @inject_tag: json:"owner_id"
	OwnerId string `protobuf:"bytes,1,opt,name=OwnerId,proto3" json:"owner_id"`
	// @inject_tag: json:"type"
	Type uint32 `protobuf:"varint,2,opt,name=Type,proto3" json:"type"`
	// @inject_tag: json:"group_id"
	GroupId uint32 `protobuf:"varint,3,opt,name=GroupId,proto3" json:"group_id"`
}

func (x *CreateDialogRequest) Reset() {
	*x = CreateDialogRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_v1_dialog_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateDialogRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateDialogRequest) ProtoMessage() {}

func (x *CreateDialogRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_v1_dialog_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateDialogRequest.ProtoReflect.Descriptor instead.
func (*CreateDialogRequest) Descriptor() ([]byte, []int) {
	return file_api_v1_dialog_proto_rawDescGZIP(), []int{0}
}

func (x *CreateDialogRequest) GetOwnerId() string {
	if x != nil {
		return x.OwnerId
	}
	return ""
}

func (x *CreateDialogRequest) GetType() uint32 {
	if x != nil {
		return x.Type
	}
	return 0
}

func (x *CreateDialogRequest) GetGroupId() uint32 {
	if x != nil {
		return x.GroupId
	}
	return 0
}

type CreateDialogResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// @inject_tag: json:"id"
	Id uint32 `protobuf:"varint,1,opt,name=Id,proto3" json:"id"`
	// @inject_tag: json:"owner_id"
	OwnerId string `protobuf:"bytes,2,opt,name=OwnerId,proto3" json:"owner_id"`
	// @inject_tag: json:"type"
	Type uint32 `protobuf:"varint,3,opt,name=Type,proto3" json:"type"`
	// @inject_tag: json:"group_id"
	GroupId uint32 `protobuf:"varint,4,opt,name=GroupId,proto3" json:"group_id"`
}

func (x *CreateDialogResponse) Reset() {
	*x = CreateDialogResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_v1_dialog_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateDialogResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateDialogResponse) ProtoMessage() {}

func (x *CreateDialogResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_v1_dialog_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateDialogResponse.ProtoReflect.Descriptor instead.
func (*CreateDialogResponse) Descriptor() ([]byte, []int) {
	return file_api_v1_dialog_proto_rawDescGZIP(), []int{1}
}

func (x *CreateDialogResponse) GetId() uint32 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *CreateDialogResponse) GetOwnerId() string {
	if x != nil {
		return x.OwnerId
	}
	return ""
}

func (x *CreateDialogResponse) GetType() uint32 {
	if x != nil {
		return x.Type
	}
	return 0
}

func (x *CreateDialogResponse) GetGroupId() uint32 {
	if x != nil {
		return x.GroupId
	}
	return 0
}

type JoinDialogRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// @inject_tag: json:"dialog_id"
	DialogId uint32 `protobuf:"varint,1,opt,name=DialogId,proto3" json:"dialog_id"`
	// @inject_tag: json:"user_id"
	UserId string `protobuf:"bytes,2,opt,name=UserId,proto3" json:"user_id"`
}

func (x *JoinDialogRequest) Reset() {
	*x = JoinDialogRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_v1_dialog_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *JoinDialogRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*JoinDialogRequest) ProtoMessage() {}

func (x *JoinDialogRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_v1_dialog_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use JoinDialogRequest.ProtoReflect.Descriptor instead.
func (*JoinDialogRequest) Descriptor() ([]byte, []int) {
	return file_api_v1_dialog_proto_rawDescGZIP(), []int{2}
}

func (x *JoinDialogRequest) GetDialogId() uint32 {
	if x != nil {
		return x.DialogId
	}
	return 0
}

func (x *JoinDialogRequest) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

type GetUserDialogListRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// @inject_tag: json:"user_id"
	UserId string `protobuf:"bytes,1,opt,name=UserId,proto3" json:"user_id"`
}

func (x *GetUserDialogListRequest) Reset() {
	*x = GetUserDialogListRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_v1_dialog_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetUserDialogListRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetUserDialogListRequest) ProtoMessage() {}

func (x *GetUserDialogListRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_v1_dialog_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetUserDialogListRequest.ProtoReflect.Descriptor instead.
func (*GetUserDialogListRequest) Descriptor() ([]byte, []int) {
	return file_api_v1_dialog_proto_rawDescGZIP(), []int{3}
}

func (x *GetUserDialogListRequest) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

type GetUserDialogListResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// @inject_tag: json:"dialog_ids"
	DialogIds []uint32 `protobuf:"varint,1,rep,packed,name=DialogIds,proto3" json:"dialog_ids"`
}

func (x *GetUserDialogListResponse) Reset() {
	*x = GetUserDialogListResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_v1_dialog_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetUserDialogListResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetUserDialogListResponse) ProtoMessage() {}

func (x *GetUserDialogListResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_v1_dialog_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetUserDialogListResponse.ProtoReflect.Descriptor instead.
func (*GetUserDialogListResponse) Descriptor() ([]byte, []int) {
	return file_api_v1_dialog_proto_rawDescGZIP(), []int{4}
}

func (x *GetUserDialogListResponse) GetDialogIds() []uint32 {
	if x != nil {
		return x.DialogIds
	}
	return nil
}

type GetDialogByIdsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// @inject_tag: json:"dialog_ids"
	DialogIds []uint32 `protobuf:"varint,1,rep,packed,name=DialogIds,proto3" json:"dialog_ids"`
}

func (x *GetDialogByIdsRequest) Reset() {
	*x = GetDialogByIdsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_v1_dialog_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetDialogByIdsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetDialogByIdsRequest) ProtoMessage() {}

func (x *GetDialogByIdsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_v1_dialog_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetDialogByIdsRequest.ProtoReflect.Descriptor instead.
func (*GetDialogByIdsRequest) Descriptor() ([]byte, []int) {
	return file_api_v1_dialog_proto_rawDescGZIP(), []int{5}
}

func (x *GetDialogByIdsRequest) GetDialogIds() []uint32 {
	if x != nil {
		return x.DialogIds
	}
	return nil
}

type GetDialogByIdsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// @inject_tag: json:"dialogs"
	Dialogs []*Dialog `protobuf:"bytes,1,rep,name=Dialogs,proto3" json:"dialogs"`
}

func (x *GetDialogByIdsResponse) Reset() {
	*x = GetDialogByIdsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_v1_dialog_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetDialogByIdsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetDialogByIdsResponse) ProtoMessage() {}

func (x *GetDialogByIdsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_v1_dialog_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetDialogByIdsResponse.ProtoReflect.Descriptor instead.
func (*GetDialogByIdsResponse) Descriptor() ([]byte, []int) {
	return file_api_v1_dialog_proto_rawDescGZIP(), []int{6}
}

func (x *GetDialogByIdsResponse) GetDialogs() []*Dialog {
	if x != nil {
		return x.Dialogs
	}
	return nil
}

type Dialog struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// @inject_tag: json:"id"
	Id uint32 `protobuf:"varint,1,opt,name=Id,proto3" json:"id"`
	// @inject_tag: json:"owner_id"
	OwnerId string `protobuf:"bytes,2,opt,name=OwnerId,proto3" json:"owner_id"`
	// @inject_tag: json:"type"
	Type uint32 `protobuf:"varint,3,opt,name=Type,proto3" json:"type"`
	// @inject_tag: json:"group_id"
	GroupId uint32 `protobuf:"varint,4,opt,name=GroupId,proto3" json:"group_id"`
}

func (x *Dialog) Reset() {
	*x = Dialog{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_v1_dialog_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Dialog) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Dialog) ProtoMessage() {}

func (x *Dialog) ProtoReflect() protoreflect.Message {
	mi := &file_api_v1_dialog_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Dialog.ProtoReflect.Descriptor instead.
func (*Dialog) Descriptor() ([]byte, []int) {
	return file_api_v1_dialog_proto_rawDescGZIP(), []int{7}
}

func (x *Dialog) GetId() uint32 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *Dialog) GetOwnerId() string {
	if x != nil {
		return x.OwnerId
	}
	return ""
}

func (x *Dialog) GetType() uint32 {
	if x != nil {
		return x.Type
	}
	return 0
}

func (x *Dialog) GetGroupId() uint32 {
	if x != nil {
		return x.GroupId
	}
	return 0
}

type GetDialogUsersByDialogIDRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// @inject_tag: json:"dialog_id"
	DialogId uint32 `protobuf:"varint,1,opt,name=DialogId,proto3" json:"dialog_id"`
}

func (x *GetDialogUsersByDialogIDRequest) Reset() {
	*x = GetDialogUsersByDialogIDRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_v1_dialog_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetDialogUsersByDialogIDRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetDialogUsersByDialogIDRequest) ProtoMessage() {}

func (x *GetDialogUsersByDialogIDRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_v1_dialog_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetDialogUsersByDialogIDRequest.ProtoReflect.Descriptor instead.
func (*GetDialogUsersByDialogIDRequest) Descriptor() ([]byte, []int) {
	return file_api_v1_dialog_proto_rawDescGZIP(), []int{8}
}

func (x *GetDialogUsersByDialogIDRequest) GetDialogId() uint32 {
	if x != nil {
		return x.DialogId
	}
	return 0
}

type GetDialogUsersByDialogIDResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// @inject_tag: json:"user_ids"
	UserIds []string `protobuf:"bytes,1,rep,name=UserIds,proto3" json:"user_ids"`
}

func (x *GetDialogUsersByDialogIDResponse) Reset() {
	*x = GetDialogUsersByDialogIDResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_v1_dialog_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetDialogUsersByDialogIDResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetDialogUsersByDialogIDResponse) ProtoMessage() {}

func (x *GetDialogUsersByDialogIDResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_v1_dialog_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetDialogUsersByDialogIDResponse.ProtoReflect.Descriptor instead.
func (*GetDialogUsersByDialogIDResponse) Descriptor() ([]byte, []int) {
	return file_api_v1_dialog_proto_rawDescGZIP(), []int{9}
}

func (x *GetDialogUsersByDialogIDResponse) GetUserIds() []string {
	if x != nil {
		return x.UserIds
	}
	return nil
}

type Empty struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *Empty) Reset() {
	*x = Empty{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_v1_dialog_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Empty) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Empty) ProtoMessage() {}

func (x *Empty) ProtoReflect() protoreflect.Message {
	mi := &file_api_v1_dialog_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Empty.ProtoReflect.Descriptor instead.
func (*Empty) Descriptor() ([]byte, []int) {
	return file_api_v1_dialog_proto_rawDescGZIP(), []int{10}
}

var File_api_v1_dialog_proto protoreflect.FileDescriptor

var file_api_v1_dialog_proto_rawDesc = []byte{
	0x0a, 0x13, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x64, 0x69, 0x61, 0x6c, 0x6f, 0x67, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x02, 0x76, 0x31, 0x22, 0x5d, 0x0a, 0x13, 0x43, 0x72, 0x65,
	0x61, 0x74, 0x65, 0x44, 0x69, 0x61, 0x6c, 0x6f, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x18, 0x0a, 0x07, 0x4f, 0x77, 0x6e, 0x65, 0x72, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x07, 0x4f, 0x77, 0x6e, 0x65, 0x72, 0x49, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x54, 0x79,
	0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x04, 0x54, 0x79, 0x70, 0x65, 0x12, 0x18,
	0x0a, 0x07, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x49, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0d, 0x52,
	0x07, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x49, 0x64, 0x22, 0x6e, 0x0a, 0x14, 0x43, 0x72, 0x65, 0x61,
	0x74, 0x65, 0x44, 0x69, 0x61, 0x6c, 0x6f, 0x67, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x0e, 0x0a, 0x02, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x02, 0x49, 0x64,
	0x12, 0x18, 0x0a, 0x07, 0x4f, 0x77, 0x6e, 0x65, 0x72, 0x49, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x07, 0x4f, 0x77, 0x6e, 0x65, 0x72, 0x49, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x54, 0x79,
	0x70, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x04, 0x54, 0x79, 0x70, 0x65, 0x12, 0x18,
	0x0a, 0x07, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x49, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0d, 0x52,
	0x07, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x49, 0x64, 0x22, 0x47, 0x0a, 0x11, 0x4a, 0x6f, 0x69, 0x6e,
	0x44, 0x69, 0x61, 0x6c, 0x6f, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1a, 0x0a,
	0x08, 0x44, 0x69, 0x61, 0x6c, 0x6f, 0x67, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52,
	0x08, 0x44, 0x69, 0x61, 0x6c, 0x6f, 0x67, 0x49, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x55, 0x73, 0x65,
	0x72, 0x49, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x55, 0x73, 0x65, 0x72, 0x49,
	0x64, 0x22, 0x32, 0x0a, 0x18, 0x47, 0x65, 0x74, 0x55, 0x73, 0x65, 0x72, 0x44, 0x69, 0x61, 0x6c,
	0x6f, 0x67, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x16, 0x0a,
	0x06, 0x55, 0x73, 0x65, 0x72, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x55,
	0x73, 0x65, 0x72, 0x49, 0x64, 0x22, 0x39, 0x0a, 0x19, 0x47, 0x65, 0x74, 0x55, 0x73, 0x65, 0x72,
	0x44, 0x69, 0x61, 0x6c, 0x6f, 0x67, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x44, 0x69, 0x61, 0x6c, 0x6f, 0x67, 0x49, 0x64, 0x73, 0x18,
	0x01, 0x20, 0x03, 0x28, 0x0d, 0x52, 0x09, 0x44, 0x69, 0x61, 0x6c, 0x6f, 0x67, 0x49, 0x64, 0x73,
	0x22, 0x35, 0x0a, 0x15, 0x47, 0x65, 0x74, 0x44, 0x69, 0x61, 0x6c, 0x6f, 0x67, 0x42, 0x79, 0x49,
	0x64, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1c, 0x0a, 0x09, 0x44, 0x69, 0x61,
	0x6c, 0x6f, 0x67, 0x49, 0x64, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0d, 0x52, 0x09, 0x44, 0x69,
	0x61, 0x6c, 0x6f, 0x67, 0x49, 0x64, 0x73, 0x22, 0x3e, 0x0a, 0x16, 0x47, 0x65, 0x74, 0x44, 0x69,
	0x61, 0x6c, 0x6f, 0x67, 0x42, 0x79, 0x49, 0x64, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x24, 0x0a, 0x07, 0x44, 0x69, 0x61, 0x6c, 0x6f, 0x67, 0x73, 0x18, 0x01, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x0a, 0x2e, 0x76, 0x31, 0x2e, 0x44, 0x69, 0x61, 0x6c, 0x6f, 0x67, 0x52, 0x07,
	0x44, 0x69, 0x61, 0x6c, 0x6f, 0x67, 0x73, 0x22, 0x60, 0x0a, 0x06, 0x44, 0x69, 0x61, 0x6c, 0x6f,
	0x67, 0x12, 0x0e, 0x0a, 0x02, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x02, 0x49,
	0x64, 0x12, 0x18, 0x0a, 0x07, 0x4f, 0x77, 0x6e, 0x65, 0x72, 0x49, 0x64, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x07, 0x4f, 0x77, 0x6e, 0x65, 0x72, 0x49, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x54,
	0x79, 0x70, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x04, 0x54, 0x79, 0x70, 0x65, 0x12,
	0x18, 0x0a, 0x07, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x49, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0d,
	0x52, 0x07, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x49, 0x64, 0x22, 0x3d, 0x0a, 0x1f, 0x47, 0x65, 0x74,
	0x44, 0x69, 0x61, 0x6c, 0x6f, 0x67, 0x55, 0x73, 0x65, 0x72, 0x73, 0x42, 0x79, 0x44, 0x69, 0x61,
	0x6c, 0x6f, 0x67, 0x49, 0x44, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1a, 0x0a, 0x08,
	0x44, 0x69, 0x61, 0x6c, 0x6f, 0x67, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x08,
	0x44, 0x69, 0x61, 0x6c, 0x6f, 0x67, 0x49, 0x64, 0x22, 0x3c, 0x0a, 0x20, 0x47, 0x65, 0x74, 0x44,
	0x69, 0x61, 0x6c, 0x6f, 0x67, 0x55, 0x73, 0x65, 0x72, 0x73, 0x42, 0x79, 0x44, 0x69, 0x61, 0x6c,
	0x6f, 0x67, 0x49, 0x44, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x18, 0x0a, 0x07,
	0x55, 0x73, 0x65, 0x72, 0x49, 0x64, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x07, 0x55,
	0x73, 0x65, 0x72, 0x49, 0x64, 0x73, 0x22, 0x07, 0x0a, 0x05, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x32,
	0x8e, 0x03, 0x0a, 0x0d, 0x44, 0x69, 0x61, 0x6c, 0x6f, 0x67, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x12, 0x43, 0x0a, 0x0c, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x44, 0x69, 0x61, 0x6c, 0x6f,
	0x67, 0x12, 0x17, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x44, 0x69, 0x61,
	0x6c, 0x6f, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x18, 0x2e, 0x76, 0x31, 0x2e,
	0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x44, 0x69, 0x61, 0x6c, 0x6f, 0x67, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x30, 0x0a, 0x0a, 0x4a, 0x6f, 0x69, 0x6e, 0x44, 0x69,
	0x61, 0x6c, 0x6f, 0x67, 0x12, 0x15, 0x2e, 0x76, 0x31, 0x2e, 0x4a, 0x6f, 0x69, 0x6e, 0x44, 0x69,
	0x61, 0x6c, 0x6f, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x09, 0x2e, 0x76, 0x31,
	0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22, 0x00, 0x12, 0x52, 0x0a, 0x11, 0x47, 0x65, 0x74, 0x55,
	0x73, 0x65, 0x72, 0x44, 0x69, 0x61, 0x6c, 0x6f, 0x67, 0x4c, 0x69, 0x73, 0x74, 0x12, 0x1c, 0x2e,
	0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x55, 0x73, 0x65, 0x72, 0x44, 0x69, 0x61, 0x6c, 0x6f, 0x67,
	0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1d, 0x2e, 0x76, 0x31,
	0x2e, 0x47, 0x65, 0x74, 0x55, 0x73, 0x65, 0x72, 0x44, 0x69, 0x61, 0x6c, 0x6f, 0x67, 0x4c, 0x69,
	0x73, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x49, 0x0a, 0x0e,
	0x47, 0x65, 0x74, 0x44, 0x69, 0x61, 0x6c, 0x6f, 0x67, 0x42, 0x79, 0x49, 0x64, 0x73, 0x12, 0x19,
	0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x44, 0x69, 0x61, 0x6c, 0x6f, 0x67, 0x42, 0x79, 0x49,
	0x64, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1a, 0x2e, 0x76, 0x31, 0x2e, 0x47,
	0x65, 0x74, 0x44, 0x69, 0x61, 0x6c, 0x6f, 0x67, 0x42, 0x79, 0x49, 0x64, 0x73, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x67, 0x0a, 0x18, 0x47, 0x65, 0x74, 0x44, 0x69,
	0x61, 0x6c, 0x6f, 0x67, 0x55, 0x73, 0x65, 0x72, 0x73, 0x42, 0x79, 0x44, 0x69, 0x61, 0x6c, 0x6f,
	0x67, 0x49, 0x44, 0x12, 0x23, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x44, 0x69, 0x61, 0x6c,
	0x6f, 0x67, 0x55, 0x73, 0x65, 0x72, 0x73, 0x42, 0x79, 0x44, 0x69, 0x61, 0x6c, 0x6f, 0x67, 0x49,
	0x44, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x24, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65,
	0x74, 0x44, 0x69, 0x61, 0x6c, 0x6f, 0x67, 0x55, 0x73, 0x65, 0x72, 0x73, 0x42, 0x79, 0x44, 0x69,
	0x61, 0x6c, 0x6f, 0x67, 0x49, 0x44, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00,
	0x42, 0x32, 0x5a, 0x30, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x63,
	0x6f, 0x73, 0x73, 0x69, 0x6d, 0x2f, 0x63, 0x6f, 0x73, 0x73, 0x2d, 0x73, 0x65, 0x72, 0x76, 0x65,
	0x72, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2f, 0x6d, 0x73, 0x67, 0x2f, 0x61, 0x70,
	0x69, 0x2f, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_api_v1_dialog_proto_rawDescOnce sync.Once
	file_api_v1_dialog_proto_rawDescData = file_api_v1_dialog_proto_rawDesc
)

func file_api_v1_dialog_proto_rawDescGZIP() []byte {
	file_api_v1_dialog_proto_rawDescOnce.Do(func() {
		file_api_v1_dialog_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_v1_dialog_proto_rawDescData)
	})
	return file_api_v1_dialog_proto_rawDescData
}

var file_api_v1_dialog_proto_msgTypes = make([]protoimpl.MessageInfo, 11)
var file_api_v1_dialog_proto_goTypes = []interface{}{
	(*CreateDialogRequest)(nil),              // 0: v1.CreateDialogRequest
	(*CreateDialogResponse)(nil),             // 1: v1.CreateDialogResponse
	(*JoinDialogRequest)(nil),                // 2: v1.JoinDialogRequest
	(*GetUserDialogListRequest)(nil),         // 3: v1.GetUserDialogListRequest
	(*GetUserDialogListResponse)(nil),        // 4: v1.GetUserDialogListResponse
	(*GetDialogByIdsRequest)(nil),            // 5: v1.GetDialogByIdsRequest
	(*GetDialogByIdsResponse)(nil),           // 6: v1.GetDialogByIdsResponse
	(*Dialog)(nil),                           // 7: v1.Dialog
	(*GetDialogUsersByDialogIDRequest)(nil),  // 8: v1.GetDialogUsersByDialogIDRequest
	(*GetDialogUsersByDialogIDResponse)(nil), // 9: v1.GetDialogUsersByDialogIDResponse
	(*Empty)(nil),                            // 10: v1.Empty
}
var file_api_v1_dialog_proto_depIdxs = []int32{
	7,  // 0: v1.GetDialogByIdsResponse.Dialogs:type_name -> v1.Dialog
	0,  // 1: v1.DialogService.CreateDialog:input_type -> v1.CreateDialogRequest
	2,  // 2: v1.DialogService.JoinDialog:input_type -> v1.JoinDialogRequest
	3,  // 3: v1.DialogService.GetUserDialogList:input_type -> v1.GetUserDialogListRequest
	5,  // 4: v1.DialogService.GetDialogByIds:input_type -> v1.GetDialogByIdsRequest
	8,  // 5: v1.DialogService.GetDialogUsersByDialogID:input_type -> v1.GetDialogUsersByDialogIDRequest
	1,  // 6: v1.DialogService.CreateDialog:output_type -> v1.CreateDialogResponse
	10, // 7: v1.DialogService.JoinDialog:output_type -> v1.Empty
	4,  // 8: v1.DialogService.GetUserDialogList:output_type -> v1.GetUserDialogListResponse
	6,  // 9: v1.DialogService.GetDialogByIds:output_type -> v1.GetDialogByIdsResponse
	9,  // 10: v1.DialogService.GetDialogUsersByDialogID:output_type -> v1.GetDialogUsersByDialogIDResponse
	6,  // [6:11] is the sub-list for method output_type
	1,  // [1:6] is the sub-list for method input_type
	1,  // [1:1] is the sub-list for extension type_name
	1,  // [1:1] is the sub-list for extension extendee
	0,  // [0:1] is the sub-list for field type_name
}

func init() { file_api_v1_dialog_proto_init() }
func file_api_v1_dialog_proto_init() {
	if File_api_v1_dialog_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_api_v1_dialog_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateDialogRequest); i {
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
		file_api_v1_dialog_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateDialogResponse); i {
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
		file_api_v1_dialog_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*JoinDialogRequest); i {
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
		file_api_v1_dialog_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetUserDialogListRequest); i {
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
		file_api_v1_dialog_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetUserDialogListResponse); i {
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
		file_api_v1_dialog_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetDialogByIdsRequest); i {
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
		file_api_v1_dialog_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetDialogByIdsResponse); i {
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
		file_api_v1_dialog_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Dialog); i {
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
		file_api_v1_dialog_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetDialogUsersByDialogIDRequest); i {
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
		file_api_v1_dialog_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetDialogUsersByDialogIDResponse); i {
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
		file_api_v1_dialog_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Empty); i {
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
			RawDescriptor: file_api_v1_dialog_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   11,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_api_v1_dialog_proto_goTypes,
		DependencyIndexes: file_api_v1_dialog_proto_depIdxs,
		MessageInfos:      file_api_v1_dialog_proto_msgTypes,
	}.Build()
	File_api_v1_dialog_proto = out.File
	file_api_v1_dialog_proto_rawDesc = nil
	file_api_v1_dialog_proto_goTypes = nil
	file_api_v1_dialog_proto_depIdxs = nil
}
