package model

import (
	"encoding/json"
	api "github.com/cossim/coss-server/internal/group/api/grpc/v1"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type UpdateGroupRequest struct {
	Type    uint32 `json:"type"`
	Name    string `json:"name"`
	Avatar  string `json:"avatar"`
	GroupId uint32 `json:"group_id"`
}

type CreateGroupRequest struct {
	Type   uint32   `json:"type"` // Type 群聊类型
	Name   string   `json:"name" binding:"required"`
	Avatar string   `json:"avatar"`
	Member []string `json:"member"`
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
	return value == api.GroupType_TypeEncrypted || value == api.GroupType_TypeDefault
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

type GroupTypeMemberLimit uint

const (
	DefaultGroup   = 1000 //默认群
	EncryptedGroup = 500  //加密群
)

type SenderInfo struct {
	UserId string `json:"user_id"`
	Avatar string `json:"avatar"`
	Name   string `json:"name"`
}

type ConversationType uint

const (
	UserConversation ConversationType = iota
	GroupConversation
)

type UserDialogListResponse struct {
	DialogId uint32 `json:"dialog_id"`
	UserId   string `json:"user_id,omitempty"`
	GroupId  uint32 `json:"group_id,omitempty"`
	// 会话类型
	DialogType ConversationType `json:"dialog_type"`
	// 会话名称
	DialogName string `json:"dialog_name"`
	// 会话头像
	DialogAvatar string `json:"dialog_avatar"`
	// 会话未读消息数
	DialogUnreadCount int     `json:"dialog_unread_count"`
	LastMessage       Message `json:"last_message"`

	DialogCreateAt int64 `json:"dialog_create_at"`
	TopAt          int64 `json:"top_at"`
}
type Message struct {
	GroupId            uint32               `json:"group_id,omitempty"`      //群聊id
	MsgType            uint                 `json:"msg_type"`                // 消息类型
	Content            string               `json:"content"`                 // 消息内容
	SenderId           string               `json:"sender_id"`               // 消息发送者
	SendAt             int64                `json:"send_at"`                 // 消息发送时间
	MsgId              uint64               `json:"msg_id"`                  // 消息id
	SenderInfo         SenderInfo           `json:"sender_info"`             // 消息发送者信息
	ReceiverInfo       SenderInfo           `json:"receiver_info,omitempty"` // 消息接受者信息
	AtAllUser          AtAllUserType        `json:"at_all_user,omitempty"`   // @全体用户
	AtUsers            []string             `json:"at_users,omitempty"`      // @用户id
	IsBurnAfterReading BurnAfterReadingType `json:"is_burn_after_reading"`   // 是否阅后即焚
	IsLabel            LabelMsgType         `json:"is_label"`                // 是否标记
	ReplyId            uint32               `json:"reply_id"`                // 回复消息id
}

type BurnAfterReadingType uint

const (
	NotBurnAfterReading BurnAfterReadingType = iota //非阅后即焚
	IsBurnAfterReading                              //阅后即焚消息
)

type LabelMsgType uint

const (
	NotLabel LabelMsgType = iota //不标注
	IsLabel                      //标注
)

type AtAllUserType uint

const (
	NotAtAllUser = iota
	AtAllUser
)

func (udlr UserDialogListResponse) MarshalBinary() ([]byte, error) {
	// 将UserDialogListResponse对象转换为二进制数据
	data, err := json.Marshal(udlr)
	if err != nil {
		return nil, err
	}
	return data, nil
}
