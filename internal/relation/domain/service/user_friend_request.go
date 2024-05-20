package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/cossim/coss-server/internal/relation/domain/entity"
	"github.com/cossim/coss-server/internal/relation/domain/repository"
	"github.com/cossim/coss-server/internal/relation/infra/persistence"
	"github.com/cossim/coss-server/pkg/code"
)

type ListFriendRequestOptions struct {
	UserID   string
	PageNum  int
	PageSize int
}

type UserFriendRequestDomain interface {
	Get(ctx context.Context, currentUserID string, targetUserID string) (*entity.UserFriendRequest, error)

	// CanSendRequest 检查是否可以发送好友请求，是否存在待处理请求
	CanSendRequest(ctx context.Context, currentUserID, targetUserID string) (bool, error)

	// HasPendingRequest 检查是否存在待处理请求
	HasPendingRequest(ctx context.Context, userID, friendID string) (bool, error)

	// IsMy 检查请求是否属于当前用户
	IsMy(ctx context.Context, requestID uint32, userID string) (bool, error)
	Delete(ctx context.Context, id uint32) error
	List(ctx context.Context, opts *ListFriendRequestOptions) (*entity.UserFriendRequestList, error)

	// Reject 拒绝好友请求
	Reject(ctx context.Context, requestID uint32) error

	// Accept 同意好友请求
	Accept(ctx context.Context, requestID uint32) error

	// Ignore 忽略好友请求
	Ignore(ctx context.Context, requestID uint32) error

	Expire(ctx context.Context, requestID uint32) error

	// IsHandled 判断好友请求是否已处理
	IsHandled(ctx context.Context, requestID uint32) (bool, error)

	// CreateFriendRequest 创建好友请求
	CreateFriendRequest(ctx context.Context, userID, friendID string) error

	// GetLatest 获取最新的好友请求
	GetLatest(ctx context.Context, currentUserID string, targetUserID string) (*entity.UserFriendRequest, error)
}

var _ UserFriendRequestDomain = &userFriendRequestService{}

func NewUserFriendRequestDomain(repos *persistence.Repositories) UserFriendRequestDomain {
	return &userFriendRequestService{
		repos: repos,
	}
}

type userFriendRequestService struct {
	repos *persistence.Repositories
	//db             *gorm.DB
	//userFriendRepo repository.UserFriendRequestRepository
}

func (s *userFriendRequestService) GetLatest(ctx context.Context, currentUserID string, targetUserID string) (*entity.UserFriendRequest, error) {
	find, err := s.repos.UserFriendRequestRepo.Find(ctx, &repository.UserFriendRequestQuery{
		SenderId:   currentUserID,
		ReceiverId: targetUserID,
		PageSize:   5,
		PageNum:    1,
	})
	if err != nil {
		return nil, err
	}

	if len(find.List) == 0 {
		return nil, code.RelationUserErrNoFriendRequestRecords
	}

	return find.List[0], nil
}

func (s *userFriendRequestService) CreateFriendRequest(ctx context.Context, userID, friendID string) error {
	latest, err := s.GetLatest(ctx, userID, friendID)
	if err != nil {
		return err
	}

	return s.repos.TXRepositories(func(txr *persistence.Repositories) error {
		// 删除最近最新的一条
		if err := txr.UserFriendRequestRepo.Delete(ctx, latest.ID); err != nil {
			return err
		}

		// 添加自己的
		_, err := txr.UserFriendRequestRepo.Create(ctx, &entity.UserFriendRequest{
			OwnerID:     userID,
			SenderID:    userID,
			RecipientID: friendID,
			Status:      entity.Pending,
		})
		if err != nil {
			return err
		}

		// 添加对方的
		_, err = txr.UserFriendRequestRepo.Create(ctx, &entity.UserFriendRequest{
			OwnerID:     friendID,
			SenderID:    userID,
			RecipientID: friendID,
			Status:      entity.Pending,
		})
		if err != nil {
			return err
		}

		return nil
	})
}

func (s *userFriendRequestService) CanSendRequest(ctx context.Context, currentUserID, targetUserID string) (bool, error) {
	r, err := s.repos.UserFriendRequestRepo.GetByUserIdAndFriendId(ctx, currentUserID, targetUserID)
	if err != nil {
		return false, err
	}

	fmt.Println("r => ", r.ID)

	switch r.Status {
	case entity.Pending:
		return false, code.RelationErrFriendRequestAlreadyPending
	default:
		return true, nil
	}
}

func (s *userFriendRequestService) HasPendingRequest(ctx context.Context, userID, friendID string) (bool, error) {
	req, err := s.repos.UserFriendRequestRepo.GetByUserIdAndFriendId(ctx, userID, friendID)
	if err != nil {
		if errors.Is(code.Cause(err), code.NotFound) {
			return false, nil
		}
		return false, err
	}

	return req != nil && (req.Status == entity.Pending || req.Status == entity.Invitation), nil
}

func (s *userFriendRequestService) Get(ctx context.Context, currentUserID string, targetUserID string) (*entity.UserFriendRequest, error) {
	return s.repos.UserFriendRequestRepo.GetByUserIdAndFriendId(ctx, currentUserID, targetUserID)
}

