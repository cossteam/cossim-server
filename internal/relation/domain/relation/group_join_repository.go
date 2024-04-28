package relation

import "context"

type GroupJoinRequestQuery struct {
	ID       []uint
	GroupID  []uint32
	UserID   []string
	PageSize int
	PageNum  int
}

type GroupJoinRequestRepository interface {
	Get(ctx context.Context, id uint32) (*GroupJoinRequest, error)
	Create(ctx context.Context, entity *GroupJoinRequest) (*GroupJoinRequest, error)
	Find(ctx context.Context, query *GroupJoinRequestQuery) ([]*GroupJoinRequest, error)
	Creates(ctx context.Context, entity []*GroupJoinRequest) ([]*GroupJoinRequest, error)
	UpdateStatus(ctx context.Context, id uint32, status RequestStatus) error
	Delete(ctx context.Context, id uint32) error

	GetByGroupIDAndUserID(ctx context.Context, groupID uint32, userID string) (*GroupJoinRequest, error)
}
