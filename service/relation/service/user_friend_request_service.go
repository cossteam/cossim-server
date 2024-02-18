package service

import (
	"context"
	"errors"
	"github.com/cossim/coss-server/pkg/code"
	v1 "github.com/cossim/coss-server/service/relation/api/v1"
	"github.com/cossim/coss-server/service/relation/domain/entity"
	"github.com/cossim/coss-server/service/relation/infrastructure/persistence"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

func (s *Service) GetFriendRequestList(ctx context.Context, request *v1.GetFriendRequestListRequest) (*v1.GetFriendRequestListResponse, error) {
	var resp = &v1.GetFriendRequestListResponse{}
	list, err := s.ufqr.GetFriendRequestList(request.UserId)
	if err != nil {
		return resp, status.Error(codes.Code(code.RelationUserErrGetRequestListFailed.Code()), err.Error())
	}
	for _, friend := range list {
		resp.FriendRequestList = append(resp.FriendRequestList, &v1.FriendRequestList{
			ID:         uint32(friend.ID),
			SenderId:   friend.SenderID,
			Remark:     friend.Remark,
			ReceiverId: friend.ReceiverID,
			Status:     v1.FriendRequestStatus(friend.Status),
			CreateAt:   uint64(friend.CreatedAt),
		})
	}
	return resp, nil
}

func (s *Service) SendFriendRequest(ctx context.Context, request *v1.SendFriendRequestStruct) (*v1.SendFriendRequestStructResponse, error) {
	var resp = &v1.SendFriendRequestStructResponse{}
	re, err := s.ufqr.AddFriendRequest(&entity.UserFriendRequest{
		SenderID:   request.SenderId,
		ReceiverID: request.ReceiverId,
		Remark:     request.Remark,
		Status:     entity.Pending,
	})
	if err != nil {
		return resp, status.Error(codes.Code(code.RelationErrSendFriendRequestFailed.Code()), err.Error())
	}
	resp.ID = uint32(re.ID)
	return resp, nil
}

func (s *Service) ManageFriendRequest(ctx context.Context, request *v1.ManageFriendRequestStruct) (*emptypb.Empty, error) {
	var resp = &emptypb.Empty{}

	if err := s.db.Transaction(func(tx *gorm.DB) error {

		npo := persistence.NewRepositories(tx)
		re, err := npo.Ufqr.GetFriendRequestByID(uint(request.ID))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return status.Error(codes.Code(code.RelationUserErrNoFriendRequestRecords.Code()), err.Error())
			}
			return status.Error(codes.Code(code.RelationErrManageFriendRequestFailed.Code()), formatErrorMessage(err))
		}
		//拒绝
		if request.Status == v1.FriendRequestStatus_FriendRequestStatus_REJECT {
			if err := npo.Ufqr.UpdateFriendRequestStatus(uint(request.ID), entity.Rejected); err != nil {
				return status.Error(codes.Code(code.RelationErrManageFriendRequestFailed.Code()), err.Error())
			}
			return nil
		} else {
			//修改状态
			if err := npo.Ufqr.UpdateFriendRequestStatus(uint(request.ID), entity.Accepted); err != nil {
				return status.Error(codes.Code(code.RelationErrManageFriendRequestFailed.Code()), err.Error())
			}
		}

		senderId := re.SenderID
		receiverId := re.ReceiverID
		_, err = npo.Urr.GetRelationByID(senderId, receiverId)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return status.Error(codes.Code(code.RelationErrAlreadyFriends.Code()), "")
		}

		re.Status = entity.RequestStatus(request.Status)

		//如果是单删
		oldrelation, err := npo.Urr.GetRelationByID(receiverId, senderId)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return status.Error(codes.Code(code.RelationErrManageFriendRequestFailed.Code()), "")
		}

		if oldrelation != nil {
			//添加关系
			_, err := npo.Urr.CreateRelation(&entity.UserRelation{
				UserID:   re.SenderID,
				FriendID: re.ReceiverID,
				Status:   entity.UserStatusNormal,
				DialogId: oldrelation.DialogId,
			})
			if err != nil {
				return status.Error(codes.Code(code.RelationErrManageFriendRequestFailed.Code()), err.Error())
			}
			//加入对话
			_, err = npo.Dr.JoinDialog(oldrelation.DialogId, senderId)
			if err != nil {
				return status.Error(codes.Code(code.RelationErrManageFriendRequestFailed.Code()), err.Error())
			}
			return nil
		}

		//创建对话
		dialog, err := npo.Dr.CreateDialog(senderId, entity.UserDialog, 0)
		if err != nil {
			return status.Error(codes.Code(code.RelationErrManageFriendRequestFailed.Code()), err.Error())
		}
		//加入对话
		_, err = npo.Dr.JoinDialog(dialog.ID, senderId)
		if err != nil {
			return status.Error(codes.Code(code.RelationErrManageFriendRequestFailed.Code()), err.Error())
		}
		//加入对话
		_, err = npo.Dr.JoinDialog(dialog.ID, receiverId)
		if err != nil {
			return status.Error(codes.Code(code.RelationErrManageFriendRequestFailed.Code()), err.Error())
		}

		//添加好友关系
		if _, err := npo.Urr.CreateRelation(&entity.UserRelation{
			UserID:   re.SenderID,
			FriendID: re.ReceiverID,
			Status:   entity.UserStatusNormal,
			DialogId: dialog.ID,
		}); err != nil {
			return status.Error(codes.Code(code.RelationErrManageFriendRequestFailed.Code()), err.Error())
		}

		if _, err := npo.Urr.CreateRelation(&entity.UserRelation{
			UserID:   re.ReceiverID,
			FriendID: re.SenderID,
			Status:   entity.UserStatusNormal,
			DialogId: dialog.ID,
		}); err != nil {
			return status.Error(codes.Code(code.RelationErrManageFriendRequestFailed.Code()), err.Error())
		}

		return nil
	}); err != nil {
		return resp, err
	}

	return resp, nil
}

