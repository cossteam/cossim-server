package rpc

import (
	"context"
	groupgrpcv1 "github.com/cossim/coss-server/internal/group/api/grpc/v1"
	"google.golang.org/grpc"
)

type GroupInfo struct {
	Type            uint8
	Status          uint8
	JoinApprove     bool
	ID              uint32
	MaxMembersLimit int32
	CreatorId       string
	Name            string
	Avatar          string
}

const (
	PrivateGroup = uint8(groupgrpcv1.GroupType_Private)
	PublicGroup  = uint8(groupgrpcv1.GroupType_Public)
)

type GroupService interface {
	// IsActiveGroup 检查群聊状态是否正常
	IsActiveGroup(ctx context.Context, groupID uint32) (bool, error)

	// IsPrivateGroup 检查群聊是否是私有群
	IsPrivateGroup(ctx context.Context, groupID uint32) (bool, error)
	GetGroupInfo(ctx context.Context, groupID uint32) (*GroupInfo, error)
	GetGroupsInfo(ctx context.Context, groupID []uint32) ([]*GroupInfo, error)
}

var _ GroupService = &groupServiceGrpc{}

func NewGroupGrpc(addr string) (GroupService, error) {
	var grpcOptions = []grpc.DialOption{grpc.WithInsecure()}
	grpcOptions = append(grpcOptions, grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`))
	conn, err := grpc.Dial(
		addr,
		grpcOptions...,
	)
	if err != nil {
		return nil, err
	}
	return &groupServiceGrpc{client: groupgrpcv1.NewGroupServiceClient(conn)}, nil
}

func NewGroupServiceGrpcWithClient(client groupgrpcv1.GroupServiceClient) GroupService {
	return &groupServiceGrpc{client: client}
}

type groupServiceGrpc struct {
	client groupgrpcv1.GroupServiceClient
}

func (s *groupServiceGrpc) GetGroupsInfo(ctx context.Context, groupID []uint32) ([]*GroupInfo, error) {
	var resp []*GroupInfo

	r, err := s.client.GetBatchGroupInfoByIDs(ctx, &groupgrpcv1.GetBatchGroupInfoRequest{GroupIds: groupID})
	if err != nil {
		return nil, err
	}

	for _, info := range r.Groups {
		resp = append(resp, &GroupInfo{
			ID:          info.Id,
			Type:        uint8(info.Type),
			Status:      uint8(info.Status),
			Name:        info.Name,
			Avatar:      info.Avatar,
			CreatorId:   info.CreatorId,
			JoinApprove: info.JoinApprove,
		})
	}
	return resp, nil
}

func (s *groupServiceGrpc) IsPrivateGroup(ctx context.Context, groupID uint32) (bool, error) {
	info, err := s.client.GetGroupInfoByGid(ctx, &groupgrpcv1.GetGroupInfoRequest{Gid: groupID})
	if err != nil {
		return false, err
	}
	return info.Type == groupgrpcv1.GroupType_Private, nil
}

func (s *groupServiceGrpc) GetGroupInfo(ctx context.Context, groupID uint32) (*GroupInfo, error) {
	info, err := s.client.GetGroupInfoByGid(ctx, &groupgrpcv1.GetGroupInfoRequest{Gid: groupID})
	if err != nil {
		return nil, err
	}

	return &GroupInfo{
		ID:              info.Id,
		JoinApprove:     info.JoinApprove,
		Type:            uint8(info.Type),
		Status:          uint8(info.Status),
		CreatorId:       info.CreatorId,
		Name:            info.Name,
		Avatar:          info.Avatar,
		MaxMembersLimit: info.MaxMembersLimit,
	}, nil
}

func (s *groupServiceGrpc) IsActiveGroup(ctx context.Context, groupID uint32) (bool, error) {
	r, err := s.client.GetGroupInfoByGid(ctx, &groupgrpcv1.GetGroupInfoRequest{Gid: groupID})
	if err != nil {
		return false, err
	}

	return r.Status == groupgrpcv1.GroupStatus_GROUP_STATUS_NORMAL, nil
}
