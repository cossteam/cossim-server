package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	groupgrpcv1 "github.com/cossim/coss-server/internal/group/api/grpc/v1"
	"github.com/cossim/coss-server/internal/group/api/http/model"
	relationgrpcv1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	storagev1 "github.com/cossim/coss-server/internal/storage/api/grpc/v1"
	usergrpcv1 "github.com/cossim/coss-server/internal/user/api/grpc/v1"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/constants"
	"github.com/cossim/coss-server/pkg/msg_queue"
	myminio "github.com/cossim/coss-server/pkg/storage/minio"
	httputil "github.com/cossim/coss-server/pkg/utils/http"
	"github.com/cossim/coss-server/pkg/utils/time"
	"github.com/cossim/coss-server/pkg/utils/usersorter"
	"github.com/dtm-labs/client/dtmcli"
	"github.com/dtm-labs/client/dtmgrpc"
	"github.com/dtm-labs/client/workflow"
	"github.com/google/uuid"
	"github.com/lithammer/shortuuid/v3"
	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"io/ioutil"
	"mime/multipart"
	"sync"
)

func (s *Service) CreateGroup(ctx context.Context, req *groupgrpcv1.Group) (*model.CreateGroupResponse, error) {
	var err error
	friends, err := s.relationUserClient.GetUserRelationByUserIds(ctx, &relationgrpcv1.GetUserRelationByUserIdsRequest{UserId: req.CreatorId, FriendIds: req.Member})
	if err != nil {
		s.logger.Error("获取好友关系失败", zap.Error(err))
		return nil, code.RelationErrCreateGroupFailed
	}

	if len(req.Member) != len(friends.Users) {
		return nil, code.RelationUserErrFriendRelationNotFound
	}
	for _, friend := range friends.Users {
		if friend.Status != relationgrpcv1.RelationStatus_RELATION_NORMAL {
			return nil, code.StatusNotAvailable
		}
	}

	r1 := &groupgrpcv1.CreateGroupRequest{Group: &groupgrpcv1.Group{
		Type:            req.Type,
		MaxMembersLimit: req.MaxMembersLimit,
		CreatorId:       req.CreatorId,
		Name:            req.Name,
		Avatar:          req.Avatar,
		Member:          req.Member,
	}}

	r22 := &relationgrpcv1.CreateGroupAndInviteUsersRequest{
		UserID: req.CreatorId,
		Member: req.Member,
	}

	resp1 := &groupgrpcv1.Group{}
	var groupID uint32
	var DialogID uint32
	// 创建 DTM 分布式事务工作流
	workflow.InitGrpc(s.dtmGrpcServer, s.groupGrpcServer, grpc.NewServer())
	gid := shortuuid.New()
	wfName := "create_group_workflow_" + gid
	if err = workflow.Register(wfName, func(wf *workflow.Workflow, data []byte) error {
		// 创建群聊
		resp1, err = s.groupClient.CreateGroup(wf.Context, r1)
		if err != nil {
			return err
		}
		wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
			_, err = s.groupClient.CreateGroupRevert(wf.Context, &groupgrpcv1.CreateGroupRequest{Group: &groupgrpcv1.Group{
				Id: resp1.Id,
			}})
			return err
		})
		groupID = resp1.Id
		r22.GroupId = groupID
		r22.Member = req.Member
		r22.UserID = req.CreatorId
		resp2, err := s.relationGroupClient.CreateGroupAndInviteUsers(wf.Context, r22)
		if err != nil {
			return err
		}
		DialogID = resp2.DialogId
		wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
			_, err = s.relationGroupClient.CreateGroupAndInviteUsersRevert(wf.Context, r22)
			return err
		})

		return err
	}); err != nil {
		s.logger.Error("WorkFlow CreateGroup", zap.Error(err))
		return nil, code.RelationErrCreateGroupFailed
	}
	if err = workflow.Execute(wfName, gid, nil); err != nil {
		s.logger.Error("WorkFlow CreateGroup", zap.Error(err))
		return nil, code.RelationErrCreateGroupFailed
	}

	if s.cache {
		err = s.insertRedisGroupList(req.CreatorId, usersorter.CustomGroupData{
			GroupID:  groupID,
			Name:     resp1.Name,
			Avatar:   resp1.Avatar,
			Status:   uint(resp1.Status),
			DialogId: DialogID,
		})
		if err != nil {
			s.logger.Error("CreateGroup", zap.Error(err))
			return nil, code.RelationErrCreateGroupFailed
		}

		dialog, err := s.relationDialogClient.GetDialogById(ctx, &relationgrpcv1.GetDialogByIdRequest{DialogId: DialogID})
		if err != nil {
			s.logger.Error("CreateGroup", zap.Error(err))
			return nil, code.RelationErrCreateGroupFailed
		}

		err = s.insertRedisUserDialogList(req.CreatorId, model.UserDialogListResponse{
			DialogId:          DialogID,
			GroupId:           groupID,
			DialogType:        model.ConversationType(dialog.Type),
			DialogName:        req.Name,
			DialogAvatar:      req.Avatar,
			DialogUnreadCount: 0,
			LastMessage:       model.Message{},
			DialogCreateAt:    dialog.CreateAt,
			TopAt:             0,
		})
		if err != nil {
			s.logger.Error("CreateGroup", zap.Error(err))
			return nil, code.RelationErrCreateGroupFailed
		}
	}

	// 给被邀请的用户推送
	for _, id := range req.Member {
		msg := constants.WsMsg{Uid: id, Event: constants.InviteJoinGroupEvent, Data: map[string]interface{}{"group_id": groupID, "inviter_id": req.CreatorId}, SendAt: time.Now()}
		//通知消息服务有消息需要发送
		err = s.rabbitMQClient.PublishServiceMessage(msg_queue.RelationService, msg_queue.MsgService, msg_queue.Service_Exchange, msg_queue.SendMessage, msg)
	}

	return &model.CreateGroupResponse{
		Id:              resp1.Id,
		Avatar:          resp1.Avatar,
		Name:            resp1.Name,
		Type:            uint32(resp1.Type),
		Status:          int32(resp1.Status),
		MaxMembersLimit: resp1.MaxMembersLimit,
		CreatorId:       resp1.CreatorId,
		DialogId:        DialogID,
	}, nil
}

