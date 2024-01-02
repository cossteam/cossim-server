package persistence

import (
	"gorm.io/gorm"
)

// UserRepo 需要实现UserRepository接口
type UserRepo struct {
	db *gorm.DB
	//sss
}