func (s *userFriendRequestService) IsHandled(ctx context.Context, requestID uint32) (bool, error) {
	r, err := s.repos.UserFriendRequestRepo.Get(ctx, requestID)
	if err != nil {
		return true, err
	}

	if r.Status == entity.Pending || r.Status == entity.Invitation {
		return false, err
	}

	return true, nil
}

func (s *userFriendRequestService) Reject(ctx context.Context, requestID uint32) error {
	return s.handleFriendRequest(ctx, requestID, entity.Rejected)
}

func (s *userFriendRequestService) Accept(ctx context.Context, requestID uint32) error {
	return s.handleFriendRequest(ctx, requestID, entity.Accepted)
}

func (s *userFriendRequestService) handleFriendRequest(ctx context.Context, requestID uint32, newStatus entity.RequestStatus) error {
	fr, err := s.repos.UserFriendRequestRepo.Get(ctx, requestID)
	if err != nil {
		return err
	}

	senderID := fr.SenderID
	recipientID := fr.RecipientID

	// 查找双方的申请记录
	reqs, err := s.repos.UserFriendRequestRepo.Find(ctx, &repository.UserFriendRequestQuery{
		SenderId:   senderID,
		ReceiverId: recipientID,
	})
	if err != nil {
		return err
	}

	return s.repos.TXRepositories(func(txr *persistence.Repositories) error {
		// 更新申请记录状态
		for _, v := range reqs.List {
			if v.Status == entity.Pending {
				if err := txr.UserFriendRequestRepo.UpdateStatus(ctx, v.ID, newStatus); err != nil {
					return err
				}
			}
		}

		// 处理接受请求的特定逻辑
		if newStatus == entity.Accepted {
			return s.handleAcceptedRequest(ctx, txr, senderID, recipientID)
		}

		return nil
	})
}

func (s *userFriendRequestService) handleAcceptedRequest(ctx context.Context, txr *persistence.Repositories, senderID, recipientID string) error {
	// 检查是否存在旧的好友关系
	oldRelation, err := txr.UserRepo.Get(ctx, recipientID, senderID)
	if err != nil && !errors.Is(err, code.NotFound) {
		return err
	}

	// 如果存在旧的好友关系，直接添加好友并加入对话
	if oldRelation != nil {
		return s.addFriendToExistingRelation(ctx, txr, oldRelation.DialogId, senderID, recipientID)
	}

	// 否则，创建新的对话和好友关系
	return s.createNewRelation(ctx, txr, senderID, recipientID)
}

// addFriendToExistingRelation 将好友添加到已有的对话和关系中
func (s *userFriendRequestService) addFriendToExistingRelation(ctx context.Context, txr *persistence.Repositories, dialogID uint32, senderID, recipientID string) error {
	// 添加好友关系
	if _, err := txr.UserRepo.Create(ctx, &entity.UserRelation{
		UserID:   senderID,
		FriendID: recipientID,
		Status:   entity.UserStatusNormal,
		DialogId: dialogID,
	}); err != nil {
		return err
	}

	// 将好友加入对话
	if _, err := txr.DialogUserRepo.Create(ctx, &repository.CreateDialogUser{
		DialogID: dialogID,
		UserID:   senderID,
	}); err != nil {
		return err
	}

	return nil
}

// createNewRelation 创建新的对话和好友关系
func (s *userFriendRequestService) createNewRelation(ctx context.Context, txr *persistence.Repositories, senderID, recipientID string) error {
	// 创建新的对话
	dialog, err := txr.DialogRepo.Create(ctx, &repository.CreateDialog{
		Type:    entity.UserDialog,
		OwnerId: senderID,
	})
	if err != nil {
		return err
	}

	// 将两个用户加入同一个对话
	if _, err := txr.DialogUserRepo.Creates(ctx, dialog.ID, []string{senderID, recipientID}); err != nil {
		return err
	}

	// 建立好友关系
	if err := txr.UserRepo.EstablishFriendship(ctx, dialog.ID, senderID, recipientID); err != nil {
		return err
	}

	return nil
}

func (s *userFriendRequestService) Ignore(ctx context.Context, requestID uint32) error {
	return nil
}

func (s *userFriendRequestService) Expire(ctx context.Context, requestID uint32) error {
	return s.repos.UserFriendRequestRepo.UpdateStatus(ctx, requestID, entity.Expired)
}

func (s *userFriendRequestService) IsMy(ctx context.Context, requestID uint32, userID string) (bool, error) {
	if userID == "" {
		return false, code.InvalidParameter
	}

	req, err := s.repos.UserFriendRequestRepo.Get(ctx, requestID)
	if err != nil {
		return false, err
	}

	return req.OwnerID == userID, nil
}

func (s *userFriendRequestService) Delete(ctx context.Context, id uint32) error {
	return s.repos.UserFriendRequestRepo.Delete(ctx, id)
}

func (s *userFriendRequestService) List(ctx context.Context, opts *ListFriendRequestOptions) (*entity.UserFriendRequestList, error) {
	if opts.UserID == "" {
		return nil, code.InvalidParameter
	}

	return s.repos.UserFriendRequestRepo.Find(ctx, &repository.UserFriendRequestQuery{
		UserID:   opts.UserID,
		PageSize: opts.PageSize,
		PageNum:  opts.PageNum,
		Force:    false,
	})
}
