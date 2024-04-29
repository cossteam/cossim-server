package grpc

import (
	"context"
	"errors"
	v1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	"github.com/cossim/coss-server/internal/relation/domain/entity"
	"github.com/cossim/coss-server/internal/relation/domain/repository"
	"github.com/cossim/coss-server/internal/relation/infra/persistence"
	"github.com/cossim/coss-server/pkg/code"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
	"log"
)

var _ v1.UserFriendRequestServiceServer = &userFriendRequestServiceServer{}

type userFriendRequestServiceServer struct {
	repos *persistence.Repositories
}

func (s *userFriendRequestServiceServer) ManageFriend(ctx context.Context, request *v1.ManageFriendRequest) (*v1.ManageFriendResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *userFriendRequestServiceServer) GetFriendRequestList(ctx context.Context, request *v1.GetFriendRequestListRequest) (*v1.GetFriendRequestListResponse, error) {
	resp := &v1.GetFriendRequestListResponse{}

	list, err := s.repos.UserFriendRequestRepo.Find(ctx, &repository.UserFriendRequestQuery{
		UserID:   request.UserId,
		PageSize: int(request.PageSize),
		PageNum:  int(request.PageNum),
		Force:    false,
	})
	if err != nil {
		return nil, err
	}

	//list, total, err := s.ufqr.GetFriendRequestList(request.UserId, int(request.PageSize), int(request.PageNum))
	//if err != nil {
	//	return resp, status.Error(codes.Code(code.RelationUserErrGetRequestListFailed.Code()), err.Error())
	//}

	for _, friend := range list.List {
		resp.FriendRequestList = append(resp.FriendRequestList, &v1.FriendRequestList{
			ID:         friend.ID,
			SenderId:   friend.SenderID,
			Remark:     friend.Remark,
			ReceiverId: friend.ReceiverID,
			Status:     v1.FriendRequestStatus(friend.Status),
			CreateAt:   uint64(friend.CreatedAt),
		})
	}
	resp.Total = uint64(list.Total)

	return resp, nil
}

func (s *userFriendRequestServiceServer) SendFriendRequest(ctx context.Context, request *v1.SendFriendRequestStruct) (*v1.SendFriendRequestStructResponse, error) {
	resp := &v1.SendFriendRequestStructResponse{}

	if err := s.repos.TXRepositories(func(txr *persistence.Repositories) error {
		// 添加自己的
		re1, err := txr.UserFriendRequestRepo.Create(ctx, &entity.UserFriendRequest{
			SenderID:   request.SenderId,
			ReceiverID: request.ReceiverId,
			Remark:     request.Remark,
			OwnerID:    request.SenderId,
			Status:     entity.Pending,
		})
		if err != nil {
			return status.Error(codes.Code(code.RelationErrSendFriendRequestFailed.Code()), err.Error())
		}

		// 对方拉黑了，不允许添加
		re2, err := s.repos.UserRepo.Get(ctx, request.ReceiverId, request.SenderId)
		if err == nil && re2.Status == entity.UserStatusBlocked {
			return nil
		}

		// 添加对方的
		_, err = txr.UserFriendRequestRepo.Create(ctx, &entity.UserFriendRequest{
			SenderID:   request.SenderId,
			ReceiverID: request.ReceiverId,
			Remark:     request.Remark,
			OwnerID:    request.ReceiverId,
			Status:     entity.Pending,
		})
		if err != nil {
			return status.Error(codes.Code(code.RelationErrSendFriendRequestFailed.Code()), err.Error())
		}
		resp.ID = re1.ID
		return nil
	}); err != nil {
		log.Printf("ManageFriendRequest err => %v", err)
		return resp, err
	}

	//// TODO 考虑不使用异步的方式，缓存设置失败了，重试或回滚
	//go func() {
	//	if s.cacheEnable {
	//		if err := s.cache.DeleteFriendRequestList(ctx, request.SenderId, request.ReceiverId); err != nil {
	//			log.Printf("delete FriendRequestList cache failed: %v", err)
	//		}
	//	}
	//}()

	return resp, nil
}

