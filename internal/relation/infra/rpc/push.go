package rpc

import (
	"context"
	pushgrpcv1 "github.com/cossim/coss-server/internal/push/api/grpc/v1"
	"github.com/cossim/coss-server/internal/relation/api/http/model"
	"github.com/cossim/coss-server/pkg/constants"
	"github.com/cossim/coss-server/pkg/utils"
	ptime "github.com/cossim/coss-server/pkg/utils/time"
	any2 "github.com/golang/protobuf/ptypes/any"
	"google.golang.org/grpc"
)

type PushService interface {
	// ManageGroupRequestPush 管理群组请求推送
	ManageGroupRequestPush(ctx context.Context, groupID uint32, targetUserID string, action uint32) error
	// AddGroupRequestPush 添加群组请求推送
	AddGroupRequestPush(ctx context.Context, groupID uint32, adminIDs []string, targetUserIDs string) error
	// AddFriendPush 添加好友事件推送
	AddFriendPush(ctx context.Context, userID, targetUserID, remark, e2ePublicKey string) error
	// ExchangeE2eKeyPush 用户公钥交换事件推送
	ExchangeE2eKeyPush(ctx context.Context, userID, targetUserID, e2ePublicKey string) error
	// InviteJoinGroupPush 邀请加入群组事件推送
	InviteJoinGroupPush(ctx context.Context, groupID uint32, inviterID string, adminIDs, memberIDs []string) error
}

func NewPushGrpc(addr string) (PushService, error) {
	var grpcOptions = []grpc.DialOption{grpc.WithInsecure()}
	grpcOptions = append(grpcOptions, grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`))
	conn, err := grpc.Dial(
		addr,
		grpcOptions...,
	)
	if err != nil {
		return nil, err
	}
	return &PushServiceGrpc{client: pushgrpcv1.NewPushServiceClient(conn)}, nil
}

func NewPushServiceWithClient(client pushgrpcv1.PushServiceClient) PushService {
	return &PushServiceGrpc{client: client}
}

type PushServiceGrpc struct {
	client pushgrpcv1.PushServiceClient
	//logger *zap.Logger
}

func (s *PushServiceGrpc) ManageGroupRequestPush(ctx context.Context, groupID uint32, targetUserID string, action uint32) error {
	data := constants.JoinGroupEventData{GroupId: groupID, UserId: targetUserID, Status: action}
	bytes, err := utils.StructToBytes(data)
	if err != nil {
		return err
	}
	msg := &pushgrpcv1.WsMsg{
		Uid:         targetUserID,
		Event:       pushgrpcv1.WSEventType_JoinGroupEvent,
		Data:        &any2.Any{Value: bytes},
		SendAt:      ptime.Now(),
		PushOffline: true,
	}

	msgBytes, err := utils.StructToBytes(msg)
	if err != nil {
		return err
	}

	_, err = s.client.Push(ctx, &pushgrpcv1.PushRequest{Data: msgBytes, Type: pushgrpcv1.Type_Ws})
	if err != nil {
		return err
	}

	return nil
}

func (s *PushServiceGrpc) buildJoinGroupPushMsgs(ctx context.Context, groupID uint32, targetUserID string, recipientIDs []string) ([]byte, error) {
	data := &constants.JoinGroupEventData{
		GroupId: groupID,
		UserId:  targetUserID,
	}
	bytes, err := utils.StructToBytes(data)
	if err != nil {
		return nil, err
	}

	var msgs []*pushgrpcv1.WsMsg
	for _, id := range recipientIDs {
		msg := &pushgrpcv1.WsMsg{
			Uid:         id,
			Event:       pushgrpcv1.WSEventType_JoinGroupEvent,
			Data:        &any2.Any{Value: bytes},
			SendAt:      ptime.Now(),
			PushOffline: true,
		}
		msgs = append(msgs, msg)
	}

	toBytes, err := utils.StructToBytes(msgs)
	if err != nil {
		return nil, err
	}

	return toBytes, nil
}

func (s *PushServiceGrpc) AddGroupRequestPush(ctx context.Context, groupID uint32, adminIDs []string, targetUserID string) error {
	toBytes, err := s.buildJoinGroupPushMsgs(ctx, groupID, targetUserID, adminIDs)
	if err != nil {
		return err
	}

	// 推送消息给管理员
	_, err = s.client.Push(ctx, &pushgrpcv1.PushRequest{
		Type: pushgrpcv1.Type_Ws_Batch_User,
		Data: toBytes,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *PushServiceGrpc) InviteJoinGroupPush(ctx context.Context, groupID uint32, inviterID string, adminIDs, memberIDs []string) error {
	toBytes, err := s.buildJoinGroupPushMsgs(ctx, groupID, inviterID, adminIDs)
	if err != nil {
		return err
	}

	// 推送消息给管理员
	_, err = s.client.Push(ctx, &pushgrpcv1.PushRequest{
		Type: pushgrpcv1.Type_Ws_Batch_User,
		Data: toBytes,
	})
	if err != nil {
		return err
	}

	data2 := map[string]interface{}{"group_id": groupID, "inviter_id": inviterID}
	bytes2, err := utils.StructToBytes(data2)
	if err != nil {
		return err
	}

	// 推送消息给被邀请的用户
	for _, id := range memberIDs {
		msg := &pushgrpcv1.WsMsg{
			Uid:    id,
			Event:  pushgrpcv1.WSEventType_InviteJoinGroupEvent,
			Data:   &any2.Any{Value: bytes2},
			SendAt: ptime.Now(),
		}
		structToBytes, err := utils.StructToBytes(msg)
		if err != nil {
			return err
		}
		_, err = s.client.Push(ctx, &pushgrpcv1.PushRequest{
			Type: pushgrpcv1.Type_Ws,
			Data: structToBytes,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *PushServiceGrpc) ExchangeE2eKeyPush(ctx context.Context, userID, targetUserID, e2ePublicKey string) error {
	bytes, err := utils.StructToBytes(&model.SwitchUserE2EPublicKeyRequest{
		UserId:    userID,
		PublicKey: e2ePublicKey,
	})
	if err != nil {
		return err
	}
	msg := &pushgrpcv1.WsMsg{
		Uid:    targetUserID,
		Event:  pushgrpcv1.WSEventType_PushE2EPublicKeyEvent,
		Data:   &any2.Any{Value: bytes},
		SendAt: ptime.Now(),
	}
	toBytes, err := utils.StructToBytes(msg)
	if err != nil {
		return err
	}

	_, err = s.client.Push(ctx, &pushgrpcv1.PushRequest{
		Type: pushgrpcv1.Type_Ws,
		Data: toBytes,
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *PushServiceGrpc) AddFriendPush(ctx context.Context, userID, friendID, remark, e2ePublicKey string) error {
	wsMsgData := constants.AddFriendEventData{
		UserId:       userID,
		Msg:          remark,
		E2EPublicKey: e2ePublicKey,
	}
	bytes, err := utils.StructToBytes(wsMsgData)
	if err != nil {
		return err
	}

	msg := &pushgrpcv1.WsMsg{Uid: friendID, Event: pushgrpcv1.WSEventType_AddFriendEvent, Data: &any2.Any{Value: bytes}}

	toBytes, err := utils.StructToBytes(msg)
	if err != nil {
		return err
	}

	_, err = s.client.Push(ctx, &pushgrpcv1.PushRequest{Data: toBytes})
	if err != nil {
		return err
	}

	return nil
}
