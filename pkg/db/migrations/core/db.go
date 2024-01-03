package core

import (
	"gorm.io/gorm"
	"im/pkg/config"
	"im/pkg/db"
)

var (
	DB *gorm.DB
)

func InitDB(c *config.AppConfig) error {
	dbConn, err := db.NewMySQLFromDSN(c.MySQL.DSN).GetConnection()
	if err != nil {
		panic(err)
	}
	DB = dbConn
	return nil
}