func (s *userFriendRequestServiceServer) ManageFriendRequest(ctx context.Context, request *v1.ManageFriendRequestStruct) (*emptypb.Empty, error) {
	resp := &emptypb.Empty{}
	var senderId, receiverId string

	if err := s.repos.TXRepositories(func(txr *persistence.Repositories) error {
		re, err := txr.UserFriendRequestRepo.Get(ctx, request.ID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return status.Error(codes.Code(code.RelationUserErrNoFriendRequestRecords.Code()), err.Error())
			}
			return status.Error(codes.Code(code.RelationErrManageFriendRequestFailed.Code()), formatErrorMessage(err))
		}

		senderId = re.SenderID
		receiverId = re.ReceiverID

		//拒绝
		if request.Status == v1.FriendRequestStatus_FriendRequestStatus_REJECT {
			st := entity.Pending
			finds, err := txr.UserFriendRequestRepo.Find(ctx, &repository.UserFriendRequestQuery{
				SenderId:   senderId,
				ReceiverId: receiverId,
				Status:     &st,
			})
			if err != nil {
				return err
			}

			for _, v := range finds.List {
				if err := txr.UserFriendRequestRepo.UpdateStatus(ctx, v.ID, entity.Rejected); err != nil {
					return status.Error(codes.Code(code.RelationErrManageFriendRequestFailed.Code()), err.Error())
				}
			}

			return nil
		} else {
			// 修改状态
			find, err := txr.UserFriendRequestRepo.Find(ctx, &repository.UserFriendRequestQuery{
				SenderId:   senderId,
				ReceiverId: receiverId,
			})
			if err != nil {
				return status.Error(codes.Code(code.RelationErrManageFriendRequestFailed.Code()), err.Error())
			}

			for _, v := range find.List {
				if v.Status == entity.Pending {
					if err := txr.UserFriendRequestRepo.UpdateStatus(ctx, v.ID, entity.Accepted); err != nil {
						return status.Error(codes.Code(code.RelationErrManageFriendRequestFailed.Code()), err.Error())
					}
				}
			}
		}

		_, err = txr.UserRepo.Get(ctx, senderId, receiverId)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return status.Error(codes.Code(code.RelationErrAlreadyFriends.Code()), "")
		}

		re.Status = entity.RequestStatus(request.Status)

		// 如果是单删
		oldrelation, err := txr.UserRepo.Get(ctx, receiverId, senderId)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return status.Error(codes.Code(code.RelationErrManageFriendRequestFailed.Code()), "")
		}

		if oldrelation != nil {
			//添加关系
			_, err := txr.UserRepo.Create(ctx, &entity.UserRelation{
				UserID:   re.SenderID,
				FriendID: re.ReceiverID,
				Status:   entity.UserStatusNormal,
				DialogId: oldrelation.DialogId,
			})
			if err != nil {
				return status.Error(codes.Code(code.RelationErrManageFriendRequestFailed.Code()), err.Error())
			}
			//加入对话
			_, err = txr.DialogUserRepo.Create(ctx, &repository.CreateDialogUser{
				DialogID: oldrelation.DialogId,
				UserID:   senderId,
			})
			if err != nil {
				return status.Error(codes.Code(code.RelationErrManageFriendRequestFailed.Code()), err.Error())
			}
			return nil
		}

		// 创建对话
		dialog, err := txr.DialogRepo.Create(ctx, &repository.CreateDialog{
			Type:    entity.UserDialog,
			OwnerId: senderId,
			GroupId: 0,
		})
		if err != nil {
			return status.Error(codes.Code(code.RelationErrManageFriendRequestFailed.Code()), err.Error())
		}

		// 将两个用户加入同一个对话
		_, err = txr.DialogUserRepo.Creates(ctx, dialog.ID, []string{senderId, receiverId})
		if err != nil {
			return status.Error(codes.Code(code.RelationErrManageFriendRequestFailed.Code()), err.Error())
		}

		// 建立好友关系
		if err := txr.UserRepo.EstablishFriendship(ctx, dialog.ID, re.SenderID, re.ReceiverID); err != nil {
			return err
		}

		return nil
	}); err != nil {
		log.Printf("ManageFriendRequest err => %v", err)
		return resp, err
	}

	//// TODO 考虑不使用异步的方式，缓存设置失败了，重试或回滚
	//go func() {
	//	if s.cacheEnable {
	//		if err := s.cache.DeleteFriendRequestList(ctx, senderId, receiverId); err != nil {
	//			log.Printf("delete FriendRequestList cache failed: %v", err)
	//		}
	//		if err := s.cache.DeleteFriendList(ctx, senderId, receiverId); err != nil {
	//			log.Printf("delete FriendRequestList cache failed: %v", err)
	//		}
	//		if err := s.cache.DeleteRelation(ctx, senderId, receiverId); err != nil {
	//			log.Printf("delete FriendRequestList cache failed: %v", err)
	//		}
	//	}
	//}()

	return resp, nil
}

