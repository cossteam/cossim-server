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

type BlacklistOptions struct {
	UserID   string
	PageNum  int
	PageSize int
}

type DeleteBlacklistOptions struct {
	UserID   string
	PageNum  int
	PageSize int
}

type FriendsListOptions struct {
}

type UserRelationDomain interface {
	GetRelation(ctx context.Context, userID, targetID string) (*entity.UserRelation, error)
	GetDialogID(ctx context.Context, userID, targetID string) (uint32, error)
	DeleteBlacklist(ctx context.Context, userID, targetID string) error
	AddBlacklist(ctx context.Context, userID, targetID string) error
	Blacklist(ctx context.Context, opts *BlacklistOptions) (*entity.Blacklist, error)
	IsFriend(ctx context.Context, userID, targetID string) (bool, error)
	IsInBlacklist(ctx context.Context, userID, targetID string) (bool, error)
	AddFriendAfterDelete(ctx context.Context, userID, targetID string) error
	EstablishFriendship(ctx context.Context, userID, friendID string) error
	FriendsList(ctx context.Context, userID string, opts *FriendsListOptions) ([]*entity.Friend, error)
	SetUserBurn(ctx context.Context, userID, targetID string, burn bool, timeout uint32) error
	SetUserRemark(ctx context.Context, userID, targetID, remark string) error
	SetUserSilent(ctx context.Context, userID, targetID string, silent bool) error
	DeleteFriend(ctx context.Context, userID, targetID string) error
	DeleteFriendRollback(ctx context.Context, userID, targetID string) error
}

var _ UserRelationDomain = &userRelationService{}

func NewUserRelationDomain(repos *persistence.Repositories) UserRelationDomain {
	return &userRelationService{
		//userRepo: userRepo,
		repos: repos,
	}
}

type userRelationService struct {
	//userRepo repository.UserRelationRepository
	repos *persistence.Repositories
}

func (s *userRelationService) GetRelation(ctx context.Context, userID, targetID string) (*entity.UserRelation, error) {
	return s.repos.UserRepo.Get(ctx, userID, targetID)
}

func (s *userRelationService) DeleteFriendRollback(ctx context.Context, userID, targetID string) error {
	rel, err := s.repos.UserRepo.Get(ctx, userID, targetID)
	if err != nil {
		return err
	}

	return s.repos.TXRepositories(func(txr *persistence.Repositories) error {
		var at int64
		if err := s.repos.DialogUserRepo.UpdateDialogStatus(ctx, &repository.UpdateDialogStatusParam{
			DialogID:  rel.DialogId,
			UserID:    []string{userID},
			DeletedAt: &at,
		}); err != nil {
			return err
		}
		if err := s.repos.UserRepo.DeleteRollback(ctx, rel.ID); err != nil {
			return err
		}

		return nil
	})
}

func (s *userRelationService) GetDialogID(ctx context.Context, userID, targetID string) (uint32, error) {
	rel, err := s.repos.UserRepo.Get(ctx, userID, targetID)
	if err != nil {
		return 0, err
	}

	return rel.DialogId, nil
}

func (s *userRelationService) DeleteFriend(ctx context.Context, userID, targetID string) error {
	rel, err := s.repos.UserRepo.Get(ctx, userID, targetID)
	if err != nil {
		return err
	}

	return s.repos.TXRepositories(func(txr *persistence.Repositories) error {
		if err := s.repos.DialogUserRepo.DeleteByDialogIDAndUserID(ctx, rel.DialogId, userID); err != nil {
			return err
		}
		if err := s.repos.UserRepo.Delete(ctx, rel.UserID, rel.FriendID); err != nil {
			return err
		}

		return nil
	})
}

func (s *userRelationService) SetUserSilent(ctx context.Context, userID, targetID string, silent bool) error {
	isFriend, err := s.IsFriend(ctx, userID, targetID)
	if err != nil {
		return err
	}
	if !isFriend {
		return code.RelationUserErrFriendRelationNotFound
	}

	return s.repos.UserRepo.SetUserFriendSilentNotification(ctx, userID, targetID, silent)
}

func (s *userRelationService) SetUserRemark(ctx context.Context, userID, targetID, remark string) error {
	return s.repos.UserRepo.SetFriendRemark(ctx, userID, targetID, remark)
}