func (s *Service) DeleteGroup(ctx context.Context, groupID uint32, userID string) (uint32, error) {
	_, err := s.groupClient.GetGroupInfoByGid(ctx, &groupgrpcv1.GetGroupInfoRequest{
		Gid: groupID,
	})
	if err != nil {
		return 0, code.GroupErrGroupNotFound
	}
	sf, err := s.relationGroupClient.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{
		UserId:  userID,
		GroupId: groupID,
	})
	if err != nil {
		return 0, err
	}
	if sf.Identity == relationgrpcv1.GroupIdentity_IDENTITY_USER {
		return 0, code.Forbidden
	}
	dialog, err := s.relationDialogClient.GetDialogByGroupId(ctx, &relationgrpcv1.GetDialogByGroupIdRequest{GroupId: groupID})
	if err != nil {
		return 0, err
	}

	//查询所有群员
	relation, err := s.relationGroupClient.GetGroupUserIDs(ctx, &relationgrpcv1.GroupIDRequest{
		GroupId: groupID,
	})
	if err != nil {
		return 0, err
	}

	r1 := &relationgrpcv1.DeleteDialogUsersByDialogIDRequest{
		DialogId: dialog.DialogId,
	}
	r2 := &relationgrpcv1.DeleteDialogByIdRequest{
		DialogId: dialog.DialogId,
	}
	r3 := &relationgrpcv1.GroupIDRequest{
		GroupId: groupID,
	}
	r4 := &groupgrpcv1.DeleteGroupRequest{
		Gid: groupID,
	}
	gid := shortuuid.New()
	if err = dtmgrpc.TccGlobalTransaction(s.dtmGrpcServer, gid, func(tcc *dtmgrpc.TccGrpc) error {
		r := &emptypb.Empty{}
		// 删除对话用户
		if err = tcc.CallBranch(r1, s.relationGrpcServer+relationgrpcv1.DialogService_DeleteDialogUsersByDialogID_FullMethodName, "", s.relationGrpcServer+relationgrpcv1.DialogService_DeleteDialogUsersByDialogIDRevert_FullMethodName, r); err != nil {
			return err
		}
		// 删除对话
		if err = tcc.CallBranch(r2, s.relationGrpcServer+relationgrpcv1.DialogService_DeleteDialogById_FullMethodName, "", s.relationGrpcServer+relationgrpcv1.DialogService_DeleteDialogByIdRevert_FullMethodName, r); err != nil {
			return err
		}
		// 删除群聊成员
		if err = tcc.CallBranch(r3, s.relationGrpcServer+relationgrpcv1.GroupRelationService_DeleteGroupRelationByGroupId_FullMethodName, "", s.relationGrpcServer+relationgrpcv1.GroupRelationService_DeleteGroupRelationByGroupIdRevert_FullMethodName, r); err != nil {
			return err
		}
		// 删除群聊
		if err = tcc.CallBranch(r4, s.groupGrpcServer+groupgrpcv1.GroupService_DeleteGroup_FullMethodName, "", s.groupGrpcServer+groupgrpcv1.GroupService_DeleteGroupRevert_FullMethodName, r); err != nil {
			return err
		}
		return err
	}); err != nil {
		s.logger.Error("WorkFlow DeleteGroup", zap.Error(err))
		return 0, code.GroupErrDeleteGroupFailed
	}

	if s.cache {
		for _, res := range relation.UserIds {
			err := s.removeRedisGroupList(res, groupID)
			if err != nil {
				return 0, err
			}

			err = s.removeRedisUserDialogList(res, dialog.DialogId)
			if err != nil {
				return 0, err
			}
		}
	}

	return groupID, err
}

