package dataTransformers

import (
	"github.com/cossim/coss-server/internal/msg/domain/entity"
)

type GroupMsgList struct {
	DialogId   uint
	UserID     string                 //发送者id（筛选条件）
	Content    string                 //消息内容(筛选条件)
	MsgType    entity.UserMessageType //消息类型(筛选条件)
	PageNumber int                    //页码
	PageSize   int                    //每页数量
}

type GroupMsgListResponse struct {
	GroupMessages []*entity.GroupMessage `json:"group_messages"`
	Total         int32                  `json:"total"`
	CurrentPage   int32                  `json:"current_page"`
}
