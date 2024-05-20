package entity

type GroupAnnouncementRead struct {
	ID             uint32
	CreatedAt      int64
	AnnouncementId uint32
	DialogId       uint32
	GroupID        uint32
	ReadAt         int64
	UserId         string
}

type GroupAnnouncementList struct {
	List  []*GroupAnnouncement
	Total int64
}

type GroupAnnouncement struct {
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

type GroupAnnouncementQuery struct {
	ID      []uint32 // 群公告 ID 列表
	GroupID []uint32 // 群聊 ID 列表
	Name    string   // 群聊名称
	Limit   int      // 限制结果数量
	Offset  int      // 结果的偏移量
}