func (s *Service) UpdateGroup(ctx context.Context, req *model.UpdateGroupRequest, userID string) (interface{}, error) {
	group, err := s.groupClient.GetGroupInfoByGid(ctx, &groupgrpcv1.GetGroupInfoRequest{
		Gid: req.GroupId,
	})
	if err != nil {
		s.logger.Error("更新群聊信息失败", zap.Error(err))
		return nil, err
	}

	sf, err := s.relationGroupClient.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{
		UserId:  userID,
		GroupId: req.GroupId,
	})
	if err != nil {
		s.logger.Error("获取用户群聊关系失败", zap.Error(err))
		return nil, err
	}

	if sf.Identity == relationgrpcv1.GroupIdentity_IDENTITY_USER {
		return nil, code.Unauthorized
	}

	group.Type = groupgrpcv1.GroupType(req.Type)
	group.Name = req.Name
	group.Avatar = req.Avatar
	group.Id = req.GroupId
	switch req.Type {
	case uint32(groupgrpcv1.GroupType_TypeEncrypted):
		group.MaxMembersLimit = model.EncryptedGroup
	default:
		group.MaxMembersLimit = model.DefaultGroup
	}

	resp, err := s.groupClient.UpdateGroup(ctx, &groupgrpcv1.UpdateGroupRequest{
		Group: group,
	})
	if err != nil {
		s.logger.Error("更新群聊信息失败", zap.Error(err))
		return nil, err
	}

	if s.cache {
		//查询所有群员
		relation, err := s.relationGroupClient.GetBatchGroupRelation(ctx, &relationgrpcv1.GetBatchGroupRelationRequest{
			GroupId: group.Id,
		})
		if err != nil {
			return 0, err
		}

		wg := sync.WaitGroup{}
		for _, res := range relation.GroupRelationResponses {
			go func(id string) {
				defer wg.Done()
				wg.Add(1)
				err := s.removeRedisGroupList(id, group.Id)
				if err != nil {
					return
				}
			}(res.UserId)
		}

		wg.Wait()
	}

	return resp, nil
}

func (s *Service) GetBatchGroupInfoByIDs(ctx context.Context, ids []uint32) (interface{}, error) {
	groups, err := s.groupClient.GetBatchGroupInfoByIDs(ctx, &groupgrpcv1.GetBatchGroupInfoRequest{
		GroupIds: ids,
	})
	if err != nil {
		s.logger.Error("批量获取群聊信息失败", zap.Error(err))
		return nil, err
	}

	return groups.Groups, nil
}

