package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/service/relation/api/v1"
	"github.com/cossim/coss-server/service/relation/domain/entity"
	"github.com/cossim/coss-server/service/relation/infrastructure/persistence"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
	"path/filepath"
	"runtime"
)

func (s *Service) AddFriend(ctx context.Context, request *v1.AddFriendRequest) (*v1.AddFriendResponse, error) {
	resp := &v1.AddFriendResponse{}

	//查询是否单删
	relation, err := s.urr.GetRelationByID(request.FriendId, request.GetUserId())
	if err != nil && err != gorm.ErrRecordNotFound {
		return resp, status.Error(codes.Code(code.RelationErrAddFriendFailed.Code()), formatErrorMessage(err))
	}
	if relation != nil {
		if _, err := s.urr.CreateRelation(&entity.UserRelation{
			UserID:   request.GetUserId(),
			FriendID: request.GetFriendId(),
			Status:   entity.UserRelationStatus(v1.RelationStatus_RELATION_NORMAL),
			DialogId: relation.DialogId,
		}); err != nil {
			return resp, status.Error(codes.Code(code.RelationErrAddFriendFailed.Code()), formatErrorMessage(err))
		}
		return resp, nil
	}
	//双方都没有好友关系
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		npo := persistence.NewRepositories(tx)

		if _, err := npo.Urr.CreateRelation(&entity.UserRelation{
			UserID:   request.GetUserId(),
			FriendID: request.GetFriendId(),
			Status:   entity.UserRelationStatus(v1.RelationStatus_RELATION_NORMAL),
			DialogId: uint(request.DialogId),
		}); err != nil {
			return err
		}
		if _, err := npo.Urr.CreateRelation(&entity.UserRelation{
			UserID:   request.GetFriendId(),
			FriendID: request.GetUserId(),
			Status:   entity.UserRelationStatus(v1.RelationStatus_RELATION_NORMAL),
			DialogId: uint(request.DialogId),
		}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return resp, status.Error(codes.Code(code.RelationErrAddFriendFailed.Code()), formatErrorMessage(err))
	}

	return resp, nil
}

func getFunctionName() string {
	pc, _, _, _ := runtime.Caller(1)
	return runtime.FuncForPC(pc).Name()
}

func formatErrorMessage(err error) string {
	funcName := getFunctionName()
	_, file := filepath.Split(funcName)
	return fmt.Sprintf("[%s] %s: %v", file, funcName, err)
}

func (s *Service) ManageFriendRevert(ctx context.Context, request *v1.ManageFriendRequest) (*v1.ManageFriendResponse, error) {
	resp := &v1.ManageFriendResponse{}

	//if err := s.db.Transaction(func(tx *gorm.DB) error {
	//	userId := request.GetUserId()
	//	friendId := request.GetFriendId()
	//
	//	relation1, err := s.urr.GetRelationByID(userId, friendId)
	//	if err != nil {
	//		if errors.Is(err, gorm.ErrRecordNotFound) {
	//			return status.Error(codes.Code(code.RelationUserErrNoFriendRequestRecords.Code()), err.Error())
	//		}
	//		return status.Error(codes.Code(code.RelationErrConfirmFriendFailed.Code()), formatErrorMessage(err))
	//	}
	//
	//	//if relation1.Status == entity.UserStatusAdded {
	//	//	return resp, status.Error(codes.Code(code.RelationErrAlreadyFriends.Code()), "已经是好友")
	//	//}
	//
	//	//relation1.Status = entity.UserRelationStatus(request.Status)
	//	//relation1.Status = entity.UserRelationStatus(request.Status)
	//	relation1.Status = entity.UserStatusPending
	//	relation1.DialogId = uint(request.DialogId)
	//	_, err = s.urr.UpdateRelation(relation1)
	//	if err != nil {
	//		return status.Error(codes.Code(code.RelationErrConfirmFriendFailed.Code()), formatErrorMessage(err))
	//	}
	//
	//	relation2, err := s.urr.GetRelationByID(friendId, userId)
	//	if err != nil {
	//		return status.Error(codes.Code(code.RelationErrConfirmFriendFailed.Code()), formatErrorMessage(err))
	//	}
	//
	//	//relation2.Status = entity.UserRelationStatus(request.Status)
	//	relation2.Status = entity.UserStatusApplying
	//	relation2.DialogId = uint(request.DialogId)
	//	_, err = s.urr.UpdateRelation(relation2)
	//	if err != nil {
	//		return status.Error(codes.Code(code.RelationErrConfirmFriendFailed.Code()), formatErrorMessage(err))
	//	}
	//
	//	return nil
	//}); err != nil {
	//	return resp, status.Error(codes.Code(code.RelationErrConfirmFriendFailed.Code()), formatErrorMessage(err))
	//}

	return resp, nil
}
func (s *Service) DeleteFriend(ctx context.Context, request *v1.DeleteFriendRequest) (*v1.DeleteFriendResponse, error) {
	resp := &v1.DeleteFriendResponse{}

	userId := request.GetUserId()
	friendId := request.GetFriendId()
	// Assuming urr is a UserRelationRepository instance in UserService
	if err := s.urr.DeleteRelationByID(userId, friendId); err != nil {
		//return resp, status.Error(codes.Code(code.RelationErrDeleteFriendFailed.Code()), fmt.Sprintf("failed to delete relation: %v", err))
		return resp, status.Error(codes.Aborted, fmt.Sprintf("failed to delete relation: %v", err))
	}

	//if err := s.urr.DeleteRelationByID(friendId, userId); err != nil {
	//	return resp, status.Error(codes.Code(code.RelationErrDeleteFriendFailed.Code()), fmt.Sprintf("failed to delete relation: %v", err))
	//}

	return resp, nil
}

