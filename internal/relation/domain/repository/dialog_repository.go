package repository

import (
	"context"
	"github.com/cossim/coss-server/internal/relation/domain/entity"
)

type DialogQuery struct {
	DialogID []uint32
	GroupID  []uint32
	UserID   []string
	PageSize int
	PageNum  int
}

type CreateDialog struct {
	Type    entity.DialogType
	OwnerId string
	GroupId uint32
}

type DialogRepository interface {
	Get(ctx context.Context, id uint32) (*entity.Dialog, error)
	Create(ctx context.Context, createDialog *CreateDialog) (*entity.Dialog, error)
	Creates(ctx context.Context, dialog []*entity.Dialog) ([]*entity.Dialog, error)
	Update(ctx context.Context, dialog *entity.Dialog) (*entity.Dialog, error)
	Delete(ctx context.Context, id ...uint32) error
	Find(ctx context.Context, query *DialogQuery) ([]*entity.Dialog, error)

	GetByGroupID(ctx context.Context, groupID uint32) (*entity.Dialog, error)

	// UpdateFields 根据会话ID更新会话信息 Dialog
	UpdateFields(ctx context.Context, dialogID uint, updateFields map[string]interface{}) error
}

type DialogUserQuery struct {
	DialogID []uint32
	UserID   []string
	IsShow   bool
	Force    bool
	PageSize int
	PageNum  int
}

type CreateDialogUser struct {
	DialogID uint32
	UserID   string
}

type UpdateDialogStatusParam struct {
	DialogID  uint32
	UserID    []string
	IsShow    *bool
	TopAt     *int64
	DeletedAt *int64
}

type DialogUserRepository interface {
	Get(ctx context.Context, id uint32) (*entity.DialogUser, error)
	Create(ctx context.Context, createDialogUser *CreateDialogUser) (*entity.DialogUser, error)
	Creates(ctx context.Context, dialogID uint32, userID []string) ([]*entity.DialogUser, error)
	Update(ctx context.Context, dialog *entity.DialogUser) (*entity.DialogUser, error)
	Delete(ctx context.Context, id ...uint32) error
	Find(ctx context.Context, query *DialogUserQuery) ([]*entity.DialogUser, error)

	// ListByDialogID 获取对话下的所有用户
	ListByDialogID(ctx context.Context, dialogID uint32) ([]*entity.DialogUser, error)

	// DeleteByDialogID 根据对话id删除用户对话
	DeleteByDialogID(ctx context.Context, dialogID uint32) error

	// DeleteByDialogIDAndUserID 根据对话id和用户id删除用户对话关系
	DeleteByDialogIDAndUserID(ctx context.Context, dialogID uint32, userID ...string) error

	UpdateDialogStatus(ctx context.Context, Param *UpdateDialogStatusParam) error

	// UpdateFields 根据id更新会话用户信息 DialogUser
	UpdateFields(ctx context.Context, id uint32, updateFields map[string]interface{}) error
}