func (s *Service) GetGroupInfoByGid(ctx context.Context, gid uint32, userID string) (interface{}, error) {
	group, err := s.groupClient.GetGroupInfoByGid(ctx, &groupgrpcv1.GetGroupInfoRequest{
		Gid: gid,
	})
	if err != nil {
		return nil, err
	}

	id, err := s.relationDialogClient.GetDialogByGroupId(ctx, &relationgrpcv1.GetDialogByGroupIdRequest{GroupId: gid})
	if err != nil {
		return nil, err
	}

	relation, err := s.relationGroupClient.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{
		GroupId: gid,
		UserId:  userID,
	})
	if err != nil {
		return nil, err
	}

	per := &model.Preferences{
		OpenBurnAfterReading: model.OpenBurnAfterReadingType(relation.OpenBurnAfterReading),
		SilentNotification:   model.SilentNotification(relation.IsSilent),
		Remark:               relation.Remark,
		EntryMethod:          model.EntryMethod(relation.JoinMethod),
		Inviter:              relation.Inviter,
		JoinedAt:             relation.JoinTime,
		MuteEndTime:          relation.MuteEndTime,
		Identity:             model.GroupIdentity(relation.Identity),
	}

	return &model.GroupInfo{
		Id:              group.Id,
		Avatar:          group.Avatar,
		Name:            group.Name,
		Type:            uint32(group.Type),
		Status:          int32(group.Status),
		MaxMembersLimit: group.MaxMembersLimit,
		CreatorId:       group.CreatorId,
		DialogId:        id.DialogId,
		Preferences:     per,
	}, nil
}