func (s *Service) DeleteFriendRevert(ctx context.Context, request *v1.DeleteFriendRequest) (*v1.DeleteFriendResponse, error) {

	resp := &v1.DeleteFriendResponse{}
	if err := s.urr.UpdateRelationDeleteAtByID(request.UserId, request.FriendId, 0); err != nil {
		return resp, status.Error(codes.Code(code.RelationErrDeleteFriendFailed.Code()), fmt.Sprintf("DeleteFriendRevert failed to update relation: %v", err))
	}
	return resp, nil
}

func (s *Service) AddBlacklist(ctx context.Context, request *v1.AddBlacklistRequest) (*v1.AddBlacklistResponse, error) {
	resp := &v1.AddBlacklistResponse{}

	userId := request.GetUserId()
	friendId := request.GetFriendId()
	// Assuming urr is a UserRelationRepository instance in UserService
	relation1, err := s.urr.GetRelationByID(userId, friendId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, status.Error(codes.Code(code.RelationUserErrFriendRelationNotFound.Code()), fmt.Sprintf("failed to retrieve relation: %v", err))
		}
		return resp, status.Error(codes.Code(code.RelationErrAddBlacklistFailed.Code()), fmt.Sprintf("failed to retrieve relation: %v", err))
	}

	if relation1.Status != entity.UserStatusNormal {
		return resp, code.RelationUserErrFriendRelationNotFound
	}

	relation1.Status = entity.UserStatusBlocked
	if _, err = s.urr.UpdateRelation(relation1); err != nil {
		return resp, status.Error(codes.Code(code.RelationErrAddBlacklistFailed.Code()), fmt.Sprintf("failed to update relation: %v", err))
	}

	return resp, nil
}

func (s *Service) DeleteBlacklist(ctx context.Context, request *v1.DeleteBlacklistRequest) (*v1.DeleteBlacklistResponse, error) {
	resp := &v1.DeleteBlacklistResponse{}

	userId := request.GetUserId()
	friendId := request.GetFriendId()
	// Assuming urr is a UserRelationRepository instance in UserService
	relation1, err := s.urr.GetRelationByID(userId, friendId)
	if err != nil {
		return resp, status.Error(codes.Code(code.RelationErrDeleteBlacklistFailed.Code()), fmt.Sprintf("failed to retrieve relation: %v", err))
	}

	relation1.Status = entity.UserStatusNormal
	if _, err = s.urr.UpdateRelation(relation1); err != nil {
		return resp, status.Error(codes.Code(code.RelationErrDeleteBlacklistFailed.Code()), fmt.Sprintf("failed to update relation: %v", err))
	}

	//relation2, err := s.urr.GetRelationByID(friendId, userId)
	//if err != nil {
	//	return resp, code.RelationErrDeleteBlacklistFailed.Reason(fmt.Errorf("failed to retrieve relation: %w", err))
	//}
	//
	//relation2.Action = entity.UserStatusAdded
	//if _, err = s.urr.UpdateRelation(relation2); err != nil {
	//	return resp, code.RelationErrDeleteBlacklistFailed.Reason(fmt.Errorf("failed to update relation: %w", err))
	//}

	return resp, nil
}

func (s *Service) GetFriendList(ctx context.Context, request *v1.GetFriendListRequest) (*v1.GetFriendListResponse, error) {
	resp := &v1.GetFriendListResponse{}

	friends, err := s.urr.GetRelationsByUserID(request.GetUserId())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, status.Error(codes.Code(code.RelationErrUserNotFound.Code()), fmt.Sprintf("failed to get friend list: %v", err))
		}
		return resp, status.Error(codes.Code(code.RelationErrGetFriendListFailed.Code()), fmt.Sprintf("failed to get friend list: %v", err))
	}

	for _, friend := range friends {
		resp.FriendList = append(resp.FriendList, &v1.Friend{UserId: friend.FriendID, DialogId: uint32(friend.DialogId)})
	}

	return resp, nil
}

