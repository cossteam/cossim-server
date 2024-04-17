package group

import "github.com/go-playground/validator/v10"

type Group struct {
	ID              uint32
	CreatedAt       int64
	Type            Type
	Status          Status
	MaxMembersLimit int
	CreatorID       string
	Name            string
	Avatar          string
}

type Type uint

const (
	TypeDefault   Type = iota // 默认群
	TypeEncrypted             // 加密群
)

type Status uint

const (
	StatusNormal  Status = iota // 正常状态
	StatusLocked                // 锁定状态
	StatusDeleted               // 删除状态
)

func (g *Group) Validate() error {
	validate := validator.New()
	return validate.Struct(g)
}
