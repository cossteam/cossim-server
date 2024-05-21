package po

type GroupAnnouncement struct {
	BaseModel
	GroupID uint32 `gorm:"column:group_id"`
	Title   string `gorm:"column:title;comment:公告标题"`
	Content string `gorm:"column:content;comment:公告内容"`
	UserID  string `gorm:"column:user_id"`
}

func (m *GroupAnnouncement) TableName() string {
	return "group_announcements"
}

type GroupAnnouncementRead struct {
	BaseModel
	AnnouncementID uint32 `gorm:"comment:公告ID" json:"announcement_id"`
	DialogID       uint32 `gorm:"default:0;comment:对话ID" json:"dialog_id"`
	GroupID        uint32 `gorm:"comment:群聊id" json:"group_id"`
	ReadAt         int64  `gorm:"comment:已读时间" json:"read_at"`
	UserID         string `gorm:"comment:用户ID" json:"user_id"`
}

func (m *GroupAnnouncementRead) TableName() string {
	return "group_announcement_reads"
}
