package rpc

import (
	"context"
	"fmt"
	relationgrpcv1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	"github.com/cossim/coss-server/internal/user/domain/entity"
	"google.golang.org/grpc"
)

type RelationUserService interface {
	GetUserRelation(ctx context.Context, userID string, friendID string) (*entity.Relation, error)
	EstablishFriendship(ctx context.Context, userID string, friendID string) error
}

var _ RelationUserService = &relationUserGrpc{}

func NewRelationUserGrpcWithClient(client relationgrpcv1.UserRelationServiceClient) RelationUserService {
	return &relationUserGrpc{client: client}
}

func NewRelationUserGrpc(addr string) (RelationUserService, error) {
	var grpcOptions = []grpc.DialOption{grpc.WithInsecure()}
	grpcOptions = append(grpcOptions, grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`))
	conn, err := grpc.Dial(
		addr,
		grpcOptions...,
	)
	if err != nil {
		return nil, err
	}

	return &relationUserGrpc{client: relationgrpcv1.NewUserRelationServiceClient(conn)}, nil
}

type relationUserGrpc struct {
	client relationgrpcv1.UserRelationServiceClient
}

func (s *relationUserGrpc) EstablishFriendship(ctx context.Context, userID string, friendID string) error {
	_, err := s.client.AddFriend(ctx, &relationgrpcv1.AddFriendRequest{
		UserId:   userID,
		FriendId: friendID,
	})

	return err
}

func (s *relationUserGrpc) GetUserRelation(ctx context.Context, userID string, friendID string) (*entity.Relation, error) {
	fmt.Println("relation.GetUserRelation => ", userID, friendID)

	relation, err := s.client.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{
		UserId:   userID,
		FriendId: friendID,
	})
	if err != nil {
		return nil, err
	}

	resp := &entity.Relation{
		DialogID: relation.DialogId,
	}

	if relation.Status == relationgrpcv1.RelationStatus_RELATION_NORMAL {
		resp.Status = entity.UserRelationStatusFriend
	} else if relation.Status == relationgrpcv1.RelationStatus_RELATION_STATUS_BLOCKED {
		resp.Status = entity.UserRelationStatusBlacked
	} else {
		resp.Status = entity.UserRelationStatusNone
	}

	fmt.Println("relation.Remark => ", relation.Remark)

	resp.Remark = relation.Remark
	resp.SilentNotification = relation.IsSilent
	resp.OpenBurnAfterReading = relation.OpenBurnAfterReading
	resp.OpenBurnAfterReadingTimeOut = relation.OpenBurnAfterReadingTimeOut

	return resp, nil
}
