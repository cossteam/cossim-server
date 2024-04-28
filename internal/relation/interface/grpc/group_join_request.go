package grpc

import (
	"context"
	"fmt"
	v1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	"github.com/cossim/coss-server/internal/relation/cache"
	"github.com/cossim/coss-server/internal/relation/domain/relation"
	"github.com/cossim/coss-server/internal/relation/infra/persistence"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/utils"
	"github.com/cossim/coss-server/pkg/utils/time"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
	"log"
)

var _ v1.GroupJoinRequestServiceServer = &groupJoinRequestServiceServer{}

type groupJoinRequestServiceServer struct {
	db          *gorm.DB
	cache       cache.RelationGroupCache
	cacheEnable bool
	dr          relation.DialogRepository
	grr         relation.GroupRepository
	gjqr        relation.GroupJoinRequestRepository
}

func (s *groupJoinRequestServiceServer) InviteJoinGroup(ctx context.Context, request *v1.InviteJoinGroupRequest) (*v1.JoinGroupResponse, error) {
	resp := &v1.JoinGroupResponse{}

	// 获取群组管理员 ID
	adminIDs, err := s.grr.ListGroupAdmin(ctx, request.GroupId)
	if err != nil {
		return resp, status.Error(codes.Code(code.RelationGroupErrInviteFailed.Code()), err.Error())
	}

	// 构建关系切片
	relations := make([]*relation.GroupJoinRequest, 0)
	var notifiy []string
	notifiy = append(notifiy, request.InviterId)
	notifiy = append(notifiy, adminIDs...)

	//去除重复
	notifiy = utils.RemoveDuplicate(notifiy)

	fmt.Println("管理员ids", notifiy)
	fmt.Println("原来的ids", request.Member)
	for _, userID := range request.Member {
		fmt.Println("添加被邀请人记录")

		userGroup := &relation.GroupJoinRequest{
			UserID:      userID,
			GroupID:     request.GroupId,
			Inviter:     request.InviterId,
			OwnerID:     userID,
			InviterTime: time.Now(),
			Status:      relation.Invitation,
		}
		relations = append(relations, userGroup)
	}

	// 将管理员添加到关系切片中
	for _, id := range notifiy {
		for _, userID := range request.Member {
			fmt.Println("添加管理员记录")
			userGroup := &relation.GroupJoinRequest{
				UserID:      userID,
				GroupID:     request.GroupId,
				Inviter:     request.InviterId,
				OwnerID:     id,
				InviterTime: time.Now(),
				Status:      relation.Invitation,
			}
			relations = append(relations, userGroup)
		}
	}

	// 添加关系切片到数据库
	_, err = s.gjqr.Creates(ctx, relations)
	if err != nil {
		return resp, status.Error(codes.Code(code.RelationGroupErrInviteFailed.Code()), err.Error())
	}

	request.Member = append(request.Member, request.InviterId)
	if s.cacheEnable {
		if err := s.cache.DeleteGroupJoinRequestListByUser(ctx, request.Member...); err != nil {
			log.Printf("delete group join request list by user failed, err: %v", err)
		}
	}

	return resp, nil
}

