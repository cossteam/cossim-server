package po

import (
	"github.com/cossim/coss-server/pkg/utils"
	ptime "github.com/cossim/coss-server/pkg/utils/time"
	"gorm.io/gorm"
)

type User struct {
	ID           string `gorm:"type:varchar(64);primary_key;comment:用户id" json:"id"`
	CossID       string `gorm:"type:varchar(64);"`
	Email        string `gorm:"type:varchar(100);uniqueIndex;comment:邮箱" json:"email"`
	Tel          string `gorm:"type:varchar(50);comment:联系电话" json:"tel"`
	NickName     string `gorm:"comment:昵称" json:"nickname"`
	Avatar       string `gorm:"type:longtext;comment:头像" json:"avatar"`
	PublicKey    string `gorm:"comment:用户pgp公钥" json:"public_key,omitempty"`
	Password     string `gorm:"type:varchar(50);comment:登录密码" json:"password,omitempty"`
	LastIp       string `gorm:"type:varchar(20);comment:最后登录IP" json:"last_ip"`
	LineIp       string `gorm:"type:varchar(20);comment:最后在线IP（接口）" json:"line_ip"`
	CreatedIp    string `gorm:"type:varchar(20);comment:注册IP" json:"created_ip"`
	Signature    string `gorm:"type:varchar(255);comment:个性签名" json:"signature"`
	LineAt       int64  `gorm:"comment:最后在线时间（接口）" json:"line_at"`
	LastAt       int64  `gorm:"comment:最后登录时间" json:"last_at"`
	Status       uint   `gorm:"type:tinyint(4);default:0;comment:用户状态" json:"status"`
	EmailVerity  bool   `gorm:"type:tinyint(1);default:0;comment:邮箱是否已验证" json:"email_verity"`
	Bot          uint   `gorm:"type:tinyint(4);default:0;comment:是否机器人" json:"bot"`
	SecretBundle string `gorm:"type:longtext;comment:用户密钥" json:"secret_bundle,omitempty"`
	CreatedAt    int64  `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt    int64  `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt    int64  `gorm:"default:0;comment:删除时间" json:"deleted_at"`
}

func (m *User) BeforeCreate(tx *gorm.DB) error {
	now := ptime.Now()
	m.CreatedAt = now
	m.UpdatedAt = now
	cid, err := utils.GenCossID()
	if err != nil {
		return err
	}
	m.CossID = cid

	return nil
}

func (m *User) BeforeUpdate(tx *gorm.DB) error {
	m.UpdatedAt = ptime.Now()
	return nil
}

func (m *User) TableName() string {
	return "users"
}
