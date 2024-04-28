package relation

type GroupAnnouncementRead struct {
	ID             uint32
	CreatedAt      int64
	AnnouncementId uint32
	DialogId       uint32
	GroupID        uint32
	ReadAt         int64
	UserId         string
}

type GroupAnnouncement struct {
	//BaseModel
	ID        uint32
	GroupID   uint32
	CreatedAt int64
	UpdatedAt int64
	Title     string
	Content   string
	UserID    string
}

type UpdateGroupAnnouncement struct {
	ID      uint32
	Title   string
	Content string
}
