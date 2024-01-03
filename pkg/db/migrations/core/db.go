package core

import (
	"errors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	DB *gorm.DB
)

func InDB(url string) (*gorm.DB, error) {
	// 创建 SQLite 数据库连接
	db, err := gorm.Open(mysql.Open(url), &gorm.Config{})
	if err != nil {
		return nil, errors.New("无法连接数据库：" + err.Error())
	}
	return db, nil

}

func InitDB() error {
	db, err := InDB("root:888888@tcp(127.0.0.1:33066)/coss?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		return err
	}
	DB = db
	return nil
}
