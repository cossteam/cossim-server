package service

import (
	"context"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/utils/time"
	v1 "github.com/cossim/coss-server/service/relation/api/v1"
	"github.com/cossim/coss-server/service/relation/domain/entity"
	"github.com/cossim/coss-server/service/relation/infrastructure/persistence"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

func (s *Service) InviteJoinGroup(ctx context.Context, request *v1.InviteJoinGroupRequest) (*v1.JoinGroupResponse, error) {
	resp := &v1.JoinGroupResponse{}
	inviterID := request.InviterId
	relations := make([]*entity.GroupJoinRequest, 0)

	for _, userID := range request.Member {
		userGroup := &entity.GroupJoinRequest{
			UserID:      userID,
			GroupID:     uint(request.GroupId),
			Inviter:     inviterID,
			InviterTime: time.Now(),
			Status:      entity.Invitation,
		}
		relations = append(relations, userGroup)
	}

	_, err := s.gjqr.AddJoinRequestBatch(relations)
	if err != nil {
		return resp, status.Error(codes.Code(code.RelationGroupErrInviteFailed.Code()), err.Error())
	}
	return resp, nil
}

func (s *Service) JoinGroup(ctx context.Context, request *v1.JoinGroupRequest) (*v1.JoinGroupResponse, error) {
	resp := &v1.JoinGroupResponse{}

	_, err := s.gjqr.AddJoinRequest(&entity.GroupJoinRequest{
		GroupID:     uint(request.GroupId),
		UserID:      request.UserId,
		Remark:      request.Msg,
		InviterTime: time.Now(),
		Status:      entity.Pending,
	})
	if err != nil {
		return resp, status.Error(codes.Code(code.RelationGroupErrRequestFailed.Code()), err.Error())
	}
	return resp, nil
}

func (s *Service) GetGroupJoinRequestListByUserId(ctx context.Context, request *v1.GetGroupJoinRequestListRequest) (*v1.GroupJoinRequestListResponse, error) {
	var resp = &v1.GroupJoinRequestListResponse{}

	//获取该用户管理的群聊
	ds, err := s.grr.GetUserManageGroupIDs(request.UserId)
	if err != nil {
		return nil, err
	}
	if len(ds) > 0 {
		//获取该用户管理的群聊的申请列表
		ids := make([]uint, 0)
		for _, v := range ds {
			ids = append(ids, uint(v))
		}
		list, err := s.gjqr.GetJoinRequestBatchListByGroupIDs(ids)
		if err != nil {
			return resp, err
		}
		if len(list) > 0 {
			for _, v := range list {
				resp.GroupJoinRequestResponses = append(resp.GroupJoinRequestResponses, &v1.GroupJoinRequestResponse{
					ID:        uint32(v.ID),
					UserId:    v.UserID,
					GroupId:   uint32(v.GroupID),
					Status:    v1.GroupRequestStatus(v.Status),
					InviterId: v.Inviter,
					CreatedAt: uint64(v.CreatedAt),
					Remark:    v.Remark,
				})
			}
		}
	}
	//获取他自己的申请列表
	list, err := s.gjqr.GetGroupJoinRequestListByUserId(request.UserId)
	if err != nil {
		return resp, err
	}
	if len(list) > 0 {
		for _, v := range list {
			resp.GroupJoinRequestResponses = append(resp.GroupJoinRequestResponses, &v1.GroupJoinRequestResponse{
				ID:        uint32(v.ID),
				UserId:    v.UserID,
				GroupId:   uint32(v.GroupID),
				Status:    v1.GroupRequestStatus(v.Status),
				InviterId: v.Inviter,
				CreatedAt: uint64(v.CreatedAt),
				Remark:    v.Remark,
			})
		}
	}
	return resp, nil
}

func (s *Service) GetGroupJoinRequestByGroupIdAndUserId(ctx context.Context, request *v1.GetGroupJoinRequestByGroupIdAndUserIdRequest) (*v1.GetGroupJoinRequestByGroupIdAndUserIdResponse, error) {
	var resp = &v1.GetGroupJoinRequestByGroupIdAndUserIdResponse{}
	req, err := s.gjqr.GetGroupJoinRequestByGroupIdAndUserId(uint(request.GroupId), request.UserId)
	if err != nil {
		return resp, err
	}
	if req != nil {
		resp.ID = uint32(req.ID)
		resp.GroupId = uint32(req.GroupID)
		resp.UserId = req.UserID
		resp.Status = v1.GroupRequestStatus(req.Status)
		resp.CreatedAt = uint64(req.CreatedAt)
		if req.Remark != "" {
			resp.Remark = req.Remark
		}
		resp.InviterId = req.Inviter
		return resp, nil
	}
	return resp, nil
}

func (s *Service) ManageGroupJoinRequestByID(ctx context.Context, in *v1.ManageGroupJoinRequestByIDRequest) (*emptypb.Empty, error) {
	var resp = &emptypb.Empty{}

	info, err := s.gjqr.GetGroupJoinRequestByRequestID(uint(in.ID))
	if err != nil {
		return nil, err
	}
	//拒绝请求
	if in.Status == v1.GroupRequestStatus_Rejected {
		if err := s.gjqr.ManageGroupJoinRequestByID(uint(in.ID), entity.RequestStatus(in.Status)); err != nil {
			return resp, status.Error(codes.Code(code.RelationGroupErrManageJoinFailed.Code()), err.Error())
		}
		return resp, nil
	}

	relation, err := s.grr.GetUserGroupByID(uint32(info.GroupID), info.UserID)
	if relation != nil {
		return resp, status.Error(codes.Code(code.RelationGroupErrAlreadyInGroup.Code()), err.Error())
	}

	//获取对话id
	dialog, err := s.dr.GetDialogByGroupId(info.GroupID)
	if err != nil {
		return resp, status.Error(codes.Code(code.RelationGroupErrManageJoinFailed.Code()), err.Error())
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		npo := persistence.NewRepositories(tx)

		//加入对话
		_, err := npo.Dr.JoinDialog(dialog.ID, info.UserID)
		if err != nil {
			return err
		}

		//修改请求状态
		if err := npo.Gjqr.ManageGroupJoinRequestByID(uint(in.ID), entity.RequestStatus(in.Status)); err != nil {
			return status.Error(codes.Code(code.RelationGroupErrManageJoinFailed.Code()), err.Error())
		}

		gr := &entity.GroupRelation{
			GroupID:     info.GroupID,
			UserID:      info.UserID,
			Identity:    entity.IdentityUser,
			JoinedAt:    time.Now(),
			EntryMethod: entity.EntrySearch,
		}

		if info.Inviter != "" {
			gr.Inviter = info.Inviter
			gr.EntryMethod = entity.EntryInvitation
		}
		//加入群聊
		_, err = npo.Grr.CreateRelation(gr)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return resp, err
	}

	return resp, nil
}