func (s *groupJoinRequestServiceServer) JoinGroup(ctx context.Context, request *v1.JoinGroupRequest) (*v1.JoinGroupResponse, error) {
	resp := &v1.JoinGroupResponse{}
	relations := make([]*relation.GroupJoinRequest, 0)

	// 获取对话id
	//dialog, err := s.dr.GetDialogByGroupId(uint(request.GroupId))
	//if err != nil {
	//	return resp, status.Error(codes.Code(code.RelationGroupErrManageJoinFailed.Code()), err.Error())
	//}

	dialog, err := s.dr.GetByGroupID(ctx, request.GroupId)
	if err != nil {
		return nil, err
	}

	if !request.JoinApprove {
		if err := s.db.Transaction(func(tx *gorm.DB) error {
			npo := persistence.NewRepositories(tx)

			// 加入对话
			//_, err := npo.Dr.JoinDialog(dialog.ID, request.UserId)
			//if err != nil {
			//	return err
			//}

			if _, err := npo.Dur.Create(ctx, &relation.CreateDialogUser{
				DialogID: uint32(dialog.ID),
				UserID:   request.UserId,
			}); err != nil {
				return err
			}

			gr := &relation.CreateGroupRelation{
				GroupID:     request.GroupId,
				UserID:      request.UserId,
				Identity:    relation.IdentityUser,
				JoinedAt:    time.Now(),
				EntryMethod: relation.EntrySearch,
			}
			// 加入群聊
			if _, err := s.grr.Create(ctx, gr); err != nil {
				return err
			}

			return nil
		}); err != nil {
			return nil, err
		}

		if s.cacheEnable {
			if err := s.cache.DeleteRelation(ctx, request.UserId, request.GroupId); err != nil {
				log.Printf("delete group relation list by user failed, err: %v", err)
			}
		}

		return nil, nil
	}

	// 添加管理员群聊申请记录
	ids, err := s.grr.ListGroupAdmin(ctx, request.GroupId)
	if err != nil {
		return resp, err
	}
	for _, id := range ids {
		userGroup := &relation.GroupJoinRequest{
			UserID:      id,
			GroupID:     request.GroupId,
			Remark:      request.Msg,
			OwnerID:     id,
			InviterTime: time.Now(),
			Status:      relation.Pending,
		}
		relations = append(relations, userGroup)
	}

	// 添加用户群聊申请记录
	ur := &relation.GroupJoinRequest{
		GroupID:     request.GroupId,
		UserID:      request.UserId,
		Remark:      request.Msg,
		OwnerID:     request.UserId,
		InviterTime: time.Now(),
		Status:      relation.Pending,
	}
	relations = append(relations, ur)

	_, err = s.gjqr.Creates(ctx, relations)
	if err != nil {
		return resp, status.Error(codes.Code(code.RelationGroupErrInviteFailed.Code()), err.Error())
	}

	return resp, nil
}

func (s *groupJoinRequestServiceServer) GetGroupJoinRequestListByUserId(ctx context.Context, request *v1.GetGroupJoinRequestListRequest) (*v1.GroupJoinRequestListResponse, error) {
	var resp = &v1.GroupJoinRequestListResponse{}

	//if s.cacheEnable {
	//	users, err := s.cache.GetGroupJoinRequestListByUser(ctx, request.UserId)
	//	if err == nil && len(users.GroupJoinRequestResponses) != 0 {
	//		return users, nil
	//	}
	//}

	//获取他自己的申请列表
	//list, total, err := s.gjqr.GetGroupJoinRequestListByUserId(request.UserId, int(request.PageSize), int(request.PageNum))
	//if err != nil {
	//	return resp, err
	//}

	list, err := s.gjqr.Find(ctx, &relation.GroupJoinRequestQuery{
		UserID:   []string{request.UserId},
		PageSize: int(request.PageSize),
		PageNum:  int(request.PageNum),
	})
	if err != nil {
		return nil, err
	}

	resp.Total = uint64(len(list))
	if len(list) > 0 {
		for _, v := range list {
			resp.GroupJoinRequestResponses = append(resp.GroupJoinRequestResponses, &v1.GroupJoinRequestResponse{
				ID:        v.ID,
				UserId:    v.UserID,
				GroupId:   v.GroupID,
				Status:    v1.GroupRequestStatus(v.Status),
				InviterId: v.Inviter,
				CreatedAt: uint64(v.CreatedAt),
				Remark:    v.Remark,
			})
		}
	}

	//if s.cacheEnable {
	//	if err := s.cache.SetGroupJoinRequestListByUser(ctx, request.UserId, resp, cache.RelationExpireTime); err != nil {
	//		log.Printf("set group join request list by user failed, err: %v", err)
	//	}
	//}

	return resp, nil
}

