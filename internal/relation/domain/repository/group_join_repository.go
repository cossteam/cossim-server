package repository

import (
	"context"
	"github.com/cossim/coss-server/internal/relation/domain/entity"
)

type GroupJoinRequestQuery struct {
	ID       []uint
	GroupID  []uint32
	UserID   []string
	PageSize int
	PageNum  int
	Force    bool
}

type GroupJoinRequestRepository interface {
	Get(ctx context.Context, id uint32) (*entity.GroupJoinRequest, error)
	Create(ctx context.Context, entity *entity.GroupJoinRequest) (*entity.GroupJoinRequest, error)
	Find(ctx context.Context, query *GroupJoinRequestQuery) ([]*entity.GroupJoinRequest, error)
	Creates(ctx context.Context, entity []*entity.GroupJoinRequest) ([]*entity.GroupJoinRequest, error)
	UpdateStatus(ctx context.Context, id uint32, status entity.RequestStatus) error
	Delete(ctx context.Context, id uint32) error

	GetByGroupIDAndUserID(ctx context.Context, groupID uint32, userID string) (*entity.GroupJoinRequest, error)
}
