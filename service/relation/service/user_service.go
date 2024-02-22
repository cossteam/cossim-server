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

		dialog, err := npo.Dr.CreateDialog(request.UserId, entity.DialogType(v1.DialogType_USER_DIALOG), 0)
		if err != nil {
			return err
		}

		if _, err := npo.Urr.CreateRelation(&entity.UserRelation{
			UserID:   request.GetUserId(),
			FriendID: request.GetFriendId(),
			Status:   entity.UserRelationStatus(v1.RelationStatus_RELATION_NORMAL),
			DialogId: dialog.ID,
		}); err != nil {
			return err
		}
		if _, err := npo.Urr.CreateRelation(&entity.UserRelation{
			UserID:   request.GetFriendId(),
			FriendID: request.GetUserId(),
			Status:   entity.UserRelationStatus(v1.RelationStatus_RELATION_NORMAL),
			DialogId: dialog.ID,
		}); err != nil {
			return err
		}
		_, err = npo.Dr.JoinDialog(dialog.ID, request.UserId)
		if err != nil {
			return err
		}

		_, err = npo.Dr.JoinDialog(dialog.ID, request.FriendId)
		if err != nil {
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
	if err = s.urr.UpdateRelationColumn(relation1.ID, "status", entity.UserStatusBlocked); err != nil {
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
		resp.FriendList = append(resp.FriendList, &v1.Friend{UserId: friend.FriendID, DialogId: uint32(friend.DialogId), Remark: friend.Remark})
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
		return resp, err
	}

	resp.Status = v1.RelationStatus(relation.Status)
	resp.DialogId = uint32(relation.DialogId)
	resp.UserId = relation.UserID
	resp.FriendId = relation.FriendID
	resp.IsSilent = v1.UserSilentNotificationType(relation.SilentNotification)
	resp.OpenBurnAfterReading = v1.OpenBurnAfterReadingType(relation.OpenBurnAfterReading)
	resp.Remark = relation.Remark
	resp.IsSilent = v1.UserSilentNotificationType(relation.SilentNotification)
	return resp, nil
}

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

func (s *Service) SetUserOpenBurnAfterReading(ctx context.Context, in *v1.SetUserOpenBurnAfterReadingRequest) (*emptypb.Empty, error) {
	var resp = &emptypb.Empty{}
	if err := s.urr.SetUserOpenBurnAfterReading(in.UserId, in.FriendId, entity.OpenBurnAfterReadingType(in.OpenBurnAfterReading)); err != nil {
		return resp, status.Error(codes.Code(code.RelationErrSetUserOpenBurnAfterReadingFailed.Code()), err.Error())
	}
	return resp, nil
}

func (s *Service) SetFriendRemark(ctx context.Context, in *v1.SetFriendRemarkRequest) (*emptypb.Empty, error) {
	var resp = &emptypb.Empty{}
	if err := s.urr.SetFriendRemarkByUserIdAndFriendId(in.UserId, in.FriendId, in.Remark); err != nil {
		return resp, status.Error(codes.Code(code.RelationErrSetFriendRemarkFailed.Code()), err.Error())
	}
	return resp, nil
}
