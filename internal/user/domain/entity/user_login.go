package entity

//type UserLogin struct {
//	BaseModel
//	UserId      string `gorm:"type:varchar(64);comment:用户id" json:"user_id"`
//	LoginCount  uint   `gorm:"default:0;comment:登录次数" json:"login_count"`
//	LastAt      int64  `gorm:"comment:最后登录时间" json:"last_at"`
//	Token       string `gorm:"type:longtext;comment:登录token" json:"token"`
//	DriverId    string `gorm:"type:longtext;comment:登录设备id" json:"driver_id"`
//	DriverToken string `gorm:"type:varchar(255);comment:登录设备token" json:"driver_token"`
//	ClientType  string `gorm:"type:varchar(20);comment:客户端类型" json:"client_type"`
//	Platform    string `gorm:"type:varchar(50);comment:手机厂商" json:"platform"`
//}
//
//type BaseModel struct {
//	ID        uint  `gorm:"primaryKey;autoIncrement;" json:"id"`
//	CreatedAt int64 `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
//	UpdatedAt int64 `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
//	DeletedAt int64 `gorm:"default:0;comment:删除时间" json:"deleted_at"`
//}
