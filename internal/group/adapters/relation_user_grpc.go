package adapters

import (
	"context"
	"fmt"
	"github.com/cossim/coss-server/internal/group/app/command"
	relationgrpcv1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	"github.com/cossim/coss-server/pkg/code"
)

//var _ command.RelationUserService = &RelationUserGrpc{}

func NewRelationUserGrpc(client relationgrpcv1.UserRelationServiceClient) *RelationUserGrpc {
	return &RelationUserGrpc{client: client}
}

type RelationUserGrpc struct {
	client relationgrpcv1.UserRelationServiceClient
	//logger *zap.Logger
}

func (s *RelationUserGrpc) IsFriendsWithAll(ctx context.Context, currentUserID string, invitedUserIDs []string) (bool, error) {
	friends, err := s.client.GetUserRelationByUserIds(ctx, &relationgrpcv1.GetUserRelationByUserIdsRequest{UserId: currentUserID, FriendIds: invitedUserIDs})
	if err != nil {
		return false, code.RelationErrCreateGroupFailed
	}

	for _, friend := range friends.Users {
		if friend.Status != relationgrpcv1.RelationStatus_RELATION_NORMAL {
			return false, code.StatusNotAvailable.CustomMessage(fmt.Sprintf("%s不是你的好友", friend.Remark))
		}
	}

	return true, nil
}

func (s *RelationUserGrpc) GetUserRelationships(ctx context.Context, currentUserID string, userIDs []string) (map[string]command.UserRelationship, error) {
	friends, err := s.client.GetUserRelationByUserIds(ctx, &relationgrpcv1.GetUserRelationByUserIdsRequest{UserId: currentUserID, FriendIds: userIDs})
	if err != nil {
		return nil, code.RelationErrCreateGroupFailed
	}

	relations := make(map[string]command.UserRelationship)

	for _, friend := range friends.Users {
		relations[friend.FriendId] = command.UserRelationship{
			ID:     friend.UserId,
			Status: uint(friend.Status),
			Remark: friend.Remark,
		}
	}

	return relations, nil
}
