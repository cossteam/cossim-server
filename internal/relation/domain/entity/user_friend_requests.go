package entity

type UserFriendRequest struct {
	BaseModel
	SenderID   string        `json:"sender_id" gorm:"column:sender_id"`
	ReceiverID string        `json:"receiver_id" gorm:"column:receiver_id"`
	Status     RequestStatus `json:"status" gorm:"column:status"`
	Remark     string        `json:"remark" gorm:"column:remark"`
	// DeletedBy 表示删除该记录的用户ID id1,id2
	DeletedBy string `json:"deleted_by" gorm:"column:deleted_by"`
}
