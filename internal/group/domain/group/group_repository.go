package group

import (
	"context"
	"time"
)

type Query struct {
	ID       []uint     // 群聊 ID 列表
	Name     string     // 群聊名称
	UserID   []uint     // 包含的用户 ID 列表
	CreateAt *time.Time // 创建时间范围
	UpdateAt *time.Time // 更新时间范围
	Limit    int        // 限制结果数量
	Offset   int        // 结果的偏移量
}

type Repository interface {
	Get(ctx context.Context, id uint32) (*Group, error)
	Update(ctx context.Context, group *Group, updateFn func(h *Group) (*Group, error)) error
	Create(ctx context.Context, group *Group, createFn func(h *Group) (*Group, error)) error
	Delete(ctx context.Context, id uint32) error
	Find(ctx context.Context, query Query) ([]*Group, error)

	// UpdateFields 根据 ID 更新 Group 对象的多个字段
	UpdateFields(ctx context.Context, id uint32, fields map[string]interface{}) error
}
