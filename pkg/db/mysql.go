package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/cossim/coss-server/pkg/utils"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"sync"
	"time"
)

// MySQL mysql配置
type MySQL struct {
	Host     string `toml:"host" env:"MYSQL_HOST"`
	Port     string `toml:"port" env:"MYSQL_PORT"`
	UserName string `toml:"username" env:"MYSQL_USERNAME"`
	Password string `toml:"password" env:"MYSQL_PASSWORD"`
	Database string `toml:"database" env:"MYSQL_DATABASE"`
	Level    logger.Interface
	Dsn      string
	// 因为使用的是 Mysql的连接池，需要对连接池做一些规划配置
	// 控制当前程序的 Mysql打开的连接数
	MaxOpenConn int `toml:"max_open_conn" env:"MYSQL_MAX_OPEN_CONN"`
	// 控制 Mysql复用，比如 5， 最多运行5个复用
	MaxIdleConn int `toml:"max_idle_conn" env:"MYSQL_MAX_IDLE_CONN"`
	// 一个连接的生命周期，这个和 Mysql Server配置有关系，必须小于 Server 配置
	// 比如一个链接用 12 h 换一个 conn，保证一点的可用性
	MaxLifeTime string `toml:"max_life_time" env:"MYSQL_MAX_LIFE_TIME"`
	// Idle 连接 最多允许存货多久
	MaxIdleTime string `toml:"max_idle_time" env:"MYSQL_MAX_idle_TIME"`
	// 作为私有变量，用于控制DetDB
	lock sync.Mutex
}

var mysqlDb *gorm.DB
var lock sync.Mutex

type Option func(*MySQL)

func WithMaxOpenConn(maxOpenConn int) Option {
	return func(m *MySQL) {
		m.MaxOpenConn = maxOpenConn
	}
}

func WithMaxIdleConn(maxIdleConn int) Option {
	return func(m *MySQL) {
		m.MaxIdleConn = maxIdleConn
	}
}

func WithMaxLifeTime(maxLifeTime string) Option {
	return func(m *MySQL) {
		m.MaxLifeTime = maxLifeTime
	}
}

func WithMaxIdleTime(maxIdleTime string) Option {
	return func(m *MySQL) {
		m.MaxIdleTime = maxIdleTime
	}
}

const (
	DefaultMaxOpenConn = 10
	DefaultMaxIdleConn = 5
	DefaultMaxLifeTime = "2h"
	DefaultMaxIdleTime = "4h"
)

func NewMySQL(host, port, username, password, database string, level int64, opts yaml.MapSlice) (*MySQL, error) {
	if host == "" || port == "" || username == "" || password == "" || database == "" {
		return nil, fmt.Errorf("required fields are missing")
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, host, port, database)

	// Check if there are options
	if len(opts) > 0 {
		dsn += "?"

		// Iterate through options
		for i, entry := range opts {
			key := entry.Key.(string)
			value := entry.Value.(string)
			dsn += fmt.Sprintf("%s=%s", key, value)

			// Add "&" if it's not the last option
			if i < len(opts)-1 {
				dsn += "&"
			}
		}
	}

	logLevel := logger.Default.LogMode(-1)
	switch level {
	case int64(zap.DebugLevel):
		logLevel = logger.Default.LogMode(logger.Info)
	case int64(zap.InfoLevel):
		logLevel = logger.Default.LogMode(logger.Info)
	case int64(zap.WarnLevel):
		logLevel = logger.Default.LogMode(logger.Warn)
	case int64(zap.ErrorLevel):
		logLevel = logger.Default.LogMode(logger.Error)
	default:
		logLevel = logger.Default.LogMode(logger.Silent)
	}

	c := &MySQL{
		Dsn:   dsn,
		Level: logLevel,
	}

	// Set default values if not provided in options
	if c.MaxOpenConn == 0 {
		c.MaxOpenConn = DefaultMaxOpenConn
	}
	if c.MaxIdleConn == 0 {
		c.MaxIdleConn = DefaultMaxIdleConn
	}
	if len(c.MaxLifeTime) == 0 {
		c.MaxLifeTime = DefaultMaxLifeTime
	}
	if len(c.MaxIdleTime) == 0 {
		c.MaxIdleTime = DefaultMaxIdleTime
	}

	return c, nil
}

