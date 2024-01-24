package core

import (
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/db"
	"gorm.io/gorm"
)

var (
	DB *gorm.DB
)

func InitDB(c *pkgconfig.AppConfig) error {
	dbConn, err := db.NewMySQLFromDSN(c.MySQL.DSN).GetConnection()
	if err != nil {
		panic(err)
	}
	DB = dbConn
	return nil
}
