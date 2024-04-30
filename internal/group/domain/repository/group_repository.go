package repository

import (
	"context"
	"github.com/cossim/coss-server/internal/group/domain/entity"
	"time"
)

type Query struct {
	ID       []uint32   // 群聊 ID 列表
	Name     string     // 群聊名称
	UserID   []string   // 包含的用户 ID 列表
	CreateAt *time.Time // 创建时间范围
	UpdateAt *time.Time // 更新时间范围
	Limit    int        // 限制结果数量
	Offset   int        // 结果的偏移量
	Cache    bool
}

type Repository interface {
	Get(ctx context.Context, id uint32) (*entity.Group, error)
	Update(ctx context.Context, group *entity.Group, updateFn func(h *entity.Group) (*entity.Group, error)) error
	Create(ctx context.Context, group *entity.Group, createFn func(h *entity.Group) (*entity.Group, error)) error
	Delete(ctx context.Context, id uint32) error
	Find(ctx context.Context, query Query) ([]*entity.Group, error)

	// UpdateFields 根据 ID 更新 Group 对象的多个字段
	UpdateFields(ctx context.Context, id uint32, fields map[string]interface{}) error
}