func NewDefaultMysqlConn() *MySQL {
	return NewMySQLFromDSN("root:Hitosea@123..@tcp(mysql:3306)/coss?allowNativePasswords=true&timeout=800ms&readTimeout=200ms&writeTimeout=800ms&parseTime=true&loc=Local&charset=utf8mb4")
	//return NewMySQLFromDSN("root:888888@tcp(mysql:33066)/coss?allowNativePasswords=true&timeout=800ms&readTimeout=200ms&writeTimeout=800ms&parseTime=true&loc=Local&charset=utf8mb4")

}

func NewMySQLFromDSN(dsn string, opts ...Option) *MySQL {
	if dsn == "" {
		return nil
	}

	c := &MySQL{
		Dsn: dsn,
	}
	for _, opt := range opts {
		opt(c)
	}

	// Set default values if not provided in options
	if c.MaxOpenConn == 0 {
		c.MaxOpenConn = DefaultMaxOpenConn
	}
	if c.MaxIdleConn == 0 {
		c.MaxIdleConn = DefaultMaxIdleConn
	}
	if len(c.MaxLifeTime) == 0 {
		c.MaxLifeTime = DefaultMaxLifeTime
	}
	if len(c.MaxIdleTime) == 0 {
		c.MaxIdleTime = DefaultMaxIdleTime
	}

	return c
}

//func GenerateMysqlDSN(rootUsername, rootPassword, addr, database string) string {
//	return fmt.Sprintf("%s:%s@tcp(%s)/%s?allowNativePasswords=true&parseTime=true&loc=Local&charset=utf8mb4&multiStatements=true", rootUsername, rootPassword, addr, database)
//}

func (m *MySQL) GetConnection() (*gorm.DB, error) {
	m.lock.Lock() // 锁住临界区，保证线程安全
	defer m.lock.Unlock()

	if mysqlDb == nil {
		conn, err := m.getDBConn()
		if err != nil {
			return nil, err
		}
		mysqlDb = conn
	}
	return mysqlDb, nil
}

// gorm获取数据库连接
func (m *MySQL) getDBConn() (*gorm.DB, error) {
	var dsn string
	if m.Dsn != "" {
		dsn = m.Dsn
	} else {
		return nil, fmt.Errorf("连接Mysql error: dsn is nil")
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: m.Level,
	})
	if err != nil {
		return nil, fmt.Errorf("连接Mysql：%s，error：%s", dsn, err.Error())
	}

	maxLifeTime, err := utils.ParseDurationFromString(m.MaxLifeTime)
	if err != nil {
		return nil, err
	}

	maxIdleTime, err := utils.ParseDurationFromString(m.MaxIdleTime)
	if err != nil {
		return nil, err
	}

	// 维护连接池
	sqlDB, err := db.DB()
	sqlDB.SetMaxOpenConns(m.MaxOpenConn)
	sqlDB.SetMaxIdleConns(m.MaxIdleConn)
	sqlDB.SetConnMaxLifetime(maxLifeTime)
	sqlDB.SetConnMaxIdleTime(maxIdleTime)

	// 用于测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping mysql %s，error：%s", dsn, err.Error())
	}
	return db, nil
}

func GetConnection() (*gorm.DB, error) {
	lock.Lock() // 锁住临界区，保证线程安全
	defer lock.Unlock()

	if mysqlDb != nil {
		return mysqlDb, nil
	}
	return mysqlDb, errors.New("mysql connection is nil")
}
