package entity

type GroupAnnouncement struct {
	BaseModel
	GroupID uint32 `gorm:"column:group_id"`
	Title   string `gorm:"column:title"`
	Content string `gorm:"column:content"`
	UserID  string `gorm:"column:user_id"`
}