func (s *userRelationService) SetUserBurn(ctx context.Context, userID, targetID string, burn bool, timeout uint32) error {
	return s.repos.UserRepo.SetUserOpenBurnAfterReading(ctx, userID, targetID, burn, timeout)
}

func (s *userRelationService) FriendsList(ctx context.Context, userID string, opts *FriendsListOptions) ([]*entity.Friend, error) {
	return s.repos.UserRepo.ListFriend(ctx, userID)
}

func (s *userRelationService) EstablishFriendship(ctx context.Context, userID, friendID string) error {
	//return s.userRepo.EstablishFriendship(ctx, userID, friendID)
	return nil
}

func (s *userRelationService) AddFriendAfterDelete(ctx context.Context, userID, targetID string) error {
	status := entity.UserStatusDeleted
	rels, err := s.repos.UserRepo.Find(ctx, &repository.UserQuery{
		UserId:   userID,
		FriendId: []string{targetID},
		Status:   &status,
		Force:    true,
	})
	if err != nil {
		return err
	}

	if len(rels) == 0 {
		return nil
	}

	return s.repos.UserRepo.RestoreFriendship(ctx, rels[0].DialogId, userID, targetID)
}

func (s *userRelationService) DeleteBlacklist(ctx context.Context, userID, targetID string) error {
	rel, err := s.repos.UserRepo.Get(ctx, userID, targetID)
	if err != nil {
		if errors.Is(code.Cause(err), code.NotFound) {
			return code.RelationUserErrFriendRelationNotFound
		}
		return err
	}

	if rel.Status != entity.UserStatusBlocked {
		return code.RelationErrNotInBlacklist
	}

	isBlacklist, err := s.IsInBlacklist(ctx, userID, targetID)
	if err != nil {
		return err
	}

	if !isBlacklist {
		return code.RelationErrNotInBlacklist
	}

	return s.repos.UserRepo.UpdateStatus(ctx, rel.ID, entity.UserStatusNormal)
}

func (s *userRelationService) AddBlacklist(ctx context.Context, userID, targetID string) error {
	isBlacklist, err := s.IsInBlacklist(ctx, userID, targetID)
	if err != nil {
		return err
	}

	if isBlacklist {
		return code.RelationErrAlreadyBlacklist
	}

	rel, err := s.repos.UserRepo.Get(ctx, userID, targetID)
	if err != nil {
		if errors.Is(code.Cause(err), code.NotFound) {
			return code.RelationUserErrFriendRelationNotFound
		}
		return err
	}

	if rel.Status != entity.UserStatusNormal {
		return code.RelationUserErrFriendRelationNotFound
	}

	return s.repos.UserRepo.UpdateStatus(ctx, rel.ID, entity.UserStatusBlocked)
}

func (s *userRelationService) IsFriend(ctx context.Context, userID, targetID string) (bool, error) {
	rel, err := s.repos.UserRepo.Get(ctx, userID, targetID)
	if err != nil {
		if errors.Is(code.Cause(err), code.NotFound) {
			return false, nil
		}
		return false, err
	}

	if rel.Status == entity.UserStatusBlocked {
		return false, code.MyCustomErrorCode.CustomMessage("对方拒收了你的消息")
	}

	return rel.Status == entity.UserStatusNormal, nil
	//return rel.Status == entity.UserStatusNormal || rel.Status == entity.UserStatusBlocked, nil

}

func (s *userRelationService) IsInBlacklist(ctx context.Context, userID, targetID string) (bool, error) {
	blacklist, err := s.repos.UserRepo.Blacklist(ctx, &entity.BlacklistOptions{
		UserID: userID,
	})
	if err != nil {
		return false, err
	}

	if len(blacklist.List) == 0 {
		return false, nil
	}

	for _, black := range blacklist.List {
		fmt.Println("black.ID => ", black.UserID)
		fmt.Println("targetID => ", targetID)
		if black.UserID == targetID {
			return true, nil
		}
	}

	return false, nil
}

func (s *userRelationService) Blacklist(ctx context.Context, opts *BlacklistOptions) (*entity.Blacklist, error) {
	return s.repos.UserRepo.Blacklist(ctx, &entity.BlacklistOptions{
		UserID:   opts.UserID,
		PageNum:  opts.PageNum,
		PageSize: opts.PageSize,
	})
}
