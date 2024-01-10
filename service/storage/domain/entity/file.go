package entity

// File 文件实体
type File struct {
	ID        string     `gorm:"type:char(64);primary_key;comment:文件id" json:"id"`
	Name      string     `gorm:"type:varchar(50);comment:文件名" json:"name"`
	Owner     string     `gorm:"type:char(64);comment:所属者id" json:"owner"`
	Content   string     `gorm:"type:text;comment:文件内容" json:"content"`
	Path      string     `gorm:"type:text;comment:文件路径" json:"path"`
	Type      FileType   `gorm:"comment:文件类型" json:"file_type"`
	Status    FileStatus `gorm:"comment:文件状态" json:"file_status"`
	Provider  string     `gorm:"default:MinIO;comment:文件供应商" json:"provider"`
	Share     bool       `gorm:"comment:是否共享" json:"share"`
	Size      uint64     `gorm:"comment:文件大小" json:"size"`
	CreatedAt int64      `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt int64      `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt int64      `gorm:"default:null;comment:删除时间" json:"deleted_at"`
}

type FileType int

const (
	FileTypeText    FileType = iota // 文本消息
	FileTypeVoice                   // 语音消息
	FileTypeImage                   // 图片消息
	FileTypeFile                    // 文件消息
	FileTypeVideo                   // 视频消息
	FileTypeEmoji                   // Emoji表情
	FileTypeSticker                 // 表情包
)

type FileStatus int

const (
	Pending FileStatus = iota
	Approved
	Rejected
	Archived
	Expired
)

type Provider string

const (
	MinIO    Provider = "MinIO"
	Ceph     Provider = "Ceph"
	Google   Provider = "Google"
	AmazonS3 Provider = "AmazonS3"
)
