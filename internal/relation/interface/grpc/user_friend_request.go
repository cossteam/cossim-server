package grpc

import (
	"context"
	"errors"
	v1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	"github.com/cossim/coss-server/internal/relation/domain/entity"
	"github.com/cossim/coss-server/internal/relation/domain/repository"
	"github.com/cossim/coss-server/internal/relation/infrastructure/persistence"
	"github.com/cossim/coss-server/pkg/cache"
	"github.com/cossim/coss-server/pkg/code"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
	"log"
)

var _ v1.UserFriendRequestServiceServer = &userFriendRequestServiceServer{}

type userFriendRequestServiceServer struct {
	db          *gorm.DB
	cache       cache.RelationUserCache
	cacheEnable bool
	ufqr        repository.UserFriendRequestRepository
}

func (s *userFriendRequestServiceServer) ManageFriend(ctx context.Context, request *v1.ManageFriendRequest) (*v1.ManageFriendResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *userFriendRequestServiceServer) GetFriendRequestList(ctx context.Context, request *v1.GetFriendRequestListRequest) (*v1.GetFriendRequestListResponse, error) {
	resp := &v1.GetFriendRequestListResponse{}

	if s.cacheEnable {
		// 尝试从缓存中获取好友请求列表
		cachedList, err := s.cache.GetFriendRequestList(ctx, request.UserId)
		if err == nil && cachedList != nil {
			// 如果缓存中存在，则直接返回缓存的结果
			return cachedList, nil
		}
	}

	list, total, err := s.ufqr.GetFriendRequestList(request.UserId, int(request.PageSize), int(request.PageNum))
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
	resp.Total = uint64(total)

	// TODO 考虑不使用异步的方式，缓存设置失败了，重试或回滚
	go func() {
		if s.cacheEnable {
			if err := s.cache.SetFriendRequestList(ctx, request.UserId, resp, cache.RelationExpireTime); err != nil {
				log.Printf("set FriendRequestList cache failed: %v", err)
			}
		}
	}()

	return resp, nil
}

func (s *userFriendRequestServiceServer) SendFriendRequest(ctx context.Context, request *v1.SendFriendRequestStruct) (*v1.SendFriendRequestStructResponse, error) {
	var resp = &v1.SendFriendRequestStructResponse{}

	// 添加自己的
	re1, err := s.ufqr.AddFriendRequest(&entity.UserFriendRequest{
		SenderID:   request.SenderId,
		ReceiverID: request.ReceiverId,
		Remark:     request.Remark,
		OwnerID:    request.SenderId,
		Status:     entity.Pending,
	})
	if err != nil {
		return resp, status.Error(codes.Code(code.RelationErrSendFriendRequestFailed.Code()), err.Error())
	}
	resp.ID = uint32(re1.ID)

	// 添加对方的
	_, err = s.ufqr.AddFriendRequest(&entity.UserFriendRequest{
		SenderID:   request.SenderId,
		ReceiverID: request.ReceiverId,
		Remark:     request.Remark,
		OwnerID:    request.ReceiverId,
		Status:     entity.Pending,
	})
	if err != nil {
		return resp, status.Error(codes.Code(code.RelationErrSendFriendRequestFailed.Code()), err.Error())
	}

	// TODO 考虑不使用异步的方式，缓存设置失败了，重试或回滚
	go func() {
		if s.cacheEnable {
			if err := s.cache.DeleteFriendRequestList(ctx, request.SenderId, request.ReceiverId); err != nil {
				log.Printf("delete FriendRequestList cache failed: %v", err)
			}
		}
	}()

	return resp, nil
}

func (s *userFriendRequestServiceServer) ManageFriendRequest(ctx context.Context, request *v1.ManageFriendRequestStruct) (*emptypb.Empty, error) {
	var resp = &emptypb.Empty{}
	var senderId, receiverId string

	if err := s.db.Transaction(func(tx *gorm.DB) error {

		npo := persistence.NewRepositories(tx)
		re, err := npo.Ufqr.GetFriendRequestByID(uint(request.ID))
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
			if err := npo.Ufqr.UpdateFriendRequestStatus(senderId, receiverId, entity.Rejected); err != nil {
				return status.Error(codes.Code(code.RelationErrManageFriendRequestFailed.Code()), err.Error())
			}
			return nil
		} else {
			//修改状态
			if err := npo.Ufqr.UpdateFriendRequestStatus(senderId, receiverId, entity.Accepted); err != nil {
				return status.Error(codes.Code(code.RelationErrManageFriendRequestFailed.Code()), err.Error())
			}
		}

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
		log.Printf("ManageFriendRequest err => %v", err)
		return resp, err
	}

	// TODO 考虑不使用异步的方式，缓存设置失败了，重试或回滚
	go func() {
		if s.cacheEnable {
			if err := s.cache.DeleteFriendRequestList(ctx, senderId, receiverId); err != nil {
				log.Printf("delete FriendRequestList cache failed: %v", err)
			}
			if err := s.cache.DeleteFriendList(ctx, senderId, receiverId); err != nil {
				log.Printf("delete FriendRequestList cache failed: %v", err)
			}
			if err := s.cache.DeleteRelation(ctx, senderId, receiverId); err != nil {
				log.Printf("delete FriendRequestList cache failed: %v", err)
			}
		}
	}()

	return resp, nil
}

func (s *userFriendRequestServiceServer) GetFriendRequestById(ctx context.Context, request *v1.GetFriendRequestByIdRequest) (*v1.FriendRequestList, error) {
	var resp = &v1.FriendRequestList{}

	if re, err := s.ufqr.GetFriendRequestByID(uint(request.ID)); err != nil {
		return resp, status.Error(codes.Code(code.RelationUserErrNoFriendRequestRecords.Code()), err.Error())
	} else {
		resp.ID = uint32(re.ID)
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
	var resp = &v1.FriendRequestList{}
	if re, err := s.ufqr.GetFriendRequestBySenderIDAndReceiverID(request.UserId, request.FriendId); err != nil {
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

func (s *userFriendRequestServiceServer) DeleteFriendRequestByUserIdAndFriendId(ctx context.Context, request *v1.DeleteFriendRequestByUserIdAndFriendIdRequest) (*emptypb.Empty, error) {
	var resp = &emptypb.Empty{}
	if err := s.ufqr.DeleteFriendRequestByUserIdAndFriendIdRequest(request.UserId, request.FriendId); err != nil {
		return resp, status.Error(codes.Code(code.RelationUserErrNoFriendRequestRecords.Code()), err.Error())
	}
	return resp, nil
}

func (s *userFriendRequestServiceServer) DeleteFriendRecord(ctx context.Context, req *v1.DeleteFriendRecordRequest) (*emptypb.Empty, error) {
	resp := &emptypb.Empty{}
	if err := s.ufqr.DeletedById(req.ID); err != nil {
		return nil, status.Error(codes.Code(code.RelationErrDeleteUserFriendRecord.Code()), err.Error())
	}

	// TODO 考虑不使用异步的方式，缓存设置失败了，重试或回滚
	go func() {
		if s.cacheEnable {
			if err := s.cache.DeleteFriendRequestList(ctx, req.UserId); err != nil {
				log.Printf("delete FriendRequestList cache failed: %v", err)
			}
		}
	}()

	return resp, nil
}
