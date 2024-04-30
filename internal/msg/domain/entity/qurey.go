package entity

type UserMsgQuery struct {
	MsgIds    []uint32
	DialogIds []uint32
	PageNum   int64
	PageSize  int64
	StartAt   int64
	EndAt     int64
	Sort      string
	Content   string
	SendID    string
	MsgType   UserMessageType
}

type UserMsgQueryResult struct {
	TotalCount    int64          // 总消息数
	Remaining     int64          // 剩余消息数
	ReturnedCount int64          // 本次查询返回的消息数
	CurrentPage   int64          // 当前页码
	TotalPages    int64          // 总页码
	Messages      []*UserMessage // 消息列表
}