func (s *userFriendRequestServiceServer) GetFriendRequestById(ctx context.Context, request *v1.GetFriendRequestByIdRequest) (*v1.FriendRequestList, error) {
	resp := &v1.FriendRequestList{}

	if re, err := s.repos.UserFriendRequestRepo.Get(ctx, request.ID); err != nil {
		return resp, status.Error(codes.Code(code.RelationUserErrNoFriendRequestRecords.Code()), err.Error())
	} else {
		resp.ID = re.ID
		resp.SenderId = re.SenderID
		resp.ReceiverId = re.ReceiverID
		resp.Remark = re.Remark
		resp.Status = v1.FriendRequestStatus(re.Status)
		resp.CreateAt = uint64(re.CreatedAt)
		resp.OwnerID = re.OwnerID
	}
	return resp, nil
}

func (s *userFriendRequestServiceServer) GetFriendRequestByUserIdAndFriendId(ctx context.Context, request *v1.GetFriendRequestByUserIdAndFriendIdRequest) (*v1.FriendRequestList, error) {
	resp := &v1.FriendRequestList{}

	rel, err := s.repos.UserFriendRequestRepo.GetByUserIdAndFriendId(ctx, request.UserId, request.FriendId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, status.Error(codes.Code(code.RelationUserErrNoFriendRequestRecords.Code()), err.Error())
		}
		return nil, err
	}

	resp.ID = rel.ID
	resp.SenderId = rel.SenderID
	resp.ReceiverId = rel.ReceiverID
	resp.Remark = rel.Remark
	resp.Status = v1.FriendRequestStatus(rel.Status)
	resp.CreateAt = uint64(rel.CreatedAt)
	return resp, nil
}

func (s *userFriendRequestServiceServer) DeleteFriendRequestByUserIdAndFriendId(ctx context.Context, request *v1.DeleteFriendRequestByUserIdAndFriendIdRequest) (*emptypb.Empty, error) {
	resp := &emptypb.Empty{}

	find, err := s.repos.UserFriendRequestRepo.Find(ctx, &repository.UserFriendRequestQuery{
		SenderId:   request.UserId,
		ReceiverId: request.FriendId,
		PageSize:   5,
		PageNum:    1,
	})
	if err != nil {
		return nil, err
	}

	if len(find.List) == 0 {
		return resp, status.Error(codes.Code(code.RelationUserErrNoFriendRequestRecords.Code()), code.RelationUserErrNoFriendRequestRecords.Message())
	}

	if err := s.repos.UserFriendRequestRepo.Delete(ctx, find.List[0].ID); err != nil {
		return resp, status.Error(codes.Code(code.RelationUserErrNoFriendRequestRecords.Code()), code.RelationUserErrNoFriendRequestRecords.Message())
	}

	//if err := s.ufqr.DeleteFriendRequestByUserIdAndFriendIdRequest(request.UserId, request.FriendId); err != nil {
	//	return resp, status.Error(codes.Code(code.RelationUserErrNoFriendRequestRecords.Code()), err.Error())
	//}
	return resp, nil
}

func (s *userFriendRequestServiceServer) DeleteFriendRecord(ctx context.Context, req *v1.DeleteFriendRecordRequest) (*emptypb.Empty, error) {
	resp := &emptypb.Empty{}
	if err := s.repos.UserFriendRequestRepo.Delete(ctx, req.ID); err != nil {
		return nil, status.Error(codes.Code(code.RelationErrDeleteUserFriendRecord.Code()), err.Error())
	}

	//// TODO 考虑不使用异步的方式，缓存设置失败了，重试或回滚
	//go func() {
	//	if s.cacheEnable {
	//		if err := s.cache.DeleteFriendRequestList(ctx, req.UserId); err != nil {
	//			log.Printf("delete FriendRequestList cache failed: %v", err)
	//		}
	//	}
	//}()

	return resp, nil
}
