package adapters

import (
	"context"
	groupgrpcv1 "github.com/cossim/coss-server/internal/group/api/grpc/v1"
	"github.com/cossim/coss-server/internal/group/app/command"
)

//var _ command.GroupService = &GroupGrpc{}

func NewGroupGrpc(client groupgrpcv1.GroupServiceClient) *GroupGrpc {
	return &GroupGrpc{client: client}
}

type GroupGrpc struct {
	client groupgrpcv1.GroupServiceClient
}

func (s *GroupGrpc) Get(ctx context.Context, id uint32) (*command.Group, error) {
	resp, err := s.client.GetGroupInfoByGid(ctx, &groupgrpcv1.GetGroupInfoRequest{
		Gid: id,
	})
	if err != nil {
		return nil, err
	}

	return &command.Group{
		ID:              resp.Id,
		Type:            uint(resp.Type),
		Name:            resp.Name,
		Avatar:          resp.Avatar,
		MaxMembersLimit: resp.MaxMembersLimit,
		CreatorID:       resp.CreatorId,
		SilenceTime:     resp.SilenceTime,
		JoinApprove:     resp.JoinApprove,
		Encrypt:         resp.Encrypt,
	}, nil
}
