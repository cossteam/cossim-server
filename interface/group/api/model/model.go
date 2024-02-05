package model

import api "github.com/cossim/coss-server/service/group/api/v1"

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type UpdateGroupRequest struct {
	Type            uint32 `json:"type"`
	Status          uint32 `json:"status"`
	MaxMembersLimit uint32 `json:"max_members_limit"`
	CreatorID       string `json:"creator_id"`
	Name            string `json:"name"`
	Avatar          string `json:"avatar"`
	GroupId         uint32 `json:"group_id"`
}

type CreateGroupRequest struct {
	Type            uint32   `json:"type"`
	MaxMembersLimit uint32   `json:"max_members_limit"`
	Name            string   `json:"name" binding:"required"`
	Avatar          string   `json:"avatar"`
	Member          []string `json:"member"`
}

type CreateGroupResponse struct {
	Id              uint32 `json:"id"`
	Avatar          string `json:"avatar"`
	Name            string `json:"name"`
	Type            uint32 `json:"type"`
	Status          int32  `json:"status"`
	MaxMembersLimit int32  `json:"max_members_limit"`
	CreatorId       string `json:"creator_id"`
	DialogId        uint32 `json:"dialog_id"`
}

type GroupInfo struct {
	Id              uint32       `json:"id"`
	Avatar          string       `json:"avatar"`
	Name            string       `json:"name"`
	Type            uint32       `json:"type"`
	Status          int32        `json:"status"`
	MaxMembersLimit int32        `json:"max_members_limit"`
	CreatorId       string       `json:"creator_id"`
	DialogId        uint32       `json:"dialog_id"`
	Preferences     *Preferences `json:"preferences,omitempty"`
}

type Preferences struct {
	EntryMethod          EntryMethod              `json:"entry_method"`
	JoinedAt             int64                    `json:"joined_at"`
	MuteEndTime          int64                    `json:"mute_end_time"`
	SilentNotification   SilentNotification       `json:"silent_notification"`
	GroupNickname        string                   ` json:"group_nickname"`
	Inviter              string                   ` json:"inviter"`
	Remark               string                   ` json:"remark"`
	OpenBurnAfterReading OpenBurnAfterReadingType `json:"open_burn_after_reading"`
	Identity             GroupIdentity            `json:"identity"`
}

type GroupIdentity uint

const (
	IdentityUser  GroupIdentity = iota // 普通用户
	IdentityAdmin                      // 管理员
	IdentityOwner                      // 群主
)

type OpenBurnAfterReadingType uint

const (
	CloseBurnAfterReading OpenBurnAfterReadingType = iota //关闭阅后即焚
	OpenBurnAfterReading                                  //开启阅后即焚消息
)

func IsValidGroupType(value api.GroupType) bool {
	return value == api.GroupType_TypePublic || value == api.GroupType_TypePrivate
}

type DeleteGroupRequest struct {
	GroupId uint32 `json:"group_id" binding:"required"`
}

type EntryMethod uint

const (
	EntryInvitation EntryMethod = iota // 邀请
	EntrySearch                        // 搜索
)

type SilentNotification uint

const (
	NotSilentNotification SilentNotification = iota //不开启静默通知
	IsSilentNotification                            //开启静默通知
)
