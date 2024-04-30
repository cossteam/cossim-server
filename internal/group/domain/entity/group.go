package entity

import "github.com/go-playground/validator/v10"

type Group struct {
	ID        uint32
	CreatedAt int64
	Type      Type
	Status    Status
	//Member          int
	MaxMembersLimit int
	CreatorID       string
	Name            string
	Avatar          string
	SilenceTime     int64 // 群禁言结束时间，不为0表示开启群聊全员禁言
	JoinApprove     bool  // 入群审批
	Encrypt         bool  // 是否开启加密
}

type Type uint

const (
	TypePrivate Type = iota // 公开群
	TypePublic              // 私密群
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
