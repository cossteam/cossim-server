package dataTransformers

import (
	"github.com/cossim/coss-server/service/msg/domain/entity"
)

type UserMsgListResponse struct {
	UserMessages []entity.UserMessage `json:"user_messages" form:"user_messages" uri:"user_messages"`
	Total        int32                `json:"total" form:"total" uri:"total"`
	CurrentPage  int32                `json:"current_page" form:"current_page" uri:"current_page"`
}

type GroupMessageResponse struct {
	Id        uint                   `json:"id"`
	GroupID   uint                   `json:"group_id"`
	Type      entity.UserMessageType `json:"type"`
	ReplyId   uint                   `json:"reply_id"`
	ReadCount int                    `json:"read_count"`
	UID       string                 `json:"uid"`
	Content   string                 `json:"content"`
}

type LastMessage struct {
	ID                 uint                        `json:"id"`
	DialogId           uint                        `json:"dialog_id"`
	Content            string                      `json:"msg"`
	Type               uint                        `json:"msg_type"`
	SenderId           string                      `json:"sender_id"`
	ReceiverId         string                      `json:"receiver_id"`
	CreateAt           int64                       `json:"create_at"`
	IsBurnAfterReading entity.BurnAfterReadingType `json:"is_burn_after_reading"`
	AtUsers            []string                    `json:"at_users"`
	AtAllUser          entity.AtAllUserType        `json:"at_all_user"`
	IsLabel            entity.MessageLabelType     `json:"is_label"`
	ReplyId            uint                        `json:"reply_id"`
}

type GroupMsgList struct {
	GroupID    uint32                 `json:"group_id"`    //群聊id
	UserID     string                 `json:"user_id"`     //发送者id（筛选条件）
	Content    string                 `json:"content"`     //消息内容(筛选条件)
	MsgType    entity.UserMessageType `json:"msg_type"`    //消息类型(筛选条件)
	PageNumber int                    `json:"page_number"` //页码
	PageSize   int                    `json:"page_size"`   //每页数量
}

type GroupMsgListResponse struct {
	GroupMessages []entity.GroupMessage `json:"group_messages"`
	Total         int32                 `json:"total"`
	CurrentPage   int32                 `json:"current_page"`
}
