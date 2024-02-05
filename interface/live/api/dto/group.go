package dto

type GroupCallRequest struct {
	GroupID uint32     `json:"group_id" binding:"required"` // 群组的ID
	Member  []string   `json:"member"`                      // 成员ids
	Option  CallOption `json:"option"`
}

type GroupJoinRequest struct {
	Room   string     `json:"room" binding:"required"` // 房间名称
	Option CallOption `json:"option"`
}
