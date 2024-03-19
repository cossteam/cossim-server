package dataTransformers

import (
	"github.com/cossim/coss-server/internal/msg/domain/entity"
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
	ID       uint   `json:"id"`
	DialogId uint   `json:"dialog_id"`
	Content  string `json:"msg"`
	Type     uint   `json:"msg_type"`
	SenderId string `json:"sender_id"`
	CreateAt int64  `json:"create_at"`
}