func (s *groupJoinRequestServiceServer) GetGroupJoinRequestByGroupIdAndUserId(ctx context.Context, request *v1.GetGroupJoinRequestByGroupIdAndUserIdRequest) (*v1.GetGroupJoinRequestByGroupIdAndUserIdResponse, error) {
	var resp = &v1.GetGroupJoinRequestByGroupIdAndUserIdResponse{}

	find, err := s.gjqr.Find(ctx, &relation.GroupJoinRequestQuery{
		UserID:  []string{request.UserId},
		GroupID: []uint32{request.GroupId},
	})
	if err != nil {
		return nil, err
	}

	if len(find) == 0 {
		return resp, nil
	}

	//req, err := s.gjqr.GetGroupJoinRequestByGroupIdAndUserId(uint(request.GroupId), request.UserId)
	//if err != nil {
	//	return resp, err
	//}
	//if req != nil {
	resp.ID = find[0].ID
	resp.GroupId = find[0].GroupID
	resp.UserId = find[0].UserID
	resp.Status = v1.GroupRequestStatus(find[0].Status)
	resp.CreatedAt = uint64(find[0].CreatedAt)
	resp.InviterId = find[0].Inviter
	resp.Remark = find[0].Remark
	//}
	return resp, nil
}

func (s *groupJoinRequestServiceServer) ManageGroupJoinRequestByID(ctx context.Context, request *v1.ManageGroupJoinRequestByIDRequest) (*emptypb.Empty, error) {
	var resp = &emptypb.Empty{}

	info, err := s.gjqr.Get(ctx, request.ID)
	if err != nil {
		return nil, err
	}

	//info, err := s.gjqr.GetGroupJoinRequestByRequestID(uint(request.ID))
	//if err != nil {
	//	return nil, err
	//}

	//拒绝请求
	if request.Status == v1.GroupRequestStatus_Rejected {
		//if err := s.gjqr.ManageGroupJoinRequestByID(info.GroupID, info.UserID, relation.RequestStatus(request.Status)); err != nil {
		//	return resp, status.Error(codes.Code(code.RelationGroupErrManageJoinFailed.Code()), err.Error())
		//}

		if err := s.gjqr.UpdateStatus(ctx, info.ID, relation.RequestStatus(request.Status)); err != nil {
			return resp, status.Error(codes.Code(code.RelationGroupErrManageJoinFailed.Code()), err.Error())
		}

		//if s.cacheEnable {
		//	if err := s.cache.DeleteGroupJoinRequestListByUser(ctx, info.UserID, info.Inviter); err != nil {
		//		log.Printf("delete group join request list by user failed, err: %v", err)
		//	}
		//}

		return resp, nil
	}

	// 判断用户是否已经加入该群聊
	rel, err := s.grr.GetUserGroupByGroupIDAndUserID(ctx, info.GroupID, info.UserID)
	if err != nil {
		return resp, status.Error(codes.Code(code.RelationGroupErrAlreadyInGroup.Code()), err.Error())
	}
	if rel != nil {
		return resp, status.Error(codes.Code(code.RelationGroupErrAlreadyInGroup.Code()), code.RelationGroupErrAlreadyInGroup.Message())
	}

	// 获取对话id
	//dialog, err := s.dr.GetDialogByGroupId(uint(info.GroupID))
	//if err != nil {
	//	return resp, status.Error(codes.Code(code.RelationGroupErrManageJoinFailed.Code()), err.Error())
	//}

	dialog, err := s.dr.GetByGroupID(ctx, info.GroupID)
	if err != nil {
		return resp, status.Error(codes.Code(code.RelationGroupErrManageJoinFailed.Code()), err.Error())
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		npo := persistence.NewRepositories(tx)

		//加入对话
		//_, err := npo.Dr.JoinDialog(dialog.ID, info.UserID)
		//if err != nil {
		//	return err
		//}

		if _, err := npo.Dur.Create(ctx, &relation.CreateDialogUser{
			DialogID: dialog.ID,
			UserID:   info.UserID,
		}); err != nil {
			return err
		}

		//修改请求状态
		//if err := npo.Gjqr.ManageGroupJoinRequestByID(info.GroupID, info.UserID, relation.RequestStatus(request.Status)); err != nil {
		//	return status.Error(codes.Code(code.RelationGroupErrManageJoinFailed.Code()), err.Error())
		//}
		if err := npo.Gjqr.UpdateStatus(ctx, info.ID, relation.RequestStatus(request.Status)); err != nil {
			return status.Error(codes.Code(code.RelationGroupErrManageJoinFailed.Code()), err.Error())
		}

		gr := &relation.CreateGroupRelation{
			GroupID:     info.GroupID,
			UserID:      info.UserID,
			Identity:    relation.IdentityUser,
			JoinedAt:    time.Now(),
			EntryMethod: relation.EntrySearch,
		}

		if info.Inviter != "" {
			gr.Inviter = info.Inviter
			gr.EntryMethod = relation.EntryInvitation
		}
		//加入群聊
		_, err = npo.Grr.Create(ctx, gr)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return resp, err
	}

	//if s.cacheEnable {
	//	if err := s.cache.DeleteGroupJoinRequestListByUser(ctx, info.UserID, info.Inviter); err != nil {
	//		log.Printf("delete group join request list by user failed, err: %v", err)
	//	}
	//}

	return resp, nil
}

