package adapters

import (
	"context"
	"github.com/cossim/coss-server/internal/group/app/command"
	relationgrpcv1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	"github.com/cossim/coss-server/pkg/code"
	"go.uber.org/zap"
)

//var _ command.RelationGroupService = &RelationGroupGrpc{}

const (
	RelationGroupServiceName = "relation_service"
)

func NewRelationGroupGrpc(client relationgrpcv1.GroupRelationServiceClient) *RelationGroupGrpc {
	return &RelationGroupGrpc{client: client}
}

type RelationGroupGrpc struct {
	client relationgrpcv1.GroupRelationServiceClient
	logger *zap.Logger
}

func (s *RelationGroupGrpc) GetRelation(ctx context.Context, groupID uint32, userID string) (*command.GroupRelationship, error) {
	resp, err := s.client.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{
		UserId:  userID,
		GroupId: groupID,
	})
	if err != nil {
		return nil, err
	}

	return &command.GroupRelationship{
		UserId:                      resp.UserId,
		ID:                          resp.GroupId,
		Identity:                    uint(resp.Identity),
		JoinMethod:                  uint(resp.JoinMethod),
		JoinTime:                    resp.JoinTime,
		MuteEndTime:                 resp.MuteEndTime,
		IsSilent:                    uint(resp.IsSilent),
		Inviter:                     resp.Inviter,
		Remark:                      resp.Remark,
		OpenBurnAfterReading:        uint(resp.OpenBurnAfterReading),
		OpenBurnAfterReadingTimeOut: uint(resp.OpenBurnAfterReadingTimeOut),
	}, nil
}

func (s *RelationGroupGrpc) DeleteGroupRelationsRevert(ctx context.Context, groupID uint32) error {
	_, err := s.client.DeleteGroupRelationByGroupIdRevert(ctx, &relationgrpcv1.GroupIDRequest{
		GroupId: groupID,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *RelationGroupGrpc) DeleteGroupRelations(ctx context.Context, groupID uint32) error {
	_, err := s.client.DeleteGroupRelationByGroupId(ctx, &relationgrpcv1.GroupIDRequest{
		GroupId: groupID,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *RelationGroupGrpc) GetGroupMembers(ctx context.Context, groupID uint32) ([]string, error) {
	ids, err := s.client.GetGroupUserIDs(ctx, &relationgrpcv1.GroupIDRequest{GroupId: groupID})
	if err != nil {
		return nil, err
	}

	return ids.UserIds, nil
}

func (s *RelationGroupGrpc) IsGroupOwner(ctx context.Context, groupID uint32, userID string) (bool, error) {
	sf, err := s.client.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{
		UserId:  userID,
		GroupId: groupID,
	})
	if err != nil {
		return false, err
	}

	if sf.Identity == relationgrpcv1.GroupIdentity_IDENTITY_OWNER {
		return true, nil
	}
	return false, code.Forbidden
}

func (s *RelationGroupGrpc) CreateGroup(ctx context.Context, groupID uint32, currentUserID string, memberIDs []string) (uint32, error) {
	resp, err := s.client.CreateGroupAndInviteUsers(ctx, &relationgrpcv1.CreateGroupAndInviteUsersRequest{
		GroupId: groupID,
		UserID:  currentUserID,
		Member:  memberIDs,
	})
	if err != nil {
		return 0, err
	}

	return resp.DialogId, nil
}

func (s *RelationGroupGrpc) CreateGroupRevert(ctx context.Context, groupID uint32, currentUserID string, memberIDs []string) error {
	_, err := s.client.CreateGroupAndInviteUsersRevert(ctx, &relationgrpcv1.CreateGroupAndInviteUsersRequest{
		GroupId: groupID,
		UserID:  currentUserID,
		Member:  memberIDs,
	})
	if err != nil {
		return err
	}
	return nil
}