func (s *Service) insertRedisGroupList(userID string, msg usersorter.CustomGroupData) error {
	key := fmt.Sprintf("group:%s", userID)
	exists, err := s.redisClient.ExistsKey(key)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	err = s.redisClient.AddToListLeft(key, []interface{}{msg})
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) removeRedisGroupList(userID string, groupID uint32) error {
	key := fmt.Sprintf("group:%s", userID)
	//判断key是否存在，存在才继续
	f, err := s.redisClient.ExistsKey(key)
	if err != nil {
		return err
	}
	if !f {
		return nil
	}

	length, err := s.redisClient.GetListLength(key)
	if err != nil {
		return err
	}

	if length > 10 {
		for i := int64(0); i < length; i += 10 {
			stop := i + 9
			if stop >= length {
				stop = length - 1
			}

			// 获取当前范围内的元素
			values, err := s.redisClient.GetList(key, i, stop)
			if err != nil {
				return err
			}

			// 遍历当前范围内的元素
			for j, v := range values {
				var group usersorter.CustomGroupData
				err := json.Unmarshal([]byte(v), &group)
				if err != nil {
					fmt.Println("Error decoding cached data:", err)
					return err
				}
				if group.GroupID == groupID {
					// 弹出指定位置的元素
					_, err := s.redisClient.PopListElement(key, i+int64(j))
					if err != nil {
						return err
					}
				}
			}
		}
	} else {
		values, err := s.redisClient.GetAllListValues(key)
		if err != nil {
			return err
		}
		for i, v := range values {
			var group usersorter.CustomGroupData
			err := json.Unmarshal([]byte(v), &group)
			if err != nil {
				fmt.Println("Error decoding cached data:", err)
				return err
			}
			if group.GroupID == groupID {
				_, err := s.redisClient.PopListElement(key, int64(i))
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// 更新redis里的对话列表数据
func (s *Service) insertRedisUserDialogList(userID string, msg model.UserDialogListResponse) error {
	key := fmt.Sprintf("dialog:%s", userID)
	//判断key是否存在，存在才继续
	f, err := s.redisClient.ExistsKey(key)
	if err != nil {
		return err
	}
	if !f {
		return nil
	}

	err = s.redisClient.AddToListLeft(key, []interface{}{msg})
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) removeRedisUserDialogList(userID string, dialogID uint32) error {
	key := fmt.Sprintf("dialog:%s", userID)
	//判断key是否存在，存在才继续
	f, err := s.redisClient.ExistsKey(key)
	if err != nil {
		return err
	}
	if !f {
		return nil
	}

	length, err := s.redisClient.GetListLength(key)
	if err != nil {
		return err
	}

	if length > 10 {
		for i := int64(0); i < length; i += 10 {
			stop := i + 9
			if stop >= length {
				stop = length - 1
			}

			// 获取当前范围内的元素
			values, err := s.redisClient.GetList(key, i, stop)
			if err != nil {
				return err
			}

			// 遍历当前范围内的元素
			for j, v := range values {
				var dialog model.UserDialogListResponse
				err := json.Unmarshal([]byte(v), &dialog)
				if err != nil {
					fmt.Println("Error decoding cached data:", err)
					return err
				}
				if dialog.DialogId == dialogID {
					// 弹出指定位置的元素
					_, err := s.redisClient.PopListElement(key, i+int64(j))
					if err != nil {
						return err
					}
				}
			}
		}
	} else {
		values, err := s.redisClient.GetAllListValues(key)
		if err != nil {
			return err
		}
		for i, v := range values {
			var dialog model.UserDialogListResponse
			err := json.Unmarshal([]byte(v), &dialog)
			if err != nil {
				fmt.Println("Error decoding cached data:", err)
				return err
			}
			if dialog.DialogId == dialogID {
				_, err := s.redisClient.PopListElement(key, int64(i))
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (s *Service) ModifyGroupAvatar(ctx context.Context, userID string, groupID uint32, avatar multipart.File) (string, error) {
	_, err := s.groupClient.GetGroupInfoByGid(ctx, &groupgrpcv1.GetGroupInfoRequest{
		Gid: groupID,
	})
	if err != nil {
		return "", err
	}

	relation, err := s.relationGroupClient.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{
		UserId:  userID,
		GroupId: groupID,
	})
	if err != nil {
		return "", err
	}

	if relation.Identity != relationgrpcv1.GroupIdentity_IDENTITY_OWNER && relation.Identity != relationgrpcv1.GroupIdentity_IDENTITY_ADMIN {
		return "", code.Forbidden
	}

	_, err = s.userClient.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
		UserId: userID,
	})
	if err != nil {
		return "", err
	}

	data, err := ioutil.ReadAll(avatar)
	if err != nil {
		return "", err
	}

	bucket, err := myminio.GetBucketName(int(storagev1.FileType_Other))
	if err != nil {
		return "", err
	}

	// 将字节数组转换为 io.Reader
	reader := bytes.NewReader(data)
	fileID := uuid.New().String()
	key := myminio.GenKey(bucket, fileID+".jpeg")
	err = s.sp.UploadAvatar(ctx, key, reader, reader.Size(), minio.PutObjectOptions{
		ContentType: "image/jpeg",
	})
	if err != nil {
		return "", err
	}

	aUrl := fmt.Sprintf("http://%s%s/%s", s.gatewayAddress, s.downloadURL, key)
	if s.conf.SystemConfig.Ssl {
		aUrl, err = httputil.ConvertToHttps(aUrl)
		if err != nil {
			return "", err
		}
	}

	_, err = s.groupClient.UpdateGroup(ctx, &groupgrpcv1.UpdateGroupRequest{
		Group: &groupgrpcv1.Group{
			Id:     groupID,
			Avatar: aUrl,
		},
	})
	if err != nil {
		return "", err
	}

	if s.cache {
		//获取所有群成员
		members, err := s.relationGroupClient.GetGroupUserIDs(ctx, &relationgrpcv1.GroupIDRequest{
			GroupId: groupID,
		})

		for _, id := range members.UserIds {
			err = s.redisClient.DelKey(fmt.Sprintf("dialog:%s", id))
			if err != nil {
				return "", err
			}

			err = s.redisClient.DelKey(fmt.Sprintf("group:%s", id))
			if err != nil {
				return "", err
			}
		}
	}

	return aUrl, nil
}