func (s *Service) GetBlacklist(ctx context.Context, request *v1.GetBlacklistRequest) (*v1.GetBlacklistResponse, error) {
	resp := &v1.GetBlacklistResponse{}

	blacklist, err := s.urr.GetBlacklistByUserID(request.GetUserId())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, status.Error(codes.Code(code.RelationErrUserNotFound.Code()), fmt.Sprintf("failed to get black list: %v", err))
		}
		return resp, status.Error(codes.Code(code.RelationErrGetBlacklistFailed.Code()), fmt.Sprintf("failed to get black list: %v", err))
	}

	for _, black := range blacklist {
		resp.Blacklist = append(resp.Blacklist, &v1.Blacklist{UserId: black.UserID})
	}

	return resp, nil
}

func (s *Service) GetUserRelation(ctx context.Context, request *v1.GetUserRelationRequest) (*v1.GetUserRelationResponse, error) {
	resp := &v1.GetUserRelationResponse{}

	relation, err := s.urr.GetRelationByID(request.GetUserId(), request.GetFriendId())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, status.Error(codes.Code(code.RelationUserErrFriendRelationNotFound.Code()), err.Error())
		}
		return resp, status.Error(codes.Code(code.RelationErrGetUserRelationFailed.Code()), fmt.Sprintf("failed to get user relation: %v", err))
	}

	resp.Status = v1.RelationStatus(relation.Status)
	resp.DialogId = uint32(relation.DialogId)
	resp.UserId = relation.UserID
	resp.FriendId = relation.FriendID
	resp.IsSilent = v1.UserSilentNotificationType(relation.SilentNotification)
	return resp, nil
}

//func (s *Service) GetFriendRequestList(ctx context.Context, request *v1.GetFriendRequestListRequest) (*v1.GetFriendRequestListResponse, error) {
//	resp := &v1.GetFriendRequestListResponse{}
//
//	friends, err := s.urr.GetFriendRequestListByUserID(request.UserId)
//	if err != nil {
//		return resp, status.Error(codes.Code(code.RelationGroupErrGetJoinRequestListFailed.Code()), err.Error())
//	}
//
//	for _, friend := range friends {
//		resp.FriendRequestList = append(resp.FriendRequestList, &v1.FriendRequestList{
//			UserId: friend.FriendID,
//			Msg:    friend.Remark,
//			Status: v1.FriendRequestStatus(friend.Status),
//		})
//	}
//
//	return resp, nil
//}

func (s *Service) GetUserRelationByUserIds(ctx context.Context, request *v1.GetUserRelationByUserIdsRequest) (*v1.GetUserRelationByUserIdsResponse, error) {
	resp := &v1.GetUserRelationByUserIdsResponse{}

	relations, err := s.urr.GetRelationByIDs(request.UserId, request.FriendIds)
	if err != nil {
		return resp, status.Error(codes.Code(code.RelationErrGetUserRelationFailed.Code()), err.Error())
	}

	for _, relation := range relations {
		resp.Users = append(resp.Users, &v1.GetUserRelationResponse{
			UserId:   relation.UserID,
			FriendId: relation.FriendID,
			Status:   v1.RelationStatus(relation.Status),
			DialogId: uint32(relation.DialogId),
		})
	}
	return resp, nil
}

func (s *Service) SetFriendSilentNotification(ctx context.Context, request *v1.SetFriendSilentNotificationRequest) (*emptypb.Empty, error) {
	var resp = &emptypb.Empty{}
	if err := s.urr.SetUserFriendSilentNotification(request.UserId, request.FriendId, entity.SilentNotification(request.IsSilent)); err != nil {
		return resp, status.Error(codes.Code(code.RelationErrSetUserFriendSilentNotificationFailed.Code()), err.Error())
	}
	return resp, nil
}

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

func (s *Service) SendFriendRequest(ctx context.Context, request *v1.SendFriendRequestStruct) (*emptypb.Empty, error) {
	var resp = &emptypb.Empty{}
	err := s.ufqr.AddFriendRequest(&entity.UserFriendRequest{
		SenderID:   request.SenderId,
		ReceiverID: request.ReceiverId,
		Remark:     request.Remark,
		Status:     entity.Pending,
	})
	if err != nil {
		return resp, status.Error(codes.Code(code.RelationErrSendFriendRequestFailed.Code()), err.Error())
	}
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
			if err := s.ufqr.UpdateFriendRequestStatus(uint(request.ID), entity.Rejected); err != nil {
				return status.Error(codes.Code(code.RelationErrManageFriendRequestFailed.Code()), err.Error())
			}
			return nil
		} else {
			//修改状态
			if err := s.ufqr.UpdateFriendRequestStatus(uint(request.ID), entity.Accepted); err != nil {
				return status.Error(codes.Code(code.RelationErrManageFriendRequestFailed.Code()), err.Error())
			}
		}

		senderId := re.SenderID
		receiverId := re.ReceiverID
		_, err = npo.Urr.GetRelationByID(senderId, receiverId)
		if err != nil {
			return status.Error(codes.Code(code.RelationErrAlreadyFriends.Code()), "已经是好友")
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
			_, err = npo.Dr.JoinDialog(oldrelation.DialogId, receiverId)
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
