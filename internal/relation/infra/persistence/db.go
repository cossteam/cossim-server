package persistence

import (
	"context"
	"github.com/cossim/coss-server/internal/relation/cache"
	"github.com/cossim/coss-server/internal/relation/domain/repository"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/db"
	"gorm.io/gorm"
	"strconv"
)

type Repositories struct {
	db *gorm.DB

	UserRepo              repository.UserRelationRepository
	GroupRepo             repository.GroupRelationRepository
	UserFriendRequestRepo repository.UserFriendRequestRepository
	GroupJoinRequestRepo  repository.GroupRequestRepository
	GroupAnnouncementRepo repository.GroupAnnouncementRepository
	DialogRepo            repository.DialogRepository
	DialogUserRepo        repository.DialogUserRepository

	userCache  cache.RelationUserCache
	groupCache cache.RelationGroupCache
}

func NewRepositories(cfg *pkgconfig.AppConfig) *Repositories {
	mysql, err := db.NewMySQL(cfg.MySQL.Address, strconv.Itoa(cfg.MySQL.Port), cfg.MySQL.Username, cfg.MySQL.Password, cfg.MySQL.Database, int64(cfg.Log.Level), cfg.MySQL.Opts)
	if err != nil {
		panic(err)
	}

	dbConn, err := mysql.GetConnection()
	if err != nil {
		panic(err)
	}

	userCache, err := cache.NewRelationUserCacheRedis(cfg.Redis.Addr(), cfg.Redis.Password, 0)
	if err != nil {
		panic(err)
	}

	groupCache, err := cache.NewRelationGroupCacheRedis(cfg.Redis.Addr(), cfg.Redis.Password, 0)
	if err != nil {
		panic(err)
	}

	return &Repositories{
		UserRepo:              NewMySQLRelationUserRepository(dbConn, userCache),
		GroupRepo:             NewMySQLRelationGroupRepository(dbConn, groupCache),
		DialogRepo:            NewMySQLMySQLDialogRepository(dbConn, nil),
		DialogUserRepo:        NewMySQLDialogUserRepository(dbConn, nil),
		GroupJoinRequestRepo:  NewMySQLGroupJoinRequestRepository(dbConn, nil),
		GroupAnnouncementRepo: NewMySQLRelationGroupAnnouncementRepository(dbConn, nil),
		UserFriendRequestRepo: NewMySQLUserFriendRequestRepository(dbConn, nil),
		db:                    dbConn,
		userCache:             userCache,
		groupCache:            groupCache,
	}
}

func (r *Repositories) TXRepositories(fc func(txr *Repositories) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 创建一个新的 Repositories 实例，确保在事务中使用的是同一个数据库连接
		txr := &Repositories{
			UserRepo:              NewMySQLRelationUserRepository(tx, r.userCache),
			GroupRepo:             NewMySQLRelationGroupRepository(tx, r.groupCache),
			DialogRepo:            NewMySQLMySQLDialogRepository(tx, nil),
			DialogUserRepo:        NewMySQLDialogUserRepository(tx, nil),
			GroupJoinRequestRepo:  NewMySQLGroupJoinRequestRepository(tx, nil),
			GroupAnnouncementRepo: NewMySQLRelationGroupAnnouncementRepository(tx, nil),
			UserFriendRequestRepo: NewMySQLUserFriendRequestRepository(tx, nil),
			db:                    tx,
		}
		if err := fc(txr); err != nil {
			return err
		}
		return nil
	})
}

func (r *Repositories) Automigrate() error {
	return r.db.AutoMigrate(
		&GroupRelationModel{},
		&UserRelationModel{},
		&DialogModel{},
		&DialogUserModel{},
		&UserFriendRequestModel{},
		&GroupJoinRequestModel{},
		&GroupAnnouncementModel{},
		&GroupAnnouncementReadModel{},
	)
}

func (r *Repositories) Close() error {
	if r.groupCache != nil {
		r.groupCache.DeleteAllCache(context.Background())
	}
	return nil
}