func (s *Service) GetFriendRequestById(ctx context.Context, in *v1.GetFriendRequestByIdRequest) (*v1.FriendRequestList, error) {
	var resp = &v1.FriendRequestList{}
	if re, err := s.ufqr.GetFriendRequestByID(uint(in.ID)); err != nil {
		return resp, status.Error(codes.Code(code.RelationUserErrNoFriendRequestRecords.Code()), err.Error())
	} else {
		resp.ID = uint32(re.ID)
		resp.SenderId = re.SenderID
		resp.ReceiverId = re.ReceiverID
		resp.Remark = re.Remark
		resp.Status = v1.FriendRequestStatus(re.Status)
		resp.CreateAt = uint64(re.CreatedAt)
	}
	return resp, nil
}

func (s *Service) GetFriendRequestByUserIdAndFriendId(ctx context.Context, in *v1.GetFriendRequestByUserIdAndFriendIdRequest) (*v1.FriendRequestList, error) {
	var resp = &v1.FriendRequestList{}
	if re, err := s.ufqr.GetFriendRequestBySenderIDAndReceiverID(in.UserId, in.FriendId); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, status.Error(codes.Code(code.RelationUserErrNoFriendRequestRecords.Code()), err.Error())
		}
		return resp, err
	} else {
		resp.ID = uint32(re.ID)
		resp.SenderId = re.SenderID
		resp.ReceiverId = re.ReceiverID
		resp.Remark = re.Remark
		resp.Status = v1.FriendRequestStatus(re.Status)
		resp.CreateAt = uint64(re.CreatedAt)
	}

	return resp, nil
}

func (s *Service) DeleteFriendRequestByUserIdAndFriendId(ctx context.Context, in *v1.DeleteFriendRequestByUserIdAndFriendIdRequest) (*emptypb.Empty, error) {
	var resp = &emptypb.Empty{}
	if err := s.ufqr.DeleteFriendRequestByUserIdAndFriendIdRequest(in.UserId, in.FriendId); err != nil {
		return resp, status.Error(codes.Code(code.RelationUserErrNoFriendRequestRecords.Code()), err.Error())
	}
	return resp, nil
}