func (s *groupJoinRequestServiceServer) GetGroupJoinRequestByID(ctx context.Context, request *v1.GetGroupJoinRequestByIDRequest) (*v1.GetGroupJoinRequestByIDResponse, error) {
	resp := &v1.GetGroupJoinRequestByIDResponse{}

	if request.ID == 0 {
		return nil, status.Error(codes.Code(code.InvalidParameter.Code()), code.InvalidParameter.Message())
	}

	r, err := s.gjqr.Get(ctx, request.ID)
	if err != nil {
		return nil, status.Error(codes.Code(code.RelationErrGetGroupJoinRequestFailed.Code()), err.Error())
	}

	//r, err := s.gjqr.GetGroupJoinRequestByRequestID(uint(request.ID))
	//if err != nil {
	//	return nil, status.Error(codes.Code(code.RelationErrGetGroupJoinRequestFailed.Code()), err.Error())
	//}
	resp.GroupId = r.GroupID
	resp.UserId = r.UserID
	resp.Status = v1.GroupRequestStatus(r.Status)
	resp.CreatedAt = uint64(r.CreatedAt)
	resp.InviterId = r.Inviter
	resp.Remark = r.Remark
	resp.OwnerID = r.OwnerID
	return resp, nil
}

func (s *groupJoinRequestServiceServer) DeleteGroupRecord(ctx context.Context, req *v1.DeleteGroupRecordRequest) (*emptypb.Empty, error) {
	resp := &emptypb.Empty{}

	if req.ID == 0 || req.UserId == "" {
		return nil, status.Error(codes.Code(code.InvalidParameter.Code()), code.InvalidParameter.Message())
	}

	if err := s.gjqr.Delete(ctx, req.ID); err != nil {
		return nil, status.Error(codes.Code(code.RelationErrDeleteGroupJoinRecord.Code()), err.Error())
	}

	//if err := s.gjqr.DeleteJoinRequestByID(uint(req.ID)); err != nil {
	//	return nil, status.Error(codes.Code(code.RelationErrDeleteGroupJoinRecord.Code()), err.Error())
	//}
	//if s.cacheEnable {
	//	if err := s.cache.DeleteGroupJoinRequestListByUser(ctx, req.UserId); err != nil {
	//		log.Printf("delete group join request list by user failed, err: %v", err)
	//	}
	//}
	return resp, nil
}
